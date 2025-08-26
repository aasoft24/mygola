package main

import (
	"database/sql"
	"fmt"
	"log"
	"mygola/app/providers"
	"mygola/config"
	"mygola/database"
	"mygola/pkg/cache"
	"mygola/pkg/foundation"
	"mygola/pkg/gola"
	"mygola/pkg/routing"
	"mygola/pkg/session"
	"mygola/pkg/view"
	"net/http"
)

func main() {
	startServer()
}

func startServer() {
	// Load config and DB
	config.LoadConfig("config.yaml")
	database.InitDB()

	// Template engine
	//this for server
	//templateEngine := view.NewTemplateEngine("/www/wwwroot/golang.jionpay.com/resources/views", "app")
	//this only local
	templateEngine := view.NewTemplateEngine("resources/views", "app")
	ctx := &gola.Context{TemplateEngine: templateEngine}
	router := routing.NewRouter(ctx)

	// Session manager
	sessionStore := session.NewMemoryStore()
	sessionManager := session.NewManager(sessionStore, "go_laravel_session")

	// Cache
	appCache := cache.NewMemoryCache()

	// Scheduler
	// scheduler := schedule.NewScheduler()
	// scheduler.Daily(func() {
	// 	log.Println("Running daily task: cleaning up expired sessions")
	// })
	// go scheduler.Start()

	// Application container
	app := foundation.NewApplication()
	app.Bind((*sql.DB)(nil), database.DB)
	app.Bind((*session.Manager)(nil), sessionManager)
	app.Bind((*view.TemplateEngine)(nil), templateEngine)
	app.Bind((*cache.Cache)(nil), appCache)

	// Register route service provider
	app.Register(providers.NewRouteServiceProvider(router, templateEngine))
	app.Boot()

	// Middleware
	router.Use(func(next func(ctx *gola.Context)) func(ctx *gola.Context) {
		return func(ctx *gola.Context) {
			sess, err := sessionManager.Start(ctx.Writer, ctx.Request)
			if err != nil {
				ctx.Error(500, "Internal Server Error")
				return
			}
			next(ctx)
			sess.Save()
		}
	})

	router.Use(func(next func(ctx *gola.Context)) func(ctx *gola.Context) {
		return func(ctx *gola.Context) {
			log.Printf("%s %s", ctx.Request.Method, ctx.Request.URL.Path)
			next(ctx)
		}
	})

	// Register web routes
	//RegisterWebRoutes(router, templateEngine)

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	url := fmt.Sprintf("http://%s", serverAddr)
	log.Printf("🚀 Server running at %s", url)
	log.Fatal(http.ListenAndServe(serverAddr, router))
}
