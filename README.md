# SkillsHub Enterprise

企业级 AI Agent Skill 私有仓库平台

## 项目概述

SkillsHub Enterprise 是面向企业团队的 AI Skill 安全治理平台，提供：

- 🔒 **安全扫描** - 三层安全扫描引擎（静态分析 + 沙箱测试 + 供应链分析）
- 📦 **私有仓库** - Skill 存储、版本管理与分发
- 👥 **权限管控** - RBAC 角色权限体系
- 📊 **审计追踪** - 完整的操作日志与合规报告
- 🚀 **CLI 工具** - 无缝集成 Claude Code

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 API | Go 1.22 + Gin |
| 扫描服务 | Python 3.12 + FastAPI |
| 前端控制台 | React 18 + TypeScript + Ant Design |
| 数据库 | PostgreSQL 15 |
| 缓存/队列 | Redis 7 |
| 对象存储 | MinIO |

## 快速开始

### 开发环境启动

```bash
# 克隆项目
git clone https://github.com/oh-Aerox/skillshub-enterprise.git
cd skillshub-enterprise

# 启动依赖服务 (PostgreSQL, Redis, MinIO)
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

### 访问地址

- Web 管理台：http://localhost:3001
- API 服务：http://localhost:3000
- 扫描服务：http://localhost:8000

## 项目结构

```
skillshub-enterprise/
├── apps/
│   ├── api/           # Go API 服务
│   ├── scanner/       # Python 扫描服务
│   ├── web/           # React 管理台
│   └── cli/           # Go CLI 工具
├── packages/
│   ├── database/      # 数据库包
│   ├── sdk/           # 共享 SDK
│   └── types/         # 类型定义
├── deploy/
│   ├── docker/        # Docker 配置
│   └── k8s/           # K8s 配置
└── docs/              # 文档
```

## 核心功能

### 1. 安全扫描引擎

三层扫描流水线：

1. **静态分析** - 提示词意图分析、工具调用白名单、敏感信息扫描
2. **沙箱测试** - Docker 隔离环境中的行为监控
3. **供应链分析** - 来源信誉、版本 Diff、许可证合规

### 2. 客户端四重拦截

- Plugin Marketplace 重定向
- 文件系统权限控制
- 实时监控守护进程
- CLAUDE.md 策略注入

### 3. 审批工作流

- 自动审批（低风险 Skill）
- 人工审核（中高风险 Skill）
- 多渠道通知（邮件/Slack/企业微信）

## 开发计划

详见 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md)

## 文档

- [数据库设计](./docs/database.md)
- [API 接口规范](./docs/api.md)
- [部署指南](./docs/deployment.md)

## License

Internal Use Only
