package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Migration 数据库迁移接口
type Migration interface {
	Name() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
	Status(db *gorm.DB) string
}

// migrationRecord 迁移记录表
type migrationRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"unique;not null"`
	AppliedAt time.Time `gorm:"not null"`
}

// getAllMigrations 获取所有迁移
func getAllMigrations() []Migration {
	return []Migration{
		&CreateUsersTable{},
		&CreateRolesTable{},
		&CreateNamespacesTable{},
		&CreateContainersTable{},
		&CreateConfigMapsTable{},
		&CreateSecretsTable{},
		&CreateServicesTable{},
		&CreateVolumesTable{},
		&CreateOperationLogsTable{},
		&CreateResourceUsageTable{},
	}
}

// CreateMigrationTable 创建迁移记录表
func CreateMigrationTable(db *gorm.DB) error {
	return db.AutoMigrate(&migrationRecord{})
}

// ensureMigrationTable 确保迁移表存在
func ensureMigrationTable(db *gorm.DB) error {
	if err := CreateMigrationTable(db); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}
	return nil
}

// isMigrationApplied 检查迁移是否已应用
func isMigrationApplied(db *gorm.DB, name string) (bool, error) {
	if err := ensureMigrationTable(db); err != nil {
		return false, err
	}

	var count int64
	err := db.Model(&migrationRecord{}).Where("name = ?", name).Count(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// recordMigration 记录迁移
func recordMigration(db *gorm.DB, name string) error {
	if err := ensureMigrationTable(db); err != nil {
		return err
	}

	record := &migrationRecord{
		Name:      name,
		AppliedAt: time.Now(),
	}
	return db.Create(record).Error
}

// removeMigrationRecord 移除迁移记录
func removeMigrationRecord(db *gorm.DB, name string) error {
	return db.Where("name = ?", name).Delete(&migrationRecord{}).Error
}

// BaseMigration 基础迁移结构
type BaseMigration struct {
	name string
}

func (m *BaseMigration) Name() string {
	return m.name
}

func (m *BaseMigration) Status(db *gorm.DB) string {
	applied, err := isMigrationApplied(db, m.Name())
	if err != nil {
		return "error"
	}
	if applied {
		return "applied"
	}
	return "pending"
}

func (m *BaseMigration) record(db *gorm.DB) error {
	return recordMigration(db, m.Name())
}

func (m *BaseMigration) removeRecord(db *gorm.DB) error {
	return removeMigrationRecord(db, m.Name())
}