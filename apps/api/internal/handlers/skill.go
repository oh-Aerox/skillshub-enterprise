package handlers

import (
	"net/http"
	"skillshub-enterprise/api/internal/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SkillHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

type CreateSkillRequest struct {
	Name        string   `json:"name" binding:"required"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description" binding:"required"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	SourceType  string   `json:"source_type" binding:"required,oneof=internal opensource"`
	SourceURL   string   `json:"source_url"`
	License     string   `json:"license"`
}

func NewSkillHandler(db *pgxpool.Pool, cfg *config.Config) *SkillHandler {
	return &SkillHandler{db: db, cfg: cfg}
}

func (h *SkillHandler) ListSkills(c *gin.Context) {
	q := c.Query("q")
	category := c.Query("category")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	// Build query
	whereClause := "WHERE status = 'active'"
	args := []interface{}{}
	argCount := 0

	if q != "" {
		argCount++
		whereClause += " AND (name ILIKE $" + string(rune('0'+argCount)) + " OR description ILIKE $" + string(rune('0'+argCount)) + ")"
		args = append(args, "%"+q+"%")
	}

	if category != "" {
		argCount++
		whereClause += " AND category = $" + string(rune('0'+argCount))
		args = append(args, category)
	}

	// Execute query
	rows, err := h.db.Query(c, "SELECT id, name, display_name, description, category, tags, source_type, author_id, license, status, install_count, created_at, updated_at FROM skills "+whereClause+" ORDER BY created_at DESC LIMIT "+limit+" OFFSET "+page, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch skills"})
		return
	}
	defer rows.Close()

	skills := []gin.H{}
	for rows.Next() {
		var skill struct {
			ID           uuid.UUID  `json:"id"`
			Name         string     `json:"name"`
			DisplayName  *string    `json:"display_name,omitempty"`
			Description  string     `json:"description"`
			Category     *string    `json:"category,omitempty"`
			Tags         []string   `json:"tags"`
			SourceType   string     `json:"source_type"`
			AuthorID     *uuid.UUID `json:"author_id,omitempty"`
			License      *string    `json:"license,omitempty"`
			Status       string     `json:"status"`
			InstallCount int        `json:"install_count"`
			CreatedAt    time.Time  `json:"created_at"`
			UpdatedAt    time.Time  `json:"updated_at"`
		}
		if err := rows.Scan(&skill.ID, &skill.Name, &skill.DisplayName, &skill.Description, &skill.Category, &skill.Tags, &skill.SourceType, &skill.AuthorID, &skill.License, &skill.Status, &skill.InstallCount, &skill.CreatedAt, &skill.UpdatedAt); err != nil {
			continue
		}
		skills = append(skills, gin.H{
			"id":            skill.ID,
			"name":          skill.Name,
			"display_name":  skill.DisplayName,
			"description":   skill.Description,
			"category":      skill.Category,
			"tags":          skill.Tags,
			"source_type":   skill.SourceType,
			"license":       skill.License,
			"status":        skill.Status,
			"install_count": skill.InstallCount,
			"created_at":    skill.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"skills": skills})
}

func (h *SkillHandler) GetSkill(c *gin.Context) {
	skillID := c.Param("skillId")

	var skill gin.H
	err := h.db.QueryRow(c, `
		SELECT s.id, s.name, s.display_name, s.description, s.category, s.tags,
		       s.source_type, s.source_url, s.license, s.status, s.install_count,
		       s.created_at, s.updated_at, u.username as author_name
		FROM skills s
		LEFT JOIN users u ON s.author_id = u.id
		WHERE s.id = $1`, skillID).Scan(
		&skill)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Skill not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"skill": skill})
}

func (h *SkillHandler) CreateSkill(c *gin.Context) {
	var req CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	var skillID uuid.UUID
	err := h.db.QueryRow(c, `
		INSERT INTO skills (name, display_name, description, category, tags, source_type, source_url, license, author_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		req.Name, req.DisplayName, req.Description, req.Category, req.Tags,
		req.SourceType, req.SourceURL, req.License, userID,
	).Scan(&skillID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Skill name already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Skill created successfully", "skill_id": skillID})
}

func (h *SkillHandler) UpdateSkill(c *gin.Context) {
	skillID := c.Param("skillId")

	var req CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(c, `
		UPDATE skills SET name = $1, display_name = $2, description = $3,
		        category = $4, tags = $5, source_type = $6, source_url = $7, license = $8
		WHERE id = $9`,
		req.Name, req.DisplayName, req.Description, req.Category, req.Tags,
		req.SourceType, req.SourceURL, req.License, skillID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill updated successfully"})
}

func (h *SkillHandler) DeleteSkill(c *gin.Context) {
	skillID := c.Param("skillId")

	_, err := h.db.Exec(c, `DELETE FROM skills WHERE id = $1`, skillID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill deleted successfully"})
}

func (h *SkillHandler) InstallSkill(c *gin.Context) {
	skillID := c.Param("skillId")
	userID, _ := c.Get("user_id")
	deviceID := c.PostForm("device_id")

	var versionID uuid.UUID
	err := h.db.QueryRow(c, `
		SELECT id FROM skill_versions WHERE skill_id = $1 AND is_latest = true`, skillID).Scan(&versionID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No available version found"})
		return
	}

	_, err = h.db.Exec(c, `
		INSERT INTO installations (skill_version_id, user_id, device_id)
		VALUES ($1, $2, $3)`, versionID, userID, deviceID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already installed"})
		return
	}

	// Update install count
	_, _ = h.db.Exec(c, `UPDATE skills SET install_count = install_count + 1 WHERE id = $1`, skillID)

	c.JSON(http.StatusOK, gin.H{"message": "Skill installed successfully"})
}

func (h *SkillHandler) UninstallSkill(c *gin.Context) {
	skillID := c.Param("skillId")
	userID, _ := c.Get("user_id")

	_, err := h.db.Exec(c, `
		UPDATE installations SET is_active = false, uninstalled_at = NOW()
		WHERE skill_version_id IN (
		    SELECT id FROM skill_versions WHERE skill_id = $1
		) AND user_id = $2`, skillID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to uninstall"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill uninstalled successfully"})
}

func (h *SkillHandler) ListVersions(c *gin.Context) {
	skillID := c.Param("skillId")

	rows, err := h.db.Query(c, `
		SELECT id, version, changelog, is_latest, status, published_at
		FROM skill_versions WHERE skill_id = $1 ORDER BY published_at DESC`, skillID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch versions"})
		return
	}
	defer rows.Close()

	versions := []gin.H{}
	for rows.Next() {
		var v struct {
			ID        uuid.UUID  `json:"id"`
			Version   string     `json:"version"`
			Changelog *string    `json:"changelog,omitempty"`
			IsLatest  bool       `json:"is_latest"`
			Status    string     `json:"status"`
			PublishedAt time.Time `json:"published_at"`
		}
		if err := rows.Scan(&v.ID, &v.Version, &v.Changelog, &v.IsLatest, &v.Status, &v.PublishedAt); err != nil {
			continue
		}
		versions = append(versions, gin.H{
			"id":         v.ID,
			"version":    v.Version,
			"changelog":  v.Changelog,
			"is_latest":  v.IsLatest,
			"status":     v.Status,
			"published_at": v.PublishedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

func (h *SkillHandler) DownloadSkill(c *gin.Context) {
	skillID := c.Param("skillId")
	version := c.Param("version")

	var storagePath string
	err := h.db.QueryRow(c, `
		SELECT storage_path FROM skill_versions
		WHERE skill_id = $1 AND version = $2`, skillID, version).Scan(&storagePath)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// In production, generate a pre-signed URL from MinIO/S3
	c.JSON(http.StatusOK, gin.H{
		"download_url": "/api/v1/files/" + storagePath,
		"expires_in":   600, // 10 minutes
	})
}

func (h *SkillHandler) ListInstallations(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rows, err := h.db.Query(c, `
		SELECT sv.skill_id, s.name, s.description, sv.version, i.installed_at, i.is_active
		FROM installations i
		JOIN skill_versions sv ON i.skill_version_id = sv.id
		JOIN skills s ON sv.skill_id = s.id
		WHERE i.user_id = $1 AND i.is_active = true`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installations"})
		return
	}
	defer rows.Close()

	installations := []gin.H{}
	for rows.Next() {
		var item gin.H
		var skillID uuid.UUID
		var name, description, version string
		var installedAt time.Time
		var isActive bool
		if err := rows.Scan(&skillID, &name, &description, &version, &installedAt, &isActive); err != nil {
			continue
		}
		installations = append(installations, gin.H{
			"skill_id":     skillID,
			"name":         name,
			"description":  description,
			"version":      version,
			"installed_at": installedAt,
			"is_active":    isActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"installations": installations})
}

func (h *SkillHandler) GetStats(c *gin.Context) {
	var stats struct {
		TotalSkills       int `json:"total_skills"`
		TotalInstallations int `json:"total_installations"`
		PendingReviews    int `json:"pending_reviews"`
		ActiveUsers       int `json:"active_users"`
	}

	_ = h.db.QueryRow(c, `SELECT COUNT(*) FROM skills`).Scan(&stats.TotalSkills)
	_ = h.db.QueryRow(c, `SELECT COUNT(*) FROM installations WHERE is_active = true`).Scan(&stats.TotalInstallations)
	_ = h.db.QueryRow(c, `SELECT COUNT(*) FROM reviews WHERE status = 'pending'`).Scan(&stats.PendingReviews)
	_ = h.db.QueryRow(c, `SELECT COUNT(*) FROM users WHERE last_login_at > NOW() - INTERVAL '7 days'`).Scan(&stats.ActiveUsers)

	c.JSON(http.StatusOK, stats)
}
