# Docker Registry 推送问题解决方案

## 问题总结

1. ✅ **HTTP Registry 配置** - 已解决
2. ✅ **Docker Desktop 配置** - 已解决  
3. ❌ **Manifest Invalid 错误** - 持续存在

## 当前状态

- Docker 配置正确：`192.168.248.200:5000` 已在 insecure registries 中
- 镜像构建成功：`container-platform-backend:local`
- 推送失败：`error from registry: manifest invalid`

## 可能的原因

1. **Registry 服务器问题**：私有 registry 可能不支持某些镜像格式
2. **镜像格式不兼容**：多架构镜像或特殊格式导致问题
3. **Registry 版本问题**：registry:2 版本可能有限制

## 解决方案

### 方案1：使用 docker-compose 直接构建（推荐）

修改 `docker-compose.yml`，移除预构建的镜像标签：

```yaml
services:
  backend:
    # 移除这行：image: 192.168.248.200:5000/container-platform-backend:1.0.0
    build:
      context: ./backend
      dockerfile: Dockerfile
    # ... 其他配置
```

然后在 Ubuntu 上直接运行：
```bash
docker-compose up --build
```

### 方案2：使用镜像文件传输

1. **在 Windows 上**：
```bash
# 构建镜像
docker build -t container-platform-backend:1.0.0 ./backend
docker build -t container-platform-frontend:1.0.0 ./frontend

# 保存镜像
docker save container-platform-backend:1.0.0 -o backend-image.tar
docker save container-platform-frontend:1.0.0 -o frontend-image.tar
```

2. **传输到 Ubuntu**：
```bash
# 使用 scp 或其他方式传输文件
scp backend-image.tar user@ubuntu-server:/path/to/destination/
scp frontend-image.tar user@ubuntu-server:/path/to/destination/
```

3. **在 Ubuntu 上加载**：
```bash
docker load -i backend-image.tar
docker load -i frontend-image.tar
docker tag container-platform-backend:1.0.0 192.168.248.200:5000/container-platform-backend:1.0.0
docker tag container-platform-frontend:1.0.0 192.168.248.200:5000/container-platform-frontend:1.0.0
```

### 方案3：修复 Registry 配置

检查 registry 服务器配置，确保支持：
- 多架构镜像
- 最新的 Docker 镜像格式
- 正确的存储后端

### 方案4：使用不同的 Registry

考虑使用：
- Docker Hub（需要登录）
- 阿里云容器镜像服务
- 腾讯云容器镜像服务
- Harbor（企业级 registry）

## 推荐操作

**立即执行**：使用方案1，在 Ubuntu 上直接构建和运行：

```bash
# 在 Ubuntu 上
git clone <your-repo>
cd code_server
docker-compose up --build
```

这样可以避免 registry 推送问题，直接在目标环境构建和运行。
