// cmd/migrate/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"mygola/config"
	"mygola/pkg/database"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	// Load configuration
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create migrator
	migrator := database.NewMigrator(db)

	// Handle commands
	switch os.Args[1] {
	case "migrate":
		if err := migrator.RunMigrations("database/migrations"); err != nil {
			log.Fatal("Migration failed:", err)
		}
		fmt.Println("Migrations completed successfully")

	case "rollback":
		if err := migrator.RollbackMigrations("database/migrations"); err != nil {
			log.Fatal("Rollback failed:", err)
		}
		fmt.Println("Rollback completed successfully")

	case "status":
		if err := showMigrationStatus(migrator); err != nil {
			log.Fatal("Failed to get migration status:", err)
		}

	case "make":
		if len(os.Args) < 3 {
			log.Fatal("Migration name is required")
		}
		// This would call the make migration function
		fmt.Println("Use the make command to create migrations")

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: migrate <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  migrate    Run all pending migrations")
	fmt.Println("  rollback   Rollback the last batch of migrations")
	fmt.Println("  status     Show migration status")
	fmt.Println("  make <name> Create a new migration (use the make command instead)")
}

func showMigrationStatus(migrator *database.Migrator) error {
	// Ensure migrations table exists
	if err := migrator.CreateMigrationsTable(); err != nil {
		return err
	}

	// Get run migrations
	runMigrations, err := migrator.GetRunMigrations()
	if err != nil {
		return err
	}

	// Get all migration files
	migrationFiles, err := migrator.GetMigrationFiles("database/migrations")
	if err != nil {
		return err
	}

	// Get pending migrations
	pendingMigrations := migrator.GetPendingMigrations(migrationFiles, runMigrations)

	fmt.Printf("Migrations Status:\n")
	fmt.Printf("  Run: %d\n", len(runMigrations))
	fmt.Printf("  Pending: %d\n", len(pendingMigrations))

	if len(runMigrations) > 0 {
		fmt.Println("\nRun Migrations:")
		for _, migration := range runMigrations {
			fmt.Printf("  %s (Batch: %d)\n", migration.Name, migration.Batch)
		}
	}

	if len(pendingMigrations) > 0 {
		fmt.Println("\nPending Migrations:")
		for _, migration := range pendingMigrations {
			fmt.Printf("  %s\n", migration)
		}
	}

	return nil
}
