package handlers

import (
	"net/http"
	"skillshub-enterprise/api/internal/config"
	"skillshub-enterprise/api/internal/middleware"
	"skillshub-enterprise/api/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db   *pgxpool.Pool
	cfg  *config.Config
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func NewAuthHandler(db *pgxpool.Pool, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user models.User
	err := h.db.QueryRow(c,
		`SELECT id, username, email, password_hash, role FROM users WHERE username = $1`,
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password (only if password_hash is set)
	if user.PasswordHash != "" {
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
	}

	// Generate JWT tokens
	accessToken, err := middleware.GenerateJWT(
		user.ID.String(), user.Username, user.Email, user.Role,
		h.cfg.JWTSecret, 8*time.Hour,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := middleware.GenerateJWT(
		user.ID.String(), user.Username, user.Email, user.Role,
		h.cfg.JWTSecret, 7*24*time.Hour,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Update last login
	_, _ = h.db.Exec(c, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, user.ID)

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(8 * time.Hour),
		User: UserInfo{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	var userID string
	err = h.db.QueryRow(c,
		`INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.Username, req.Email, string(hashedPassword), "developer",
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_id": userID})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse and validate refresh token
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Generate new access token
	accessToken, err := middleware.GenerateJWT(
		claims.UserID, claims.Username, claims.Email, claims.Role,
		h.cfg.JWTSecret, 8*time.Hour,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_at":   time.Now().Add(8 * time.Hour),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// In a real implementation, you might want to blacklist the token
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
