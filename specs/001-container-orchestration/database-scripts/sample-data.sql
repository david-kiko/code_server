-- 容器管理平台示例数据脚本
-- 创建时间: 2025-10-25
-- 版本: 1.0

-- 插入示例用户
INSERT INTO users (username, email, password_hash, full_name, status, last_login_at) VALUES
('john_doe', 'john.doe@example.com', '$2a$10$rOzJqQhQqQhQqQhQqQhQhOzJqQhQqQhQqQhQqQhQqQhQqQhQqQhQ', 'John Doe', 'active', CURRENT_TIMESTAMP - INTERVAL '2 hours'),
('jane_smith', 'jane.smith@example.com', '$2a$10$rOzJqQhQqQhQqQhQqQhQhOzJqQhQqQhQqQhQqQhQqQhQqQhQqQhQ', 'Jane Smith', 'active', CURRENT_TIMESTAMP - INTERVAL '1 day'),
('bob_wilson', 'bob.wilson@example.com', '$2a$10$rOzJqQhQqQhQqQhQqQhQhOzJqQhQqQhQqQhQqQhQqQhQqQhQqQhQ', 'Bob Wilson', 'active', CURRENT_TIMESTAMP - INTERVAL '3 days'),
('alice_brown', 'alice.brown@example.com', '$2a$10$rOzJqQhQqQhQqQhQqQhQhOzJqQhQqQhQqQhQqQhQqQhQqQhQqQhQ', 'Alice Brown', 'inactive', CURRENT_TIMESTAMP - INTERVAL '1 week')
ON CONFLICT (username) DO NOTHING;

-- 分配用户角色
INSERT INTO user_roles (user_id, role_id, assigned_by)
SELECT u.id, r.id, 1
FROM users u, roles r
WHERE u.username IN ('john_doe', 'jane_smith')
AND r.name = 'operator'
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO user_roles (user_id, role_id, assigned_by)
SELECT u.id, r.id, 1
FROM users u, roles r
WHERE u.username = 'bob_wilson'
AND r.name = 'viewer'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- 插入示例容器镜像
INSERT INTO container_images (name, tag, digest, repository, size_bytes, architecture, os, created_at, status) VALUES
('nginx', '1.21.6', 'sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890', 'library/nginx', 142000000, 'amd64', 'linux', CURRENT_TIMESTAMP - INTERVAL '1 week', 'available'),
('redis', '7.0.5', 'sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 'library/redis', 112000000, 'amd64', 'linux', CURRENT_TIMESTAMP - INTERVAL '2 weeks', 'available'),
('postgres', '15.2', 'sha256:bcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890a', 'library/postgres', 378000000, 'amd64', 'linux', CURRENT_TIMESTAMP - INTERVAL '3 days', 'available'),
('mysql', '8.0.32', 'sha256:cdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab', 'library/mysql', 456000000, 'amd64', 'linux', CURRENT_TIMESTAMP - INTERVAL '5 days', 'available'),
('python', '3.11-slim', 'sha256:def1234567890abcdef1234567890abcdef1234567890abcdef1234567890abc', 'library/python', 125000000, 'amd64', 'linux', CURRENT_TIMESTAMP - INTERVAL '1 month', 'available'),
('node', '18-alpine', 'sha256:ef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcd', 'library/node', 170000000, 'amd64', 'linux', CURRENT_TIMESTAMP - INTERVAL '2 weeks', 'available')
ON CONFLICT (name, tag) DO NOTHING;

-- 插入示例容器
INSERT INTO containers (
    name, display_name, namespace_id, image_id, kubernetes_name, pod_name,
    status, phase, cpu_request, cpu_limit, memory_request, memory_limit,
    restart_count, pod_ip, host_ip, started_at, created_at, created_by
) VALUES
-- 运行中的 Nginx 容器
('web-server-1', 'Web Server 1', 2, 1, 'web-server-1', 'web-server-1-pod-abc123',
 'running', 'Running', '100m', '500m', '128Mi', '512Mi',
 0, '10.244.1.10', '192.168.1.100', CURRENT_TIMESTAMP - INTERVAL '2 hours', CURRENT_TIMESTAMP - INTERVAL '2 hours', 2),

-- 运行中的 Redis 容器
('redis-cache-1', 'Redis Cache 1', 2, 2, 'redis-cache-1', 'redis-cache-1-pod-def456',
 'running', 'Running', '50m', '200m', '64Mi', '256Mi',
 0, '10.244.1.11', '192.168.1.100', CURRENT_TIMESTAMP - INTERVAL '3 hours', CURRENT_TIMESTAMP - INTERVAL '3 hours', 2),

-- 停止的 PostgreSQL 容器
('postgres-db-1', 'PostgreSQL Database 1', 3, 3, 'postgres-db-1', 'postgres-db-1-pod-ghi789',
 'stopped', 'Succeeded', '200m', '1000m', '256Mi', '1Gi',
 1, '10.244.2.15', '192.168.1.101', CURRENT_TIMESTAMP - INTERVAL '1 day', CURRENT_TIMESTAMP - INTERVAL '1 day', 2),

-- 失败的 MySQL 容器
('mysql-db-1', 'MySQL Database 1', 3, 4, 'mysql-db-1', 'mysql-db-1-pod-jkl012',
 'failed', 'Failed', '200m', '1000m', '256Mi', '1Gi',
 3, NULL, '192.168.1.101', NULL, CURRENT_TIMESTAMP - INTERVAL '6 hours', 2),

-- 运行中的 Python 应用容器
('python-app-1', 'Python Application 1', 2, 5, 'python-app-1', 'python-app-1-pod-mno345',
 'running', 'Running', '100m', '500m', '256Mi', '512Mi',
 2, '10.244.1.12', '192.168.1.100', CURRENT_TIMESTAMP - INTERVAL '4 hours', CURRENT_TIMESTAMP - INTERVAL '4 hours', 3),

-- 暂停的 Node.js 应用容器
('node-app-1', 'Node.js Application 1', 2, 6, 'node-app-1', 'node-app-1-pod-pqr678',
 'paused', 'Paused', '100m', '500m', '128Mi', '512Mi',
 0, '10.244.1.13', '192.168.1.100', CURRENT_TIMESTAMP - INTERVAL '5 hours', CURRENT_TIMESTAMP - INTERVAL '5 hours', 3)
ON CONFLICT (namespace_id, name) DO NOTHING;

-- 插入示例存储卷
INSERT INTO volumes (
    name, namespace_id, type, storage_class, size, access_mode,
    mount_path, kubernetes_name, status, created_at, created_by
) VALUES
('web-data', 2, 'persistent_volume_claim', 'fast-ssd', '10Gi', 'ReadWriteOnce',
 '/usr/share/nginx/html', 'web-data', 'bound', CURRENT_TIMESTAMP - INTERVAL '2 hours', 2),

('redis-data', 2, 'persistent_volume_claim', 'standard', '5Gi', 'ReadWriteOnce',
 '/data', 'redis-data', 'bound', CURRENT_TIMESTAMP - INTERVAL '3 hours', 2),

('postgres-data', 3, 'persistent_volume_claim', 'fast-ssd', '20Gi', 'ReadWriteOnce',
 '/var/lib/postgresql/data', 'postgres-data', 'bound', CURRENT_TIMESTAMP - INTERVAL '1 day', 2),

('mysql-data', 3, 'persistent_volume_claim', 'fast-ssd', '20Gi', 'ReadWriteOnce',
 '/var/lib/mysql', 'mysql-data', 'failed', CURRENT_TIMESTAMP - INTERVAL '6 hours', 2),

('app-config', 2, 'config_map', NULL, NULL, 'ReadOnlyMany',
 '/app/config', 'app-config', 'available', CURRENT_TIMESTAMP - INTERVAL '4 hours', 3),

('app-secrets', 2, 'secret', NULL, NULL, 'ReadOnlyMany',
 '/app/secrets', 'app-secrets', 'available', CURRENT_TIMESTAMP - INTERVAL '4 hours', 3),

('temp-storage', 2, 'empty_dir', NULL, '1Gi', 'ReadWriteOnce',
 '/tmp', 'temp-storage', 'available', CURRENT_TIMESTAMP - INTERVAL '5 hours', 3),

('logs-volume', 2, 'host_path', NULL, NULL, 'ReadWriteOnce',
 '/var/log/app', 'logs-volume', 'available', CURRENT_TIMESTAMP - INTERVAL '5 hours', 3)
ON CONFLICT (namespace_id, name) DO NOTHING;

-- 插入容器存储卷关联
INSERT INTO container_volumes (container_id, volume_id, mount_path, read_only, created_at) VALUES
(1, 1, '/usr/share/nginx/html', false, CURRENT_TIMESTAMP - INTERVAL '2 hours'),
(2, 2, '/data', false, CURRENT_TIMESTAMP - INTERVAL '3 hours'),
(3, 3, '/var/lib/postgresql/data', false, CURRENT_TIMESTAMP - INTERVAL '1 day'),
(4, 4, '/var/lib/mysql', false, CURRENT_TIMESTAMP - INTERVAL '6 hours'),
(5, 5, '/app/config', true, CURRENT_TIMESTAMP - INTERVAL '4 hours'),
(5, 6, '/app/secrets', true, CURRENT_TIMESTAMP - INTERVAL '4 hours'),
(5, 7, '/tmp', false, CURRENT_TIMESTAMP - INTERVAL '4 hours'),
(6, 8, '/var/log/app', false, CURRENT_TIMESTAMP - INTERVAL '5 hours')
ON CONFLICT (container_id, volume_id, mount_path) DO NOTHING;

-- 插入端口映射
INSERT INTO port_mappings (
    container_id, name, container_port, host_port, protocol,
    service_name, service_port, node_port, created_at
) VALUES
(1, 'http', 80, 8080, 'TCP', 'web-server-service', 80, 30080, CURRENT_TIMESTAMP - INTERVAL '2 hours'),
(1, 'https', 443, 8443, 'TCP', 'web-server-service', 443, 30443, CURRENT_TIMESTAMP - INTERVAL '2 hours'),
(2, 'redis', 6379, 6379, 'TCP', 'redis-service', 6379, NULL, CURRENT_TIMESTAMP - INTERVAL '3 hours'),
(3, 'postgres', 5432, 5432, 'TCP', 'postgres-service', 5432, NULL, CURRENT_TIMESTAMP - INTERVAL '1 day'),
(4, 'mysql', 3306, 3306, 'TCP', 'mysql-service', 3306, NULL, CURRENT_TIMESTAMP - INTERVAL '6 hours'),
(5, 'http', 8000, 8000, 'TCP', 'python-app-service', 8000, 30800, CURRENT_TIMESTAMP - INTERVAL '4 hours'),
(6, 'http', 3000, 3000, 'TCP', 'node-app-service', 3000, 30300, CURRENT_TIMESTAMP - INTERVAL '5 hours')
ON CONFLICT DO NOTHING;

-- 插入环境变量
INSERT INTO environment_variables (
    container_id, name, value, value_from, source_name, source_key, created_at
) VALUES
(1, 'NGINX_VERSION', '1.21.6', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '2 hours'),
(1, 'SERVER_NAME', 'example.com', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '2 hours'),

(2, 'REDIS_PASSWORD', NULL, 'secret', 'app-secrets', 'redis-password', CURRENT_TIMESTAMP - INTERVAL '3 hours'),

(3, 'POSTGRES_DB', 'myapp', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '1 day'),
(3, 'POSTGRES_USER', 'admin', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '1 day'),
(3, 'POSTGRES_PASSWORD', NULL, 'secret', 'app-secrets', 'postgres-password', CURRENT_TIMESTAMP - INTERVAL '1 day'),

(4, 'MYSQL_ROOT_PASSWORD', NULL, 'secret', 'app-secrets', 'mysql-root-password', CURRENT_TIMESTAMP - INTERVAL '6 hours'),
(4, 'MYSQL_DATABASE', 'myapp', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '6 hours'),
(4, 'MYSQL_USER', 'appuser', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '6 hours'),

(5, 'PYTHON_VERSION', '3.11', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '4 hours'),
(5, 'APP_ENV', 'development', 'config_map', 'app-config', 'environment', CURRENT_TIMESTAMP - INTERVAL '4 hours'),
(5, 'DEBUG_MODE', 'true', 'config_map', 'app-config', 'debug', CURRENT_TIMESTAMP - INTERVAL '4 hours'),

(6, 'NODE_ENV', 'development', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '5 hours'),
(6, 'PORT', '3000', 'literal', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '5 hours'),
(6, 'API_URL', 'http://api.example.com', 'config_map', 'app-config', 'api-url', CURRENT_TIMESTAMP - INTERVAL '5 hours')
ON CONFLICT (container_id, name) DO NOTHING;

-- 插入配置映射
INSERT INTO config_maps (
    name, namespace_id, kubernetes_name, data, created_at, created_by, updated_by
) VALUES
('app-config', 2, 'app-config',
 '{"environment": "development", "debug": "true", "api-url": "http://api.example.com", "database-url": "postgresql://localhost:5432/myapp"}',
 CURRENT_TIMESTAMP - INTERVAL '4 hours', 3, 3),

('nginx-config', 2, 'nginx-config',
 '{"nginx.conf": "user nginx;\\nworker_processes auto;\\nerror_log /var/log/nginx/error.log;\\npid /run/nginx.pid;"}',
 CURRENT_TIMESTAMP - INTERVAL '2 hours', 2, 2),

('redis-config', 2, 'redis-config',
 '{"redis.conf": "maxmemory 256mb\\nmaxmemory-policy allkeys-lru\\nsave 900 1\\nsave 300 10"}',
 CURRENT_TIMESTAMP - INTERVAL '3 hours', 2, 2)
ON CONFLICT (namespace_id, name) DO NOTHING;

-- 插入密钥
INSERT INTO secrets (
    name, namespace_id, kubernetes_name, type, data, created_at, created_by, updated_by
) VALUES
('app-secrets', 2, 'app-secrets', 'Opaque',
 '{"redis-password": "cmVkaXNfcGFzc3dvcmQxMjM=", "postgres-password": "cG9zdGdyZXNfcGFzc3dvcmQ0NTY=", "mysql-root-password": "bXlzcWxfcm9vdF9wYXNzd29yZDc4OQ=="}',
 CURRENT_TIMESTAMP - INTERVAL '4 hours', 3, 3),

('registry-secret', 2, 'registry-secret', 'kubernetes.io/dockerconfigjson',
 '{"\\.dockerconfigjson": "eyJhdXRocyI6IHsicmVnaXN0cnkuZXhhbXBsZS5jb20iOiB7InVzZXJuYW1lIjogInVzZXIiLCAicGFzc3dvcmQiOiAicGFzc3dvcmQifX19"}',
 CURRENT_TIMESTAMP - INTERVAL '1 day', 2, 2),

('ssl-certs', 4, 'ssl-certs', 'kubernetes.io/tls',
 '{"tls.crt": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t...", "tls.key": "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0t..."}',
 CURRENT_TIMESTAMP - INTERVAL '3 days', 1, 1)
ON CONFLICT (namespace_id, name) DO NOTHING;

-- 插入示例操作日志
INSERT INTO operation_logs (
    operation_id, user_id, username, action, resource_type, resource_id, resource_name,
    namespace_id, request_method, request_path, request_body, request_headers,
    client_ip, user_agent, status_code, response_body, duration_ms, started_at, completed_at
) VALUES
-- 创建容器操作
(uuid_generate_v4(), 2, 'john_doe', 'create', 'container', 1, 'web-server-1',
 2, 'POST', '/api/v1/containers',
 '{"name": "web-server-1", "image": "nginx:1.21.6", "namespace": "development"}',
 '{"Content-Type": "application/json", "Authorization": "Bearer token123"}',
 '192.168.1.10', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
 201, '{"id": 1, "status": "success"}', 1250,
 CURRENT_TIMESTAMP - INTERVAL '2 hours', CURRENT_TIMESTAMP - INTERVAL '2 hours' + INTERVAL '1.25 seconds'),

-- 启动容器操作
(uuid_generate_v4(), 2, 'john_doe', 'start', 'container', 1, 'web-server-1',
 2, 'POST', '/api/v1/containers/1/start',
 '{}',
 '{"Content-Type": "application/json", "Authorization": "Bearer token123"}',
 '192.168.1.10', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
 200, '{"status": "running"}', 800,
 CURRENT_TIMESTAMP - INTERVAL '2 hours' + INTERVAL '30 seconds', CURRENT_TIMESTAMP - INTERVAL '2 hours' + INTERVAL '30.8 seconds'),

-- 创建另一个容器
(uuid_generate_v4(), 3, 'jane_smith', 'create', 'container', 2, 'redis-cache-1',
 2, 'POST', '/api/v1/containers',
 '{"name": "redis-cache-1", "image": "redis:7.0.5", "namespace": "development"}',
 '{"Content-Type": "application/json", "Authorization": "Bearer token456"}',
 '192.168.1.20', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
 201, '{"id": 2, "status": "success"}', 1100,
 CURRENT_TIMESTAMP - INTERVAL '3 hours', CURRENT_TIMESTAMP - INTERVAL '3 hours' + INTERVAL '1.1 seconds'),

-- 停止容器操作
(uuid_generate_v4(), 2, 'john_doe', 'stop', 'container', 3, 'postgres-db-1',
 3, 'POST', '/api/v1/containers/3/stop',
 '{}',
 '{"Content-Type": "application/json", "Authorization": "Bearer token123"}',
 '192.168.1.10', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
 200, '{"status": "stopped"}', 600,
 CURRENT_TIMESTAMP - INTERVAL '1 day', CURRENT_TIMESTAMP - INTERVAL '1 day' + INTERVAL '0.6 seconds'),

-- 删除容器操作
(uuid_generate_v4(), 3, 'jane_smith', 'delete', 'container', 4, 'mysql-db-1',
 3, 'DELETE', '/api/v1/containers/4',
 '{}',
 '{"Content-Type": "application/json", "Authorization": "Bearer token456"}',
 '192.168.1.20', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
 200, '{"status": "deleted"}', 900,
 CURRENT_TIMESTAMP - INTERVAL '6 hours', CURRENT_TIMESTAMP - INTERVAL '6 hours' + INTERVAL '0.9 seconds'),

-- 更新容器配置操作
(uuid_generate_v4(), 4, 'bob_wilson', 'update', 'container', 5, 'python-app-1',
 2, 'PUT', '/api/v1/containers/5',
 '{"cpu_limit": "600m", "memory_limit": "768Mi"}',
 '{"Content-Type": "application/json", "Authorization": "Bearer token789"}',
 '192.168.1.30', 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36',
 200, '{"status": "updated"}', 1500,
 CURRENT_TIMESTAMP - INTERVAL '4 hours', CURRENT_TIMESTAMP - INTERVAL '4 hours' + INTERVAL '1.5 seconds'),

-- 暂停容器操作
(uuid_generate_v4(), 4, 'bob_wilson', 'pause', 'container', 6, 'node-app-1',
 2, 'POST', '/api/v1/containers/6/pause',
 '{}',
 '{"Content-Type": "application/json", "Authorization": "Bearer token789"}',
 '192.168.1.30', 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36',
 200, '{"status": "paused"}', 450,
 CURRENT_TIMESTAMP - INTERVAL '5 hours', CURRENT_TIMESTAMP - INTERVAL '5 hours' + INTERVAL '0.45 seconds'),

-- 错误操作示例
(uuid_generate_v4(), 4, 'bob_wilson', 'create', 'container', NULL, 'invalid-container',
 2, 'POST', '/api/v1/containers',
 '{"name": "invalid-container", "image": "nonexistent:latest"}',
 '{"Content-Type": "application/json", "Authorization": "Bearer token789"}',
 '192.168.1.30', 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36',
 400, '{"error": "Image not found", "code": "IMAGE_NOT_FOUND"}', 300,
 CURRENT_TIMESTAMP - INTERVAL '30 minutes', CURRENT_TIMESTAMP - INTERVAL '30 minutes' + INTERVAL '0.3 seconds');

-- 插入容器状态变更历史
INSERT INTO container_status_history (
    container_id, previous_status, current_status, reason, message,
    changed_at, triggered_by, user_id, operation_id
) VALUES
(1, NULL, 'pending', 'ContainerCreating', 'Container is being created',
 CURRENT_TIMESTAMP - INTERVAL '2 hours', 'user', 2, (SELECT operation_id FROM operation_logs WHERE resource_name = 'web-server-1' AND action = 'create' LIMIT 1)),

(1, 'pending', 'running', 'Started', 'Container started successfully',
 CURRENT_TIMESTAMP - INTERVAL '2 hours' + INTERVAL '30 seconds', 'user', 2, (SELECT operation_id FROM operation_logs WHERE resource_name = 'web-server-1' AND action = 'start' LIMIT 1)),

(2, NULL, 'pending', 'ContainerCreating', 'Container is being created',
 CURRENT_TIMESTAMP - INTERVAL '3 hours', 'user', 3, (SELECT operation_id FROM operation_logs WHERE resource_name = 'redis-cache-1' AND action = 'create' LIMIT 1)),

(2, 'pending', 'running', 'Started', 'Container started successfully',
 CURRENT_TIMESTAMP - INTERVAL '3 hours' + INTERVAL '45 seconds', 'system', NULL, NULL),

(3, NULL, 'pending', 'ContainerCreating', 'Container is being created',
 CURRENT_TIMESTAMP - INTERVAL '1 day', 'user', 2, (SELECT operation_id FROM operation_logs WHERE resource_name = 'postgres-db-1' AND action = 'create' LIMIT 1)),

(3, 'pending', 'running', 'Started', 'Container started successfully',
 CURRENT_TIMESTAMP - INTERVAL '1 day' + INTERVAL '1 minute', 'system', NULL, NULL),

(3, 'running', 'stopped', 'UserInitiated', 'Container stopped by user',
 CURRENT_TIMESTAMP - INTERVAL '1 day', 'user', 2, (SELECT operation_id FROM operation_logs WHERE resource_name = 'postgres-db-1' AND action = 'stop' LIMIT 1)),

(4, NULL, 'pending', 'ContainerCreating', 'Container is being created',
 CURRENT_TIMESTAMP - INTERVAL '6 hours', 'user', 3, (SELECT operation_id FROM operation_logs WHERE resource_name = 'mysql-db-1' AND action = 'create' LIMIT 1)),

(4, 'pending', 'failed', 'ImagePullBackOff', 'Failed to pull image "nonexistent:latest"',
 CURRENT_TIMESTAMP - INTERVAL '6 hours' + INTERVAL '2 minutes', 'system', NULL, NULL),

(5, NULL, 'pending', 'ContainerCreating', 'Container is being created',
 CURRENT_TIMESTAMP - INTERVAL '4 hours', 'user', 4, (SELECT operation_id FROM operation_logs WHERE resource_name = 'python-app-1' AND action = 'create' LIMIT 1)),

(5, 'pending', 'running', 'Started', 'Container started successfully',
 CURRENT_TIMESTAMP - INTERVAL '4 hours' + INTERVAL '1 minute', 'system', NULL, NULL),

(6, NULL, 'pending', 'ContainerCreating', 'Container is being created',
 CURRENT_TIMESTAMP - INTERVAL '5 hours', 'user', 4, (SELECT operation_id FROM operation_logs WHERE resource_name = 'node-app-1' AND action = 'create' LIMIT 1)),

(6, 'pending', 'running', 'Started', 'Container started successfully',
 CURRENT_TIMESTAMP - INTERVAL '5 hours' + INTERVAL '1 minute', 'system', NULL, NULL),

(6, 'running', 'paused', 'UserInitiated', 'Container paused by user',
 CURRENT_TIMESTAMP - INTERVAL '5 hours' + INTERVAL '30 minutes', 'user', 4, (SELECT operation_id FROM operation_logs WHERE resource_name = 'node-app-1' AND action = 'pause' LIMIT 1));

-- 插入容器资源使用数据（模拟最近24小时的数据）
INSERT INTO container_resource_usage (
    container_id, timestamp, cpu_cores_used, cpu_cores_request, cpu_cores_limit, cpu_usage_percent,
    memory_bytes_used, memory_bytes_request, memory_bytes_limit, memory_usage_percent,
    network_bytes_rx, network_bytes_tx, network_packets_rx, network_packets_tx,
    disk_bytes_used, disk_bytes_total, disk_usage_percent
)
SELECT
    container_id,
    CURRENT_TIMESTAMP - (n * INTERVAL '1 hour'),
    -- CPU 数据（模拟波动）
    ROUND(0.05 + (RANDOM() * 0.15)::numeric, 4) as cpu_cores_used,
    0.1 as cpu_cores_request,
    0.5 as cpu_cores_limit,
    ROUND(((0.05 + (RANDOM() * 0.15)) / 0.5 * 100)::numeric, 2) as cpu_usage_percent,

    -- 内存数据（模拟波动）
    (100000000 + (RANDOM() * 100000000)::bigint) as memory_bytes_used,
    134217728 as memory_bytes_request, -- 128Mi
    536870912 as memory_bytes_limit, -- 512Mi
    ROUND(((100000000 + (RANDOM() * 100000000)) / 536870912 * 100)::numeric, 2) as memory_usage_percent,

    -- 网络数据（累计增长）
    (n * 1000000)::bigint + (RANDOM() * 1000000)::bigint as network_bytes_rx,
    (n * 2000000)::bigint + (RANDOM() * 1000000)::bigint as network_bytes_tx,
    (n * 1000)::integer + (RANDOM() * 1000)::integer as network_packets_rx,
    (n * 800)::integer + (RANDOM() * 800)::integer as network_packets_tx,

    -- 磁盘数据
    500000000::bigint + (n * 10000000)::bigint as disk_bytes_used,
    2147483648::bigint as disk_bytes_total, -- 2Gi
    ROUND(((500000000 + (n * 10000000)) / 2147483648 * 100)::numeric, 2) as disk_usage_percent,

    '{}'::jsonb as metadata
FROM generate_series(0, 23) n
CROSS JOIN (SELECT id as container_id FROM containers WHERE status = 'running') running_containers;

-- 为其他状态的容器插入少量资源数据
INSERT INTO container_resource_usage (
    container_id, timestamp, cpu_cores_used, cpu_cores_request, cpu_cores_limit, cpu_usage_percent,
    memory_bytes_used, memory_bytes_request, memory_bytes_limit, memory_usage_percent
)
SELECT
    id as container_id,
    CURRENT_TIMESTAMP - INTERVAL '30 minutes',
    0 as cpu_cores_used,
    CASE WHEN name = 'postgres-db-1' THEN 0.2 ELSE 0.1 END as cpu_cores_request,
    CASE WHEN name = 'postgres-db-1' THEN 1.0 ELSE 0.5 END as cpu_cores_limit,
    0 as cpu_usage_percent,
    0 as memory_bytes_used,
    CASE WHEN name = 'postgres-db-1' THEN 268435456 ELSE 134217728 END as memory_bytes_request,
    CASE WHEN name = 'postgres-db-1' THEN 1073741824 ELSE 536870912 END as memory_bytes_limit,
    0 as memory_usage_percent
FROM containers
WHERE status IN ('stopped', 'failed', 'paused');

-- 输出数据插入完成信息
DO $$
BEGIN
    RAISE NOTICE '示例数据插入完成！';
    RAISE NOTICE '已插入的数据包括：';
    RAISE NOTICE '- 4个示例用户';
    RAISE NOTICE '- 6个容器镜像';
    RAISE NOTICE '- 6个容器实例';
    RAISE NOTICE '- 8个存储卷';
    RAISE NOTICE '- 7个端口映射';
    RAISE NOTICE '- 17个环境变量';
    RAISE NOTICE '- 3个配置映射';
    RAISE NOTICE '- 3个密钥';
    RAISE NOTICE '- 9条操作日志';
    RAISE NOTICE '- 13条状态变更记录';
    RAISE NOTICE '- 144条资源使用记录';
END $$;