// pkg/database/migrations/migration.go
package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Migration struct {
	ID        int
	Name      string
	Batch     int
	CreatedAt time.Time
}

type MigrationRunner struct {
	db         *sql.DB
	migrations map[string]MigrationFile
}

type MigrationFile interface {
	Up() string
	Down() string
	GetName() string
}

func NewMigrationRunner(db *sql.DB) *MigrationRunner {
	runner := &MigrationRunner{
		db:         db,
		migrations: make(map[string]MigrationFile),
	}
	runner.createMigrationsTable()
	return runner
}

func (r *MigrationRunner) createMigrationsTable() {
	query := `CREATE TABLE IF NOT EXISTS migrations (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		batch INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create migrations table:", err)
	}
}

func (r *MigrationRunner) RegisterMigration(migration MigrationFile) {
	r.migrations[migration.GetName()] = migration
}

func (r *MigrationRunner) RunMigrations() error {
	// Get already run migrations
	runMigrations, err := r.getRunMigrations()
	if err != nil {
		return err
	}

	// Determine which migrations to run
	var toRun []MigrationFile
	for name, migration := range r.migrations {
		if !r.migrationRun(runMigrations, name) {
			toRun = append(toRun, migration)
		}
	}

	// Sort migrations by name (which should include timestamp)
	sort.Slice(toRun, func(i, j int) bool {
		return toRun[i].GetName() < toRun[j].GetName()
	})

	// Get current batch number
	batch, err := r.getNextBatchNumber()
	if err != nil {
		return err
	}

	// Run migrations
	for _, migration := range toRun {
		log.Printf("Running migration: %s", migration.GetName())
		_, err := r.db.Exec(migration.Up())
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %v", migration.GetName(), err)
		}

		// Record migration
		_, err = r.db.Exec("INSERT INTO migrations (name, batch) VALUES (?, ?)", migration.GetName(), batch)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %v", migration.GetName(), err)
		}
	}

	return nil
}

func (r *MigrationRunner) getRunMigrations() ([]Migration, error) {
	rows, err := r.db.Query("SELECT id, name, batch, created_at FROM migrations ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var m Migration
		err := rows.Scan(&m.ID, &m.Name, &m.Batch, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, nil
}

func (r *MigrationRunner) migrationRun(migrations []Migration, name string) bool {
	for _, m := range migrations {
		if m.Name == name {
			return true
		}
	}
	return false
}

func (r *MigrationRunner) getNextBatchNumber() (int, error) {
	var maxBatch sql.NullInt64
	err := r.db.QueryRow("SELECT MAX(batch) FROM migrations").Scan(&maxBatch)
	if err != nil {
		return 0, err
	}

	if !maxBatch.Valid {
		return 1, nil
	}

	return int(maxBatch.Int64) + 1, nil
}

// Create a migration file template
func CreateMigration(name string) error {
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.go", timestamp, strings.ToLower(name))

	content := fmt.Sprintf(`package migrations

import "github.com/google/uuid"

type %s struct {
	name string
}

func New%s() *%s {
	return &%s{
		name: "%s_%s",
	}
}

func (m *%s) GetName() string {
	return m.name
}

func (m *%s) Up() string {
	return `+"`"+`CREATE TABLE examples (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`+"`"+`
}

func (m *%s) Down() string {
	return "DROP TABLE examples"
}
`, name, name, name, name, timestamp, strings.ToLower(name), name, name, name)

	// Create migrations directory if it doesn't exist
	err := os.MkdirAll("database/migrations", 0755)
	if err != nil {
		return err
	}

	// Write migration file
	path := filepath.Join("database/migrations", filename)
	return os.WriteFile(path, []byte(content), 0644)
}
