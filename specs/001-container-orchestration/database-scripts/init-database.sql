-- 容器编排管理平台数据库初始化脚本
-- 创建时间: 2025-10-25
-- 版本: 1.0

-- 创建数据库（如果不存在）
-- CREATE DATABASE container_management
-- WITH
--     ENCODING = 'UTF8'
--     LC_COLLATE = 'en_US.UTF-8'
--     LC_CTYPE = 'en_US.UTF-8'
--     TEMPLATE = template0;

-- 连接到数据库
-- \c container_management;

-- 创建必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- 创建枚举类型
DO $$ BEGIN
    CREATE TYPE user_status AS ENUM ('active', 'inactive', 'locked');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE container_status AS ENUM ('pending', 'running', 'paused', 'stopped', 'failed', 'crash_loop_back_off');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE volume_type AS ENUM ('persistent_volume_claim', 'config_map', 'secret', 'host_path', 'empty_dir');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE access_mode AS ENUM ('ReadWriteOnce', 'ReadOnlyMany', 'ReadWriteMany');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE protocol_type AS ENUM ('TCP', 'UDP');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 创建触发器函数（用于自动更新 updated_at 字段）
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 用户和权限管理表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url VARCHAR(500),
    status user_status DEFAULT 'active',
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id)
);

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_by BIGINT REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, role_id)
);

-- 命名空间和容器管理表
CREATE TABLE IF NOT EXISTS namespaces (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(63) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    kubernetes_name VARCHAR(63) NOT NULL,
    cluster_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    resource_quota JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id)
);

CREATE TRIGGER update_namespaces_updated_at BEFORE UPDATE ON namespaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS container_images (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    tag VARCHAR(128) NOT NULL,
    digest VARCHAR(128),
    repository VARCHAR(255),
    size_bytes BIGINT,
    architecture VARCHAR(20),
    os VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE,
    pulled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'deleted', 'error')),
    UNIQUE(name, tag)
);

CREATE TABLE IF NOT EXISTS containers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(253) NOT NULL,
    display_name VARCHAR(253),
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id),
    image_id BIGINT NOT NULL REFERENCES container_images(id),

    -- Kubernetes 相关信息
    kubernetes_name VARCHAR(253) NOT NULL,
    pod_name VARCHAR(253),
    deployment_name VARCHAR(253),
    replica_set_name VARCHAR(253),

    -- 状态信息
    status container_status DEFAULT 'pending',
    phase VARCHAR(20),
    reason VARCHAR(100),
    message TEXT,

    -- 资源信息
    cpu_request VARCHAR(20),
    cpu_limit VARCHAR(20),
    memory_request VARCHAR(20),
    memory_limit VARCHAR(20),

    -- 网络信息
    restart_count INTEGER DEFAULT 0,
    pod_ip VARCHAR(45),
    host_ip VARCHAR(45),

    -- 时间信息
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 创建者信息
    created_by BIGINT REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),

    -- 约束
    UNIQUE(namespace_id, name)
);

CREATE TRIGGER update_containers_updated_at BEFORE UPDATE ON containers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 存储和网络配置表
CREATE TABLE IF NOT EXISTS volumes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(253) NOT NULL,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id),
    type volume_type NOT NULL,
    storage_class VARCHAR(100),
    size VARCHAR(20),
    access_mode access_mode,
    mount_path VARCHAR(500),
    host_path VARCHAR(500),
    kubernetes_name VARCHAR(253),
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'bound', 'failed', 'deleted')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),

    UNIQUE(namespace_id, name)
);

CREATE TRIGGER update_volumes_updated_at BEFORE UPDATE ON volumes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS container_volumes (
    id BIGSERIAL PRIMARY KEY,
    container_id BIGINT NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    volume_id BIGINT NOT NULL REFERENCES volumes(id) ON DELETE CASCADE,
    mount_path VARCHAR(500) NOT NULL,
    read_only BOOLEAN DEFAULT FALSE,
    sub_path VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(container_id, volume_id, mount_path)
);

CREATE TABLE IF NOT EXISTS port_mappings (
    id BIGSERIAL PRIMARY KEY,
    container_id BIGINT NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    name VARCHAR(100),
    container_port INTEGER NOT NULL,
    host_port INTEGER,
    protocol protocol_type DEFAULT 'TCP',
    service_name VARCHAR(253),
    service_port INTEGER,
    node_port INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 配置和环境变量表
CREATE TABLE IF NOT EXISTS environment_variables (
    id BIGSERIAL PRIMARY KEY,
    container_id BIGINT NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    name VARCHAR(253) NOT NULL,
    value TEXT,
    value_from VARCHAR(50), -- 'literal', 'config_map', 'secret', 'field_ref'
    source_name VARCHAR(253),
    source_key VARCHAR(253),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(container_id, name)
);

CREATE TABLE IF NOT EXISTS config_maps (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(253) NOT NULL,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id),
    kubernetes_name VARCHAR(253) NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    binary_data JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),

    UNIQUE(namespace_id, name)
);

CREATE TRIGGER update_config_maps_updated_at BEFORE UPDATE ON config_maps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS secrets (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(253) NOT NULL,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id),
    kubernetes_name VARCHAR(253) NOT NULL,
    type VARCHAR(100) DEFAULT 'Opaque',
    data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),

    UNIQUE(namespace_id, name)
);

CREATE TRIGGER update_secrets_updated_at BEFORE UPDATE ON secrets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 操作日志和审计表
CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGSERIAL PRIMARY KEY,
    operation_id UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    user_id BIGINT REFERENCES users(id),
    username VARCHAR(50),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id BIGINT,
    resource_name VARCHAR(253),
    namespace_id BIGINT REFERENCES namespaces(id),

    -- 请求信息
    request_method VARCHAR(10),
    request_path VARCHAR(500),
    request_body JSONB,
    request_headers JSONB,
    client_ip VARCHAR(45),
    user_agent TEXT,

    -- 响应信息
    status_code INTEGER,
    response_body JSONB,
    duration_ms INTEGER,

    -- 错误信息
    error_code VARCHAR(50),
    error_message TEXT,

    -- 时间信息
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- 元数据
    metadata JSONB DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS container_status_history (
    id BIGSERIAL PRIMARY KEY,
    container_id BIGINT NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    previous_status container_status,
    current_status container_status NOT NULL,
    reason VARCHAR(100),
    message TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    triggered_by VARCHAR(50), -- 'system', 'user', 'kubernetes'
    user_id BIGINT REFERENCES users(id),
    operation_id UUID REFERENCES operation_logs(operation_id)
);

-- 资源监控表
CREATE TABLE IF NOT EXISTS container_resource_usage (
    id BIGSERIAL PRIMARY KEY,
    container_id BIGINT NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- CPU 使用情况
    cpu_cores_used DECIMAL(10,4),
    cpu_cores_request DECIMAL(10,4),
    cpu_cores_limit DECIMAL(10,4),
    cpu_usage_percent DECIMAL(5,2),

    -- 内存使用情况
    memory_bytes_used BIGINT,
    memory_bytes_request BIGINT,
    memory_bytes_limit BIGINT,
    memory_usage_percent DECIMAL(5,2),

    -- 网络使用情况
    network_bytes_rx BIGINT,
    network_bytes_tx BIGINT,
    network_packets_rx BIGINT,
    network_packets_tx BIGINT,

    -- 磁盘使用情况
    disk_bytes_used BIGINT,
    disk_bytes_total BIGINT,
    disk_usage_percent DECIMAL(5,2),

    -- 文件系统使用情况
    filesystem_reads BIGINT,
    filesystem_writes BIGINT,

    metadata JSONB DEFAULT '{}'
);

-- 数据库版本管理表
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64) NOT NULL,
    execution_time_ms INTEGER,
    success BOOLEAN DEFAULT TRUE
);

-- 创建索引
-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- 角色表索引
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_roles_permissions ON roles USING GIN(permissions);

-- 用户角色关联表索引
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);

-- 命名空间表索引
CREATE INDEX IF NOT EXISTS idx_namespaces_name ON namespaces(name);
CREATE INDEX IF NOT EXISTS idx_namespaces_k8s_name ON namespaces(kubernetes_name);
CREATE INDEX IF NOT EXISTS idx_namespaces_cluster ON namespaces(cluster_name);
CREATE INDEX IF NOT EXISTS idx_namespaces_status ON namespaces(status);

-- 容器镜像表索引
CREATE INDEX IF NOT EXISTS idx_container_images_name_tag ON container_images(name, tag);
CREATE INDEX IF NOT EXISTS idx_container_images_repository ON container_images(repository);
CREATE INDEX IF NOT EXISTS idx_container_images_status ON container_images(status);

-- 容器表索引
CREATE INDEX IF NOT EXISTS idx_containers_namespace_id ON containers(namespace_id);
CREATE INDEX IF NOT EXISTS idx_containers_k8s_name ON containers(kubernetes_name);
CREATE INDEX IF NOT EXISTS idx_containers_status ON containers(status);
CREATE INDEX IF NOT EXISTS idx_containers_pod_name ON containers(pod_name);
CREATE INDEX IF NOT EXISTS idx_containers_created_by ON containers(created_by);
CREATE INDEX IF NOT EXISTS idx_containers_created_at ON containers(created_at);
CREATE INDEX IF NOT EXISTS idx_containers_pod_ip ON containers(pod_ip);

-- 复合索引
CREATE INDEX IF NOT EXISTS idx_containers_namespace_status ON containers(namespace_id, status);

-- 存储卷表索引
CREATE INDEX IF NOT EXISTS idx_volumes_namespace_id ON volumes(namespace_id);
CREATE INDEX IF NOT EXISTS idx_volumes_type ON volumes(type);
CREATE INDEX IF NOT EXISTS idx_volumes_k8s_name ON volumes(kubernetes_name);
CREATE INDEX IF NOT EXISTS idx_volumes_status ON volumes(status);

-- 容器存储卷关联表索引
CREATE INDEX IF NOT EXISTS idx_container_volumes_container_id ON container_volumes(container_id);
CREATE INDEX IF NOT EXISTS idx_container_volumes_volume_id ON container_volumes(volume_id);

-- 端口映射表索引
CREATE INDEX IF NOT EXISTS idx_port_mappings_container_id ON port_mappings(container_id);
CREATE INDEX IF NOT EXISTS idx_port_mappings_host_port ON port_mappings(host_port);
CREATE INDEX IF NOT EXISTS idx_port_mappings_service_name ON port_mappings(service_name);

-- 环境变量表索引
CREATE INDEX IF NOT EXISTS idx_env_vars_container_id ON environment_variables(container_id);
CREATE INDEX IF NOT EXISTS idx_env_vars_name ON environment_variables(name);

-- 配置映射表索引
CREATE INDEX IF NOT EXISTS idx_config_maps_namespace_id ON config_maps(namespace_id);
CREATE INDEX IF NOT EXISTS idx_config_maps_k8s_name ON config_maps(kubernetes_name);
CREATE INDEX IF NOT EXISTS idx_config_maps_data ON config_maps USING GIN(data);

-- 密钥表索引
CREATE INDEX IF NOT EXISTS idx_secrets_namespace_id ON secrets(namespace_id);
CREATE INDEX IF NOT EXISTS idx_secrets_k8s_name ON secrets(kubernetes_name);

-- 操作日志表索引
CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_action ON operation_logs(action);
CREATE INDEX IF NOT EXISTS idx_operation_logs_resource_type ON operation_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_operation_logs_resource_id ON operation_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_namespace_id ON operation_logs(namespace_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_started_at ON operation_logs(started_at);
CREATE INDEX IF NOT EXISTS idx_operation_logs_status_code ON operation_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_operation_logs_operation_id ON operation_logs(operation_id);

-- 容器状态变更记录表索引
CREATE INDEX IF NOT EXISTS idx_container_status_history_container_id ON container_status_history(container_id);
CREATE INDEX IF NOT EXISTS idx_container_status_history_changed_at ON container_status_history(changed_at);
CREATE INDEX IF NOT EXISTS idx_container_status_history_current_status ON container_status_history(current_status);

-- 容器资源使用记录表索引
CREATE INDEX IF NOT EXISTS idx_container_resource_usage_container_id ON container_resource_usage(container_id);
CREATE INDEX IF NOT EXISTS idx_container_resource_usage_timestamp ON container_resource_usage(timestamp);
CREATE INDEX IF NOT EXISTS idx_container_resource_usage_cpu_usage ON container_resource_usage(cpu_usage_percent);
CREATE INDEX IF NOT EXISTS idx_container_resource_usage_memory_usage ON container_resource_usage(memory_usage_percent);

-- 插入初始数据
-- 插入默认角色
INSERT INTO roles (name, display_name, description, is_system, permissions) VALUES
('admin', '系统管理员', '拥有所有权限的系统管理员', true, '["*"]'),
('operator', '运维人员', '可以管理容器和查看监控信息', true, '["container:*", "namespace:read", "volume:*", "config_map:*", "secret:*", "logs:read"]'),
('viewer', '只读用户', '只能查看信息，不能进行操作', true, '["container:read", "namespace:read", "volume:read", "config_map:read", "logs:read"]')
ON CONFLICT (name) DO NOTHING;

-- 插入默认管理员用户（密码需要在应用中设置）
INSERT INTO users (username, email, full_name, status) VALUES
('admin', 'admin@example.com', '系统管理员', 'active')
ON CONFLICT (username) DO NOTHING;

-- 为管理员分配角色
INSERT INTO user_roles (user_id, role_id, assigned_by)
SELECT u.id, r.id, u.id
FROM users u, roles r
WHERE u.username = 'admin' AND r.name = 'admin'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- 插入默认命名空间
INSERT INTO namespaces (name, display_name, kubernetes_name, cluster_name, created_by) VALUES
('default', '默认命名空间', 'default', 'main-cluster', 1),
('development', '开发环境', 'development', 'main-cluster', 1),
('staging', '测试环境', 'staging', 'main-cluster', 1),
('production', '生产环境', 'production', 'main-cluster', 1)
ON CONFLICT (name) DO NOTHING;

-- 创建视图
CREATE OR REPLACE VIEW container_details AS
SELECT
    c.id,
    c.name,
    c.display_name,
    c.status,
    c.phase,
    c.reason,
    c.message,
    c.kubernetes_name,
    c.pod_name,
    c.pod_ip,
    c.host_ip,
    c.restart_count,
    c.cpu_request,
    c.cpu_limit,
    c.memory_request,
    c.memory_limit,
    c.started_at,
    c.finished_at,
    c.created_at,
    c.updated_at,

    -- 关联信息
    n.name as namespace_name,
    n.kubernetes_name as namespace_k8s_name,
    ci.name as image_name,
    ci.tag as image_tag,
    ci.repository as image_repository,
    u.username as created_by_username,
    u.full_name as created_by_full_name,

    -- 计算字段
    CASE
        WHEN c.status = 'running' THEN EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - c.started_at))
        ELSE NULL
    END as uptime_seconds,

    -- 资源使用情况 (最新)
    latest.cpu_usage_percent,
    latest.memory_usage_percent,
    latest.timestamp as last_metric_timestamp

FROM containers c
LEFT JOIN namespaces n ON c.namespace_id = n.id
LEFT JOIN container_images ci ON c.image_id = ci.id
LEFT JOIN users u ON c.created_by = u.id
LEFT JOIN LATERAL (
    SELECT DISTINCT ON (container_id)
        cpu_usage_percent,
        memory_usage_percent,
        timestamp
    FROM container_resource_usage
    WHERE container_id = c.id
    ORDER BY timestamp DESC
) latest ON true;

-- 创建用户操作统计视图
CREATE OR REPLACE VIEW user_operation_stats AS
SELECT
    u.id as user_id,
    u.username,
    u.full_name,

    -- 操作统计
    COUNT(*) as total_operations,
    COUNT(CASE WHEN ol.status_code >= 200 AND ol.status_code < 300 THEN 1 END) as successful_operations,
    COUNT(CASE WHEN ol.status_code >= 400 THEN 1 END) as failed_operations,

    -- 按操作类型统计
    COUNT(CASE WHEN ol.action = 'create' THEN 1 END) as create_operations,
    COUNT(CASE WHEN ol.action = 'start' THEN 1 END) as start_operations,
    COUNT(CASE WHEN ol.action = 'stop' THEN 1 END) as stop_operations,
    COUNT(CASE WHEN ol.action = 'delete' THEN 1 END) as delete_operations,

    -- 时间统计
    MIN(ol.started_at) as first_operation_at,
    MAX(ol.started_at) as last_operation_at,
    AVG(ol.duration_ms) as avg_duration_ms,

    -- 最近30天统计
    COUNT(CASE WHEN ol.started_at >= CURRENT_TIMESTAMP - INTERVAL '30 days' THEN 1 END) as operations_last_30_days

FROM users u
LEFT JOIN operation_logs ol ON u.id = ol.user_id
GROUP BY u.id, u.username, u.full_name;

-- 记录数据库初始化版本
INSERT INTO schema_migrations (version, checksum) VALUES
('001_initial_schema', 'initial_checksum_placeholder')
ON CONFLICT (version) DO NOTHING;

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE '容器管理平台数据库初始化完成！';
    RAISE NOTICE '请执行以下命令：';
    RAISE NOTICE '1. 更新管理员用户密码';
    RAISE NOTICE '2. 配置数据库连接池';
    RAISE NOTICE '3. 设置数据库备份策略';
END $$;