// pkg/database/migrator.go
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	ID        int
	Name      string
	Batch     int
	CreatedAt time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// CreateMigrationsTable creates the migrations table if it doesn't exist
func (m *Migrator) CreateMigrationsTable() error {
	query := `CREATE TABLE IF NOT EXISTS migrations (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		batch INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := m.db.Exec(query)
	return err
}

// GetRunMigrations retrieves all migrations that have been run
func (m *Migrator) GetRunMigrations() ([]Migration, error) {
	rows, err := m.db.Query("SELECT id, name, batch, created_at FROM migrations ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.ID, &migration.Name, &migration.Batch, &migration.CreatedAt)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// GetNextBatchNumber gets the next batch number for migrations
func (m *Migrator) GetNextBatchNumber() (int, error) {
	var maxBatch sql.NullInt64
	err := m.db.QueryRow("SELECT MAX(batch) FROM migrations").Scan(&maxBatch)
	if err != nil {
		return 0, err
	}

	if !maxBatch.Valid {
		return 1, nil
	}

	return int(maxBatch.Int64) + 1, nil
}

// RunMigrations runs all pending migrations
func (m *Migrator) RunMigrations(migrationsPath string) error {
	// Ensure migrations table exists
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Get already run migrations
	runMigrations, err := m.GetRunMigrations()
	if err != nil {
		return fmt.Errorf("failed to get run migrations: %v", err)
	}

	// Get all migration files
	migrationFiles, err := m.GetMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}

	// Filter out already run migrations
	pendingMigrations := m.GetPendingMigrations(migrationFiles, runMigrations)
	if len(pendingMigrations) == 0 {
		fmt.Println("No pending migrations.")
		return nil
	}

	// Get next batch number
	batch, err := m.GetNextBatchNumber()
	if err != nil {
		return fmt.Errorf("failed to get next batch number: %v", err)
	}

	// Run each pending migration
	for _, migrationFile := range pendingMigrations {
		fmt.Printf("Running migration: %s\n", migrationFile)

		// Read and execute migration SQL
		if err := m.RunMigrationFile(filepath.Join(migrationsPath, migrationFile), batch); err != nil {
			return fmt.Errorf("failed to run migration %s: %v", migrationFile, err)
		}
	}

	fmt.Printf("Ran %d migrations successfully.\n", len(pendingMigrations))
	return nil
}

// RunMigrationFile runs a single migration file
func (m *Migrator) RunMigrationFile(filePath string, batch int) error {
	// Read SQL file
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Split into individual statements
	sql := string(sqlBytes)
	statements := strings.Split(sql, ";")

	// Execute each statement
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		if _, err := m.db.Exec(statement); err != nil {
			return fmt.Errorf("failed to execute SQL: %v\nSQL: %s", err, statement)
		}
	}

	// Record migration in database
	migrationName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	_, err = m.db.Exec("INSERT INTO migrations (name, batch) VALUES (?, ?)", migrationName, batch)
	return err
}

// RollbackMigrations rolls back the last batch of migrations
func (m *Migrator) RollbackMigrations(migrationsPath string) error {
	// Get the last batch number
	var lastBatch int
	err := m.db.QueryRow("SELECT MAX(batch) FROM migrations").Scan(&lastBatch)
	if err != nil {
		return fmt.Errorf("failed to get last batch: %v", err)
	}

	// Get migrations from the last batch
	rows, err := m.db.Query("SELECT name FROM migrations WHERE batch = ? ORDER BY id DESC", lastBatch)
	if err != nil {
		return fmt.Errorf("failed to get migrations for batch %d: %v", lastBatch, err)
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		migrations = append(migrations, name)
	}

	if len(migrations) == 0 {
		fmt.Println("No migrations to rollback.")
		return nil
	}

	// Rollback each migration
	for _, migration := range migrations {
		fmt.Printf("Rolling back migration: %s\n", migration)

		// Find and run the down migration file
		downFile := fmt.Sprintf("%s_down.sql", migration)
		downPath := filepath.Join(migrationsPath, downFile)

		if _, err := os.Stat(downPath); os.IsNotExist(err) {
			return fmt.Errorf("down migration file not found: %s", downPath)
		}

		// Read and execute down migration SQL
		sqlBytes, err := os.ReadFile(downPath)
		if err != nil {
			return err
		}

		sql := string(sqlBytes)
		statements := strings.Split(sql, ";")

		for _, statement := range statements {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}

			if _, err := m.db.Exec(statement); err != nil {
				return fmt.Errorf("failed to execute SQL: %v\nSQL: %s", err, statement)
			}
		}

		// Remove migration record
		_, err = m.db.Exec("DELETE FROM migrations WHERE name = ?", migration)
		if err != nil {
			return fmt.Errorf("failed to delete migration record: %v", err)
		}
	}

	fmt.Printf("Rolled back %d migrations successfully.\n", len(migrations))
	return nil
}

// GetMigrationFiles gets all migration files in the directory
func (m *Migrator) GetMigrationFiles(migrationsPath string) ([]string, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}

	// Sort migrations by name (which should include timestamp)
	sort.Strings(migrations)
	return migrations, nil
}

// GetPendingMigrations filters out migrations that have already been run
func (m *Migrator) GetPendingMigrations(allMigrations []string, runMigrations []Migration) []string {
	runMap := make(map[string]bool)
	for _, migration := range runMigrations {
		runMap[migration.Name+".sql"] = true
	}

	var pending []string
	for _, migration := range allMigrations {
		if !runMap[migration] {
			pending = append(pending, migration)
		}
	}

	return pending
}
