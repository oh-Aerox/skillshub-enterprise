# 部署指南

**版本：** 1.0.0
**最后更新：** 2026-03-20

---

## 1. 部署架构

### 1.1 最小化部署（开发/测试环境）

```
┌─────────────────────────────────────────┐
│            Docker Compose               │
├─────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐ │
│  │   API   │  │ Scanner │  │   Web   │ │
│  │  :3000  │  │  :8000  │  │  :3001  │ │
│  └────┬────┘  └────┬────┘  └────┬────┘ │
│       │            │            │       │
│  ┌────┴────────────┴────────────┴────┐ │
│  │        Nginx (Reverse Proxy)      │ │
│  └─────────────────┬─────────────────┘ │
│                    │                    │
│  ┌─────┐  ┌───────┴───────┐  ┌──────┐ │
│  │ PG  │  │     Redis     │  │MinIO │ │
│  │:5432│  │     :6379     │  │:9000 │ │
│  └─────┘  └───────────────┘  └──────┘ │
└─────────────────────────────────────────┘
```

### 1.2 生产环境部署（Kubernetes）

详见 [deploy/k8s/README.md](./deploy/k8s/README.md)

---

## 2. 快速开始（Docker Compose）

### 2.1 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 4GB 可用内存
- 至少 10GB 可用磁盘空间

### 2.2 启动服务

```bash
# 克隆项目
git clone https://github.com/your-org/skillshub-enterprise.git
cd skillshub-enterprise

# 复制环境变量文件
cp .env.example .env

# 修改配置（可选）
vim .env

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 2.3 访问服务

| 服务 | 地址 | 说明 |
|------|------|------|
| Web 管理台 | http://localhost:3001 | 前端控制台 |
| API 服务 | http://localhost:3000 | REST API |
| 扫描服务 | http://localhost:8000 | 安全扫描引擎 |
| PostgreSQL | localhost:5432 | 数据库 |
| Redis | localhost:6379 | 缓存/队列 |
| MinIO | http://localhost:9000 | 对象存储 |
| MinIO Console | http://localhost:9001 | MinIO 管理台 |

### 2.4 默认账号

```
用户名：admin
密码：admin123
```

---

## 3. 配置说明

### 3.1 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `PORT` | API 服务端口 | 3000 |
| `DATABASE_URL` | PostgreSQL 连接串 | postgresql://skillshub:skillshub_password@postgres:5432/skillshub |
| `REDIS_URL` | Redis 连接串 | redis://redis:6379 |
| `STORAGE_TYPE` | 存储类型 | minio |
| `MINIO_ENDPOINT` | MinIO 地址 | minio:9000 |
| `MINIO_ACCESS_KEY` | MinIO 访问密钥 | minioadmin |
| `MINIO_SECRET_KEY` | MinIO 密钥 | minioadmin |
| `MINIO_BUCKET` | MinIO Bucket 名称 | skillshub-files |
| `JWT_SECRET` | JWT 密钥 | 请修改为随机字符串 |
| `JWT_EXPIRY` | JWT 过期时间 | 8h |
| `SCANNER_SERVICE_URL` | 扫描服务地址 | http://scanner:8000 |
| `AUTO_APPROVE_MAX_SCORE` | 自动审批最大风险分 | 30 |

### 3.2 生产环境配置

```bash
# .env.production
# 数据库
DATABASE_URL=postgresql://user:password@prod-db:5432/skillshub

# Redis（建议配置密码）
REDIS_URL=redis://:password@prod-redis:6379

# JWT（必须修改为随机密钥）
JWT_SECRET=<使用 openssl rand -hex 32 生成>

# MinIO（生产环境请使用独立部署）
MINIO_ENDPOINT=prod-minio.internal:9000
MINIO_ACCESS_KEY=<生产密钥>
MINIO_SECRET_KEY=<生产密钥>

# OIDC/SSO（可选）
OIDC_ENABLED=true
OIDC_ISSUER=https://sso.company.com
OIDC_CLIENT_ID=skillshub
OIDC_CLIENT_SECRET=<SSO 密钥>
```

---

## 4. 数据库初始化

### 4.1 自动初始化

首次启动时，数据库会自动执行初始化脚本：

```bash
docker-compose exec postgres psql -U skillshub -d skillshub -f /docker-entrypoint-initdb.d/001_initial_schema.sql
```

### 4.2 手动迁移

```bash
# 查看迁移状态
docker-compose exec postgres psql -U skillshub -d skillshub -c "\d"

# 检查表结构
docker-compose exec postgres psql -U skillshub -d skillshub -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public';"
```

---

## 5. 健康检查

```bash
# 检查所有服务状态
docker-compose ps

# 检查 API 健康
curl http://localhost:3000/health

# 检查扫描服务健康
curl http://localhost:8000/health

# 检查数据库连接
docker-compose exec postgres pg_isready -U skillshub

# 检查 Redis 连接
docker-compose exec redis redis-cli ping
```

---

## 6. 日志管理

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f api
docker-compose logs -f scanner

# 按时间过滤
docker-compose logs --since="2024-01-01" api

# 导出日志
docker-compose logs api > api.log
```

---

## 7. 备份与恢复

### 7.1 数据库备份

```bash
# 备份数据库
docker-compose exec postgres pg_dump -U skillshub skillshub > backup_$(date +%Y%m%d).sql

# 恢复数据库
docker-compose exec -T postgres psql -U skillshub skillshub < backup_20240101.sql
```

### 7.2 MinIO 备份

```bash
# 使用 mc 工具备摄 MinIO 数据
mc alias set myminio http://localhost:9000 minioadmin minioadmin
mc mirror myminio/skillshub-files ./backup/minio
```

---

## 8. 故障排查

### 8.1 API 服务无法启动

```bash
# 检查日志
docker-compose logs api

# 常见原因：
# 1. 数据库未就绪 - 等待 postgres 健康检查通过
# 2. 端口被占用 - 修改 docker-compose.yml 中的端口映射
# 3. 配置错误 - 检查 .env 文件
```

### 8.2 扫描服务超时

```bash
# 增加沙箱超时时间
# 编辑 .env:
SCAN_SANDBOX_TIMEOUT=180000  # 3 分钟

# 重启服务
docker-compose restart scanner
```

### 8.3 数据库连接失败

```bash
# 检查数据库状态
docker-compose ps postgres

# 检查连接串
docker-compose exec api env | grep DATABASE

# 重启数据库
docker-compose restart postgres
```

---

## 9. 性能优化

### 9.1 数据库优化

```sql
-- 添加索引（根据实际查询需求）
CREATE INDEX CONCURRENTLY idx_skills_category_status ON skills(category, status);
CREATE INDEX CONCURRENTLY idx_scans_created_desc ON scans(created_at DESC);

-- 定期清理
VACUUM ANALYZE skills;
VACUUM ANALYZE scans;
```

### 9.2 Redis 优化

```bash
# 配置 Redis 内存限制
docker-compose exec redis redis-cli CONFIG SET maxmemory 512mb
docker-compose exec redis redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

---

## 10. 安全建议

1. **修改默认密码** - 立即修改 admin 账号密码
2. **配置 HTTPS** - 生产环境必须使用 HTTPS
3. **限制网络访问** - 仅开放必要端口
4. **定期更新** - 及时应用安全补丁
5. **启用审计日志** - 记录所有关键操作
6. **配置备份** - 定期备份数据库和文件

---

*文档结束*
