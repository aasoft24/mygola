package controllers

import (
	"mygola/pkg/gola"
)

type HomeController struct{}

func NewHomeController() *HomeController {
	return &HomeController{}
}

// GET /
func (c *HomeController) Index(ctx *gola.Context) {

	ctx.View("home", map[string]interface{}{
		"title":  "Homepage",
		"msg":    "Welcome to the homepage!",
		"footer": "Homepage footer",
	}, "app")

	// ctx.HTML(200, "<h1>Welcome to the homepage!</h1>")
}

func (uc *HomeController) About(ctx *gola.Context) {

	data := map[string]any{
		"msg":   "Users list meber about",
		"title": "About",
	}
	ctx.View("sayed", data, "member")
	// ctx.View("dashboard/index.blade", map[string]interface{}{
	// 	"UserName":  "Sayed",
	// 	"title":     "About page",
	// 	"getmsg":    "Welcome to the about page!",
	// 	"getfooter": "About footer",

	// 	"Now":            time.Now().Format("15:04:05"),
	// 	"FlashMessage":   "Welcome aboute!",
	// 	"SuccessMessage": "Success message",
	// })
}

// GET /about
// func (c *HomeController) About(ctx *gola.Context) {
// 	data := map[string]interface{}{
// 		"message": "About page",
// 		"title":   "About",
// 		"content": "This is the about page",
// 		"footer":  "About footer",
// 	}

// 	ctx.JSON(200, data)
// }

// GET /show
func (c *HomeController) Show(ctx *gola.Context) {
	ctx.String(200, "Show page with plain text")
}

// GET /show
func (c *HomeController) Test(ctx *gola.Context) {
	ctx.String(200, "Show page with plain text")
}
