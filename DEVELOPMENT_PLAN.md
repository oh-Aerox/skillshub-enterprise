# SkillsHub Enterprise - 开发计划文档

**文档版本：** v1.0.0
**创建日期：** 2026-03-20
**项目状态：** 启动中

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术架构](#2-技术架构)
3. [开发阶段规划](#3-开发阶段规划)
4. [项目结构](#4-项目结构)
5. [数据库设计](#5-数据库设计)
6. [API 接口规范](#6-api-接口规范)
7. [部署指南](#7-部署指南)

---

## 1. 项目概述

### 1.1 产品名称

**SkillsHub Enterprise** —— 企业级 AI Agent Skill 私有仓库平台

### 1.2 产品定位

面向企业团队的 AI Skill 安全治理平台，定位于 Claude Code 用户与开源 Skill 生态之间的安全网关。

### 1.3 核心价值

| 价值维度 | 描述 |
|----------|------|
| **安全可控** | 所有 Skill 须经多层安全扫描后方可入库，从源头消除供应链风险 |
| **管控有据** | 完整记录 Skill 来源、审查过程、安装使用情况，满足合规审计需求 |
| **效率提升** | 自动从开源仓库引入并扫描，减少人工干预 |
| **团队协作** | 统一 Skill 版本，消除成员间环境不一致问题 |

### 1.4 目标用户

| 用户角色 | 核心需求 | 主要使用场景 |
|----------|----------|--------------|
| 企业安全管理员 | 对团队 Skill 使用有完整管控 | 配置扫描规则、审批入库申请、处理安全告警 |
| 研发工程师 | 快速找到并安装所需 Skill | 搜索 Skill、发起入库申请、查看审批进度 |
| 平台运维人员 | 保证私服高可用和数据安全 | 监控平台健康、管理备份、处理故障 |
| 团队技术负责人 | 了解团队 Skill 使用全貌 | 查看使用报告、审批中高风险 Skill |

---

## 2. 技术架构

### 2.1 系统分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                    客户端层 (Client Layer)                    │
│  ┌──────────────┐    ┌─────────────────────────────────┐   │
│  │  Claude Code  │───▶│  SkillsHub CLI / Agent Plugin   │   │
│  └──────────────┘    └───────────────┬─────────────────┘   │
└──────────────────────────────────────│─────────────────────┘
                                       │ HTTPS
┌──────────────────────────────────────▼─────────────────────┐
│                    API 网关层 (Gateway Layer)                  │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────┐   │
│  │  API Gateway │  │  Web 管理台  │  │  后台任务队列    │   │
│  │  (Go/Gin)   │  │  (React)    │  │  (Redis/BullMQ)  │   │
│  └──────┬──────┘  └──────┬──────┘  └────────┬─────────┘   │
└─────────│────────────────│──────────────────│─────────────┘
          │                │                   │
┌─────────▼────────────────▼──────────────────▼─────────────┐
│                    业务服务层 (Service Layer)                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │仓库服务  │  │扫描服务  │  │审核服务  │  │用户服务  │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │通知服务  │  │审计服务  │  │策略服务  │  │同步服务  │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└────────────────────────┬───────────────────────────────────┘
                         │
┌────────────────────────▼───────────────────────────────────┐
│                      数据层 (Data Layer)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐ │
│  │  PostgreSQL  │  │    Redis     │  │   MinIO/S3       │ │
│  │  (元数据)    │  │  (缓存/队列)  │  │  (Skill 文件)      │ │
│  └──────────────┘  └──────────────┘  └──────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 技术栈选型

| 层次 | 技术 | 理由 |
|------|------|------|
| 后端 API | Go 1.22 + Gin | 高性能、低内存、单二进制部署 |
| 扫描服务 | Python 3.12 + FastAPI | 丰富的 NLP/安全分析生态 |
| 客户端代理 | Go | 跨平台、无运行时依赖 |
| 前端控制台 | React 18 + TypeScript + Ant Design | 企业级组件库，类型安全 |
| 数据库 | PostgreSQL 15+ | 可靠的关系型存储，JSONB 灵活字段 |
| 缓存/队列 | Redis 7+ | 扫描任务队列，热数据缓存 |
| 对象存储 | MinIO | 自托管，S3 兼容 |
| 沙箱隔离 | gVisor/Docker | 轻量快速，系统调用级隔离 |

### 2.3 客户端四重拦截

```
┌─────────────────────────────────────────────────────────────┐
│  拦截层 1: Plugin Marketplace 重定向                          │
│  配置文件 ~/\.claude/settings.json，强制指向私服地址            │
├─────────────────────────────────────────────────────────────┤
│  拦截层 2: 文件系统权限控制                                    │
│  OS 级权限，仅 skillhub-agent 可写入 skills 目录               │
├─────────────────────────────────────────────────────────────┤
│  拦截层 3: 实时监控守护进程                                    │
│  inotify/FSEvents 监控，未授权文件立即删除+告警                 │
├─────────────────────────────────────────────────────────────┤
│  拦截层 4: CLAUDE.md 策略注入                                 │
│  AI 行为层软控制，拒绝执行未授权 Skill 安装                      │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. 开发阶段规划

### Phase 1: 基础框架 (第 1-4 周)

**目标：** 完成项目骨架、数据库设计、基础 API

| 任务 | 描述 | 优先级 |
|------|------|--------|
| P1.1 | 项目初始化，Monorepo 结构搭建 | P0 |
| P1.2 | PostgreSQL 数据库设计与迁移 | P0 |
| P1.3 | 用户认证服务 (JWT + OIDC) | P0 |
| P1.4 | Skill 基础 CRUD API | P0 |
| P1.5 | MinIO 文件存储集成 | P1 |
| P1.6 | 基础 Web 管理台框架 | P1 |

**验收标准：** 用户可以登录系统，上传/下载 Skill 文件包

### Phase 2: 安全扫描引擎 (第 5-10 周)

**目标：** 实现三层安全扫描流水线

| 任务 | 描述 | 优先级 |
|------|------|--------|
| P2.1 | Layer 1: 静态分析扫描 (结构/格式) | P0 |
| P2.2 | Layer 2: 提示词意图分析 (LLM) | P0 |
| P2.3 | Layer 2: 工具调用白名单检查 | P0 |
| P2.4 | Layer 2: 敏感信息/密钥扫描 | P0 |
| P2.5 | Layer 3: Docker 沙箱环境搭建 | P0 |
| P2.6 | Layer 3: 行为监控 (syscall/网络/文件) | P0 |
| P2.7 | Layer 4: 供应链溯源分析 | P1 |
| P2.8 | 综合风险评分算法 | P0 |

**验收标准：** 扫描引擎可正确识别高风险 Skill，评分准确

### Phase 3: 审批工作流 (第 11-12 周)

**目标：** 实现人工审核流程与通知系统

| 任务 | 描述 | 优先级 |
|------|------|--------|
| P3.1 | 审批状态机 (FSM) 实现 | P0 |
| P3.2 | 审核工作台 UI | P0 |
| P3.3 | 通知服务 (邮件/Slack/企业微信) | P1 |
| P3.4 | 审批 SLA 管理与超时告警 | P1 |

**验收标准：** 完整的申请 - 扫描 - 审批 - 入库流程跑通

### Phase 4: 客户端集成 (第 13-14 周)

**目标：** 实现客户端拦截与 CLI 工具

| 任务 | 描述 | 优先级 |
|------|------|--------|
| P4.1 | SkillsHub CLI 工具 | P0 |
| P4.2 | skillhub-agent 守护进程 | P0 |
| P4.3 | 文件系统监控 (inotify/FSEvents) | P0 |
| P4.4 | CLAUDE.md 策略注入工具 | P1 |
| P4.5 | MDM/Ansible 部署脚本 | P1 |

**验收标准：** 开发者可通过 CLI 安装 Skill，无法绕过私服

### Phase 5: 管理控制台 (第 15-17 周)

**目标：** 完成 Web 管理台全部功能

| 任务 | 描述 | 优先级 |
|------|------|--------|
| P5.1 | 仪表盘 (Dashboard) | P0 |
| P5.2 | Skill 管理页 | P0 |
| P5.3 | 审批中心 | P0 |
| P5.4 | 审计日志查询 | P1 |
| P5.5 | 统计报表 | P1 |
| P5.6 | RBAC 权限管理 | P0 |

**验收标准：** 管理员可通过 Web 台完成所有管理操作

### Phase 6: 集成测试与上线 (第 18-20 周)

**目标：** 端到端测试、性能压测、试点部署

| 任务 | 描述 | 优先级 |
|------|------|--------|
| P6.1 | 端到端测试 (Playwright) | P0 |
| P6.2 | 性能压测 (k6) | P0 |
| P6.3 | 安全渗透测试 | P0 |
| P6.4 | 文档完善 (用户手册/运维手册) | P1 |
| P6.5 | 试点团队灰度发布 | P0 |

**验收标准：** 通过试点团队验证，准备 GA 发布

---

## 4. 项目结构

```
skillshub-enterprise/
├── apps/
│   ├── api/                    # Go API 服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── handlers/
│   │   │   ├── services/
│   │   │   ├── models/
│   │   │   └── middleware/
│   │   └── pkg/
│   ├── scanner/                # Python 扫描服务
│   │   ├── app/
│   │   ├── analyzers/
│   │   │   ├── static/
│   │   │   ├── sandbox/
│   │   │   └── supply_chain/
│   │   └── requirements.txt
│   ├── web/                    # React 管理台
│   │   ├── src/
│   │   │   ├── components/
│   │   │   ├── pages/
│   │   │   ├── services/
│   │   │   └── hooks/
│   │   └── package.json
│   └── cli/                    # Go CLI 工具
│       ├── cmd/
│       └── internal/
├── packages/
│   ├── database/               # 共享数据库包
│   │   ├── migrations/
│   │   └── models/
│   ├── sdk/                    # 共享 SDK
│   └── types/                  # 共享类型定义
├── deploy/
│   ├── docker/
│   │   ├── api/
│   │   ├── scanner/
│   │   └── web/
│   ├── k8s/
│   └── docker-compose.yml
├── docs/
│   ├── api.md
│   ├── database.md
│   └── deployment.md
├── Makefile
└── README.md
```

---

## 5. 数据库设计

### 5.1 核心数据表

#### users 表

```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    role          VARCHAR(20) NOT NULL DEFAULT 'developer',  -- admin/security/tech_lead/developer/viewer
    team_id       UUID REFERENCES teams(id),
    mfa_enabled   BOOLEAN DEFAULT false,
    mfa_secret    VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### teams 表

```sql
CREATE TABLE teams (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### skills 表

```sql
CREATE TABLE skills (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL UNIQUE,
    display_name  VARCHAR(200),
    description   TEXT NOT NULL,
    category      VARCHAR(50),
    tags          TEXT[],
    source_type   VARCHAR(20) NOT NULL,  -- internal/opensource
    source_url    TEXT,
    author_id     UUID REFERENCES users(id),
    license       VARCHAR(50) DEFAULT 'Internal',
    status        VARCHAR(20) NOT NULL DEFAULT 'active',  -- active/deprecated/blacklisted
    install_count INTEGER DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_skills_name ON skills(name);
CREATE INDEX idx_skills_status ON skills(status);
CREATE INDEX idx_skills_category ON skills(category);
```

#### skill_versions 表

```sql
CREATE TABLE skill_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id        UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version         VARCHAR(20) NOT NULL,
    changelog       TEXT,
    storage_path    TEXT NOT NULL,      -- MinIO 路径
    file_hash       VARCHAR(64) NOT NULL,  -- SHA-256
    file_size       BIGINT,
    is_latest       BOOLEAN DEFAULT false,
    status          VARCHAR(20) DEFAULT 'stable',  -- stable/beta/deprecated
    scan_id         UUID REFERENCES scans(id),
    published_by    UUID REFERENCES users(id),
    published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(skill_id, version)
);

CREATE INDEX idx_versions_skill ON skill_versions(skill_id);
CREATE INDEX idx_versions_hash ON skill_versions(file_hash);
```

#### scans 表

```sql
CREATE TABLE scans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/running/completed/failed
    risk_level      CHAR(1),              -- A/B/C/D/F
    risk_score      SMALLINT,             -- 0-100
    layer1_result   JSONB,                -- 结构扫描结果
    layer2_result   JSONB,                -- 静态分析结果
    layer3_result   JSONB,                -- 沙箱测试结果
    layer4_result   JSONB,                -- 供应链分析结果
    summary         TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scans_version ON scans(skill_version_id);
CREATE INDEX idx_scans_status ON scans(status);
```

#### reviews 表

```sql
CREATE TABLE reviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id         UUID NOT NULL REFERENCES scans(id),
    applicant_id    UUID REFERENCES users(id),
    assignee_id     UUID REFERENCES users(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/approved/rejected/escalated
    decision        VARCHAR(20),   -- approved/rejected/escalated
    comment         TEXT,
    conditions      JSONB,         -- 附加限制条件
    due_at          TIMESTAMPTZ,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_status ON reviews(status);
CREATE INDEX idx_reviews_assignee ON reviews(assignee_id);
```

#### installations 表

```sql
CREATE TABLE installations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_version_id UUID NOT NULL REFERENCES skill_versions(id),
    user_id         UUID REFERENCES users(id),
    device_id       VARCHAR(255),
    installed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    uninstalled_at  TIMESTAMPTZ,
    is_active       BOOLEAN DEFAULT true
);

CREATE INDEX idx_installations_user ON installations(user_id);
CREATE INDEX idx_installations_skill ON installations(skill_version_id);
```

#### audit_logs 表

```sql
CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type  VARCHAR(50) NOT NULL,
    actor_id    UUID,
    actor_meta  JSONB,    -- IP、设备信息
    resource    JSONB,    -- 操作对象信息
    result      VARCHAR(20),  -- success/failure
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- 注意：此表不设 UPDATE/DELETE 权限
);

CREATE INDEX idx_audit_event ON audit_logs(event_type);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at);
```

#### sync_sources 表

```sql
CREATE TABLE sync_sources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    source_type     VARCHAR(20) NOT NULL,  -- github/npm/git/http
    url             TEXT NOT NULL,
    config          JSONB,
    is_enabled      BOOLEAN DEFAULT true,
    last_sync_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 5.2 ER 图

```
┌─────────────┐       ┌─────────────┐
│    users    │       │    teams    │
├─────────────┤       ├─────────────┤
│ id          │       │ id          │
│ username    │◄──────│ name        │
│ email       │       │ description │
│ role        │       └─────────────┘
│ team_id     │───┐
└─────────────┘   │
                  │
┌─────────────┐   │
│   skills    │   │
├─────────────┤   │
│ id          │   │
│ name        │   │
│ author_id   │───┘
│ status      │
└──────┬──────┘
       │
       │ 1:N
       ▼
┌─────────────┐       ┌─────────────┐
│skill_versions├──────►│   scans     │
├─────────────┤       ├─────────────┤
│ id          │       │ id          │
│ skill_id    │       │ skill_ver_id│
│ version     │       │ risk_score  │
│ file_hash   │       │ risk_level  │
└──────┬──────┘       └──────┬──────┘
       │                     │
       │ 1:N                 │ 1:1
       ▼                     ▼
┌─────────────┐       ┌─────────────┐
│installations│       │   reviews   │
├─────────────┤       ├─────────────┤
│ id          │       │ id          │
│ version_id  │       │ scan_id     │
│ user_id     │       │ status      │
└─────────────┘       │ decision    │
                      └─────────────┘
```

---

## 6. API 接口规范

### 6.1 通用规范

- **Base URL**: `/api/v1`
- **认证**: `Authorization: Bearer <jwt-token>`
- **内容类型**: `Content-Type: application/json`
- **错误格式**:
```json
{
  "error": {
    "code": "SKILL_NOT_FOUND",
    "message": "Skill 不存在",
    "details": {}
  }
}
```

### 6.2 核心接口

#### 认证接口

```http
POST /api/v1/auth/login
Body: { "username": "", "password": "" }

POST /api/v1/auth/refresh
Body: { "refresh_token": "" }

POST /api/v1/auth/logout
```

#### Skills 查询

```http
GET /api/v1/skills?q={keyword}&category={cat}&page={n}&limit={n}
GET /api/v1/skills/{skillId}
GET /api/v1/skills/{skillId}/versions
GET /api/v1/skills/{skillId}/{version}/download
```

#### Skill 安装

```http
POST /api/v1/skills/{skillId}/install
DELETE /api/v1/skills/{skillId}/install
GET /api/v1/installations
```

#### 开源引入

```http
POST /api/v1/imports
Body: { "source": "github", "url": "...", "reason": "..." }
GET /api/v1/imports/{importId}
```

#### 扫描结果

```http
GET /api/v1/scans/{scanId}
GET /api/v1/skills/{skillId}/scans/latest
```

#### 审核接口

```http
GET /api/v1/reviews?status=pending
PUT /api/v1/reviews/{reviewId}
Body: { "decision": "approve/reject", "comment": "" }
```

### 6.3 Webhook 事件

支持推送的事件：
- `skill.published`
- `import.requested`
- `scan.completed`
- `review.completed`
- `skill.installed`
- `security.alert`

---

## 7. 部署指南

### 7.1 开发环境部署

```bash
# 克隆项目
git clone https://github.com/your-org/skillshub-enterprise.git
cd skillshub-enterprise

# 启动依赖服务
docker-compose up -d postgres redis minio

# 初始化数据库
make db-migrate

# 启动 API 服务
cd apps/api && go run cmd/main.go

# 启动扫描服务
cd apps/scanner && python -m uvicorn app.main:app --reload

# 启动 Web 管理台
cd apps/web && npm install && npm run dev
```

### 7.2 环境变量配置

```bash
# .env
NODE_ENV=development
PORT=3000

# 数据库
DATABASE_URL=postgresql://skillshub:password@localhost:5432/skillshub

# Redis
REDIS_URL=redis://localhost:6379

# 对象存储
STORAGE_TYPE=minio
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=skillshub-files

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=8h

# 扫描配置
SCAN_SANDBOX_TIMEOUT=120000
AUTO_APPROVE_MAX_SCORE=30
```

### 7.3 生产部署

详见 [deploy/k8s/README.md](./deploy/k8s/README.md)

---

## 附录

### A. 术语表

| 术语 | 定义 |
|------|------|
| Skill | Claude Code 的功能扩展模块，由 SKILL.md 和附属脚本组成 |
| SKILL.md | Skill 的核心文件，包含 YAML 前置元数据和指令内容 |
| 私服 | 企业内部部署的 Skill 托管与分发平台 |
| 安全扫描 | 对 Skill 进行静态分析和动态沙箱测试的过程 |
| 提示词注入 | 通过精心构造的文本劫持 AI 模型行为的攻击方式 |

### B. 参考资料

- Claude Code 官方文档
- OWASP LLM Top 10
- gVisor 容器安全隔离文档

---

*文档结束*
