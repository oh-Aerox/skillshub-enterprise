package handlers

import (
	"net/http"
	"skillshub-enterprise/api/internal/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScanHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

type TriggerScanRequest struct {
	SkillVersionID string `json:"skill_version_id" binding:"required"`
	Priority       string `json:"priority"`
}

func NewScanHandler(db *pgxpool.Pool, cfg *config.Config) *ScanHandler {
	return &ScanHandler{db: db, cfg: cfg}
}

func (h *ScanHandler) GetScan(c *gin.Context) {
	scanID := c.Param("scanId")

	var scan struct {
		ID          uuid.UUID      `json:"id"`
		SkillVerID  uuid.UUID      `json:"skill_version_id"`
		Status      string         `json:"status"`
		RiskLevel   *string        `json:"risk_level,omitempty"`
		RiskScore   *int           `json:"risk_score,omitempty"`
		Summary     string         `json:"summary,omitempty"`
		StartedAt   *time.Time     `json:"started_at,omitempty"`
		CompletedAt *time.Time     `json:"completed_at,omitempty"`
		CreatedAt   time.Time      `json:"created_at"`
	}

	err := h.db.QueryRow(c, `
		SELECT id, skill_version_id, status, risk_level, risk_score, summary, started_at, completed_at, created_at
		FROM scans WHERE id = $1`, scanID).Scan(
		&scan.ID, &scan.SkillVerID, &scan.Status, &scan.RiskLevel, &scan.RiskScore,
		&scan.Summary, &scan.StartedAt, &scan.CompletedAt, &scan.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scan": scan})
}

func (h *ScanHandler) TriggerScan(c *gin.Context) {
	var req TriggerScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	skillVerID, err := uuid.Parse(req.SkillVersionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skill_version_id"})
		return
	}

	var scanID uuid.UUID
	err = h.db.QueryRow(c, `
		INSERT INTO scans (skill_version_id, status) VALUES ($1, 'pending') RETURNING id`,
		skillVerID).Scan(&scanID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scan"})
		return
	}

	// In production, push to Redis queue for scanner service
	go func() {
		// Call scanner service
		// POST to ScannerServiceURL/api/scan
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"scan_id": scanID,
		"message": "Scan triggered successfully",
	})
}
