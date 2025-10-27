# 容器编排管理平台数据模型设计

**文档版本**: 1.0
**创建日期**: 2025-10-25
**适用项目**: 容器编排管理平台 (001-container-orchestration)

## 1. 数据库架构概述

### 1.1 技术选型
- **数据库**: PostgreSQL 14+
- **ORM框架**: GORM (Go)
- **字符集**: UTF-8
- **时区**: UTC

### 1.2 设计原则
- **数据一致性**: 使用外键约束确保引用完整性
- **性能优化**: 合理设计索引，支持高并发查询
- **扩展性**: 支持水平分片和读写分离
- **安全性**: 敏感数据加密存储
- **审计能力**: 完整的操作日志记录

## 2. 核心数据表设计

### 2.1 用户和权限管理

#### 2.1.1 用户表 (users)
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'locked')),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id)
);

-- 索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);
```

#### 2.1.2 角色表 (roles)
```sql
CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_permissions ON roles USING GIN(permissions);
```

#### 2.1.3 用户角色关联表 (user_roles)
```sql
CREATE TABLE user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_by BIGINT REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, role_id)
);

-- 索引
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

### 2.2 容器信息管理

#### 2.2.1 命名空间表 (namespaces)
```sql
CREATE TABLE namespaces (
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

-- 索引
CREATE INDEX idx_namespaces_name ON namespaces(name);
CREATE INDEX idx_namespaces_k8s_name ON namespaces(kubernetes_name);
CREATE INDEX idx_namespaces_cluster ON namespaces(cluster_name);
CREATE INDEX idx_namespaces_status ON namespaces(status);
```

#### 2.2.2 容器镜像表 (container_images)
```sql
CREATE TABLE container_images (
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

-- 索引
CREATE INDEX idx_container_images_name_tag ON container_images(name, tag);
CREATE INDEX idx_container_images_repository ON container_images(repository);
CREATE INDEX idx_container_images_status ON container_images(status);
```

#### 2.2.3 容器实例表 (containers)
```sql
CREATE TABLE containers (
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
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'paused', 'stopped', 'failed', 'crash_loop_back_off')),
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

-- 索引
CREATE INDEX idx_containers_namespace_id ON containers(namespace_id);
CREATE INDEX idx_containers_k8s_name ON containers(kubernetes_name);
CREATE INDEX idx_containers_status ON containers(status);
CREATE INDEX idx_containers_pod_name ON containers(pod_name);
CREATE INDEX idx_containers_created_by ON containers(created_by);
CREATE INDEX idx_containers_created_at ON containers(created_at);
CREATE INDEX idx_containers_pod_ip ON containers(pod_ip);
```

### 2.3 存储和网络配置

#### 2.3.1 存储卷表 (volumes)
```sql
CREATE TABLE volumes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(253) NOT NULL,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id),
    type VARCHAR(50) NOT NULL, -- 'persistent_volume_claim', 'config_map', 'secret', 'host_path', 'empty_dir'
    storage_class VARCHAR(100),
    size VARCHAR(20),
    access_mode VARCHAR(20), -- 'ReadWriteOnce', 'ReadOnlyMany', 'ReadWriteMany'
    mount_path VARCHAR(500),
    host_path VARCHAR(500),
    kubernetes_name VARCHAR(253),
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'bound', 'failed', 'deleted')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),

    UNIQUE(namespace_id, name)
);

-- 索引
CREATE INDEX idx_volumes_namespace_id ON volumes(namespace_id);
CREATE INDEX idx_volumes_type ON volumes(type);
CREATE INDEX idx_volumes_k8s_name ON volumes(kubernetes_name);
CREATE INDEX idx_volumes_status ON volumes(status);
```

#### 2.3.2 容器存储卷关联表 (container_volumes)
```sql
CREATE TABLE container_volumes (
    id BIGSERIAL PRIMARY KEY,
    container_id BIGINT NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    volume_id BIGINT NOT NULL REFERENCES volumes(id) ON DELETE CASCADE,
    mount_path VARCHAR(500) NOT NULL,
    read_only BOOLEAN DEFAULT FALSE,
    sub_path VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(container_id, volume_id, mount_path)
);

-- 索引
CREATE INDEX idx_container_volumes_container_id ON container_volumes(container_id);
CREATE INDEX idx_container_volumes_volume_id ON container_volumes(volume_id);
```

#### 2.3.3 端口映射表 (port_mappings)
```sql
CREATE TABLE port_mappings (
    id BIGSERIAL PRIMARY KEY,
    container_id BIGINT NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    name VARCHAR(100),
    container_port INTEGER NOT NULL,
    host_port INTEGER,
    protocol VARCHAR(10) DEFAULT 'TCP' CHECK (protocol IN ('TCP', 'UDP')),
    service_name VARCHAR(253),
    service_port INTEGER,
    node_port INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_port_mappings_container_id ON port_mappings(container_id);
CREATE INDEX idx_port_mappings_host_port ON port_mappings(host_port);
CREATE INDEX idx_port_mappings_service_name ON port_mappings(service_name);
```

### 2.4 环境变量和配置

#### 2.4.1 环境变量表 (environment_variables)
```sql
CREATE TABLE environment_variables (
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

-- 索引
CREATE INDEX idx_env_vars_container_id ON environment_variables(container_id);
CREATE INDEX idx_env_vars_name ON environment_variables(name);
```

#### 2.4.2 配置映射表 (config_maps)
```sql
CREATE TABLE config_maps (
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

-- 索引
CREATE INDEX idx_config_maps_namespace_id ON config_maps(namespace_id);
CREATE INDEX idx_config_maps_k8s_name ON config_maps(kubernetes_name);
CREATE INDEX idx_config_maps_data ON config_maps USING GIN(data);
```

#### 2.4.3 密钥表 (secrets)
```sql
CREATE TABLE secrets (
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

-- 索引
CREATE INDEX idx_secrets_namespace_id ON secrets(namespace_id);
CREATE INDEX idx_secrets_k8s_name ON secrets(kubernetes_name);
```

### 2.5 操作日志和审计

#### 2.5.1 操作日志表 (operation_logs)
```sql
CREATE TABLE operation_logs (
    id BIGSERIAL PRIMARY KEY,
    operation_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    user_id BIGINT REFERENCES users(id),
    username VARCHAR(50),
    action VARCHAR(50) NOT NULL, -- 'create', 'start', 'stop', 'pause', 'restart', 'delete', 'update'
    resource_type VARCHAR(50) NOT NULL, -- 'container', 'namespace', 'volume', 'config_map', 'secret'
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

-- 索引
CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_operation_logs_action ON operation_logs(action);
CREATE INDEX idx_operation_logs_resource_type ON operation_logs(resource_type);
CREATE INDEX idx_operation_logs_resource_id ON operation_logs(resource_id);
CREATE INDEX idx_operation_logs_namespace_id ON operation_logs(namespace_id);
CREATE INDEX idx_operation_logs_started_at ON operation_logs(started_at);
CREATE INDEX idx_operation_logs_status_code ON operation_logs(status_code);
CREATE INDEX idx_operation_logs_operation_id ON operation_logs(operation_id);

-- 分区表 (按月分区，提高查询性能)
CREATE TABLE operation_logs_y2025m10 PARTITION OF operation_logs
FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
```

#### 2.5.2 容器状态变更记录表 (container_status_history)
```sql
CREATE TABLE container_status_history (
    id BIGSERIAL PRIMARY KEY,
    container_id BIGINT NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    previous_status VARCHAR(20),
    current_status VARCHAR(20) NOT NULL,
    reason VARCHAR(100),
    message TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    triggered_by VARCHAR(50), -- 'system', 'user', 'kubernetes'
    user_id BIGINT REFERENCES users(id),
    operation_id UUID REFERENCES operation_logs(operation_id)
);

-- 索引
CREATE INDEX idx_container_status_history_container_id ON container_status_history(container_id);
CREATE INDEX idx_container_status_history_changed_at ON container_status_history(changed_at);
CREATE INDEX idx_container_status_history_current_status ON container_status_history(current_status);
```

### 2.6 资源监控和统计

#### 2.6.1 容器资源使用记录表 (container_resource_usage)
```sql
CREATE TABLE container_resource_usage (
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

-- 索引
CREATE INDEX idx_container_resource_usage_container_id ON container_resource_usage(container_id);
CREATE INDEX idx_container_resource_usage_timestamp ON container_resource_usage(timestamp);
CREATE INDEX idx_container_resource_usage_cpu_usage ON container_resource_usage(cpu_usage_percent);
CREATE INDEX idx_container_resource_usage_memory_usage ON container_resource_usage(memory_usage_percent);

-- 时间序列数据分区 (按日分区)
CREATE TABLE container_resource_usage_y2025m10d25 PARTITION OF container_resource_usage
FOR VALUES FROM ('2025-10-25') TO ('2025-10-26');
```

## 3. 数据库视图设计

### 3.1 容器详细信息视图
```sql
CREATE VIEW container_details AS
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
```

### 3.2 用户操作统计视图
```sql
CREATE VIEW user_operation_stats AS
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
```

## 4. 数据库索引优化策略

### 4.1 主要索引设计原则
1. **外键索引**: 为所有外键字段创建索引
2. **查询索引**: 根据常用查询条件创建复合索引
3. **唯一索引**: 确保业务约束的唯一性
4. **部分索引**: 对特定条件的子集创建索引
5. **GIN索引**: 为JSONB字段创建GIN索引

### 4.2 复合索引设计
```sql
-- 容器状态和命名空间复合查询
CREATE INDEX idx_containers_namespace_status ON containers(namespace_id, status);

-- 用户操作日志复合查询
CREATE INDEX idx_operation_logs_user_action_time ON operation_logs(user_id, action, started_at DESC);

-- 容器资源使用时间序列查询
CREATE INDEX idx_resource_usage_container_time ON container_resource_usage(container_id, timestamp DESC);

-- 存储卷类型和状态复合查询
CREATE INDEX idx_volumes_type_status ON volumes(type, status);

-- 端口映射复合查询
CREATE INDEX idx_port_mappings_container_port ON port_mappings(container_id, container_port, protocol);
```

### 4.3 部分索引设计
```sql
-- 只为运行中的容器创建性能监控索引
CREATE INDEX idx_containers_running ON containers(namespace_id)
WHERE status = 'running';

-- 只为失败的操作日志创建索引
CREATE INDEX idx_operation_logs_failed ON operation_logs(user_id, started_at)
WHERE status_code >= 400;

-- 只为活跃的用户创建会话索引
CREATE INDEX idx_users_active ON users(id, last_login_at)
WHERE status = 'active';
```

## 5. 数据迁移和版本管理

### 5.1 数据库版本控制表
```sql
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64) NOT NULL,
    execution_time_ms INTEGER,
    success BOOLEAN DEFAULT TRUE
);
```

### 5.2 迁移脚本命名规范
```
migrations/
├── 001_create_users_table.up.sql
├── 001_create_users_table.down.sql
├── 002_create_roles_and_permissions.up.sql
├── 002_create_roles_and_permissions.down.sql
├── 003_create_containers_table.up.sql
├── 003_create_containers_table.down.sql
└── ...
```

### 5.3 数据迁移最佳实践

#### 5.3.1 迁移脚本示例
```sql
-- 004_add_container_labels.up.sql
-- 添加容器标签支持
ALTER TABLE containers ADD COLUMN labels JSONB DEFAULT '{}';
CREATE INDEX idx_containers_labels ON containers USING GIN(labels);

-- 更新现有容器的默认标签
UPDATE containers SET labels = '{}' WHERE labels IS NULL;
```

```sql
-- 004_add_container_labels.down.sql
-- 回滚容器标签支持
DROP INDEX IF EXISTS idx_containers_labels;
ALTER TABLE containers DROP COLUMN IF EXISTS labels;
```

#### 5.3.2 数据迁移管理工具配置
```go
// gorm-migrator 配置示例
type Migration struct {
    ID          string `gorm:"primaryKey"`
    AppliedAt   time.Time
    Checksum    string
    ExecTime    int
    Success     bool
}

func RunMigrations(db *gorm.DB) error {
    m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
        {
            ID: "001_create_users_table",
            Migrate: func(tx *gorm.DB) error {
                return tx.AutoMigrate(&User{})
            },
            Rollback: func(tx *gorm.DB) error {
                return tx.Migrator().DropTable("users")
            },
        },
        // 更多迁移...
    })

    return m.Migrate()
}
```

## 6. 性能优化建议

### 6.1 查询优化
```sql
-- 使用物化视图缓存复杂查询结果
CREATE MATERIALIZED VIEW container_summary AS
SELECT
    n.name as namespace_name,
    COUNT(*) as total_containers,
    COUNT(CASE WHEN c.status = 'running' THEN 1 END) as running_containers,
    COUNT(CASE WHEN c.status = 'stopped' THEN 1 END) as stopped_containers,
    COUNT(CASE WHEN c.status = 'failed' THEN 1 END) as failed_containers,
    SUM(CASE WHEN ci.size_bytes IS NOT NULL THEN ci.size_bytes ELSE 0 END) as total_size_bytes
FROM namespaces n
LEFT JOIN containers c ON n.id = c.namespace_id
LEFT JOIN container_images ci ON c.image_id = ci.id
GROUP BY n.id, n.name;

-- 定期刷新物化视图
CREATE OR REPLACE FUNCTION refresh_container_summary()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY container_summary;
END;
$$ LANGUAGE plpgsql;

-- 创建定时任务每小时刷新一次
SELECT cron.schedule('refresh-container-summary', '0 * * * *', 'SELECT refresh_container_summary();');
```

### 6.2 连接池配置
```go
// GORM 数据库连接配置
func NewDBConnection(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
    })

    if err != nil {
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // 连接池配置
    sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
    sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
    sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
    sqlDB.SetConnMaxIdleTime(time.Minute * 30) // 空闲连接最大生存时间

    return db, nil
}
```

### 6.3 分区表维护
```sql
-- 自动创建分区函数
CREATE OR REPLACE FUNCTION create_monthly_partitions(table_name text, start_date date)
RETURNS void AS $$
DECLARE
    partition_name text;
    end_date date;
BEGIN
    FOR i IN 0..11 LOOP
        partition_name := table_name || '_y' || to_char(start_date + interval '1 month' * i, 'YYYY') || 'm' || to_char(start_date + interval '1 month' * i, 'MM');
        end_date := start_date + interval '1 month' * (i + 1);

        EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
                      partition_name, table_name, start_date + interval '1 month' * i, end_date);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 清理旧分区函数
CREATE OR REPLACE FUNCTION drop_old_partitions(table_name text, retention_months integer)
RETURNS void AS $$
DECLARE
    partition_name text;
    cutoff_date date;
BEGIN
    cutoff_date := CURRENT_DATE - interval '1 month' * retention_months;

    FOR partition_name IN
        SELECT tablename
        FROM pg_tables
        WHERE tablename LIKE table_name || '_%'
        AND split_part(tablename, '_', 3) < to_char(cutoff_date, 'YYYYmm')
    LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || partition_name || ' CASCADE';
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

## 7. 数据备份和恢复策略

### 7.1 备份策略
```bash
#!/bin/bash
# 数据库备份脚本

DB_NAME="container_management"
DB_USER="postgres"
BACKUP_DIR="/var/backups/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR

# 全量备份
pg_dump -U $DB_USER -h localhost -d $DB_NAME -f $BACKUP_DIR/full_backup_$DATE.sql

# 压缩备份文件
gzip $BACKUP_DIR/full_backup_$DATE.sql

# 删除7天前的备份
find $BACKUP_DIR -name "full_backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_DIR/full_backup_$DATE.sql.gz"
```

### 7.2 数据恢复脚本
```bash
#!/bin/bash
# 数据库恢复脚本

if [ $# -eq 0 ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

BACKUP_FILE=$1
DB_NAME="container_management"
DB_USER="postgres"

# 检查备份文件是否存在
if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: Backup file $BACKUP_FILE not found"
    exit 1
fi

# 如果是压缩文件，先解压
if [[ $BACKUP_FILE == *.gz ]]; then
    gunzip -c $BACKUP_FILE | psql -U $DB_USER -h localhost -d $DB_NAME
else
    psql -U $DB_USER -h localhost -d $DB_NAME -f $BACKUP_FILE
fi

echo "Database restored from $BACKUP_FILE"
```

## 8. 监控和维护

### 8.1 数据库监控指标
```sql
-- 创建数据库监控视图
CREATE VIEW database_stats AS
SELECT
    schemaname,
    tablename,
    attname as column_name,
    n_distinct,
    correlation
FROM pg_stats
WHERE schemaname = 'public'
ORDER BY tablename, attname;

-- 表大小统计
CREATE VIEW table_sizes AS
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) as index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### 8.2 性能优化建议
1. **定期VACUUM和ANALYZE**: 保持统计信息准确
2. **监控慢查询**: 识别和优化性能瓶颈
3. **索引维护**: 定期重建和优化索引
4. **连接监控**: 监控数据库连接池使用情况
5. **存储空间监控**: 跟踪表和索引增长趋势

## 9. 安全考虑

### 9.1 数据加密
```sql
-- 敏感数据加密存储
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 加密用户敏感信息
CREATE TABLE user_secrets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret_type VARCHAR(50) NOT NULL, -- 'api_key', 'password', 'token'
    encrypted_value BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 加密函数
CREATE OR REPLACE FUNCTION encrypt_secret(value text, key text)
RETURNS bytea AS $$
BEGIN
    RETURN pgp_sym_encrypt(value, key);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 解密函数
CREATE OR REPLACE FUNCTION decrypt_secret(encrypted_value bytea, key text)
RETURNS text AS $$
BEGIN
    RETURN pgp_sym_decrypt(encrypted_value, key);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 9.2 行级安全策略
```sql
-- 启用行级安全
ALTER TABLE containers ENABLE ROW LEVEL SECURITY;

-- 创建安全策略：用户只能访问自己有权限的命名空间下的容器
CREATE POLICY container_access_policy ON containers
USING (
    namespace_id IN (
        SELECT ns.id
        FROM namespaces ns
        JOIN user_roles ur ON ns.id = ur.namespace_id
        WHERE ur.user_id = current_setting('app.current_user_id')::bigint
        AND ur.expires_at > CURRENT_TIMESTAMP
    )
);

-- 创建策略：用户只能查看自己的操作日志
CREATE POLICY operation_log_policy ON operation_logs
USING (
    user_id = current_setting('app.current_user_id')::bigint
    OR
    EXISTS (
        SELECT 1 FROM user_roles ur
        WHERE ur.user_id = current_setting('app.current_user_id')::bigint
        AND JSON_EXTRACT(ur.permissions, '$.admin') = true
    )
);
```

这个数据模型设计提供了容器管理平台所需的所有核心功能，包括用户权限管理、容器生命周期管理、资源配置、监控统计和审计跟踪。设计考虑了性能、安全性和可扩展性，并提供了完整的SQL示例和最佳实践建议。