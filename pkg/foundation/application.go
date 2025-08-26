// vendor/foundation/application.go
package foundation

import (
	"reflect"
)

type Application struct {
	providers []ServiceProvider
	services  map[reflect.Type]interface{}
}

func NewApplication() *Application {
	return &Application{
		services: make(map[reflect.Type]interface{}),
	}
}

func (app *Application) Register(provider ServiceProvider) {
	provider.Register(app)
	app.providers = append(app.providers, provider)
}

func (app *Application) Boot() {
	for _, provider := range app.providers {
		provider.Boot(app)
	}
}

func (app *Application) Bind(abstract interface{}, concrete interface{}) {
	app.services[reflect.TypeOf(abstract)] = concrete
}

func (app *Application) Make(abstract interface{}) interface{} {
	return app.services[reflect.TypeOf(abstract)]
}

type ServiceProvider interface {
	Register(app *Application)
	Boot(app *Application)
}
