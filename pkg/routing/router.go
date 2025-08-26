package routing

import (
	"mygola/pkg/gola"
	"net/http"
	"regexp"
)

type MiddlewareFunc func(func(ctx *gola.Context)) func(ctx *gola.Context)

type Route struct {
	method      string
	pattern     *regexp.Regexp
	paramNames  []string
	handler     func(ctx *gola.Context)
	middlewares []MiddlewareFunc
}

type Router struct {
	routes         []Route
	middleware     []MiddlewareFunc
	TemplateEngine *gola.Context // inject template engine to each context

}

func oldNewRouter() *Router {
	return &Router{}
}

func NewRouter(templateEngine *gola.Context) *Router {
	return &Router{
		TemplateEngine: templateEngine,
	}
}

// AddRoute adds a route with pattern
func (r *Router) AddRoute(method string, pattern string, handler func(ctx *gola.Context), middlewares ...MiddlewareFunc) {
	// Convert :param to regex
	paramNames := []string{}
	regexPattern := regexp.MustCompile(`:([a-zA-Z0-9_]+)`)
	replacer := regexPattern.ReplaceAllStringFunc(pattern, func(m string) string {
		paramNames = append(paramNames, m[1:])
		return "([^/]+)"
	})
	regex := regexp.MustCompile("^" + replacer + "$")

	r.routes = append(r.routes, Route{
		method:      method,
		pattern:     regex,
		paramNames:  paramNames,
		handler:     handler,
		middlewares: middlewares,
	})
}

func (r *Router) Get(pattern string, handler func(ctx *gola.Context), middlewares ...MiddlewareFunc) {
	r.AddRoute("GET", pattern, handler, middlewares...)
}

func (r *Router) Post(pattern string, handler func(ctx *gola.Context), middlewares ...MiddlewareFunc) {
	r.AddRoute("POST", pattern, handler, middlewares...)
}

// Use adds global middleware
func (r *Router) Use(middleware MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware)
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	for _, route := range r.routes {
		if route.method != method {
			continue
		}

		matches := route.pattern.FindStringSubmatch(path)
		if matches == nil {
			continue
		}

		// build context
		ctx := &gola.Context{
			Writer:         w,
			Request:        req,
			Params:         map[string]string{},
			TemplateEngine: r.TemplateEngine.TemplateEngine,
		}

		// extract params
		for i, name := range route.paramNames {
			ctx.Params[name] = matches[i+1]
		}

		handler := route.handler

		// apply route middleware
		for i := len(route.middlewares) - 1; i >= 0; i-- {
			handler = route.middlewares[i](handler)
		}

		// apply global middleware
		for i := len(r.middleware) - 1; i >= 0; i-- {
			handler = r.middleware[i](handler)
		}

		handler(ctx)
		return
	}

	http.NotFound(w, req)
}
