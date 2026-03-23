package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	TeamID       uuid.UUID  `json:"team_id,omitempty"`
	MfaEnabled   bool       `json:"mfa_enabled"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Team struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Skill struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	DisplayName  string     `json:"display_name,omitempty"`
	Description  string     `json:"description"`
	Category     string     `json:"category,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	SourceType   string     `json:"source_type"`
	SourceURL    string     `json:"source_url,omitempty"`
	AuthorID     uuid.UUID  `json:"author_id,omitempty"`
	License      string     `json:"license"`
	Status       string     `json:"status"`
	InstallCount int        `json:"install_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SkillVersion struct {
	ID            uuid.UUID  `json:"id"`
	SkillID       uuid.UUID  `json:"skill_id"`
	Version       string     `json:"version"`
	Changelog     string     `json:"changelog,omitempty"`
	StoragePath   string     `json:"storage_path"`
	FileHash      string     `json:"file_hash"`
	FileSize      int64      `json:"file_size"`
	IsLatest      bool       `json:"is_latest"`
	Status        string     `json:"status"`
	ScanID        *uuid.UUID `json:"scan_id,omitempty"`
	PublishedBy   uuid.UUID  `json:"published_by,omitempty"`
	PublishedAt   time.Time  `json:"published_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Scan struct {
	ID               uuid.UUID  `json:"id"`
	SkillVersionID   uuid.UUID  `json:"skill_version_id"`
	Status           string     `json:"status"`
	RiskLevel        *string    `json:"risk_level,omitempty"`
	RiskScore        *int       `json:"risk_score,omitempty"`
	Layer1Result     []byte     `json:"layer1_result,omitempty"`
	Layer2Result     []byte     `json:"layer2_result,omitempty"`
	Layer3Result     []byte     `json:"layer3_result,omitempty"`
	Layer4Result     []byte     `json:"layer4_result,omitempty"`
	Summary          string     `json:"summary,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type Review struct {
	ID          uuid.UUID  `json:"id"`
	ScanID      uuid.UUID  `json:"scan_id"`
	ApplicantID uuid.UUID  `json:"applicant_id,omitempty"`
	AssigneeID  uuid.UUID  `json:"assignee_id,omitempty"`
	Status      string     `json:"status"`
	Decision    *string    `json:"decision,omitempty"`
	Comment     string     `json:"comment,omitempty"`
	Conditions  []byte     `json:"conditions,omitempty"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Installation struct {
	ID            uuid.UUID  `json:"id"`
	SkillVersionID uuid.UUID `json:"skill_version_id"`
	UserID        uuid.UUID  `json:"user_id"`
	DeviceID      string     `json:"device_id"`
	InstalledAt   time.Time  `json:"installed_at"`
	UninstalledAt *time.Time `json:"uninstalled_at,omitempty"`
	IsActive      bool       `json:"is_active"`
}

type AuditLog struct {
	ID        uuid.UUID      `json:"id"`
	EventType string         `json:"event_type"`
	ActorID   *uuid.UUID     `json:"actor_id,omitempty"`
	ActorMeta map[string]any `json:"actor_meta,omitempty"`
	Resource  map[string]any `json:"resource,omitempty"`
	Result    string         `json:"result"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type SyncSource struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	SourceType string         `json:"source_type"`
	URL        string         `json:"url"`
	Config     map[string]any `json:"config"`
	IsEnabled  bool           `json:"is_enabled"`
	LastSyncAt *time.Time     `json:"last_sync_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
