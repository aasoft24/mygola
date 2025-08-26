package providers

import (
	"mygola/pkg/foundation"
	"mygola/pkg/routing"
	"mygola/pkg/view"
	"mygola/routes"
)

type RouteServiceProvider struct {
	router         *routing.Router
	templateEngine *view.TemplateEngine
}

func NewRouteServiceProvider(router *routing.Router, templateEngine *view.TemplateEngine) *RouteServiceProvider {
	return &RouteServiceProvider{
		router:         router,
		templateEngine: templateEngine,
	}
}

// Bind router to app container
func (p *RouteServiceProvider) Register(app *foundation.Application) {
	app.Bind((*routing.Router)(nil), p.router)
}

// Boot routes
func (p *RouteServiceProvider) Boot(app *foundation.Application) {
	routes.RegisterWebRoutes(p.router, p.templateEngine)
}
