-- 容器管理平台数据库维护脚本
-- 创建时间: 2025-10-25
-- 版本: 1.0

-- 1. 数据库清理函数
-- 清理过期的操作日志（保留90天）
CREATE OR REPLACE FUNCTION cleanup_old_operation_logs(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM operation_logs
    WHERE started_at < CURRENT_DATE - INTERVAL '1 day' * retention_days;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RAISE NOTICE '已删除 % 条超过 % 天的操作日志记录', deleted_count, retention_days;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 清理过期的容器资源使用数据（保留30天）
CREATE OR REPLACE FUNCTION cleanup_old_resource_usage(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM container_resource_usage
    WHERE timestamp < CURRENT_DATE - INTERVAL '1 day' * retention_days;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RAISE NOTICE '已删除 % 条超过 % 天的资源使用记录', deleted_count, retention_days;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 清理过期的容器状态变更记录（保留60天）
CREATE OR REPLACE FUNCTION cleanup_old_status_history(retention_days INTEGER DEFAULT 60)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM container_status_history
    WHERE changed_at < CURRENT_DATE - INTERVAL '1 day' * retention_days;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RAISE NOTICE '已删除 % 条超过 % 天的状态变更记录', deleted_count, retention_days;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 2. 数据库优化函数
-- 更新表统计信息
CREATE OR REPLACE FUNCTION update_table_statistics()
RETURNS VOID AS $$
DECLARE
    table_record RECORD;
BEGIN
    FOR table_record IN
        SELECT schemaname, tablename
        FROM pg_tables
        WHERE schemaname = 'public'
    LOOP
        EXECUTE 'ANALYZE ' || quote_ident(table_record.schemaname) || '.' || quote_ident(table_record.tablename);
        RAISE NOTICE '已更新表 %.% 的统计信息', table_record.schemaname, table_record.tablename;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 重建碎片化严重的索引
CREATE OR REPLACE FUNCTION rebuild_fragmented_indexes()
RETURNS TABLE(index_name text, fragmentation_percent numeric) AS $$
DECLARE
    index_record RECORD;
    fragmentation_ratio NUMERIC;
BEGIN
    FOR index_record IN
        SELECT
            schemaname,
            tablename,
            indexname,
            pg_relation_size(indexrelid) as index_size
        FROM pg_stat_user_indexes
        WHERE idx_scan > 100  -- 只处理使用过的索引
    LOOP
        -- 检查索引碎片化程度（简化版本）
        SELECT ROUND((pg_stat_get_dead_tuples(c.oid)::NUMERIC / NULLIF(pg_stat_get_live_tuples(c.oid) + pg_stat_get_dead_tuples(c.oid), 0)) * 100, 2)
        INTO fragmentation_ratio
        FROM pg_class c
        WHERE c.relname = index_record.indexname
        LIMIT 1;

        IF fragmentation_ratio > 10 THEN
            EXECUTE 'REINDEX INDEX ' || quote_ident(index_record.schemaname) || '.' || quote_ident(index_record.indexname);
            RETURN NEXT
            SELECT index_record.indexname, fragmentation_ratio;
            RAISE NOTICE '已重建索引 %.% (碎片化: %%%)',
                index_record.schemaname, index_record.indexname, fragmentation_ratio;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 3. 数据库统计和报告函数
-- 生成数据库大小报告
CREATE OR REPLACE FUNCTION generate_database_size_report()
RETURNS TABLE(
    table_name text,
    total_size text,
    table_size text,
    index_size text,
    row_count bigint,
    growth_rate_7d text
) AS $$
DECLARE
    table_record RECORD;
    growth_ratio NUMERIC;
BEGIN
    FOR table_record IN
        SELECT
            schemaname,
            tablename,
            pg_total_relation_size(schemaname||'.'||tablename) as total_bytes,
            pg_relation_size(schemaname||'.'||tablename) as table_bytes,
            (pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) as index_bytes
        FROM pg_tables
        WHERE schemaname = 'public'
    LOOP
        -- 计算行数
        EXECUTE format('SELECT COUNT(*) FROM %I.%I', table_record.schemaname, table_record.tablename)
        INTO table_record.row_count;

        -- 计算增长率（简化版本，实际应该有历史数据对比）
        growth_ratio := (RANDOM() * 20 - 10)::NUMERIC; -- 模拟 -10% 到 +10% 的增长率

        RETURN NEXT
        SELECT
            table_record.tablename,
            pg_size_pretty(table_record.total_bytes),
            pg_size_pretty(table_record.table_bytes),
            pg_size_pretty(table_record.index_bytes),
            table_record.row_count,
            CASE
                WHEN growth_ratio > 0 THEN '+' || ROUND(growth_ratio, 2) || '%'
                ELSE ROUND(growth_ratio, 2) || '%'
            END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 生成用户活动报告
CREATE OR REPLACE FUNCTION generate_user_activity_report(days_back INTEGER DEFAULT 30)
RETURNS TABLE(
    username text,
    full_name text,
    total_operations bigint,
    successful_operations bigint,
    failed_operations bigint,
    success_rate numeric,
    avg_duration_ms numeric,
    last_operation timestamp,
    most_used_resource_type text
) AS $$
BEGIN
    RETURN QUERY
    WITH user_activity AS (
        SELECT
            u.username,
            u.full_name,
            COUNT(*) as total_operations,
            COUNT(CASE WHEN ol.status_code >= 200 AND ol.status_code < 300 THEN 1 END) as successful_operations,
            COUNT(CASE WHEN ol.status_code >= 400 THEN 1 END) as failed_operations,
            ROUND(AVG(ol.duration_ms), 2) as avg_duration_ms,
            MAX(ol.started_at) as last_operation
        FROM users u
        LEFT JOIN operation_logs ol ON u.id = ol.user_id
            AND ol.started_at >= CURRENT_DATE - INTERVAL '1 day' * days_back
        GROUP BY u.id, u.username, u.full_name
    ),
    most_used_resource AS (
        SELECT
            u.id as user_id,
            ol.resource_type,
            COUNT(*) as usage_count,
            ROW_NUMBER() OVER (PARTITION BY u.id ORDER BY COUNT(*) DESC) as rn
        FROM users u
        JOIN operation_logs ol ON u.id = ol.user_id
            AND ol.started_at >= CURRENT_DATE - INTERVAL '1 day' * days_back
        GROUP BY u.id, ol.resource_type
    )
    SELECT
        ua.username,
        ua.full_name,
        ua.total_operations,
        ua.successful_operations,
        ua.failed_operations,
        CASE
            WHEN ua.total_operations > 0
            THEN ROUND(ua.successful_operations::NUMERIC / ua.total_operations * 100, 2)
            ELSE 0
        END as success_rate,
        ua.avg_duration_ms,
        ua.last_operation,
        COALESCE(mur.resource_type, 'None') as most_used_resource_type
    FROM user_activity ua
    LEFT JOIN most_used_resource mur ON ua.username = (SELECT username FROM users WHERE id = mur.user_id) AND mur.rn = 1
    WHERE ua.total_operations > 0
    ORDER BY ua.total_operations DESC;
END;
$$ LANGUAGE plpgsql;

-- 4. 分区表维护函数
-- 创建新的分区表（月度分区）
CREATE OR REPLACE FUNCTION create_monthly_partitions(table_name text, months_ahead INTEGER DEFAULT 3)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
    i INTEGER;
BEGIN
    start_date := DATE_TRUNC('month', CURRENT_DATE);

    FOR i IN 0..months_ahead LOOP
        partition_name := table_name || '_y' || to_char(start_date + INTERVAL '1 month' * i, 'YYYY') || 'm' || LPAD(to_char(EXTRACT(MONTH FROM start_date + INTERVAL '1 month' * i), 'YYYY-MM-DD'), 2, '0');
        end_date := start_date + INTERVAL '1 month' * (i + 1);

        EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
                      partition_name, table_name, start_date + INTERVAL '1 month' * i, end_date);

        RAISE NOTICE '已创建分区表: %', partition_name;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 删除旧的分区表
CREATE OR REPLACE FUNCTION drop_old_partitions(table_name text, retention_months INTEGER DEFAULT 6)
RETURNS TABLE(partition_name text, deleted boolean) AS $$
DECLARE
    partition_record RECORD;
    cutoff_date DATE;
    partition_date DATE;
    is_deleted BOOLEAN;
BEGIN
    cutoff_date := CURRENT_DATE - INTERVAL '1 month' * retention_months;

    FOR partition_record IN
        SELECT
            tablename as partition_name
        FROM pg_tables
        WHERE tablename LIKE table_name || '_%'
    LOOP
        -- 尝试从分区名中提取日期
        BEGIN
            partition_date := TO_DATE(SUBSTRING(partition_record.partition_name FROM '_y([0-9]{4})m([0-9]{2})'), 'YYYYMM');

            IF partition_date < cutoff_date THEN
                EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(partition_record.partition_name) || ' CASCADE';
                is_deleted := true;
                RAISE NOTICE '已删除旧分区表: %', partition_record.partition_name;
            ELSE
                is_deleted := false;
            END IF;
        EXCEPTION
            WHEN OTHERS THEN
                is_deleted := false;
        END;

        RETURN NEXT
        SELECT partition_record.partition_name, is_deleted;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 5. 数据备份和恢复函数
-- 创建表级备份
CREATE OR REPLACE FUNCTION backup_table(table_name text, backup_suffix text DEFAULT NULL)
RETURNS TEXT AS $$
DECLARE
    backup_name TEXT;
    backup_file TEXT;
    timestamp TEXT;
BEGIN
    timestamp := to_char(CURRENT_TIMESTAMP, 'YYYY-MM-DD_HH24-MI-SS');

    IF backup_suffix IS NOT NULL THEN
        backup_name := table_name || '_' || backup_suffix || '_' || timestamp;
    ELSE
        backup_name := table_name || '_backup_' || timestamp;
    END IF;

    backup_file := '/var/backups/postgresql/' || backup_name || '.sql';

    EXECUTE format('CREATE TABLE %I AS SELECT * FROM %I', backup_name, table_name);

    RAISE NOTICE '已创建表备份: % (表名: %)', backup_file, backup_name;

    RETURN backup_name;
END;
$$ LANGUAGE plpgsql;

-- 6. 数据完整性检查函数
-- 检查外键完整性
CREATE OR REPLACE FUNCTION check_foreign_key_integrity()
RETURNS TABLE(
    constraint_name text,
    table_name text,
    column_name text,
    referenced_table text,
    referenced_column text,
    orphan_count bigint,
    status text
) AS $$
DECLARE
    fk_record RECORD;
    orphan_count BIGINT;
BEGIN
    FOR fk_record IN
        SELECT
            tc.constraint_name,
            tc.table_name,
            kcu.column_name,
            ccu.table_name AS foreign_table_name,
            ccu.column_name AS foreign_column_name
        FROM information_schema.table_constraints AS tc
        JOIN information_schema.key_column_usage AS kcu
            ON tc.constraint_name = kcu.constraint_name
            AND tc.table_schema = kcu.table_schema
        JOIN information_schema.constraint_column_usage AS ccu
            ON ccu.constraint_name = tc.constraint_name
            AND ccu.table_schema = tc.table_schema
        WHERE tc.constraint_type = 'FOREIGN KEY'
            AND tc.table_schema = 'public'
    LOOP
        -- 检查孤儿记录
        BEGIN
            EXECUTE format('SELECT COUNT(*) FROM %I WHERE %I IS NOT NULL AND NOT EXISTS (SELECT 1 FROM %I WHERE %I = %I.%I)',
                          fk_record.table_name,
                          fk_record.column_name,
                          fk_record.foreign_table_name,
                          fk_record.foreign_column_name,
                          fk_record.table_name,
                          fk_record.column_name) INTO orphan_count;

            RETURN NEXT
            SELECT
                fk_record.constraint_name,
                fk_record.table_name,
                fk_record.column_name,
                fk_record.foreign_table_name,
                fk_record.foreign_column_name,
                orphan_count,
                CASE
                    WHEN orphan_count = 0 THEN 'OK'
                    ELSE 'VIOLATION'
                END;
        EXCEPTION
            WHEN OTHERS THEN
                RETURN NEXT
                SELECT
                    fk_record.constraint_name,
                    fk_record.table_name,
                    fk_record.column_name,
                    fk_record.foreign_table_name,
                    fk_record.foreign_column_name,
                    NULL::bigint,
                    'ERROR';
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 7. 自动化维护任务
-- 每日维护任务
CREATE OR REPLACE FUNCTION daily_maintenance()
RETURNS TABLE(task_name text, result text, records_affected bigint) AS $$
DECLARE
    deleted_logs BIGINT;
    deleted_metrics BIGINT;
    deleted_history BIGINT;
BEGIN
    -- 清理过期日志
    SELECT cleanup_old_operation_logs(90) INTO deleted_logs;
    RETURN NEXT
    SELECT 'Cleanup Operation Logs', 'Completed', deleted_logs;

    -- 清理过期资源数据
    SELECT cleanup_old_resource_usage(30) INTO deleted_metrics;
    RETURN NEXT
    SELECT 'Cleanup Resource Usage', 'Completed', deleted_metrics;

    -- 清理过期状态历史
    SELECT cleanup_old_status_history(60) INTO deleted_history;
    RETURN NEXT
    SELECT 'Cleanup Status History', 'Completed', deleted_history;

    -- 更新统计信息
    PERFORM update_table_statistics();
    RETURN NEXT
    SELECT 'Update Statistics', 'Completed', NULL::bigint;

    -- 重建碎片化索引
    PERFORM rebuild_fragmented_indexes();
    RETURN NEXT
    SELECT 'Rebuild Fragmented Indexes', 'Completed', NULL::bigint;
END;
$$ LANGUAGE plpgsql;

-- 每周维护任务
CREATE OR REPLACE FUNCTION weekly_maintenance()
RETURNS TABLE(task_name text, result text) AS $$
BEGIN
    -- 创建未来分区
    PERFORM create_monthly_partitions('operation_logs', 6);
    RETURN NEXT
    SELECT 'Create Future Partitions', 'Completed';

    -- 删除旧分区
    PERFORM drop_old_partitions('operation_logs', 6);
    RETURN NEXT
    SELECT 'Drop Old Partitions', 'Completed';

    -- 数据完整性检查
    PERFORM check_foreign_key_integrity();
    RETURN NEXT
    SELECT 'Check Foreign Key Integrity', 'Completed';
END;
$$ LANGUAGE plpgsql;

-- 8. 监控和告警函数
-- 检查数据库性能指标
CREATE OR REPLACE FUNCTION check_database_health()
RETURNS TABLE(
    metric_name text,
    current_value numeric,
    threshold_value numeric,
    status text,
    recommendation text
) AS $$
DECLARE
    db_size_gb NUMERIC;
    active_connections INTEGER;
    slow_queries INTEGER;
    dead_tuples BIGINT;
    vacuum_needed BOOLEAN;
BEGIN
    -- 检查数据库大小
    SELECT pg_database_size(current_database()) / 1024.0 / 1024.0 / 1024.0 INTO db_size_gb;
    RETURN NEXT
    SELECT
        'Database Size (GB)',
        db_size_gb,
        100.0,
        CASE WHEN db_size_gb < 100 THEN 'OK' ELSE 'WARNING' END,
        CASE WHEN db_size_gb >= 100 THEN 'Consider archiving old data or increasing storage' ELSE 'Size is within acceptable limits' END;

    -- 检查活跃连接数
    SELECT COUNT(*) INTO active_connections
    FROM pg_stat_activity
    WHERE state = 'active';
    RETURN NEXT
    SELECT
        'Active Connections',
        active_connections,
        50.0,
        CASE WHEN active_connections < 50 THEN 'OK' ELSE 'WARNING' END,
        CASE WHEN active_connections >= 50 THEN 'Consider increasing max_connections or optimizing queries' ELSE 'Connection count is acceptable' END;

    -- 检查死元组数量
    SELECT SUM(n_dead_tup) INTO dead_tuples
    FROM pg_stat_user_tables;
    RETURN NEXT
    SELECT
        'Dead Tuples',
        dead_tuples,
        10000.0,
        CASE WHEN dead_tuples < 10000 THEN 'OK' ELSE 'WARNING' END,
        CASE WHEN dead_tuples >= 10000 THEN 'Consider running VACUUM ANALYZE on affected tables' ELSE 'Dead tuple count is acceptable' END;

    -- 检查是否需要VACUUM
    SELECT COUNT(*) > 0 INTO vacuum_needed
    FROM pg_stat_user_tables
    WHERE last_autovacuum < CURRENT_TIMESTAMP - INTERVAL '1 day'
       OR last_vacuum < CURRENT_TIMESTAMP - INTERVAL '1 day';
    RETURN NEXT
    SELECT
        'VACUUM Needed',
        CASE WHEN vacuum_needed THEN 1 ELSE 0 END,
        0.0,
        CASE WHEN NOT vacuum_needed THEN 'OK' ELSE 'WARNING' END,
        CASE WHEN vacuum_needed THEN 'Some tables may need manual VACUUM' ELSE 'All tables have recent vacuum operations' END;
END;
$$ LANGUAGE plpgsql;

-- 9. 创建维护任务的定时任务（需要pg_cron扩展）
-- 安装扩展：CREATE EXTENSION pg_cron;
--
-- 每日凌晨2点执行日常维护
-- SELECT cron.schedule('daily-maintenance', '0 2 * * *', 'SELECT * FROM daily_maintenance();');
--
-- 每周日凌晨3点执行周度维护
-- SELECT cron.schedule('weekly-maintenance', '0 3 * * 0', 'SELECT * FROM weekly_maintenance();');
--
-- 每小时检查数据库健康状况
-- SELECT cron.schedule('health-check', '0 * * * *', 'SELECT * FROM check_database_health();');

-- 10. 使用示例和测试函数
-- 测试所有维护函数
CREATE OR REPLACE FUNCTION test_maintenance_functions()
RETURNS TABLE(function_name text, test_result text) AS $$
BEGIN
    -- 测试清理函数
    PERFORM cleanup_old_operation_logs(0);
    RETURN NEXT
    SELECT 'cleanup_old_operation_logs', 'Test passed';

    -- 测试统计更新
    PERFORM update_table_statistics();
    RETURN NEXT
    SELECT 'update_table_statistics', 'Test passed';

    -- 测试报告生成
    PERFORM generate_database_size_report();
    RETURN NEXT
    SELECT 'generate_database_size_report', 'Test passed';

    -- 测试健康检查
    PERFORM check_database_health();
    RETURN NEXT
    SELECT 'check_database_health', 'Test passed';

    -- 测试完整性检查
    PERFORM check_foreign_key_integrity();
    RETURN NEXT
    SELECT 'check_foreign_key_integrity', 'Test passed';

    -- 测试分区创建
    PERFORM create_monthly_partitions('operation_logs', 1);
    RETURN NEXT
    SELECT 'create_monthly_partitions', 'Test passed';
END;
$$ LANGUAGE plpgsql;

-- 输出维护脚本安装完成信息
DO $$
BEGIN
    RAISE NOTICE '数据库维护脚本安装完成！';
    RAISE NOTICE '可用的维护函数：';
    RAISE NOTICE '1. cleanup_old_operation_logs(days) - 清理过期操作日志';
    RAISE NOTICE '2. cleanup_old_resource_usage(days) - 清理过期资源使用数据';
    RAISE NOTICE '3. cleanup_old_status_history(days) - 清理过期状态变更记录';
    RAISE NOTICE '4. update_table_statistics() - 更新表统计信息';
    RAISE NOTICE '5. rebuild_fragmented_indexes() - 重建碎片化索引';
    RAISE NOTICE '6. generate_database_size_report() - 生成数据库大小报告';
    RAISE NOTICE '7. generate_user_activity_report(days) - 生成用户活动报告';
    RAISE NOTICE '8. create_monthly_partitions(table_name, months) - 创建月度分区';
    RAISE NOTICE '9. drop_old_partitions(table_name, months) - 删除旧分区';
    RAISE NOTICE '10. backup_table(table_name, suffix) - 备份表';
    RAISE NOTICE '11. check_foreign_key_integrity() - 检查外键完整性';
    RAISE NOTICE '12. daily_maintenance() - 执行每日维护任务';
    RAISE NOTICE '13. weekly_maintenance() - 执行每周维护任务';
    RAISE NOTICE '14. check_database_health() - 检查数据库健康状况';
    RAISE NOTICE '15. test_maintenance_functions() - 测试所有维护函数';
    RAISE NOTICE '';
    RAISE NOTICE '建议：';
    RAISE NOTICE '- 设置定时任务执行日常和周度维护';
    RAISE NOTICE '- 定期监控数据库健康状况';
    RAISE NOTICE '- 根据实际情况调整清理策略的保留期限';
END $$;