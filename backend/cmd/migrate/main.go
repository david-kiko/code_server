package main

import (
	"container-platform-backend/internal/database"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "migrate",
		Usage: "Database migration tool",
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Run all pending migrations",
				Action: runMigrationsUp,
			},
			{
				Name:  "down",
				Usage: "Rollback last migration",
				Action: runMigrationsDown,
			},
			{
				Name:  "status",
				Usage: "Show migration status",
				Action: showMigrationStatus,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runMigrationsUp(c *cli.Context) error {
	config := getDatabaseConfig()
	db, err := database.NewDatabase(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.CloseDatabase(db)

	// 运行迁移
	migrations := getAllMigrations()
	for _, migration := range migrations {
		if err := migration.Up(db); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Name(), err)
		}
		log.Printf("Migration %s completed successfully", migration.Name())
	}

	log.Println("All migrations completed successfully")
	return nil
}

func runMigrationsDown(c *cli.Context) error {
	config := getDatabaseConfig()
	db, err := database.NewDatabase(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.CloseDatabase(db)

	// 回滚最后一个迁移
	migrations := getAllMigrations()
	if len(migrations) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	lastMigration := migrations[len(migrations)-1]
	if err := lastMigration.Down(db); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", lastMigration.Name(), err)
	}

	log.Printf("Migration %s rolled back successfully", lastMigration.Name())
	return nil
}

func showMigrationStatus(c *cli.Context) error {
	config := getDatabaseConfig()
	db, err := database.NewDatabase(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.CloseDatabase(db)

	// 检查迁移状态
	migrations := getAllMigrations()
	for _, migration := range migrations {
		status := migration.Status(db)
		log.Printf("Migration %s: %s", migration.Name(), status)
	}

	return nil
}

func getDatabaseConfig() *database.Config {
	return &database.Config{
		Host:            getEnvOrDefault("DATABASE_HOST", "localhost"),
		Port:            getEnvOrDefault("DATABASE_PORT", "5432"),
		Name:            getEnvOrDefault("DATABASE_NAME", "container_management"),
		User:            getEnvOrDefault("DATABASE_USER", "postgres"),
		Password:        getEnvOrDefault("DATABASE_PASSWORD", "postgres"),
		SSLMode:         getEnvOrDefault("DATABASE_SSLMODE", "disable"),
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 0, // 使用默认值
		ConnMaxIdleTime: 0, // 使用默认值
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}