package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	registryURL string
	configFile  string
)

// Config represents CLI configuration
type Config struct {
	RegistryURL string `json:"registry_url"`
	Token       string `json:"token,omitempty"`
	LastSync    string `json:"last_sync,omitempty"`
}

func init() {
	home, _ := os.UserHomeDir()
	configFile = filepath.Join(home, ".skillshub", "config.json")

	rootCmd.PersistentFlags().StringVar(&registryURL, "registry", "http://localhost:3000", "SkillsHub Registry URL")
}

var rootCmd = &cobra.Command{
	Use:   "skillshub",
	Short: "SkillsHub CLI - Enterprise Skills Management",
	Long:  `SkillsHub CLI is a command-line tool for managing AI Skills in enterprise environment.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// loadConfig loads CLI configuration
func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{RegistryURL: registryURL}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveConfig saves CLI configuration
func saveConfig(config *Config) error {
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}

// getAuthHeader returns authorization header
func getAuthHeader(config *Config) string {
	if config.Token != "" {
		return fmt.Sprintf("Bearer %s", config.Token)
	}
	return ""
}

// apiRequest makes HTTP request to API
func apiRequest(method, path string, config *Config) ([]byte, error) {
	url := strings.TrimSuffix(config.RegistryURL, "/") + path

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", getAuthHeader(config))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	return body, nil
}

// ========== Commands ==========

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to SkillsHub Registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		fmt.Print("Username: ")
		var username string
		fmt.Scanln(&username)

		fmt.Print("Password: ")
		var password string
		fmt.Scanln(&password)

		// Login request
		loginData := map[string]string{
			"username": username,
			"password": password,
		}

		loginJSON, _ := json.Marshal(loginData)
		url := strings.TrimSuffix(config.RegistryURL, "/") + "/api/v1/auth/login"

		resp, err := http.Post(url, "application/json", strings.NewReader(string(loginJSON)))
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("login failed: invalid credentials")
		}

		var result struct {
			AccessToken string `json:"access_token"`
			User        struct {
				Username string `json:"username"`
				Role     string `json:"role"`
			} `json:"user"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		config.Token = result.AccessToken
		config.RegistryURL = registryURL
		if err := saveConfig(config); err != nil {
			return err
		}

		fmt.Printf("✓ Logged in as %s (%s)\n", result.User.Username, result.User.Role)
		return nil
	},
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search for skills",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		query := ""
		if len(args) > 0 {
			query = "?q=" + args[0]
		}

		body, err := apiRequest("GET", "/api/v1/skills"+query, config)
		if err != nil {
			return err
		}

		var result struct {
			Skills []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Category    string `json:"category"`
				Status      string `json:"status"`
			} `json:"skills"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return err
		}

		if len(result.Skills) == 0 {
			fmt.Println("No skills found.")
			return nil
		}

		fmt.Printf("Found %d skills:\n\n", len(result.Skills))
		for _, skill := range result.Skills {
			status := "✓"
			if skill.Status != "active" {
				status = "○"
			}
			fmt.Printf("  %s %s\n     %s\n\n", status, skill.Name, skill.Description)
		}

		return nil
	},
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install <skill-name>",
	Short: "Install a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		skillName := args[0]

		// Get skill info
		body, err := apiRequest("GET", fmt.Sprintf("/api/v1/skills/%s", skillName), config)
		if err != nil {
			return fmt.Errorf("skill not found: %s", skillName)
		}

		var skillInfo struct {
			Skill struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"skill"`
		}

		if err := json.Unmarshal(body, &skillInfo); err != nil {
			return err
		}

		// Get download URL
		downloadBody, err := apiRequest("GET", fmt.Sprintf("/api/v1/skills/%s/latest/download", skillName), config)
		if err != nil {
			return err
		}

		var downloadInfo struct {
			URL       string `json:"download_url"`
			ExpiresIn int    `json:"expires_in"`
		}

		if err := json.Unmarshal(downloadBody, &downloadInfo); err != nil {
			return err
		}

		// In production, download and install the skill
		fmt.Printf("Installing skill: %s\n", skillInfo.Skill.Name)
		fmt.Printf("Download URL: %s\n", downloadInfo.URL)

		// Simulate installation
		home, _ := os.UserHomeDir()
		skillsDir := filepath.Join(home, ".claude", "skills", "approved", skillName)

		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return err
		}

		// Create placeholder SKILL.md
		skillMd := fmt.Sprintf(`---
name: %s
description: %s
---

# %s

Skill installed successfully!
`, skillInfo.Skill.Name, skillInfo.Skill.Description, skillInfo.Skill.Name)

		if err := os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillMd), 0644); err != nil {
			return err
		}

		fmt.Printf("✓ Skill '%s' installed to %s\n", skillName, skillsDir)
		return nil
	},
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		skillsDir := filepath.Join(home, ".claude", "skills", "approved")

		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No skills installed.")
				return nil
			}
			return err
		}

		fmt.Printf("Installed skills (%d):\n\n", len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				skillMd := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
				description := "No description"

				if data, err := os.ReadFile(skillMd); err == nil {
					var skill struct {
						Description string `yaml:"description"`
					}
					if err := yaml.Unmarshal(data, &skill); err == nil && skill.Description != "" {
						description = skill.Description
					}
				}

				fmt.Printf("  • %s\n    %s\n\n", entry.Name(), description)
			}
		}

		return nil
	},
}

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall <skill-name>",
	Short: "Uninstall a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		skillDir := filepath.Join(home, ".claude", "skills", "approved", args[0])

		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			return fmt.Errorf("skill not found: %s", args[0])
		}

		if err := os.RemoveAll(skillDir); err != nil {
			return err
		}

		fmt.Printf("✓ Skill '%s' uninstalled\n", args[0])
		return nil
	},
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show SkillsHub status",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		fmt.Println("SkillsHub CLI Status")
		fmt.Println("====================")
		fmt.Printf("Registry:   %s\n", config.RegistryURL)

		if config.Token != "" {
			token, _ := jwt.Parse(config.Token, func(t *jwt.Token) (interface{}, error) {
				return []byte{}, nil
			})
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if exp, ok := claims["exp"].(float64); ok {
					expiresAt := time.Unix(int64(exp), 0)
					fmt.Printf("Token:      Valid until %s\n", expiresAt.Format(time.RFC3339))
				}
			}
			fmt.Printf("Status:     Logged in\n")
		} else {
			fmt.Printf("Status:     Not logged in\n")
		}

		return nil
	},
}

// Register commands
func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(statusCmd)
}
