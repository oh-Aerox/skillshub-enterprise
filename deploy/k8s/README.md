# Kubernetes 部署配置

## 前置要求

- Kubernetes 1.25+
- Helm 3.0+ (可选)
- StorageClass (用于持久化存储)
- Ingress Controller (可选，用于外部访问)

## 快速部署

### 1. 创建命名空间

```bash
kubectl create namespace skillshub
```

### 2. 创建 Secret

```bash
kubectl create secret generic skillshub-secrets \
  --from-literal=jwt-secret=$(openssl rand -hex 32) \
  --from-literal=database-password=$(openssl rand -hex 16) \
  --from-literal=minio-root-user=minioadmin \
  --from-literal=minio-root-password=$(openssl rand -hex 16) \
  -n skillshub
```

### 3. 应用配置

```bash
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/postgres.yaml
kubectl apply -f deploy/k8s/redis.yaml
kubectl apply -f deploy/k8s/minio.yaml
kubectl apply -f deploy/k8s/api.yaml
kubectl apply -f deploy/k8s/scanner.yaml
kubectl apply -f deploy/k8s/web.yaml
```

### 4. 验证部署

```bash
kubectl get pods -n skillshub
kubectl get svc -n skillshub
```

## 文件列表

- `namespace.yaml` - 命名空间定义
- `configmap.yaml` - 配置项
- `postgres.yaml` - PostgreSQL 状态ful 集
- `redis.yaml` - Redis 状态ful 集
- `minio.yaml` - MinIO 状态ful 集
- `api.yaml` - API 服务部署
- `scanner.yaml` - 扫描服务部署
- `web.yaml` - Web 前端部署
- `ingress.yaml` - Ingress 规则（可选）

## 水平扩展

```bash
# 扩展 API 服务
kubectl scale deployment skillshub-api --replicas=3 -n skillshub

# 扩展扫描服务
kubectl scale deployment skillshub-scanner --replicas=5 -n skillshub
```

## 监控

```bash
# 查看资源使用
kubectl top pods -n skillshub

# 查看日志
kubectl logs -f deployment/skillshub-api -n skillshub
```
