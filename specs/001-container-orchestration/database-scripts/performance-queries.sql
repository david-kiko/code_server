-- 容器管理平台性能优化查询脚本
-- 创建时间: 2025-10-25
-- 版本: 1.0

-- 1. 容器状态统计查询
-- 用途：获取所有命名空间的容器状态分布
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    n.name as namespace_name,
    COUNT(*) as total_containers,
    COUNT(CASE WHEN c.status = 'running' THEN 1 END) as running_containers,
    COUNT(CASE WHEN c.status = 'stopped' THEN 1 END) as stopped_containers,
    COUNT(CASE WHEN c.status = 'failed' THEN 1 END) as failed_containers,
    COUNT(CASE WHEN c.status = 'paused' THEN 1 END) as paused_containers,
    COUNT(CASE WHEN c.status = 'pending' THEN 1 END) as pending_containers,
    ROUND(AVG(c.restart_count), 2) as avg_restart_count
FROM namespaces n
LEFT JOIN containers c ON n.id = c.namespace_id
GROUP BY n.id, n.name
ORDER BY n.name;

-- 2. 用户操作统计查询（优化版本）
-- 用途：统计用户操作次数和成功率
EXPLAIN (ANALYZE, BUFFERS)
WITH user_stats AS (
    SELECT
        u.id as user_id,
        u.username,
        u.full_name,
        COUNT(*) as total_operations,
        COUNT(CASE WHEN ol.status_code >= 200 AND ol.status_code < 300 THEN 1 END) as successful_operations,
        COUNT(CASE WHEN ol.status_code >= 400 THEN 1 END) as failed_operations,
        ROUND(AVG(ol.duration_ms), 2) as avg_duration_ms
    FROM users u
    LEFT JOIN operation_logs ol ON u.id = ol.user_id
        AND ol.started_at >= CURRENT_DATE - INTERVAL '30 days'
    GROUP BY u.id, u.username, u.full_name
)
SELECT
    username,
    full_name,
    total_operations,
    successful_operations,
    failed_operations,
    CASE
        WHEN total_operations > 0
        THEN ROUND(successful_operations::numeric / total_operations * 100, 2)
        ELSE 0
    END as success_rate_percent,
    avg_duration_ms
FROM user_stats
WHERE total_operations > 0
ORDER BY total_operations DESC;

-- 3. 容器资源使用统计查询
-- 用途：获取运行中容器的资源使用情况
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    c.id,
    c.name,
    c.kubernetes_name,
    n.name as namespace_name,
    ci.name as image_name,
    c.status,

    -- CPU 信息
    c.cpu_request,
    c.cpu_limit,
    latest.cpu_cores_used,
    latest.cpu_usage_percent,

    -- 内存信息
    c.memory_request,
    c.memory_limit,
    latest.memory_bytes_used,
    latest.memory_usage_percent,

    -- 网络信息
    latest.network_bytes_rx,
    latest.network_bytes_tx,

    -- 运行时间
    CASE
        WHEN c.status = 'running' AND c.started_at IS NOT NULL
        THEN CURRENT_TIMESTAMP - c.started_at
        ELSE NULL
    END as uptime,

    -- 重启次数
    c.restart_count,

    -- 最后指标时间
    latest.timestamp as last_metric_timestamp

FROM containers c
JOIN namespaces n ON c.namespace_id = n.id
JOIN container_images ci ON c.image_id = ci.id
LEFT JOIN LATERAL (
    SELECT DISTINCT ON (container_id)
        cpu_cores_used,
        cpu_usage_percent,
        memory_bytes_used,
        memory_usage_percent,
        network_bytes_rx,
        network_bytes_tx,
        timestamp
    FROM container_resource_usage
    WHERE container_id = c.id
    ORDER BY timestamp DESC
) latest ON true
WHERE c.status IN ('running', 'paused')
ORDER BY latest.cpu_usage_percent DESC NULLS LAST;

-- 4. 存储卷使用情况查询
-- 用途：统计存储卷的使用状态和类型分布
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    v.type,
    v.status,
    COUNT(*) as total_volumes,
    COUNT(DISTINCT v.namespace_id) as namespace_count,
    SUM(CASE WHEN v.status = 'bound' THEN 1 ELSE 0 END) as bound_volumes,
    SUM(CASE WHEN v.status = 'available' THEN 1 ELSE 0 END) as available_volumes,
    SUM(CASE WHEN v.status = 'failed' THEN 1 ELSE 0 END) as failed_volumes,

    -- 计算绑定容器的数量
    COUNT(DISTINCT cv.container_id) as containers_using_volumes,

    -- 按大小统计（如果有的话）
    AVG(CASE WHEN v.size ~ '^[0-9]+$' THEN v.size::INTEGER ELSE NULL END) as avg_size_gb,
    MAX(CASE WHEN v.size ~ '^[0-9]+$' THEN v.size::INTEGER ELSE NULL END) as max_size_gb

FROM volumes v
LEFT JOIN container_volumes cv ON v.id = cv.volume_id
GROUP BY v.type, v.status
ORDER BY v.type, v.status;

-- 5. 操作日志聚合查询（按时间）
-- 用途：分析操作日志的时间分布和性能趋势
EXPLAIN (ANALYZE, BUFFERS)
WITH hourly_stats AS (
    SELECT
        DATE_TRUNC('hour', started_at) as hour_bucket,
        COUNT(*) as total_operations,
        COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) as successful_operations,
        COUNT(CASE WHEN status_code >= 400 THEN 1 END) as failed_operations,
        ROUND(AVG(duration_ms), 2) as avg_duration_ms,
        ROUND(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms), 2) as p95_duration_ms,
        COUNT(DISTINCT user_id) as unique_users,
        COUNT(DISTINCT resource_type) as resource_types
    FROM operation_logs
    WHERE started_at >= CURRENT_DATE - INTERVAL '7 days'
    GROUP BY DATE_TRUNC('hour', started_at)
)
SELECT
    hour_bucket,
    total_operations,
    successful_operations,
    failed_operations,
    CASE
        WHEN total_operations > 0
        THEN ROUND(successful_operations::numeric / total_operations * 100, 2)
        ELSE 0
    END as success_rate_percent,
    avg_duration_ms,
    p95_duration_ms,
    unique_users,
    resource_types
FROM hourly_stats
ORDER BY hour_bucket DESC;

-- 6. 容器镜像使用统计
-- 用途：分析容器镜像的使用情况和大小分布
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    ci.name,
    ci.tag,
    ci.repository,
    ci.size_bytes,
    pg_size_pretty(ci.size_bytes) as size_pretty,
    COUNT(*) as container_count,
    COUNT(DISTINCT c.namespace_id) as namespace_count,
    MIN(c.created_at) as first_used_at,
    MAX(c.created_at) as last_used_at,
    COUNT(CASE WHEN c.status = 'running' THEN 1 END) as running_containers
FROM container_images ci
LEFT JOIN containers c ON ci.id = c.image_id
GROUP BY ci.id, ci.name, ci.tag, ci.repository, ci.size_bytes
ORDER BY container_count DESC, ci.size_bytes DESC;

-- 7. 端口使用情况查询
-- 用途：检查端口冲突和使用统计
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    pm.container_port,
    pm.protocol,
    COUNT(*) as usage_count,
    COUNT(DISTINCT pm.container_id) as container_count,
    COUNT(DISTINCT c.namespace_id) as namespace_count,
    array_agg(DISTINCT pm.host_port) FILTER (WHERE pm.host_port IS NOT NULL) as host_ports,
    COUNT(DISTINCT pm.service_name) as service_count
FROM port_mappings pm
JOIN containers c ON pm.container_id = c.id
GROUP BY pm.container_port, pm.protocol
ORDER BY usage_count DESC, pm.container_port;

-- 8. 命名空间资源配额检查
-- 用途：检查命名空间的资源使用情况
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    n.name,
    n.kubernetes_name,
    n.status,

    -- 容器统计
    COUNT(c.id) as total_containers,
    COUNT(CASE WHEN c.status = 'running' THEN 1 END) as running_containers,

    -- CPU 资源统计
    SUM(CASE WHEN c.cpu_request IS NOT NULL
        THEN SUBSTRING(c.cpu_request FROM '([0-9.]+)')::numeric
        ELSE 0 END) as total_cpu_request,
    SUM(CASE WHEN c.cpu_limit IS NOT NULL
        THEN SUBSTRING(c.cpu_limit FROM '([0-9.]+)')::numeric
        ELSE 0 END) as total_cpu_limit,

    -- 内存资源统计
    SUM(CASE WHEN c.memory_request IS NOT NULL
        THEN pg_size_bytes(c.memory_request)
        ELSE 0 END) as total_memory_request_bytes,
    SUM(CASE WHEN c.memory_limit IS NOT NULL
        THEN pg_size_bytes(c.memory_limit)
        ELSE 0 END) as total_memory_limit_bytes,

    -- 存储统计
    COUNT(DISTINCT v.id) as total_volumes,
    SUM(CASE WHEN v.size IS NOT NULL
        THEN pg_size_bytes(v.size)
        ELSE 0 END) as total_storage_bytes,

    -- 配置和密钥
    COUNT(DISTINCT cm.id) as config_maps,
    COUNT(DISTINCT s.id) as secrets

FROM namespaces n
LEFT JOIN containers c ON n.id = c.namespace_id
LEFT JOIN volumes v ON n.id = v.namespace_id
LEFT JOIN config_maps cm ON n.id = cm.namespace_id
LEFT JOIN secrets s ON n.id = s.namespace_id
GROUP BY n.id, n.name, n.kubernetes_name, n.status
ORDER BY n.name;

-- 9. 容器重启趋势分析
-- 用途：分析容器重启的模式和原因
EXPLAIN (ANALYZE, BUFFERS)
WITH restart_analysis AS (
    SELECT
        c.id,
        c.name,
        n.name as namespace_name,
        c.restart_count,
        c.status,
        c.started_at,
        COUNT(csh.id) as status_changes,
        MAX(csh.changed_at) as last_status_change,
        COUNT(CASE WHEN csh.previous_status = 'running' AND csh.current_status != 'running' THEN 1 END) as crash_count
    FROM containers c
    JOIN namespaces n ON c.namespace_id = n.id
    LEFT JOIN container_status_history csh ON c.id = csh.container_id
        AND csh.changed_at >= CURRENT_DATE - INTERVAL '7 days'
    GROUP BY c.id, c.name, n.name, c.restart_count, c.status, c.started_at
)
SELECT
    name,
    namespace_name,
    restart_count,
    status,
    status_changes,
    crash_count,
    last_status_change,
    CASE
        WHEN started_at IS NOT NULL
        THEN CURRENT_TIMESTAMP - started_at
        ELSE NULL
    END as uptime,
    CASE
        WHEN restart_count > 5 THEN 'High Restart Frequency'
        WHEN restart_count > 2 THEN 'Medium Restart Frequency'
        WHEN restart_count > 0 THEN 'Low Restart Frequency'
        ELSE 'No Restarts'
    END as restart_risk_level
FROM restart_analysis
WHERE restart_count > 0 OR crash_count > 0
ORDER BY restart_count DESC, crash_count DESC;

-- 10. 数据库性能监控查询
-- 用途：监控数据库表的性能和大小
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) as index_size,
    pg_stat_get_numscans(c.oid) as seq_scans,
    pg_stat_get_tuples_returned(c.oid) as seq_tup_read,
    pg_stat_get_tuples_inserted(c.oid) as n_tup_ins,
    pg_stat_get_tuples_updated(c.oid) as n_tup_upd,
    pg_stat_get_tuples_deleted(c.oid) as n_tup_del,
    pg_stat_get_live_tuples(c.oid) as n_live_tup,
    pg_stat_get_dead_tuples(c.oid) as n_dead_tup,
    pg_stat_get_last_vacuum_time(c.oid) as last_vacuum,
    pg_stat_get_last_autovacuum_time(c.oid) as last_autovacuum,
    pg_stat_get_last_analyze_time(c.oid) as last_analyze,
    pg_stat_get_last_autoanalyze_time(c.oid) as last_autoanalyze
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public'
    AND c.relkind = 'r'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 11. 慢查询检测（示例）
-- 用途：检测执行时间超过阈值的查询
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    query,
    calls,
    total_time,
    mean_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements
WHERE mean_time > 100  -- 超过100ms的查询
    AND calls > 10      -- 至少调用10次
ORDER BY mean_time DESC
LIMIT 20;

-- 12. 索引使用情况分析
-- 用途：分析索引的使用效率
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    CASE
        WHEN idx_scan = 0 THEN 'Unused Index'
        WHEN idx_scan < 100 THEN 'Low Usage Index'
        WHEN idx_scan < 1000 THEN 'Medium Usage Index'
        ELSE 'High Usage Index'
    END as usage_category
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC, pg_relation_size(indexrelid) DESC;

-- 13. 系统资源使用查询
-- 用途：检查数据库连接和锁的情况
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    datname as database_name,
    numbackends as active_connections,
    xact_commit as transactions_committed,
    xact_rollback as transactions_rolled_back,
    blks_read as blocks_read,
    blks_hit as blocks_hit,
    tup_returned as tuples_returned,
    tup_fetched as tuples_fetched,
    tup_inserted as tuples_inserted,
    tup_updated as tuples_updated,
    tup_deleted as tuples_deleted,
    deadlocks
FROM pg_stat_database
WHERE datname = current_database();

-- 14. 锁等待情况查询
-- 用途：检测可能的锁等待问题
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    pg_class.relname as table_name,
    pg_locks.locktype,
    pg_locks.mode,
    pg_locks.granted,
    pg_stat_activity.query,
    pg_stat_activity.pid,
    pg_stat_activity.usename,
    pg_stat_activity.application_name,
    pg_stat_activity.client_addr,
    age(clock_timestamp(), pg_stat_activity.query_start) as age
FROM pg_locks
JOIN pg_class ON pg_locks.relation = pg_class.oid
JOIN pg_stat_activity ON pg_locks.pid = pg_stat_activity.pid
WHERE NOT pg_locks.granted
ORDER BY pg_stat_activity.query_start;

-- 性能建议函数
CREATE OR REPLACE FUNCTION analyze_table_performance(table_name text)
RETURNS TABLE (
    column_name text,
    data_type text,
    null_percent numeric,
    distinct_count bigint,
    correlation numeric,
    recommendation text
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.attname as column_name,
        pg_catalog.format_type(a.atttypid, a.atttypmod) as data_type,
        ROUND((s.null_cnt * 100.0 / g.total_cnt), 2) as null_percent,
        s.n_distinct,
        s.correlation,
        CASE
            WHEN s.null_cnt > 0 AND a.attnotnull THEN 'Consider making this column nullable'
            WHEN s.null_cnt = 0 AND NOT a.attnotnull AND a.attlen > 0 THEN 'Consider making this column NOT NULL'
            WHEN s.n_distinct < 10 AND g.total_cnt > 1000 THEN 'Consider adding index on this column'
            WHEN s.correlation < 0.1 AND s.correlation > -0.1 AND a.attlen > 0 THEN 'Consider reordering table by this column'
            ELSE 'No recommendations'
        END as recommendation
    FROM pg_attribute a
    JOIN pg_class c ON a.attrelid = c.oid
    JOIN pg_namespace n ON c.relnamespace = n.oid
    JOIN LATERAL (
        SELECT
            COUNT(*) as total_cnt
        FROM pg_statistic st
        JOIN pg_class pc ON st.starelid = pc.oid
        WHERE pc.relname = table_name
        LIMIT 1
    ) g ON true
    JOIN LATERAL (
        SELECT
            COALESCE(stanullfrac, 0) * g.total_cnt as null_cnt,
            CASE
                WHEN stadistinct > 0 THEN stadistinct
                WHEN stadistinct < 0 THEN (-stadistinct) * g.total_cnt
                ELSE 0
            END as n_distinct,
            stacorrelation as correlation
        FROM pg_statistic st
        WHERE st.starelid = c.oid AND st.staattnum = a.attnum
    ) s ON true
    WHERE c.relname = table_name
        AND n.nspname = 'public'
        AND a.attnum > 0
        AND NOT a.attisdropped
    ORDER BY a.attnum;
END;
$$ LANGUAGE plpgsql;

-- 使用示例：分析containers表的性能
SELECT * FROM analyze_table_performance('containers');

-- 输出性能分析完成信息
DO $$
BEGIN
    RAISE NOTICE '性能优化查询脚本执行完成！';
    RAISE NOTICE '包含的查询类型：';
    RAISE NOTICE '1. 容器状态统计查询';
    RAISE NOTICE '2. 用户操作统计查询';
    RAISE NOTICE '3. 容器资源使用统计';
    RAISE NOTICE '4. 存储卷使用情况查询';
    RAISE NOTICE '5. 操作日志聚合查询';
    RAISE NOTICE '6. 容器镜像使用统计';
    RAISE NOTICE '7. 端口使用情况查询';
    RAISE NOTICE '8. 命名空间资源配额检查';
    RAISE NOTICE '9. 容器重启趋势分析';
    RAISE NOTICE '10. 数据库性能监控查询';
    RAISE NOTICE '11. 慢查询检测';
    RAISE NOTICE '12. 索引使用情况分析';
    RAISE NOTICE '13. 系统资源使用查询';
    RAISE NOTICE '14. 锁等待情况查询';
    RAISE NOTICE '15. 性能建议分析函数';
    RAISE NOTICE '';
    RAISE NOTICE '注意：这些查询都包含 EXPLAIN (ANALYZE, BUFFERS) 来分析执行计划';
    RAISE NOTICE '在生产环境中使用时，请移除 EXPLAIN 分析以获得实际结果';
END $$;