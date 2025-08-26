// main.go
package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"mygola/app/providers"
	"mygola/config"
	"mygola/pkg/cache"
	"mygola/pkg/database"
	"mygola/pkg/foundation"
	"mygola/pkg/routing"
	"mygola/pkg/schedule"
	"mygola/pkg/session"
	"mygola/pkg/view"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mygola",
	Short: "MyGola CLI Tool",
}

// ------------------------
// Commands
// ------------------------
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run all pending migrations",
	Run: func(cmd *cobra.Command, args []string) {
		db := connectDB()
		defer db.Close()
		
		migrator := database.NewMigrator(db)
		if err := migrator.RunMigrations("database/migrations"); err != nil {
			log.Fatal("Migration failed:", err)
		}
		log.Println("✅ Migrations completed successfully")
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "migrate:rollback",
	Short: "Rollback last batch of migrations",
	Run: func(cmd *cobra.Command, args []string) {
		db := connectDB()
		defer db.Close()
		
		migrator := database.NewMigrator(db)
		if err := migrator.RollbackMigrations("database/migrations"); err != nil {
			log.Fatal("Rollback failed:", err)
		}
		log.Println("✅ Rollback completed successfully")
	},
}

var statusCmd = &cobra.Command{
	Use:   "migrate:status",
	Short: "Show migration status",
	Run: func(cmd *cobra.Command, args []string) {
		db := connectDB()
		defer db.Close()
		
		migrator := database.NewMigrator(db)
		if err := showMigrationStatus(migrator); err != nil {
			log.Fatal(err)
		}
	},
}

var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Generate code (controller, model, migration, etc.)",
}

var makeControllerCmd = &cobra.Command{
	Use:   "controller [name]",
	Short: "Create a new controller",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		isResource, _ := cmd.Flags().GetBool("resource")
		makeController(name, isResource)
	},
}

var makeModelCmd = &cobra.Command{
	Use:   "model [name]",
	Short: "Create a new model",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		createMigration, _ := cmd.Flags().GetBool("migration")
		makeModel(name, createMigration)
	},
}

var makeMigrationCmd = &cobra.Command{
	Use:   "migration [name]",
	Short: "Create a new migration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		makeMigration(name)
	},
}

var makeMiddlewareCmd = &cobra.Command{
	Use:   "middleware [name]",
	Short: "Create a new middleware",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		makeMiddleware(name)
	},
}

var makeProviderCmd = &cobra.Command{
	Use:   "provider [name]",
	Short: "Create a new service provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		makeProvider(name)
	},
}

var makeRequestCmd = &cobra.Command{
	Use:   "request [name]",
	Short: "Create a new form request",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		makeRequest(name)
	},
}

var makeSeedCmd = &cobra.Command{
	Use:   "seed [name]",
	Short: "Create a new database seeder",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		makeSeed(name)
	},
}

var makeViewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "Create a new view",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		makeView(name)
	},
}

// ------------------------
// Main
// ------------------------
func main() {
	// Add commands to root
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(makeCmd)
	
	// Add subcommands to make command
	makeCmd.AddCommand(makeControllerCmd)
	makeCmd.AddCommand(makeModelCmd)
	makeCmd.AddCommand(makeMigrationCmd)
	makeCmd.AddCommand(makeMiddlewareCmd)
	makeCmd.AddCommand(makeProviderCmd)
	makeCmd.AddCommand(makeRequestCmd)
	makeCmd.AddCommand(makeSeedCmd)
	makeCmd.AddCommand(makeViewCmd)
	
	// Add flags to make commands
	makeControllerCmd.Flags().BoolP("resource", "r", false, "Create a resource controller")
	makeModelCmd.Flags().BoolP("migration", "m", false, "Create a migration for the model")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// ------------------------
// Helper Functions
// ------------------------
func connectDB() *sql.DB {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	
	return db
}

func startServer() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db := connectDB()
	defer db.Close()

	router := routing.NewRouter()

	// Session manager
	sessionStore := session.NewMemoryStore()
	sessionManager := session.NewManager(sessionStore, "mygola_session")

	// Template engine
	templateEngine := view.NewTemplateEngine("views", "app")

	// Cache
	appCache := cache.NewMemoryCache()

	// Scheduler
	scheduler := schedule.NewScheduler()
	scheduler.Daily(func() {
		log.Println("Running daily task: cleaning up expired sessions")
	})
	go scheduler.Start()

	// Application container
	app := foundation.NewApplication()
	app.Bind((*sql.DB)(nil), db)
	app.Bind((*session.Manager)(nil), sessionManager)
	app.Bind((*view.TemplateEngine)(nil), templateEngine)
	app.Bind((*cache.Cache)(nil), appCache)

	app.Register(providers.NewRouteServiceProvider(router))
	app.Boot()

	// Middleware
	router.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sess, err := sessionManager.Start(w, r)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			next(w, r)
			sess.Save()
		}
	})

	router.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s", r.Method, r.URL.Path)
			next(w, r)
		}
	})

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	url := fmt.Sprintf("http://%s", serverAddr)
	log.Printf("🚀 Server running at %s", url)
	log.Fatal(http.ListenAndServe(serverAddr, router))
}

// ------------------------
// Migration Status Helper
// ------------------------
func showMigrationStatus(migrator *database.Migrator) error {
	if err := migrator.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	runMigrations, err := migrator.GetRunMigrations()
	if err != nil {
		return fmt.Errorf("failed to get run migrations: %v", err)
	}

	migrationFiles, err := migrator.GetMigrationFiles("database/migrations")
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}

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

// ------------------------
// Make Command Functions
// ------------------------
type TemplateData struct {
	Name      string
	TableName string
	Timestamp string
	Package   string
}

func makeController(name string, isResource bool) {
	data := TemplateData{
		Name:    toCamelCase(name),
		Package: "controllers",
	}

	tmplContent := controllerTemplate
	if isResource {
		tmplContent = resourceControllerTemplate
	}

	// Create directory if it doesn't exist
	dir := "app/http/controllers"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create controllers directory:", err)
	}

	// Create file
	filename := fmt.Sprintf("%s_controller.go", strings.ToLower(name))
	path := filepath.Join(dir, filename)

	if err := createFileFromTemplate(path, tmplContent, data); err != nil {
		log.Fatal("Failed to create controller:", err)
	}

	fmt.Printf("✅ Controller created: %s\n", path)
}

func makeModel(name string, createMigration bool) {
	data := TemplateData{
		Name:      toCamelCase(name),
		TableName: toSnakeCase(name),
		Package:   "models",
	}

	// Create models directory
	dir := "app/models"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create models directory:", err)
	}

	// Create model file
	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	path := filepath.Join(dir, filename)
	if err := createFileFromTemplate(path, modelTemplate, data); err != nil {
		log.Fatal("Failed to create model:", err)
	}
	fmt.Printf("✅ Model created: %s\n", path)

	// If -m flag is passed, create migration
	if createMigration {
		makeMigration(name)
	}
}

func makeMigration(name string) {
	timestamp := time.Now().Format("20060102150405")
	migrationName := fmt.Sprintf("%s_%s", timestamp, strings.ToLower(name))

	// Create directory if it doesn't exist
	dir := "database/migrations"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create migrations directory:", err)
	}

	// Create up migration file
	upFilename := fmt.Sprintf("%s_up.sql", migrationName)
	upPath := filepath.Join(dir, upFilename)

	upContent := fmt.Sprintf(`-- Migration: %s
-- Created at: %s

-- Add your SQL here
CREATE TABLE %s (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
`, name, time.Now().Format(time.RFC3339), toSnakeCase(name)+"s")

	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		log.Fatal("Failed to create up migration:", err)
	}

	// Create down migration file
	downFilename := fmt.Sprintf("%s_down.sql", migrationName)
	downPath := filepath.Join(dir, downFilename)

	downContent := fmt.Sprintf(`-- Rollback migration: %s
-- Created at: %s

-- Add your rollback SQL here
DROP TABLE IF EXISTS %s;
`, name, time.Now().Format(time.RFC3339), toSnakeCase(name)+"s")

	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		log.Fatal("Failed to create down migration:", err)
	}

	fmt.Printf("✅ Migration created:\n  Up: %s\n  Down: %s\n", upPath, downPath)
}

func makeMiddleware(name string) {
	data := TemplateData{
		Name:    toCamelCase(name),
		Package: "middleware",
	}

	// Create directory if it doesn't exist
	dir := "internal/app/http/middleware"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create middleware directory:", err)
	}

	// Create file
	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	path := filepath.Join(dir, filename)

	if err := createFileFromTemplate(path, middlewareTemplate, data); err != nil {
		log.Fatal("Failed to create middleware:", err)
	}

	fmt.Printf("✅ Middleware created: %s\n", path)
}

func makeProvider(name string) {
	data := TemplateData{
		Name:    toCamelCase(name),
		Package: "providers",
	}

	// Create directory if it doesn't exist
	dir := "app/providers"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create providers directory:", err)
	}

	// Create file
	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	path := filepath.Join(dir, filename)

	if err := createFileFromTemplate(path, providerTemplate, data); err != nil {
		log.Fatal("Failed to create provider:", err)
	}

	fmt.Printf("✅ Provider created: %s\n", path)
}

func makeRequest(name string) {
	data := TemplateData{
		Name:    toCamelCase(name),
		Package: "requests",
	}

	// Create directory if it doesn't exist
	dir := "app/http/requests"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create requests directory:", err)
	}

	// Create file
	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	path := filepath.Join(dir, filename)

	if err := createFileFromTemplate(path, requestTemplate, data); err != nil {
		log.Fatal("Failed to create request:", err)
	}

	fmt.Printf("✅ Request created: %s\n", path)
}

func makeSeed(name string) {
	data := TemplateData{
		Name:    toCamelCase(name),
		Package: "seeds",
	}

	// Create directory if it doesn't exist
	dir := "database/seeds"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create seeds directory:", err)
	}

	// Create file
	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	path := filepath.Join(dir, filename)

	if err := createFileFromTemplate(path, seedTemplate, data); err != nil {
		log.Fatal("Failed to create seed:", err)
	}

	fmt.Printf("✅ Seed created: %s\n", path)
}

func makeView(name string) {
	// Create directory if it doesn't exist
	dir := filepath.Join("views", strings.ToLower(name))
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create view directory:", err)
	}

	// Create index view
	indexPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(indexPath, []byte(viewIndexTemplate), 0644); err != nil {
		log.Fatal("Failed to create view:", err)
	}

	// Create show view
	showPath := filepath.Join(dir, "show.html")
	if err := os.WriteFile(showPath, []byte(viewShowTemplate), 0644); err != nil {
		log.Fatal("Failed to create view:", err)
	}

	fmt.Printf("✅ Views created in: %s\n", dir)
}

func createFileFromTemplate(path, tmplContent string, data TemplateData) error {
	// Parse template
	tmpl, err := template.New("").Parse(tmplContent)
	if err != nil {
		return err
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Execute template
	return tmpl.Execute(file, data)
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Convert first character to uppercase
	result := strings.ToUpper(string(s[0])) + s[1:]

	// Remove any underscores and capitalize next letter
	for {
		i := strings.Index(result, "_")
		if i == -1 || i >= len(result)-1 {
			break
		}
		result = result[:i] + strings.ToUpper(string(result[i+1])) + result[i+2:]
	}

	return result
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// ------------------------
// Template definitions
// ------------------------
const controllerTemplate = `package {{.Package}}

import (
	"net/http"
)

type {{.Name}}Controller struct {
	// Add dependencies here
}

func New{{.Name}}Controller() *{{.Name}}Controller {
	return &{{.Name}}Controller{}
}

func (c *{{.Name}}Controller) Index(w http.ResponseWriter, r *http.Request) {
	// Implement index action
}
`

const resourceControllerTemplate = `package {{.Package}}

import (
	"net/http"
)

type {{.Name}}Controller struct {
	// Add dependencies here
}

func New{{.Name}}Controller() *{{.Name}}Controller {
	return &{{.Name}}Controller{}
}

func (c *{{.Name}}Controller) Index(w http.ResponseWriter, r *http.Request) {
	// List all resources
}

func (c *{{.Name}}Controller) Show(w http.ResponseWriter, r *http.Request) {
	// Show a specific resource
}

func (c *{{.Name}}Controller) Create(w http.ResponseWriter, r *http.Request) {
	// Show form to create a new resource
}

func (c *{{.Name}}Controller) Store(w http.ResponseWriter, r *http.Request) {
	// Store a new resource
}

func (c *{{.Name}}Controller) Edit(w http.ResponseWriter, r *http.Request) {
	// Show form to edit a resource
}

func (c *{{.Name}}Controller) Update(w http.ResponseWriter, r *http.Request) {
	// Update a resource
}

func (c *{{.Name}}Controller) Destroy(w http.ResponseWriter, r *http.Request) {
	// Delete a resource
}
`

const modelTemplate = `package {{.Package}}

import (
	"time"
	"mygola/pkg/helpers"
)

type {{.Name}} struct {
	ID        string    ` + "`" + `db:"id"` + "`" + `
	CreatedAt time.Time ` + "`" + `db:"created_at"` + "`" + `
	UpdatedAt time.Time ` + "`" + `db:"updated_at"` + "`" + `
}

func (m *{{.Name}}) TableName() string {
	return "{{.TableName}}"
}

func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{
		ID:        helpers.GenerateID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Add your model methods here
`

const middlewareTemplate = `package {{.Package}}

import (
	"net/http"
)

func {{.Name}}Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Your middleware logic here
		
		// Call the next handler
		next(w, r)
	}
}
`

const providerTemplate = `package {{.Package}}

import (
	"mygola/pkg/foundation"
)

type {{.Name}}Provider struct {}

func New{{.Name}}Provider() *{{.Name}}Provider {
	return &{{.Name}}Provider{}
}

func (p *{{.Name}}Provider) Register(app *foundation.Application) {
	// Register your services here
}

func (p *{{.Name}}Provider) Boot(app *foundation.Application) {
	// Boot your services here
}
`

const requestTemplate = `package {{.Package}}

import (
	"net/http"
	"mygola/pkg/validation"
)

type {{.Name}}Request struct {
	// Add your form fields here
}

func (r *{{.Name}}Request) Validate(req *http.Request) (*validation.Validator, error) {
	// Parse form data
	if err := req.ParseForm(); err != nil {
		return nil, err
	}
	
	// Create validator
	validator := validation.NewValidator(map[string]interface{}{
		// Add your form data here
	})
	
	// Define validation rules
	rules := map[string]string{
		// Add your validation rules here
	}
	
	// Validate
	if !validator.Validate(rules) {
		return validator, validation.ErrValidationFailed
	}
	
	return validator, nil
}
`

const seedTemplate = `package {{.Package}}

import (
	"database/sql"
	"fmt"
)

type {{.Name}}Seeder struct {}

func New{{.Name}}Seeder() *{{.Name}}Seeder {
	return &{{.Name}}Seeder{}
}

func (s *{{.Name}}Seeder) Run(db *sql.DB) error {
	// Implement your seeder logic here
	fmt.Println("Running {{.Name}} seeder...")
	
	// Example:
	// _, err := db.Exec("INSERT INTO table (column) VALUES (?)", "value")
	// if err != nil {
	//     return fmt.Errorf("failed to seed: %v", err)
	// }
	
	return nil
}
`

const viewIndexTemplate = `{{define "content"}}
<h1>{{.Title}}</h1>
<table>
    <thead>
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Actions</th>
        </tr>
    </thead>
    <tbody>
        {{range .Items}}
        <tr>
            <td>{{.ID}}</td>
            <td>{{.Name}}</td>
            <td>
                <a href="/{{$.Name}}/{{.ID}}">Show</a>
                <a href="/{{$.Name}}/{{.ID}}/edit">Edit</a>
                <form action="/{{$.Name}}/{{.ID}}" method="POST" style="display:inline;">
                    <input type="hidden" name="_method" value="DELETE">
                    <button type="submit">Delete</button>
                </form>
            </td>
        </tr>
        {{end}}
    </tbody>
</table>
<a href="/{{.Name}}/create">Create New</a>
{{end}}
`

const viewShowTemplate = `{{define "content"}}
<h1>{{.Item.Name}}</h1>
<p>ID: {{.Item.ID}}</p>
<a href="/{{.Name}}/{{.Item.ID}}/edit">Edit</a>
<form action="/{{.Name}}/{{.Item.ID}}" method="POST" style="display:inline;">
    <input type="hidden" name="_method" value="DELETE">
    <button type="submit">Delete</button>
</form>
<a href="/{{.Name}}">Back to list</a>
{{end}}