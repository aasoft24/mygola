// cmd/make/main.go
package main

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type TemplateData struct {
	Name      string
	TableName string
	Timestamp string
	Package   string
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "controller":
		if len(args) < 1 {
			log.Fatal("Controller name is required")
		}
		name := args[0]
		// Resource flag
		isResource := getOption(args, "-r", false) || getOption(args, "--resource", false)
		makeController(name, isResource)

		// Model + Migration flag
		if getOption(args, "-m", false) || getOption(args, "--model", false) {
			createMigration := getOption(args, "-m", false)
			makeModel(args[0], createMigration)
		}

	case "model":
		if len(args) < 1 {
			log.Fatal("Model name is required")
		}
		createMigration := getOption(args, "-m", false)
		makeModel(args[0], createMigration)

	case "migration":
		if len(args) < 1 {
			log.Fatal("Migration name is required")
		}
		makeMigration(args[0])

	case "middleware":
		if len(args) < 1 {
			log.Fatal("Middleware name is required")
		}
		makeMiddleware(args[0])

	case "provider":
		if len(args) < 1 {
			log.Fatal("Provider name is required")
		}
		makeProvider(args[0])

	case "request":
		if len(args) < 1 {
			log.Fatal("Request name is required")
		}
		makeRequest(args[0])

	case "seed":
		if len(args) < 1 {
			log.Fatal("Seed name is required")
		}
		makeSeed(args[0])

	case "view":
		if len(args) < 1 {
			log.Fatal("View name is required")
		}
		makeView(args[0])

	case "help", "--help", "-h":
		printHelp()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: make <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  controller <name> [--resource]  Create a new controller")
	fmt.Println("  model <name>                    Create a new model")
	fmt.Println("  migration <name>                Create a new migration")
	fmt.Println("  middleware <name>               Create a new middleware")
	fmt.Println("  provider <name>                 Create a new service provider")
	fmt.Println("  request <name>                  Create a new form request")
	fmt.Println("  seed <name>                     Create a new database seeder")
	fmt.Println("  view <name>                     Create a new view")
	fmt.Println("  help                            Show this help message")
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

	fmt.Printf("Controller created: %s\n", path)
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
	fmt.Printf("Model created: %s\n", path)

	// If -m flag is passed, create migration
	if createMigration {
		makeMigration(name)
	}
}

// func makeModel(name string) {
// 	data := TemplateData{
// 		Name:      toCamelCase(name),
// 		TableName: toSnakeCase(name),
// 		Package:   "models",
// 	}

// 	// Create directory if it doesn't exist
// 	dir := "app/models"
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		log.Fatal("Failed to create models directory:", err)
// 	}

// 	// Create file
// 	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
// 	path := filepath.Join(dir, filename)

// 	if err := createFileFromTemplate(path, modelTemplate, data); err != nil {
// 		log.Fatal("Failed to create model:", err)
// 	}

// 	fmt.Printf("Model created: %s\n", path)
// }

// Update the makeMigration function in cmd/make/main.go
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
CREATE TABLE example_table (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
`, name, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		log.Fatal("Failed to create up migration:", err)
	}

	// Create down migration file
	downFilename := fmt.Sprintf("%s_down.sql", migrationName)
	downPath := filepath.Join(dir, downFilename)

	downContent := fmt.Sprintf(`-- Rollback migration: %s
-- Created at: %s

-- Add your rollback SQL here
DROP TABLE IF EXISTS example_table;
`, name, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		log.Fatal("Failed to create down migration:", err)
	}

	fmt.Printf("Migration created:\n  Up: %s\n  Down: %s\n", upPath, downPath)
}

// func oldmakeMigration(name string) {
// 	timestamp := time.Now().Format("20060102150405")
// 	data := TemplateData{
// 		Name:      toCamelCase(name),
// 		TableName: toSnakeCase(name),
// 		Timestamp: timestamp,
// 		Package:   "migrations",
// 	}

// 	// Create directory if it doesn't exist
// 	dir := "database/migrations"
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		log.Fatal("Failed to create migrations directory:", err)
// 	}

// 	// Create file
// 	filename := fmt.Sprintf("%s_%s.go", timestamp, strings.ToLower(name))
// 	path := filepath.Join(dir, filename)

// 	if err := createFileFromTemplate(path, migrationTemplate, data); err != nil {
// 		log.Fatal("Failed to create migration:", err)
// 	}

// 	fmt.Printf("Migration created: %s\n", path)
// }

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

	fmt.Printf("Middleware created: %s\n", path)
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

	fmt.Printf("Provider created: %s\n", path)
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

	fmt.Printf("Request created: %s\n", path)
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

	fmt.Printf("Seed created: %s\n", path)
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

	fmt.Printf("Views created in: %s\n", dir)
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
	if err := tmpl.Execute(file, data); err != nil {
		return err
	}

	// Format the Go code
	formatted, err := formatSourceFile(path)
	if err != nil {
		// If formatting fails, it's not a critical error
		fmt.Printf("Warning: Could not format Go file: %v\n", err)
		return nil
	}

	// Write formatted code back to file
	return os.WriteFile(path, formatted, 0644)
}

func getOption(args []string, option string, defaultValue bool) bool {
	for _, arg := range args {
		if arg == option {
			return true
		}
	}
	return defaultValue
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

func snakeCase(name string) string {
	result := ""
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += string(r)
	}
	return stringLower(result)
}

func stringLower(s string) string {
	out := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			out += string(r + 32) // ছোট হাতের অক্ষর বানাই
		} else {
			out += string(r)
		}
	}
	return out
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

// Helper function to format Go source files
func formatSourceFile(filename string) ([]byte, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(src)
	if err != nil {
		return nil, err
	}

	return formatted, nil
}

// Template definitions
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

const migrationTemplate = `package {{.Package}}

type {{.Name}} struct {
	name string
}

func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{
		name: "{{.Timestamp}}_{{.Name}}",
	}
}

func (m *{{.Name}}) GetName() string {
	return m.name
}

func (m *{{.Name}}) Up() string {
	return ` + "`" + `CREATE TABLE {{.TableName}} (
		id VARCHAR(36) PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)` + "`" + `
}

func (m *{{.Name}}) Down() string {
	return "DROP TABLE {{.TableName}}"
}
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
	"go-laravel/pkg/foundation"
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
`
