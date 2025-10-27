package database

import (
	"gorm.io/gorm"
)

// CreateUsersTable 创建用户表
type CreateUsersTable struct {
	BaseMigration
}

func (m *CreateUsersTable) Name() string {
	return "create_users_table"
}

func (m *CreateUsersTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.User{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateUsersTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("users"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateRolesTable 创建角色表
type CreateRolesTable struct {
	BaseMigration
}

func (m *CreateRolesTable) Name() string {
	return "create_roles_table"
}

func (m *CreateRolesTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.Role{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateRolesTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("roles"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateNamespacesTable 创建命名空间表
type CreateNamespacesTable struct {
	BaseMigration
}

func (m *CreateNamespacesTable) Name() string {
	return "create_namespaces_table"
}

func (m *CreateNamespacesTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.Namespace{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateNamespacesTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("namespaces"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateContainersTable 创建容器表
type CreateContainersTable struct {
	BaseMigration
}

func (m *CreateContainersTable) Name() string {
	return "create_containers_table"
}

func (m *CreateContainersTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.ContainerImage{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&model.Container{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateContainersTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("containers"); err != nil {
		return err
	}
	if err := db.Migrator().DropTable("container_images"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateConfigMapsTable 创建配置映射表
type CreateConfigMapsTable struct {
	BaseMigration
}

func (m *CreateConfigMapsTable) Name() string {
	return "create_configmaps_table"
}

func (m *CreateConfigMapsTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.ConfigMap{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateConfigMapsTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("config_maps"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateSecretsTable 创建密钥表
type CreateSecretsTable struct {
	BaseMigration
}

func (m *CreateSecretsTable) Name() string {
	return "create_secrets_table"
}

func (m *CreateSecretsTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.Secret{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateSecretsTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("secrets"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateServicesTable 创建服务表
type CreateServicesTable struct {
	BaseMigration
}

func (m *CreateServicesTable) Name() string {
	return "create_services_table"
}

func (m *CreateServicesTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.Service{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateServicesTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("services"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateVolumesTable 创建存储卷表
type CreateVolumesTable struct {
	BaseMigration
}

func (m *CreateVolumesTable) Name() string {
	return "create_volumes_table"
}

func (m *CreateVolumesTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.Volume{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateVolumesTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("volumes"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateOperationLogsTable 创建操作日志表
type CreateOperationLogsTable struct {
	BaseMigration
}

func (m *CreateOperationLogsTable) Name() string {
	return "create_operation_logs_table"
}

func (m *CreateOperationLogsTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.OperationLog{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateOperationLogsTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("operation_logs"); err != nil {
		return err
	}
	return m.removeRecord(db)
}

// CreateResourceUsageTable 创建资源使用表
type CreateResourceUsageTable struct {
	BaseMigration
}

func (m *CreateResourceUsageTable) Name() string {
	return "create_resource_usage_table"
}

func (m *CreateResourceUsageTable) Up(db *gorm.DB) error {
	err := db.AutoMigrate(&model.ContainerResourceUsage{})
	if err != nil {
		return err
	}
	return m.record(db)
}

func (m *CreateResourceUsageTable) Down(db *gorm.DB) error {
	if err := db.Migrator().DropTable("container_resource_usage"); err != nil {
		return err
	}
	return m.removeRecord(db)
}