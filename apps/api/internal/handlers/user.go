package handlers

import (
	"net/http"
	"skillshub-enterprise/api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewUserHandler(db *pgxpool.Pool, cfg *config.Config) *UserHandler {
	return &UserHandler{db: db, cfg: cfg}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user struct {
		ID          string     `json:"id"`
		Username    string     `json:"username"`
		Email       string     `json:"email"`
		Role        string     `json:"role"`
		TeamID      *string    `json:"team_id,omitempty"`
		MfaEnabled  bool       `json:"mfa_enabled"`
		LastLoginAt *string    `json:"last_login_at,omitempty"`
		CreatedAt   string     `json:"created_at"`
	}

	err := h.db.QueryRow(c, `
		SELECT id, username, email, role, team_id, mfa_enabled, last_login_at, created_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, &user.TeamID,
		&user.MfaEnabled, &user.LastLoginAt, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		Email string `json:"email" binding:"omitempty,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(c, `UPDATE users SET email = $1 WHERE id = $2`, req.Email, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	rows, err := h.db.Query(c, `
		SELECT id, username, email, role, team_id, mfa_enabled, last_login_at, created_at
		FROM users ORDER BY created_at DESC LIMIT `+limit+` OFFSET `+page)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	users := []gin.H{}
	for rows.Next() {
		var u struct {
			ID          string     `json:"id"`
			Username    string     `json:"username"`
			Email       string     `json:"email"`
			Role        string     `json:"role"`
			TeamID      *string    `json:"team_id,omitempty"`
			MfaEnabled  bool       `json:"mfa_enabled"`
			LastLoginAt *string    `json:"last_login_at,omitempty"`
			CreatedAt   string     `json:"created_at"`
		}
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.TeamID, &u.MfaEnabled, &u.LastLoginAt, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, gin.H{
			"id":           u.ID,
			"username":     u.Username,
			"email":        u.Email,
			"role":         u.Role,
			"team_id":      u.TeamID,
			"mfa_enabled":  u.MfaEnabled,
			"last_login_at": u.LastLoginAt,
			"created_at":   u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}
