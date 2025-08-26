package view

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------
// Embed templates for deployment
// ----------------------------
//

var templatesFS embed.FS

type TemplateEngine struct {
	templates     map[string]*template.Template
	DefaultLayout string
	viewsPath     string
	useEmbed      bool
}

// ----------------------------
// NewTemplateEngine
// ----------------------------
func NewTemplateEngine(viewsPath, defaultLayout string) *TemplateEngine {
	engine := &TemplateEngine{
		templates:     make(map[string]*template.Template),
		DefaultLayout: defaultLayout,
		viewsPath:     viewsPath,
		useEmbed:      false,
	}

	// check if embedded templates exist
	if _, err := templatesFS.ReadFile("resources/views/" + defaultLayout + ".html"); err == nil {
		engine.useEmbed = true
	}

	engine.loadTemplates()
	return engine
}

// ----------------------------
// Load templates
// ----------------------------
func (e *TemplateEngine) loadTemplates() {
	var partials []string
	if e.useEmbed {
		// collect embedded partials
		entries, _ := templatesFS.ReadDir("resources/views/partials")
		for _, f := range entries {
			if strings.HasSuffix(f.Name(), ".html") {
				partials = append(partials, "resources/views/partials/"+f.Name())
			}
		}
	} else {
		// filesystem partials
		partials, _ = filepath.Glob(filepath.Join(e.viewsPath, "partials", "*.html"))
	}

	if e.useEmbed {
		// read all embed templates
		entries, _ := templatesFS.ReadDir("resources/views")
		for _, f := range entries {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".html") {
				continue
			}
			viewName := strings.TrimSuffix(f.Name(), ".html")
			files := append([]string{"resources/views/" + f.Name()}, partials...)
			tmpl := template.Must(template.ParseFS(templatesFS, files...))
			e.templates[viewName] = tmpl
		}
	} else {
		// filesystem templates
		filepath.Walk(e.viewsPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".html") {
				return nil
			}
			if strings.Contains(path, "partials") {
				return nil
			}
			relPath, _ := filepath.Rel(e.viewsPath, path)
			viewName := strings.TrimSuffix(relPath, filepath.Ext(relPath))

			files := append([]string{path}, partials...)
			layouts, _ := filepath.Glob(filepath.Join(e.viewsPath, "layouts", "*.html"))
			files = append(files, layouts...)

			tmpl := template.Must(template.ParseFiles(files...))
			e.templates[viewName] = tmpl
			return nil
		})
	}
}

// ----------------------------
// Render with default layout
// ----------------------------
func (e *TemplateEngine) Render(w io.Writer, name string, data interface{}) error {
	tmpl, ok := e.templates[name]
	if !ok {
		return fmt.Errorf("template not found: %s", name)
	}
	return tmpl.ExecuteTemplate(w, e.DefaultLayout, data)
}

// Render with specific layout
func (e *TemplateEngine) RenderWithLayout(w io.Writer, viewName, layoutName string, data interface{}) error {
	tmpl, ok := e.templates[viewName]
	if !ok {
		return fmt.Errorf("template not found: %s", viewName)
	}
	layout := layoutName
	if layout == "" {
		layout = e.DefaultLayout
	}
	return tmpl.ExecuteTemplate(w, layout, data)
}

// Render only content (HTMX style)
func (e *TemplateEngine) RenderWithoutLayout(w io.Writer, viewName string, data interface{}) error {
	tmpl, ok := e.templates[viewName]
	if !ok {
		return fmt.Errorf("template not found: %s", viewName)
	}
	return tmpl.ExecuteTemplate(w, "content", data)
}

// package view

// import (
// 	"fmt"
// 	"html/template"
// 	"io"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

// type TemplateEngine struct {
// 	templates     map[string]*template.Template
// 	DefaultLayout string
// 	viewsPath     string
// }

// func NewTemplateEngine(viewsPath, defaultLayout string) *TemplateEngine {
// 	engine := &TemplateEngine{
// 		templates:     make(map[string]*template.Template),
// 		DefaultLayout: defaultLayout,
// 		viewsPath:     viewsPath,
// 	}
// 	engine.loadTemplates()
// 	return engine
// }

// func (e *TemplateEngine) loadTemplates() {
// 	fmt.Println("viewsPath:", e.viewsPath)
// 	partials, _ := filepath.Glob(filepath.Join(e.viewsPath, "partials", "*.html"))

// 	filepath.Walk(e.viewsPath, func(path string, info os.FileInfo, err error) error {
// 		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".html") {
// 			return nil
// 		}

// 		// Skip partials
// 		if strings.Contains(path, "partials") {
// 			return nil
// 		}

// 		relPath, _ := filepath.Rel(e.viewsPath, path)
// 		viewName := strings.TrimSuffix(relPath, filepath.Ext(relPath))

// 		files := append([]string{path}, partials...)

// 		// always include default + other layouts
// 		layouts, _ := filepath.Glob(filepath.Join(e.viewsPath, "layouts", "*.html"))
// 		files = append(files, layouts...)

// 		tmpl := template.Must(template.ParseFiles(files...))
// 		e.templates[viewName] = tmpl
// 		return nil
// 	})
// }

// // Default render → execute default layout block
// func (e *TemplateEngine) Render(w io.Writer, name string, data interface{}) error {
// 	tmpl, ok := e.templates[name]
// 	if !ok {
// 		return fmt.Errorf("template not found: %s", name)
// 	}
// 	return tmpl.ExecuteTemplate(w, e.DefaultLayout, data)
// }

// // Render with specific layout
// func (e *TemplateEngine) RenderWithLayout(w io.Writer, viewName, layoutName string, data interface{}) error {
// 	tmpl, ok := e.templates[viewName]
// 	if !ok {
// 		return fmt.Errorf("template not found: %s", viewName)
// 	}

// 	layout := layoutName
// 	if layout == "" {
// 		layout = e.DefaultLayout
// 	}

// 	// এখানে layoutName আসলে layout.html এ define এর নাম হতে হবে
// 	return tmpl.ExecuteTemplate(w, layout, data)
// }

// // Only content (HTMX style)
// func (e *TemplateEngine) RenderWithoutLayout(w io.Writer, viewName string, data interface{}) error {
// 	tmpl, ok := e.templates[viewName]
// 	if !ok {
// 		return fmt.Errorf("template not found: %s", viewName)
// 	}
// 	return tmpl.ExecuteTemplate(w, "content", data)
// }
