package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/benschs/go-api/rest"
	"github.com/go-chi/chi/middleware"
)

const (
	ADDRESS     = "localhost:8080"
	ROUTES_FILE = "routes.yaml"
)

type Test struct {
	Name string `json:"name"`
}

func main() {

	jsonString := `{
	"name": "test"
}`

	var body interface{}
	e := json.Unmarshal([]byte(jsonString), &body)
	if e != nil {
		log.Fatal(e)
	}

	testBody := Test{Name: "test"}

	v, e := rest.ParseType(testBody, reflect.TypeOf(Test{}))
	if e != nil {
		log.Fatal(e)
	}

	fmt.Println(v.(Test))

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
