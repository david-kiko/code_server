# 容器编排管理平台数据库设计总结

**文档版本**: 1.0
**创建日期**: 2025-10-25
**项目**: 容器编排管理平台 (001-container-orchestration)

## 1. 设计概述

本数据库设计专为容器编排管理平台量身定制，采用PostgreSQL作为主数据库，支持完整的容器生命周期管理、用户权限控制、操作审计和资源监控。设计遵循第三范式，确保数据一致性，同时通过适当的冗余和索引优化查询性能。

### 1.1 核心设计原则
- **数据一致性**: 使用外键约束确保引用完整性
- **性能优化**: 合理设计索引，支持高并发查询
- **扩展性**: 支持水平分片和读写分离
- **安全性**: 敏感数据加密存储，行级安全策略
- **审计能力**: 完整的操作日志记录和状态变更追踪
- **维护性**: 自动化数据清理和性能优化

### 1.2 技术特性
- **数据库**: PostgreSQL 14+
- **字符集**: UTF-8
- **时区**: UTC
- **ORM**: GORM (Go)
- **索引策略**: B-tree、GIN、部分索引、复合索引
- **分区策略**: 时间分区（月度）
- **备份策略**: 全量备份 + 增量备份

## 2. 核心数据模型架构

### 2.1 用户权限管理模块

#### 核心实体关系
```
users (用户) ←→ user_roles (用户角色关联) ←→ roles (角色)
```

#### 设计要点
- **RBAC模型**: 基于角色的访问控制
- **权限存储**: JSONB格式，支持灵活的权限配置
- **角色过期**: 支持临时角色分配
- **系统角色**: 内置管理员、运维、只读用户角色

#### 最佳实践
```sql
-- 角色权限示例
{
  "containers": ["read", "create", "update", "delete"],
  "namespaces": ["read", "create", "update"],
  "volumes": ["read", "create", "update", "delete"],
  "config_maps": ["read", "create", "update", "delete"],
  "secrets": ["read", "create", "update", "delete"],
  "logs": ["read"],
  "monitoring": ["read"]
}
```

### 2.2 容器生命周期管理模块

#### 核心实体关系
```
namespaces (命名空间) ←→ containers (容器) ←→ container_images (镜像)
                                      ↓
                              container_volumes (容器卷关联)
                                      ↓
                              port_mappings (端口映射)
```

#### 设计要点
- **容器状态管理**: 支持完整的容器生命周期状态
- **Kubernetes集成**: 保存K8s相关元数据
- **资源配置**: CPU、内存的请求和限制
- **网络配置**: 端口映射、服务配置
- **存储配置**: 多种存储类型支持

#### 状态流转
```
pending → running → paused → stopped
    ↓        ↓        ↓        ↓
  failed   running   running  deleted
```

### 2.3 存储和配置管理模块

#### 學心实体关系
```
volumes (存储卷) ←→ container_volumes (关联)
config_maps (配置映射)
secrets (密钥)
environment_variables (环境变量)
```

#### 设计要点
- **多存储类型**: PVC、ConfigMap、Secret、HostPath、EmptyDir
- **访问模式**: ReadWriteOnce、ReadOnlyMany、ReadWriteMany
- **配置分离**: 配置和密钥独立管理
- **环境变量**: 支持字面量和引用配置

### 2.4 操作审计和监控模块

#### 核心实体关系
```
operation_logs (操作日志) ←→ container_status_history (状态变更历史)
                              ↓
                        container_resource_usage (资源使用记录)
```

#### 设计要点
- **完整审计**: 记录所有用户操作和系统事件
- **状态追踪**: 容器状态变更的完整历史
- **性能监控**: CPU、内存、网络、磁盘使用情况
- **时间序列**: 支持历史趋势分析

## 3. 关键设计决策

### 3.1 为什么选择PostgreSQL？

1. **JSONB支持**: 完美支持灵活的权限和配置存储
2. **高级索引**: GIN、部分索引、复合索引优化
3. **分区表**: 自动管理历史数据
4. **扩展性**: 支持行级安全、加密函数
5. **可靠性**: ACID事务，数据一致性保障

### 3.2 为什么采用时间分区？

1. **性能提升**: 查询范围缩小，索引更高效
2. **维护简化**: 旧数据快速清理
3. **存储优化**: 不同分区可使用不同存储策略
4. **备份灵活**: 可按分区进行备份恢复

### 3.3 为什么使用JSONB存储权限？

1. **灵活性**: 权限模型变更无需修改表结构
2. **查询效率**: GIN索引支持高效查询
3. **扩展性**: 支持细粒度权限控制
4. **版本兼容**: 向后兼容新权限字段

## 4. 索引优化策略

### 4.1 主要索引类型

#### B-tree索引（默认）
- 外键字段
- 状态字段
- 时间字段
- 唯一约束字段

#### GIN索引
- JSONB字段（权限、配置）
- 全文搜索字段

#### 部分索引
- 运行中容器的状态查询
- 失败操作的日志查询
- 活跃用户的会话查询

#### 复合索引
- 命名空间+状态查询
- 用户+操作+时间查询
- 容器+资源+时间查询

### 4.2 索引使用原则

1. **选择性高的字段优先**
2. **常用查询条件优先**
3. **避免过度索引**
4. **定期分析索引使用情况**
5. **及时清理无用索引**

## 5. 性能优化建议

### 5.1 查询优化

#### 使用物化视图
```sql
CREATE MATERIALIZED VIEW container_summary AS
SELECT
    n.name as namespace_name,
    COUNT(*) as total_containers,
    COUNT(CASE WHEN c.status = 'running' THEN 1 END) as running_containers
FROM namespaces n
LEFT JOIN containers c ON n.id = c.namespace_id
GROUP BY n.id, n.name;
```

#### 使用CTE优化复杂查询
```sql
WITH user_operations AS (
    SELECT user_id, COUNT(*) as op_count
    FROM operation_logs
    WHERE started_at >= CURRENT_DATE - INTERVAL '30 days'
    GROUP BY user_id
)
SELECT u.username, u.full_name, uo.op_count
FROM users u
JOIN user_operations uo ON u.id = uo.user_id;
```

### 5.2 连接池配置

```go
sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
sqlDB.SetConnMaxIdleTime(time.Minute * 30) // 空闲连接最大生存时间
```

### 5.3 缓存策略

1. **应用层缓存**: Redis缓存热点数据
2. **查询缓存**: PostgreSQL查询计划缓存
3. **物化视图**: 定期刷新的汇总数据
4. **连接池**: 复用数据库连接

## 6. 数据安全和隐私

### 6.1 数据加密

#### 敏感数据加密
```sql
-- 使用pgcrypto扩展加密敏感数据
CREATE EXTENSION pgcrypto;

-- 加密存储示例
INSERT INTO user_secrets (user_id, secret_type, encrypted_value)
VALUES (1, 'api_key', pgp_sym_encrypt('secret_api_key', 'encryption_key'));
```

#### 传输加密
- SSL/TLS数据库连接
- API接口HTTPS
- 内部服务mTLS

### 6.2 访问控制

#### 行级安全策略
```sql
-- 用户只能访问有权限的命名空间下的容器
CREATE POLICY container_access_policy ON containers
USING (
    namespace_id IN (
        SELECT ns.id FROM namespaces ns
        WHERE has_namespace_permission(current_user_id(), ns.id)
    )
);
```

#### 列级权限控制
```sql
-- 敏感列只允许特定角色访问
REVOKE SELECT ON users FROM PUBLIC;
GRANT SELECT(id, username, full_name, status, created_at) ON users TO operators;
GRANT SELECT ON users TO admin_role;
```

## 7. 监控和告警

### 7.1 关键监控指标

#### 数据库指标
- 连接数
- 查询响应时间
- 锁等待时间
- 磁盘使用率
- 复制延迟

#### 业务指标
- 容器创建成功率
- 操作响应时间
- 用户活跃度
- 资源使用趋势

### 7.2 告警策略

#### 数据库告警
- 连接数超过阈值
- 慢查询增加
- 磁盘空间不足
- 复制中断

#### 业务告警
- 容器创建失败率高
- 用户操作异常
- 资源使用异常
- 审计日志缺失

## 8. 灾难恢复

### 8.1 备份策略

#### 全量备份
- 每日凌晨执行全量备份
- 保留30天备份文件
- 异地备份存储

#### 增量备份
- 每小时执行WAL归档
- 实时流复制
- 点对点恢复

### 8.2 恢复流程

1. **评估损坏范围**
2. **选择恢复点**
3. **停止应用服务**
4. **恢复数据库**
5. **验证数据完整性**
6. **恢复应用服务**
7. **监控运行状态**

## 9. 扩展性考虑

### 9.1 水平扩展

#### 读写分离
- 主库处理写操作
- 多个从库处理读操作
- 使用ProxySQL进行路由

#### 分库分表
- 按命名空间分库
- 按时间分表
- 使用分布式中间件

### 9.2 垂直扩展

#### 资源配置
- CPU资源充足
- 内存足够大
- 高速SSD存储
- 高速网络连接

## 10. 最佳实践总结

### 10.1 开发阶段

1. **遵循数据库设计规范**
2. **编写高效的SQL查询**
3. **使用适当的事务隔离级别**
4. **实现完善的错误处理**
5. **编写单元测试和集成测试**

### 10.2 运维阶段

1. **定期执行维护任务**
2. **监控数据库性能指标**
3. **及时处理告警信息**
4. **定期测试备份恢复**
5. **持续优化查询性能**

### 10.3 安全阶段

1. **定期更新数据库版本**
2. **实施最小权限原则**
3. **定期进行安全审计**
4. **加密敏感数据**
5. **监控异常访问**

## 11. 文件清单

本设计包含以下文件：

1. **D:\work\github\code_server\specs\001-container-orchestration\data-model.md**
   - 完整的数据模型设计文档
   - 详细的表结构和关系说明
   - 索引优化策略
   - 数据迁移方案

2. **D:\work\github\code_server\specs\001-container-orchestration\database-scripts\init-database.sql**
   - 数据库初始化脚本
   - 表结构创建
   - 索引创建
   - 触发器和函数定义

3. **D:\work\github\code_server\specs\001-container-orchestration\database-scripts\sample-data.sql**
   - 示例数据插入脚本
   - 测试数据生成
   - 场景数据模拟

4. **D:\work\github\code_server\specs\001-container-orchestration\database-scripts\performance-queries.sql**
   - 性能优化查询示例
   - 执行计划分析
   - 监控查询模板

5. **D:\work\github\code_server\specs\001-container-orchestration\database-scripts\maintenance-scripts.sql**
   - 数据库维护脚本
   - 自动化任务函数
   - 清理和优化工具

## 12. 实施建议

### 12.1 部署顺序

1. **数据库环境准备**
   - 安装PostgreSQL 14+
   - 配置参数优化
   - 创建数据库和用户

2. **初始化数据库结构**
   - 执行init-database.sql
   - 创建必要的扩展
   - 配置权限

3. **插入基础数据**
   - 执行sample-data.sql
   - 创建管理员账户
   - 配置基础角色

4. **部署应用服务**
   - 配置数据库连接
   - 测试基本功能
   - 验证数据完整性

5. **配置监控和维护**
   - 设置定时任务
   - 配置监控告警
   - 测试备份恢复

### 12.2 性能调优

1. **数据库参数调优**
   - memory配置
   - 连接数配置
   - WAL配置

2. **应用层优化**
   - 连接池配置
   - 查询缓存
   - 批量操作

3. **索引优化**
   - 监控索引使用情况
   - 调整复合索引
   - 清理无用索引

### 12.3 监控指标

1. **基础指标**
   - 数据库连接数
   - 查询响应时间
   - 事务吞吐量

2. **业务指标**
   - 容器操作成功率
   - 用户活跃度
   - 资源使用情况

3. **告警阈值**
   - 连接数 > 80%
   - 响应时间 > 200ms
   - 错误率 > 5%

## 13. 结论

本数据库设计充分考虑了容器管理平台的业务需求和技术特点，通过合理的数据建模、索引优化、安全策略和维护方案，为平台提供了稳定、高效、安全的数据存储基础。设计具有良好的扩展性和维护性，能够支撑平台的长期发展需求。

建议在实施过程中严格按照设计文档执行，并结合实际运行情况持续优化调整。定期进行性能评估和安全审计，确保数据库系统的稳定运行。