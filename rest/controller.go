package rest

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/go-chi/chi"
)

// Controller handles HTTP requests by parsing request information and passing
// that information to a method of a (business logic) module.
//
// The controller must be provided with a IModule that has methods for the Controller to call.
//
// The controller must first be setup with routes using one of the following
// 		- LoadRequestsJSON(filePath string)
//		- LoadRequestsYAML(filePath string)
//
//
type Controller struct {
	*chi.Mux
	Modules  []IModule
	Requests []Request

	logger *log.Logger

	rw IResponseWriter
}

// NewController creates a new controller instance with default settings
func NewController() *Controller {
	return &Controller{
		Mux:    chi.NewMux(),
		logger: log.New(os.Stderr, "go-api", log.LstdFlags),
		rw:     &StdJSONWriter{},
	}
}

// IModule represents implementations of business logic.
// The methods of IModule will be called by the Controlle.
type IModule interface{}

// SetModule sets a new module that should be called when a request is handled.
func (c *Controller) AddModule(m IModule) {
	c.Modules = append(c.Modules, m)
}

// SetLogger changes the logger used by the controller
func (c *Controller) SetLogger(l *log.Logger) {
	c.logger = l
}

// SetLogger changes the logger used by the controller
func (c *Controller) SetWriter(r IResponseWriter) {
	c.rw = r
}

// AddRoutes returns an HTTP route multiplexer, setup with the routes of the Controller.
func (c *Controller) Routes() *chi.Mux {

	for _, rqst := range c.Requests {
		c.MethodFunc(rqst.Method, rqst.URI, c.HandleRequest(rqst))
	}

	return c.Mux
}

// AddRequestConfigFromJSON reads and unmarshals JSON in the provided file path
// to add route configuration
func (c *Controller) AddRequestConfigFromJSON(filePath string) error {
	data := readFileData(filePath)

	requests := []Request{}
	if err := json.Unmarshal(data, &requests); err != nil {
		return err
	}

	c.Requests = append(c.Requests, requests...)
	return nil
}

// AddRequestConfigFromYAML reads and unmarshals YAML in the provided file path
// to add route configuration
func (c *Controller) AddRequestConfigFromYAML(filePath string) error {
	data := readFileData(filePath)

	requests := []Request{}
	err := yaml.Unmarshal(data, &requests)
	if err != nil {
		return err
	}

	c.Requests = append(c.Requests, requests...)
	return nil
}
