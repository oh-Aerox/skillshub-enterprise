-- SkillsHub Enterprise Database Schema
-- Version: 1.0.0

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Teams table
CREATE TABLE teams (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Users table
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    role          VARCHAR(20) NOT NULL DEFAULT 'developer',
    team_id       UUID REFERENCES teams(id) ON DELETE SET NULL,
    mfa_enabled   BOOLEAN DEFAULT false,
    mfa_secret    VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- Skills table
CREATE TABLE skills (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL UNIQUE,
    display_name  VARCHAR(200),
    description   TEXT NOT NULL,
    category      VARCHAR(50),
    tags          TEXT[],
    source_type   VARCHAR(20) NOT NULL DEFAULT 'internal',
    source_url    TEXT,
    author_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    license       VARCHAR(50) DEFAULT 'Internal',
    status        VARCHAR(20) NOT NULL DEFAULT 'active',
    install_count INTEGER DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_skills_name ON skills(name);
CREATE INDEX idx_skills_status ON skills(status);
CREATE INDEX idx_skills_category ON skills(category);
CREATE INDEX idx_skills_tags ON skills USING GIN(tags);

-- Skill versions table
CREATE TABLE skill_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id        UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version         VARCHAR(20) NOT NULL,
    changelog       TEXT,
    storage_path    TEXT NOT NULL,
    file_hash       VARCHAR(64) NOT NULL,
    file_size       BIGINT,
    is_latest       BOOLEAN DEFAULT false,
    status          VARCHAR(20) DEFAULT 'stable',
    scan_id         UUID,
    published_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(skill_id, version)
);

CREATE INDEX idx_versions_skill ON skill_versions(skill_id);
CREATE INDEX idx_versions_hash ON skill_versions(file_hash);
CREATE INDEX idx_versions_latest ON skill_versions(is_latest) WHERE is_latest = true;

-- Scans table
CREATE TABLE scans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    risk_level      CHAR(1),
    risk_score      SMALLINT,
    layer1_result   JSONB,
    layer2_result   JSONB,
    layer3_result   JSONB,
    layer4_result   JSONB,
    summary         TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scans_version ON scans(skill_version_id);
CREATE INDEX idx_scans_status ON scans(status);
CREATE INDEX idx_scans_created ON scans(created_at);

-- Add scan_id foreign key to skill_versions after scans table is created
ALTER TABLE skill_versions ADD CONSTRAINT fk_scan
    FOREIGN KEY (scan_id) REFERENCES scans(id) ON DELETE SET NULL;

-- Reviews table
CREATE TABLE reviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id         UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    applicant_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    assignee_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    decision        VARCHAR(20),
    comment         TEXT,
    conditions      JSONB,
    due_at          TIMESTAMPTZ,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_status ON reviews(status);
CREATE INDEX idx_reviews_assignee ON reviews(assignee_id);
CREATE INDEX idx_reviews_created ON reviews(created_at);

-- Installations table
CREATE TABLE installations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    device_id       VARCHAR(255),
    installed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    uninstalled_at  TIMESTAMPTZ,
    is_active       BOOLEAN DEFAULT true
);

CREATE INDEX idx_installations_user ON installations(user_id);
CREATE INDEX idx_installations_skill ON installations(skill_version_id);

-- Audit logs table (append-only)
CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type  VARCHAR(50) NOT NULL,
    actor_id    UUID,
    actor_meta  JSONB,
    resource    JSONB,
    result      VARCHAR(20),
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_event ON audit_logs(event_type);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at);

-- Sync sources table
CREATE TABLE sync_sources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    source_type     VARCHAR(20) NOT NULL,
    url             TEXT NOT NULL,
    config          JSONB,
    is_enabled      BOOLEAN DEFAULT true,
    last_sync_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- API tokens table
CREATE TABLE api_tokens (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          VARCHAR(100) NOT NULL,
    token_hash    VARCHAR(64) NOT NULL UNIQUE,
    permissions   TEXT[],
    expires_at    TIMESTAMPTZ,
    last_used_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tokens_user ON api_tokens(user_id);
CREATE INDEX idx_tokens_hash ON api_tokens(token_hash);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers
CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_skills_updated_at BEFORE UPDATE ON skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sync_sources_updated_at BEFORE UPDATE ON sync_sources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user (password: admin123)
INSERT INTO teams (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Platform Team', 'Platform Engineering Team'),
    ('00000000-0000-0000-0000-000000000002', 'Security Team', 'Security Team');

INSERT INTO users (id, username, email, password_hash, role, team_id) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', 'admin@company.com', '$2a$10$rBWJfL0zQh3VKxqR.XxqZeOYQh3z1Xn3s5kN9J5L5L5L5L5L5L5L5', 'admin',
     '00000000-0000-0000-0000-000000000001');
