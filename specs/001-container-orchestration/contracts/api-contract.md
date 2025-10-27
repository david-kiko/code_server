# API契约文档

**创建日期**: 2025-10-25
**版本**: 1.0.0
**API基础路径**: `/api/v1`

## 概述

本文档定义了容器编排管理平台的RESTful API接口。API遵循REST设计原则，使用JSON格式进行数据交换，支持标准的HTTP状态码和错误处理。

## 通用规范

### 请求格式
- Content-Type: `application/json`
- Accept: `application/json`
- 认证: Bearer Token (JWT)

### 响应格式
```json
{
  "success": true,
  "data": {},
  "message": "操作成功",
  "timestamp": "2025-10-25T10:30:00Z"
}
```

### 错误响应格式
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "name",
        "message": "容器名称不能为空"
      }
    ]
  },
  "timestamp": "2025-10-25T10:30:00Z"
}
```

### 分页响应格式
```json
{
  "success": true,
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 100,
      "totalPages": 5
    }
  }
}
```

## 认证接口

### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "refresh_token_here",
    "expiresIn": 3600,
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "fullName": "系统管理员",
      "role": "admin"
    }
  }
}
```

### 刷新令牌
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refreshToken": "refresh_token_here"
}
```

### 用户登出
```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

## 用户管理接口

### 获取当前用户信息
```http
GET /api/v1/users/me
Authorization: Bearer <token>
```

### 获取用户列表
```http
GET /api/v1/users?page=1&pageSize=20&search=admin
Authorization: Bearer <token>
```

### 创建用户
```http
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "password123",
  "fullName": "新用户",
  "role": "operator"
}
```

### 更新用户
```http
PUT /api/v1/users/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "updated@example.com",
  "fullName": "更新的用户名",
  "role": "operator"
}
```

### 删除用户
```http
DELETE /api/v1/users/{id}
Authorization: Bearer <token>
```

## 命名空间管理接口

### 获取命名空间列表
```http
GET /api/v1/namespaces?page=1&pageSize=20
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "default",
        "displayName": "默认命名空间",
        "description": "系统默认命名空间",
        "clusterName": "cluster-1",
        "isActive": true,
        "createdAt": "2025-10-25T10:00:00Z",
        "createdBy": {
          "id": 1,
          "username": "admin",
          "fullName": "系统管理员"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 1,
      "totalPages": 1
    }
  }
}
```

### 创建命名空间
```http
POST /api/v1/namespaces
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "development",
  "displayName": "开发环境",
  "description": "开发测试环境",
  "clusterName": "cluster-1"
}
```

### 更新命名空间
```http
PUT /api/v1/namespaces/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "displayName": "更新的显示名称",
  "description": "更新的描述"
}
```

### 删除命名空间
```http
DELETE /api/v1/namespaces/{id}
Authorization: Bearer <token>
```

## 容器管理接口

### 获取容器列表
```http
GET /api/v1/containers?namespaceId=1&status=running&page=1&pageSize=20
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "nginx-web",
        "displayName": "Nginx Web服务器",
        "namespace": {
          "id": 1,
          "name": "default"
        },
        "image": {
          "name": "nginx",
          "tag": "1.21",
          "repository": "library/nginx"
        },
        "status": "running",
        "phase": "Running",
        "podName": "nginx-web-abc123",
        "podIp": "10.244.1.10",
        "hostIp": "192.168.1.100",
        "node": "worker-node-1",
        "restartCount": 0,
        "cpuRequest": "100m",
        "cpuLimit": "500m",
        "memoryRequest": "128Mi",
        "memoryLimit": "512Mi",
        "createdAt": "2025-10-25T09:00:00Z",
        "startedAt": "2025-10-25T09:01:00Z",
        "createdBy": {
          "id": 1,
          "username": "admin",
          "fullName": "系统管理员"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 1,
      "totalPages": 1
    }
  }
}
```

### 获取容器详情
```http
GET /api/v1/containers/{id}
Authorization: Bearer <token>
```

### 创建容器
```http
POST /api/v1/containers
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "my-app",
  "displayName": "我的应用",
  "namespaceId": 1,
  "image": {
    "name": "nginx",
    "tag": "1.21"
  },
  "resources": {
    "cpuRequest": "100m",
    "cpuLimit": "500m",
    "memoryRequest": "128Mi",
    "memoryLimit": "512Mi"
  },
  "ports": [
    {
      "name": "http",
      "containerPort": 80,
      "hostPort": 8080,
      "protocol": "TCP"
    }
  ],
  "volumes": [
    {
      "name": "data",
      "mountPath": "/data",
      "volumeName": "my-pvc",
      "readOnly": false
    }
  ],
  "environmentVariables": [
    {
      "name": "ENV",
      "value": "production"
    },
    {
      "name": "DATABASE_URL",
      "valueFrom": {
        "type": "secret",
        "name": "db-secret",
        "key": "url"
      }
    }
  ],
  "command": ["/bin/sh"],
  "args": ["-c", "echo 'Hello World' && sleep 3600"]
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "my-app",
    "status": "pending",
    "operationId": "op-123456",
    "message": "容器创建请求已提交，正在处理中"
  }
}
```

### 启动容器
```http
POST /api/v1/containers/{id}/start
Authorization: Bearer <token>
```

### 停止容器
```http
POST /api/v1/containers/{id}/stop
Authorization: Bearer <token>
```

### 暂停容器
```http
POST /api/v1/containers/{id}/pause
Authorization: Bearer <token>
```

### 重启容器
```http
POST /api/v1/containers/{id}/restart
Authorization: Bearer <token>
```

### 删除容器
```http
DELETE /api/v1/containers/{id}
Authorization: Bearer <token>
```

### 获取容器日志
```http
GET /api/v1/containers/{id}/logs?lines=100&follow=false
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "timestamp": "2025-10-25T10:30:00Z",
        "level": "INFO",
        "message": "Server started on port 80"
      }
    ],
    "hasMore": true
  }
}
```

### 获取容器资源使用情况
```http
GET /api/v1/containers/{id}/metrics?period=1h
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "current": {
      "cpuUsagePercent": 25.5,
      "memoryUsagePercent": 45.2,
      "cpuCoresUsed": 0.255,
      "memoryBytesUsed": 234567890,
      "networkBytesRx": 1024000,
      "networkBytesTx": 512000
    },
    "history": [
      {
        "timestamp": "2025-10-25T10:00:00Z",
        "cpuUsagePercent": 20.1,
        "memoryUsagePercent": 42.8
      }
    ]
  }
}
```

## 存储管理接口

### 获取存储卷列表
```http
GET /api/v1/volumes?namespaceId=1&type=persistent_volume_claim&page=1&pageSize=20
Authorization: Bearer <token>
```

### 创建存储卷
```http
POST /api/v1/volumes
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "my-volume",
  "namespaceId": 1,
  "type": "persistent_volume_claim",
  "size": "10Gi",
  "accessMode": "ReadWriteOnce",
  "storageClass": "standard"
}
```

### 删除存储卷
```http
DELETE /api/v1/volumes/{id}
Authorization: Bearer <token>
```

## 配置管理接口

### 获取ConfigMap列表
```http
GET /api/v1/configmaps?namespaceId=1&page=1&pageSize=20
Authorization: Bearer <token>
```

### 创建ConfigMap
```http
POST /api/v1/configmaps
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "my-config",
  "namespaceId": 1,
  "data": {
    "database.host": "localhost",
    "database.port": "5432",
    "app.properties": "debug=true\nport=8080"
  }
}
```

### 获取Secret列表
```http
GET /api/v1/secrets?namespaceId=1&page=1&pageSize=20
Authorization: Bearer <token>
```

### 创建Secret
```http
POST /api/v1/secrets
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "my-secret",
  "namespaceId": 1,
  "type": "Opaque",
  "data": {
    "username": "YWRtaW4=",  # base64编码
    "password": "cGFzc3dvcmQxMjM="
  }
}
```

## 服务管理接口

### 获取服务列表
```http
GET /api/v1/services?namespaceId=1&page=1&pageSize=20
Authorization: Bearer <token>
```

### 创建服务
```http
POST /api/v1/services
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "my-service",
  "namespaceId": 1,
  "type": "ClusterIP",
  "ports": [
    {
      "port": 80,
      "targetPort": 8080,
      "protocol": "TCP"
    }
  ],
  "selector": {
    "app": "my-app"
  }
}
```

## 操作日志接口

### 获取操作日志
```http
GET /api/v1/operations?userId=1&resourceType=container&action=create&page=1&pageSize=20
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "op-123456",
        "user": {
          "id": 1,
          "username": "admin",
          "fullName": "系统管理员"
        },
        "action": "create",
        "resourceType": "container",
        "resourceId": 1,
        "resourceName": "nginx-web",
        "namespace": {
          "id": 1,
          "name": "default"
        },
        "status": "completed",
        "requestBody": {},
        "durationMs": 2500,
        "startedAt": "2025-10-25T10:30:00Z",
        "completedAt": "2025-10-25T10:30:02Z",
        "clientIp": "192.168.1.100",
        "userAgent": "Mozilla/5.0..."
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 100,
      "totalPages": 5
    }
  }
}
```

## 统计和监控接口

### 获取仪表板统计
```http
GET /api/v1/dashboard/stats?namespaceId=1
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "containers": {
      "total": 50,
      "running": 35,
      "stopped": 10,
      "failed": 5
    },
    "namespaces": {
      "total": 5,
      "active": 4
    },
    "volumes": {
      "total": 20,
      "bound": 15,
      "available": 5
    },
    "services": {
      "total": 10,
      "clusterIP": 8,
      "nodePort": 2
    },
    "resourceUsage": {
      "totalCpuCores": 20.5,
      "usedCpuCores": 12.3,
      "totalMemoryGi": 64,
      "usedMemoryGi": 28.5
    },
    "recentOperations": [
      {
        "id": "op-123456",
        "action": "create",
        "resourceType": "container",
        "resourceName": "new-app",
        "user": "admin",
        "timestamp": "2025-10-25T10:30:00Z",
        "status": "completed"
      }
    ]
  }
}
```

## WebSocket接口

### 实时状态推送
```javascript
// 连接WebSocket
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/containers?token=<jwt_token>');

// 订阅容器状态变化
ws.send(JSON.stringify({
  type: 'subscribe',
  resource: 'containers',
  namespaceId: 1
}));

// 接收状态更新
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Container status update:', data);
};
```

**推送消息格式**:
```json
{
  "type": "container_status_update",
  "data": {
    "id": 1,
    "name": "nginx-web",
    "status": "running",
    "phase": "Running",
    "podIp": "10.244.1.10",
    "timestamp": "2025-10-25T10:30:00Z"
  }
}
```

## 错误代码

| 错误代码 | HTTP状态码 | 描述 |
|---------|-----------|------|
| VALIDATION_ERROR | 400 | 请求参数验证失败 |
| UNAUTHORIZED | 401 | 未授权访问 |
| FORBIDDEN | 403 | 权限不足 |
| NOT_FOUND | 404 | 资源不存在 |
| CONFLICT | 409 | 资源冲突 |
| RATE_LIMIT_EXCEEDED | 429 | 请求频率超限 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |
| SERVICE_UNAVAILABLE | 503 | 服务不可用 |
| KUBERNETES_ERROR | 502 | Kubernetes操作失败 |

## 限流规则

- 普通用户: 100 请求/分钟
- 管理员: 1000 请求/分钟
- WebSocket连接: 每用户最多5个并发连接

## 版本控制

API版本通过URL路径进行控制：
- v1: 当前稳定版本
- v2: 下一个主要版本（向后兼容）

## 测试环境

- 开发环境: `http://localhost:8080/api/v1`
- 测试环境: `https://test-api.example.com/api/v1`
- 生产环境: `https://api.example.com/api/v1`