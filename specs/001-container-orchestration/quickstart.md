# 快速开始指南

**创建日期**: 2025-10-25
**版本**: 1.0.0
**适用项目**: 容器编排管理平台 (001-container-orchestration)

## 概述

本指南将帮助您快速搭建和运行容器编排管理平台。平台提供可视化的Kubernetes容器管理界面，支持容器的完整生命周期管理。

## 系统要求

### 开发环境
- **Go**: 1.21 或更高版本
- **Node.js**: 18.x 或更高版本
- **PostgreSQL**: 14.x 或更高版本
- **Kubernetes**: 1.25 或更高版本（用于开发和测试）
- **Docker**: 20.x 或更高版本
- **Git**: 最新版本

### 生产环境
- **Kubernetes集群**: 1.25+，至少3个节点
- **PostgreSQL**: 14.x，建议使用托管服务
- **负载均衡器**: 支持HTTP/HTTPS
- **存储**: 至少100GB可用空间

## 快速部署

### 1. 克隆项目
```bash
git clone https://github.com/your-org/container-orchestration-platform.git
cd container-orchestration-platform
git checkout 001-container-orchestration
```

### 2. 环境配置

#### 后端配置
```bash
# 复制环境配置文件
cp backend/config/config.example.yaml backend/config/config.yaml

# 编辑配置文件
vim backend/config/config.yaml
```

**config.yaml 示例**:
```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s

database:
  host: "localhost"
  port: 5432
  name: "container_management"
  user: "postgres"
  password: "your_password"
  ssl_mode: "disable"
  max_connections: 20

kubernetes:
  config_path: ""  # 留空使用集群内配置
  namespace: "default"
  timeout: 30s

auth:
  jwt_secret: "your_jwt_secret_key_here"
  token_expiry: 24h
  refresh_token_expiry: 168h  # 7天

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

#### 前端配置
```bash
# 复制环境配置文件
cp frontend/.env.example frontend/.env.local

# 编辑配置文件
vim frontend/.env.local
```

**.env.local 示例**:
```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/v1/ws
NEXT_PUBLIC_APP_NAME=容器编排管理平台
NEXT_PUBLIC_VERSION=1.0.0
```

### 3. 数据库初始化

#### 创建数据库
```bash
# 连接到PostgreSQL
psql -U postgres -h localhost

# 创建数据库
CREATE DATABASE container_management;
CREATE USER container_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE container_management TO container_user;
\q
```

#### 运行数据库迁移
```bash
cd backend
go run cmd/migrate/main.go up
```

### 4. 启动后端服务

#### 安装依赖
```bash
cd backend
go mod download
```

#### 运行服务
```bash
# 开发模式
go run cmd/server/main.go

# 或使用make命令
make dev
```

验证后端服务：
```bash
curl http://localhost:8080/health
```

### 5. 启动前端服务

#### 安装依赖
```bash
cd frontend
npm install
```

#### 运行开发服务器
```bash
npm run dev
```

访问前端应用: http://localhost:3000

### 6. 创建管理员用户

```bash
# 使用后端API创建管理员用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123",
    "fullName": "系统管理员",
    "role": "admin"
  }'
```

## Docker部署

### 1. 构建镜像

```bash
# 构建后端镜像
cd backend
docker build -t container-platform-backend:latest .

# 构建前端镜像
cd frontend
docker build -t container-platform-frontend:latest .
```

### 2. 使用Docker Compose部署

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

**docker-compose.yml 示例**:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: container_management
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"

  backend:
    image: container-platform-backend:latest
    environment:
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_NAME=container_management
      - DATABASE_USER=postgres
      - DATABASE_PASSWORD=postgres
      - JWT_SECRET=your_jwt_secret_here
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    volumes:
      - ~/.kube:/root/.kube:ro  # 挂载Kubernetes配置

  frontend:
    image: container-platform-frontend:latest
    environment:
      - NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
    ports:
      - "3000:3000"
    depends_on:
      - backend

volumes:
  postgres_data:
```

## Kubernetes部署

### 1. 准备Kubernetes配置

```bash
# 创建命名空间
kubectl create namespace container-platform

# 应用数据库配置
kubectl apply -f deployments/database/
```

**database-deployment.yaml 示例**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: container-platform
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14
        env:
        - name: POSTGRES_DB
          value: container_management
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
```

### 2. 部署后端服务

```bash
# 创建配置和密钥
kubectl apply -f deployments/backend/config/
kubectl apply -f deployments/backend/secrets/

# 部署后端服务
kubectl apply -f deployments/backend/
```

### 3. 部署前端服务

```bash
kubectl apply -f deployments/frontend/
```

### 4. 配置Ingress

```bash
kubectl apply -f deployments/ingress/
```

**ingress.yaml 示例**:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: container-platform-ingress
  namespace: container-platform
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: platform.example.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 3000
```

## 验证部署

### 1. 检查服务状态

```bash
# 检查Pod状态
kubectl get pods -n container-platform

# 检查服务状态
kubectl get services -n container-platform

# 检查Ingress状态
kubectl get ingress -n container-platform
```

### 2. 健康检查

```bash
# 后端健康检查
curl http://your-domain/api/health

# 前端访问
curl http://your-domain/
```

### 3. 功能测试

1. **登录测试**: 使用管理员账户登录系统
2. **容器操作**: 创建一个简单的Nginx容器
3. **监控查看**: 验证容器状态和资源使用情况
4. **日志查看**: 查看容器运行日志

## 常见问题

### Q: 后端启动失败，数据库连接错误
**A**: 检查数据库配置和连接信息：
```bash
# 测试数据库连接
psql -h localhost -U postgres -d container_management

# 检查数据库迁移状态
go run cmd/migrate/main.go status
```

### Q: 前端无法连接后端API
**A**: 检查CORS配置和API地址：
1. 确认后端服务正常运行
2. 检查前端环境变量配置
3. 验证防火墙和网络设置

### Q: Kubernetes操作失败
**A**: 检查Kubernetes配置：
1. 确认kubeconfig文件正确配置
2. 验证RBAC权限设置
3. 检查目标命名空间是否存在

### Q: 容器创建卡在pending状态
**A**: 检查集群资源：
```bash
# 查看节点资源
kubectl describe nodes

# 查看Pod事件
kubectl describe pod <pod-name> -n <namespace>

# 检查资源配额
kubectl describe resourcequota -n <namespace>
```

## 开发指南

### 后端开发

#### 项目结构
```
backend/
├── cmd/
│   └── server/          # 应用入口
├── internal/
│   ├── api/             # HTTP处理器
│   ├── service/         # 业务逻辑
│   ├── repository/      # 数据访问
│   └── k8s/            # Kubernetes客户端
├── pkg/                 # 公共包
├── migrations/          # 数据库迁移
└── tests/              # 测试文件
```

#### 运行测试
```bash
# 单元测试
go test ./...

# 集成测试
go test -tags=integration ./...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 前端开发

#### 项目结构
```
frontend/
├── src/
│   ├── components/      # 可重用组件
│   ├── pages/          # 页面组件
│   ├── store/          # Redux store
│   ├── services/       # API服务
│   └── utils/          # 工具函数
├── public/             # 静态资源
└── tests/              # 测试文件
```

#### 运行测试
```bash
# 单元测试
npm test

# E2E测试
npm run test:e2e

# 代码格式检查
npm run lint

# 类型检查
npm run type-check
```

## 性能优化

### 后端优化
1. **数据库优化**: 定期执行VACUUM和ANALYZE
2. **连接池配置**: 合理设置数据库连接池大小
3. **缓存策略**: 使用Redis缓存热点数据
4. **并发控制**: 使用goroutine池控制并发数量

### 前端优化
1. **代码分割**: 使用React.lazy实现组件懒加载
2. **资源优化**: 压缩图片和静态资源
3. **缓存策略**: 合理设置浏览器缓存
4. **监控指标**: 使用性能监控工具

## 监控和日志

### 应用监控
```bash
# 查看应用指标
curl http://localhost:8080/metrics

# 查看健康状态
curl http://localhost:8080/health
```

### 日志查看
```bash
# 查看应用日志
docker-compose logs -f backend

# Kubernetes环境
kubectl logs -f deployment/backend -n container-platform
```

## 安全配置

### 生产环境安全检查清单
- [ ] 更改默认密码和密钥
- [ ] 配置HTTPS证书
- [ ] 设置防火墙规则
- [ ] 启用API访问限制
- [ ] 配置备份策略
- [ ] 设置监控告警
- [ ] 定期安全扫描

## 支持和帮助

- **文档**: [项目Wiki](https://github.com/your-org/container-orchestration-platform/wiki)
- **问题反馈**: [GitHub Issues](https://github.com/your-org/container-orchestration-platform/issues)
- **社区讨论**: [GitHub Discussions](https://github.com/your-org/container-orchestration-platform/discussions)

## 下一步

1. **阅读架构文档**: 了解系统整体架构设计
2. **查看API文档**: 学习API接口使用方法
3. **尝试示例**: 运行提供的示例应用
4. **参与开发**: 贡献代码或报告问题