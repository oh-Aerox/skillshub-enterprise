# API 接口文档

**版本：** 1.0.0
**Base URL:** `/api/v1`

---

## 1. 认证接口

### 1.1 用户登录

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**响应：**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-03-20T20:00:00Z",
  "user": {
    "id": "uuid",
    "username": "admin",
    "email": "admin@company.com",
    "role": "admin"
  }
}
```

### 1.2 用户注册

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "newuser",
  "email": "user@company.com",
  "password": "password123"
}
```

### 1.3 刷新 Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 1.4 登出

```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

---

## 2. Skills 接口

### 2.1 获取 Skills 列表

```http
GET /api/v1/skills?q=pdf&category=document&page=1&limit=20
Authorization: Bearer <token>
```

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `q` | string | 搜索关键词 |
| `category` | string | 分类过滤 |
| `page` | integer | 页码，默认 1 |
| `limit` | integer | 每页数量，默认 20 |

**响应：**

```json
{
  "skills": [
    {
      "id": "uuid",
      "name": "pdf-processor",
      "display_name": "PDF 处理器",
      "description": "处理 PDF 文档的 Skill",
      "category": "document",
      "tags": ["pdf", "document"],
      "source_type": "internal",
      "status": "active",
      "install_count": 100,
      "created_at": "2026-03-20T10:00:00Z"
    }
  ]
}
```

### 2.2 获取 Skill 详情

```http
GET /api/v1/skills/{skillId}
Authorization: Bearer <token>
```

### 2.3 创建 Skill

```http
POST /api/v1/skills
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "new-skill",
  "display_name": "新 Skill",
  "description": "描述",
  "category": "code",
  "tags": ["code", "generation"],
  "source_type": "internal",
  "license": "Internal"
}
```

### 2.4 更新 Skill

```http
PUT /api/v1/skills/{skillId}
Authorization: Bearer <token>
Content-Type: application/json

{
  "display_name": "更新后的名称",
  "description": "更新后的描述"
}
```

### 2.5 删除 Skill

```http
DELETE /api/v1/skills/{skillId}
Authorization: Bearer <token>
```

### 2.6 安装 Skill

```http
POST /api/v1/skills/{skillId}/install
Authorization: Bearer <token>
Content-Type: application/x-www-form-urlencoded

device_id=mac-001
```

### 2.7 卸载 Skill

```http
DELETE /api/v1/skills/{skillId}/install
Authorization: Bearer <token>
```

### 2.8 获取版本列表

```http
GET /api/v1/skills/{skillId}/versions
Authorization: Bearer <token>
```

### 2.9 下载 Skill

```http
GET /api/v1/skills/{skillId}/{version}/download
Authorization: Bearer <token>
```

**响应：**

```json
{
  "download_url": "/api/v1/files/skills/xxx.tar.gz",
  "expires_in": 600
}
```

---

## 3. 扫描接口

### 3.1 触发扫描

```http
POST /api/v1/scans/trigger
Authorization: Bearer <token>
Content-Type: application/json

{
  "skill_version_id": "uuid",
  "priority": "normal"
}
```

**响应：**

```json
{
  "scan_id": "uuid",
  "message": "Scan triggered successfully"
}
```

### 3.2 获取扫描结果

```http
GET /api/v1/scans/{scanId}
Authorization: Bearer <token>
```

**响应：**

```json
{
  "scan": {
    "id": "uuid",
    "skill_version_id": "uuid",
    "status": "completed",
    "risk_level": "A",
    "risk_score": 15,
    "summary": "Scan completed. Risk Level: A, Score: 15",
    "started_at": "2026-03-20T10:00:00Z",
    "completed_at": "2026-03-20T10:02:00Z",
    "created_at": "2026-03-20T10:00:00Z"
  }
}
```

---

## 4. 审批接口

### 4.1 获取审批列表

```http
GET /api/v1/reviews?status=pending&assignee=me
Authorization: Bearer <token>
```

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `status` | string | pending/approved/rejected/escalated |
| `assignee` | string | me - 分配给我的 |

**响应：**

```json
{
  "reviews": [
    {
      "id": "uuid",
      "scan_id": "uuid",
      "status": "pending",
      "risk_level": "B",
      "risk_score": 45,
      "skill_id": "uuid",
      "skill_name": "new-skill",
      "created_at": "2026-03-20T10:00:00Z"
    }
  ],
  "total": 1
}
```

### 4.2 获取审批详情

```http
GET /api/v1/reviews/{reviewId}
Authorization: Bearer <token>
```

### 4.3 提交审批决定

```http
PUT /api/v1/reviews/{reviewId}
Authorization: Bearer <token>
Content-Type: application/json

{
  "decision": "approved",
  "comment": "审批通过，符合安全要求"
}
```

**decision 可选值：**
- `approved` - 通过
- `rejected` - 拒绝
- `escalated` - 升级审核

---

## 5. 用户接口

### 5.1 获取个人信息

```http
GET /api/v1/user/profile
Authorization: Bearer <token>
```

### 5.2 更新个人信息

```http
PUT /api/v1/user/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "newemail@company.com"
}
```

### 5.3 获取安装列表

```http
GET /api/v1/installations
Authorization: Bearer <token>
```

---

## 6. 管理接口

### 6.1 获取用户列表

```http
GET /api/admin/v1/users?page=1&limit=20
Authorization: Bearer <token>
```

### 6.2 获取统计数据

```http
GET /api/admin/v1/stats
Authorization: Bearer <token>
```

**响应：**

```json
{
  "total_skills": 50,
  "total_installations": 500,
  "pending_reviews": 5,
  "active_users": 30
}
```

### 6.3 获取审计日志

```http
GET /api/admin/v1/audit-logs?startTime=2026-01-01&endTime=2026-03-20&eventType=SKILL_INSTALL
Authorization: Bearer <token>
```

---

## 7. 错误响应

所有错误响应遵循统一格式：

```json
{
  "error": {
    "code": "SKILL_NOT_FOUND",
    "message": "Skill 不存在",
    "details": {}
  }
}
```

**常见错误码：**

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `UNAUTHORIZED` | 401 | 未认证或 Token 无效 |
| `FORBIDDEN` | 403 | 权限不足 |
| `SKILL_NOT_FOUND` | 404 | Skill 不存在 |
| `VERSION_NOT_FOUND` | 404 | 版本不存在 |
| `ALREADY_EXISTS` | 409 | 资源已存在 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

---

## 8. 速率限制

| 接口类型 | 限制 |
|----------|------|
| 客户端 API | 100 次/分钟 |
| 管理 API | 30 次/分钟 |
| 文件下载 | 10 次/分钟 |

超过限制返回 `429 Too Many Requests`

---

*文档结束*
