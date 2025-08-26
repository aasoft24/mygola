package controllers

import (
	"mygola/app/models"
	"mygola/pkg/gola"
	"mygola/pkg/view"
)

type BlogController struct {
	templateEngine *view.TemplateEngine
}

func NewBlogController(templateEngine *view.TemplateEngine) *BlogController {
	return &BlogController{templateEngine: templateEngine}
}

func (c *BlogController) Index(ctx *gola.Context) {
	posts := []models.Post{
		{ID: "1", Title: "First Post", Content: "This is the first post"},
		{ID: "2", Title: "Second Post", Content: "This is the second post"},
	}

	data := map[string]interface{}{
		"Title": "Blog",
		"Posts": posts,
	}
	ctx.JSON(200, data)

	//c.templateEngine.Render(ctx.Writer, "blog/index", data)
}

func (c *BlogController) Show(ctx *gola.Context) {
	// TODO: extract post ID from URL (e.g., with regex group or router params)
	// For now, just using dummy data
	post := models.Post{ID: "1", Title: "First Post", Content: "This is the first post"}

	data := map[string]interface{}{
		"Title": post.Title,
		"Post":  post,
	}

	ctx.Render(200, "blog/show", data)

}
