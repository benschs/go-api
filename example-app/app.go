package main

import (
	"fmt"
	"net/http"

	"github.com/benschs/go-api/rest"
	"github.com/go-chi/chi/middleware"
)

const (
	ADDRESS     = "localhost:8080"
	ROUTES_FILE = "routes.yaml"
)

func main() {
	controller := createRestfulController()
	controller.Use(middleware.Recoverer)

	cWriter := &CustomWriter{}
	controller.SetWriter(cWriter)

	businessLogicImplementation := NewModule()
	controller.AddModule(businessLogicImplementation)

	srv := http.Server{
		Addr:    ADDRESS,
		Handler: controller.Routes(),
	}

	fmt.Printf("Listening on %s\n", ADDRESS)
	srv.ListenAndServe()
}

func createRestfulController() *rest.Controller {

	ctrl := rest.NewController()
	ctrl.AddRequestConfigFromYAML(ROUTES_FILE)

	return ctrl
}
