package handlers

import (
	"net/http"
	"skillshub-enterprise/api/internal/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

type DecideReviewRequest struct {
	Decision string `json:"decision" binding:"required,oneof=approved rejected escalated"`
	Comment  string `json:"comment"`
}

func NewReviewHandler(db *pgxpool.Pool, cfg *config.Config) *ReviewHandler {
	return &ReviewHandler{db: db, cfg: cfg}
}

func (h *ReviewHandler) ListReviews(c *gin.Context) {
	status := c.Query("status")
	assignee := c.Query("assignee")

	query := `
		SELECT r.id, r.scan_id, r.status, r.decision, r.comment, r.due_at, r.created_at,
		       s.risk_level, s.risk_score, sv.skill_id, sk.name as skill_name
		FROM reviews r
		JOIN scans s ON r.scan_id = s.id
		JOIN skill_versions sv ON s.skill_version_id = sv.id
		JOIN skills sk ON sv.skill_id = sk.id
		WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	if status != "" {
		argCount++
		query += " AND r.status = $" + string(rune('0'+argCount))
		args = append(args, status)
	}

	if assignee == "me" {
		userID, exists := c.Get("user_id")
		if exists {
			argCount++
			query += " AND r.assignee_id = $" + string(rune('0'+argCount))
			args = append(args, userID)
		}
	}

	query += " ORDER BY r.created_at DESC"

	rows, err := h.db.Query(c, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}
	defer rows.Close()

	reviews := []gin.H{}
	for rows.Next() {
		var item struct {
			ID         uuid.UUID   `json:"id"`
			ScanID     uuid.UUID   `json:"scan_id"`
			Status     string      `json:"status"`
			Decision   *string     `json:"decision,omitempty"`
			Comment    string      `json:"comment"`
			DueAt      *time.Time  `json:"due_at,omitempty"`
			CreatedAt  time.Time   `json:"created_at"`
			RiskLevel  *string     `json:"risk_level,omitempty"`
			RiskScore  *int        `json:"risk_score,omitempty"`
			SkillID    uuid.UUID   `json:"skill_id"`
			SkillName  string      `json:"skill_name"`
		}
		if err := rows.Scan(&item.ID, &item.ScanID, &item.Status, &item.Decision, &item.Comment,
			&item.DueAt, &item.CreatedAt, &item.RiskLevel, &item.RiskScore, &item.SkillID, &item.SkillName); err != nil {
			continue
		}
		reviews = append(reviews, gin.H{
			"id":          item.ID,
			"scan_id":     item.ScanID,
			"status":      item.Status,
			"decision":    item.Decision,
			"comment":     item.Comment,
			"due_at":      item.DueAt,
			"created_at":  item.CreatedAt,
			"risk_level":  item.RiskLevel,
			"risk_score":  item.RiskScore,
			"skill_id":    item.SkillID,
			"skill_name":  item.SkillName,
		})
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews, "total": len(reviews)})
}

func (h *ReviewHandler) GetReview(c *gin.Context) {
	reviewID := c.Param("reviewId")

	var review struct {
		ID        uuid.UUID      `json:"id"`
		ScanID    uuid.UUID      `json:"scan_id"`
		Status    string         `json:"status"`
		Decision  *string        `json:"decision,omitempty"`
		Comment   string         `json:"comment"`
		Conditions []byte        `json:"conditions,omitempty"`
		DueAt     *time.Time     `json:"due_at,omitempty"`
		ReviewedAt *time.Time    `json:"reviewed_at,omitempty"`
		CreatedAt time.Time      `json:"created_at"`
	}

	err := h.db.QueryRow(c, `
		SELECT id, scan_id, status, decision, comment, conditions, due_at, reviewed_at, created_at
		FROM reviews WHERE id = $1`, reviewID).Scan(
		&review.ID, &review.ScanID, &review.Status, &review.Decision, &review.Comment,
		&review.Conditions, &review.DueAt, &review.ReviewedAt, &review.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": review})
}

func (h *ReviewHandler) DecideReview(c *gin.Context) {
	reviewID := c.Param("reviewId")
	userID, _ := c.Get("user_id")

	var req DecideReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reviewUUID, err := uuid.Parse(reviewID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	now := time.Now()
	_, err = h.db.Exec(c, `
		UPDATE reviews
		SET status = CASE
		    WHEN $2 = 'approved' THEN 'approved'
		    WHEN $2 = 'rejected' THEN 'rejected'
		    WHEN $2 = 'escalated' THEN 'escalated'
		END,
		decision = $2,
		comment = $3,
		reviewed_at = $4,
		assignee_id = $5
		WHERE id = $1`,
		reviewUUID, req.Decision, req.Comment, now, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	// If approved, update skill status
	if req.Decision == "approved" {
		var scanID uuid.UUID
		_ = h.db.QueryRow(c, `SELECT scan_id FROM reviews WHERE id = $1`, reviewID).Scan(&scanID)

		var skillVerID uuid.UUID
		_ = h.db.QueryRow(c, `SELECT skill_version_id FROM scans WHERE id = $1`, scanID).Scan(&skillVerID)

		var skillID uuid.UUID
		_ = h.db.QueryRow(c, `SELECT skill_id FROM skill_versions WHERE id = $1`, skillVerID).Scan(&skillID)

		_, _ = h.db.Exec(c, `UPDATE skills SET status = 'active' WHERE id = $1`, skillID)
		_, _ = h.db.Exec(c, `UPDATE skill_versions SET is_latest = true WHERE id = $1`, skillVerID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review decision saved successfully"})
}
