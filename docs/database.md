# 数据库设计文档

**版本：** 1.0.0
**最后更新：** 2026-03-20

---

## 1. 数据库概览

SkillsHub Enterprise 使用 PostgreSQL 15 作为主数据库，存储所有核心业务数据。

### 1.1 数据表列表

| 表名 | 描述 | 数据量预估 |
|------|------|------------|
| users | 用户信息 | < 10,000 |
| teams | 团队信息 | < 1,000 |
| skills | Skill 基本信息 | < 50,000 |
| skill_versions | Skill 版本信息 | < 200,000 |
| scans | 安全扫描记录 | < 500,000 |
| reviews | 审批工单 | < 100,000 |
| installations | 安装记录 | < 1,000,000 |
| audit_logs | 审计日志 | < 10,000,000 |
| sync_sources | 同步源配置 | < 100 |
| api_tokens | API Token | < 5,000 |

---

## 2. 表结构详情

### 2.1 users

用户信息表

```sql
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
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| username | VARCHAR(50) | 用户名，唯一 |
| email | VARCHAR(255) | 邮箱，唯一 |
| password_hash | VARCHAR(255) | bcrypt 加密密码 |
| role | VARCHAR(20) | 角色：admin/security/tech_lead/developer/viewer |
| team_id | UUID | 所属团队 |
| mfa_enabled | BOOLEAN | 是否启用 MFA |
| mfa_secret | VARCHAR(255) | MFA 密钥 |
| last_login_at | TIMESTAMPTZ | 最后登录时间 |

**索引：**
- `idx_users_username` - username 列
- `idx_users_email` - email 列

---

### 2.2 teams

团队信息表

```sql
CREATE TABLE teams (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

### 2.3 skills

Skill 基本信息表

```sql
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
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | Skill 唯一标识符 |
| display_name | VARCHAR(200) | 显示名称 |
| description | TEXT | 功能描述 |
| category | VARCHAR(50) | 分类：document/analysis/code/api/other |
| tags | TEXT[] | 标签数组 |
| source_type | VARCHAR(20) | 来源：internal/opensource |
| source_url | TEXT | 原始来源 URL |
| author_id | UUID | 作者 ID |
| license | VARCHAR(50) | 许可证类型 |
| status | VARCHAR(20) | 状态：active/deprecated/blacklisted |
| install_count | INTEGER | 累计安装次数 |

**索引：**
- `idx_skills_name` - name 列
- `idx_skills_status` - status 列
- `idx_skills_category` - category 列
- `idx_skills_tags` - tags 列 (GIN)

---

### 2.4 skill_versions

Skill 版本信息表

```sql
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
    scan_id         UUID REFERENCES scans(id) ON DELETE SET NULL,
    published_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(skill_id, version)
);
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| skill_id | UUID | 所属 Skill ID |
| version | VARCHAR(20) | 语义化版本号 |
| changelog | TEXT | 变更日志 |
| storage_path | TEXT | MinIO 存储路径 |
| file_hash | VARCHAR(64) | SHA-256 哈希值 |
| file_size | BIGINT | 文件大小 (字节) |
| is_latest | BOOLEAN | 是否为最新版 |
| status | VARCHAR(20) | 状态：stable/beta/deprecated |
| scan_id | UUID | 关联的扫描记录 ID |
| published_by | UUID | 发布者 ID |

**索引：**
- `idx_versions_skill` - skill_id 列
- `idx_versions_hash` - file_hash 列
- `idx_versions_latest` - is_latest 列 (部分索引)

---

### 2.5 scans

安全扫描记录表

```sql
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
```

**风险等级说明：**

| 等级 | 分数范围 | 处理方式 |
|------|----------|----------|
| A | 0-30 | 自动审批通过 |
| B | 31-50 | 建议人工复核 |
| C | 51-70 | 必须人工审核 |
| D | 71-85 | 安全团队审查 |
| F | 86-100 | 自动拒绝 |

---

### 2.6 reviews

审批工单表

```sql
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
```

---

### 2.7 installations

安装记录表

```sql
CREATE TABLE installations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    device_id       VARCHAR(255),
    installed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    uninstalled_at  TIMESTAMPTZ,
    is_active       BOOLEAN DEFAULT true
);
```

---

### 2.8 audit_logs

审计日志表（只增不改）

```sql
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
```

**注意：** 此表不设 UPDATE/DELETE 权限，日志保留 12 个月

---

## 3. ER 图

```
┌──────────────┐       ┌──────────────┐
│    teams     │       │    users     │
├──────────────┤       ├──────────────┤
│ id (PK)      │◄──────│ id (PK)      │
│ name         │       │ username     │
│ description  │       │ email        │
│ created_at   │       │ role         │
│ updated_at   │       │ team_id (FK) │
└──────────────┘       └──────┬───────┘
                              │
                              │ 1:N
                              ▼
                        ┌──────────────┐
                        │    skills    │
                        ├──────────────┤
                        │ id (PK)      │
                        │ name         │
                        │ author_id(FK)│
                        │ status       │
                        └──────┬───────┘
                               │
                               │ 1:N
                               ▼
                        ┌──────────────┐       ┌──────────────┐
                        │skill_versions├──────►│    scans     │
                        ├──────────────┤       ├──────────────┤
                        │ id (PK)      │       │ id (PK)      │
                        │ skill_id(FK) │       │skill_ver(FK) │
                        │ version      │       │ risk_score   │
                        │ file_hash    │       │ risk_level   │
                        │ is_latest    │       └──────┬───────┘
                        └──────────────┘              │
                               │                      │ 1:1
                               │                      ▼
                               │              ┌──────────────┐
                               │              │   reviews    │
                               │              ├──────────────┤
                               │              │ id (PK)      │
                               │              │ scan_id (FK) │
                               │              │ status       │
                               ▼              └──────────────┘
                        ┌──────────────┐
                        │installations │
                        ├──────────────┤
                        │ id (PK)      │
                        │version_id(FK)│
                        │ user_id (FK) │
                        └──────────────┘
```

---

## 4. 初始化数据

```sql
-- 默认团队
INSERT INTO teams (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Platform Team', '平台工程团队'),
    ('00000000-0000-0000-0000-000000000002', 'Security Team', '安全团队');

-- 默认管理员 (密码：admin123)
INSERT INTO users (id, username, email, password_hash, role, team_id) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', 'admin@company.com',
     '$2a$10$rBWJfL0zQh3VKxqR.XxqZeOYQh3z1Xn3s5kN9J5L5L5L5L5L5L5L5',
     'admin', '00000000-0000-0000-0000-000000000001');
```

---

## 5. 性能优化建议

1. **定期清理审计日志** - 将超过 12 个月的日志归档到冷存储
2. **Skill 版本限制** - 每个 Skill 保留最近 10 个版本
3. **安装记录归档** - 已卸载的记录可定期归档
4. **索引维护** - 每周执行 `REINDEX` 维护

---

*文档结束*
