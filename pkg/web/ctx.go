package web

import (
	"net/http"

	"mygola/pkg/views"
)

type Ctx struct {
	W      http.ResponseWriter
	R      *http.Request
	Params map[string]string

	ViewEngine *views.Engine
	Flash      string // example: put flash in ctx
}

func (c *Ctx) Header(k, v string) { c.W.Header().Set(k, v) }
func (c *Ctx) Status(code int)    { c.W.WriteHeader(code) }

func (c *Ctx) View(name string, data any) error {
	// Merge default data into view
	payload := map[string]any{
		"Flash": c.Flash,
	}
	// if user passes map merge; else attach under "Data"
	switch d := data.(type) {
	case map[string]any:
		for k, v := range d {
			payload[k] = v
		}
	default:
		payload["Data"] = data
	}
	return c.ViewEngine.Render(c.W, http.StatusOK, name, payload)
}

func (c *Ctx) ViewError(status int, name string, err error) {
	_ = c.ViewEngine.Render(c.W, status, name, map[string]any{
		"Error": err,
		"Flash": c.Flash,
	})
}
