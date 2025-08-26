// In pkg/gola/context.go
package gola

import (
	"encoding/json"
	"fmt"
	"mygola/pkg/view"
	"net/http"
)

type Context struct {
	Writer         http.ResponseWriter
	Request        *http.Request
	Params         map[string]string
	TemplateEngine *view.TemplateEngine
	Flash          string
	FlashType      string
}

// JSON response
func (c *Context) JSON(code int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(code)
	_ = json.NewEncoder(c.Writer).Encode(data)
}

// HTML string
func (c *Context) HTML(status int, html string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, _ = c.Writer.Write([]byte(html))
}

// Param gets route parameter
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Error sends error response
func (c *Context) Error(code int, msg string) {
	http.Error(c.Writer, msg, code)
}

// String sends plain text
func (c *Context) String(status int, message string) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, _ = c.Writer.Write([]byte(message))
}

// Render renders a template with optional layout
func (c *Context) Render(status int, view string, data interface{}, layout ...string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)

	if c.TemplateEngine == nil {
		http.Error(c.Writer, "Template engine not configured", http.StatusInternalServerError)
		return
	}

	useLayout := ""
	if len(layout) > 0 {
		useLayout = layout[0]
	}

	err := c.TemplateEngine.RenderWithLayout(c.Writer, view, useLayout, data)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// View renders a template with default data merging
func (c *Context) View(name string, data interface{}, layout ...string) error {
	if c.TemplateEngine == nil {
		return fmt.Errorf("template engine not configured")
	}

	payload := map[string]any{
		"Flash":     c.Flash,
		"FlashType": c.FlashType,
	}

	switch d := data.(type) {
	case map[string]any:
		for k, v := range d {
			payload[k] = v
		}
	default:
		payload["Data"] = data
	}

	useLayout := ""
	if len(layout) > 0 {
		useLayout = layout[0]
	}

	fmt.Println("ltss ", useLayout)

	if useLayout == "remove" {
		return c.TemplateEngine.RenderWithoutLayout(c.Writer, name, payload)
	}
	//RenderWithoutLayout(c.Writer, name, payload)

	if useLayout != "" {
		return c.TemplateEngine.RenderWithLayout(c.Writer, name, useLayout, payload)
	}
	return c.TemplateEngine.Render(c.Writer, name, payload)
}

func (c *Context) RenderPartial(status int, view string, data interface{}) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)

	if c.TemplateEngine == nil {
		http.Error(c.Writer, "Template engine not configured", http.StatusInternalServerError)
		return
	}

	err := c.TemplateEngine.RenderWithoutLayout(c.Writer, view, data)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}
