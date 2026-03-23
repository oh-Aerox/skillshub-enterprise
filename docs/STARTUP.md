# 项目启动文档

**版本：** 1.0.0
**创建日期：** 2026-03-20

---

## 1. 开发环境启动

### 1.1 前置要求

确保本地安装以下工具：

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Python 3.12+](https://www.python.org/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

### 1.2 快速启动

```bash
# 1. 克隆项目
git clone https://github.com/your-org/skillshub-enterprise.git
cd skillshub-enterprise

# 2. 启动基础设施服务（PostgreSQL, Redis, MinIO）
docker-compose up -d postgres redis minio

# 3. 等待服务就绪（约 30 秒）
docker-compose ps

# 4. 初始化数据库
make db-migrate
```

### 1.3 启动各服务

#### API 服务（Go）

```bash
cd apps/api
go mod download
go run cmd/main.go
```

访问 http://localhost:3000/health 验证

#### 扫描服务（Python）

```bash
cd apps/scanner
pip install -r requirements.txt
python -m uvicorn app.main:app --reload
```

访问 http://localhost:8000/health 验证

#### Web 管理台（React）

```bash
cd apps/web
npm install
npm run dev
```

访问 http://localhost:3001 验证

### 1.4 默认账号

```
用户名：admin
密码：admin123
```

---

## 2. 生产环境部署

### 2.1 Docker Compose 部署

```bash
# 复制并修改配置
cp .env.example .env
vim .env

# 一键启动
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 2.2 Kubernetes 部署

详见 [deploy/k8s/README.md](./deploy/k8s/README.md)

---

## 3. 开发工作流

### 3.1 代码提交

```bash
# 拉取最新代码
git pull

# 创建功能分支
git checkout -b feature/your-feature

# 提交代码
git add .
git commit -m "feat: add your feature"
git push origin feature/your-feature
```

### 3.2 运行测试

```bash
# API 测试
cd apps/api && go test ./...

# 扫描服务测试
cd apps/scanner && python -m pytest

# Web 测试
cd apps/web && npm test
```

### 3.3 代码风格

**Go:**
```bash
go fmt ./...
go vet ./...
```

**TypeScript:**
```bash
npm run lint
```

---

## 4. 数据库操作

### 4.1 连接数据库

```bash
docker-compose exec postgres psql -U skillshub -d skillshub
```

### 4.2 常用查询

```sql
-- 查看所有表
\dt

-- 查看 Skill 列表
SELECT id, name, status, install_count FROM skills;

-- 查看待审核记录
SELECT r.id, sk.name, r.status FROM reviews r
JOIN scans s ON r.scan_id = s.id
JOIN skill_versions sv ON s.skill_version_id = sv.id
JOIN skills sk ON sv.skill_id = sk.id
WHERE r.status = 'pending';

-- 查看审计日志
SELECT event_type, actor_meta, created_at FROM audit_logs
ORDER BY created_at DESC LIMIT 20;
```

---

## 5. 常见问题

### Q: API 服务启动失败

**A:** 检查数据库是否就绪

```bash
docker-compose logs postgres
docker-compose exec postgres pg_isready -U skillshub
```

### Q: 无法登录 Web 管理台

**A:** 确认 API 服务运行正常

```bash
curl http://localhost:3000/health
```

### Q: 扫描服务超时

**A:** 增加沙箱超时时间

```bash
# .env
SCAN_SANDBOX_TIMEOUT=180000
```

---

## 6. 项目结构

```
skillshub-enterprise/
├── apps/
│   ├── api/           # Go API 服务
│   ├── scanner/       # Python 扫描服务
│   ├── web/           # React 管理台
│   └── cli/           # Go CLI 工具
├── packages/
│   ├── database/      # 数据库包
│   └── models/        # 数据模型
├── deploy/
│   ├── docker/        # Docker 配置
│   └── k8s/           # K8s 配置
├── docs/
│   ├── api.md         # API 文档
│   ├── database.md    # 数据库设计
│   └── deployment.md  # 部署指南
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## 7. 下一步

1. 修改默认密码
2. 配置 OIDC/SSO（可选）
3. 发布第一个 Skill
4. 邀请团队成员

---

*文档结束*
