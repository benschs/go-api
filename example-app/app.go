package main

import (
	"encoding/json"
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

type CustomWriter struct {
}

func (c *CustomWriter) Write(w http.ResponseWriter, content interface{}) error {
	js, err := json.MarshalIndent(content, "", "   ")
	if err != nil {
		http.Error(w, "JSON Error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Expires", "0")
	w.Write(js)
	return nil
}

func createRestfulController() *rest.Controller {

	ctrl := rest.NewController()
	ctrl.AddRequestConfigFromYAML(ROUTES_FILE)

	return ctrl
}
