// internal/routes/web.go
package routes

import (
	"mygola/app/http/controllers"
	"mygola/pkg/routing"
	"mygola/pkg/view"
	"mygola/pkg/web"
)

func RegisterRoutes(r *web.Router) {

	// hController := controllers.NewHomeController()
	// r.Get("/about", hController.About)

}

func RegisterWebRoutes(router *routing.Router, templateEngine *view.TemplateEngine) {
	homeController := controllers.NewHomeController()

	// Get template engine from container (in a real app, you'd use dependency injection)
	//templateEngine := view.NewTemplateEngine("views", "app")
	blogController := controllers.NewBlogController(templateEngine)

	router.Get("/", homeController.Index)
	router.Get("/about", homeController.About)
	router.Get("/show", homeController.Show)
	router.Get("/test", homeController.Test)

	// Blog routes
	router.Get("/blog", blogController.Index)
	router.Get("/blog/:id", blogController.Show)
}
