package rest

import (
	"fmt"
	"net/http"
	"reflect"
)

// HandleRequest is the function called by all routes that are defined.
//
// It will first check to see if the route method is found in the module.
// It then proceeds to parse headers, any URL parameters, query parameters and body
// The parsed values will then be passed as arguments to the method in the following order:
//		1. Headers
// 		2. URL parameters, each as single argument, type according to configurations.
// 		3. Query parameters as a map[string]string
//		4. Body as struct (if configured as json) or as
//			parameters in the order they are defined in the configuration (typed as either FileInfo or string).
//
// Two responses from the method call are expected: structure for response and an error.
// If the error is nil, the response will be sent as JSON.
func (c *Controller) HandleRequest(request Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mod, fnValue reflect.Value

		for _, module := range c.Modules {
			// Get modules method by name
			mod = reflect.ValueOf(module)
			if !mod.IsValid() {
				c.internalError(w, fmt.Errorf("controller module not set"))
				return
			}
			fnValue = mod.MethodByName(request.Func)
			if !fnValue.IsValid() {
				c.internalError(w, fmt.Errorf("method '%s' not found", request.Func))
				return
			}
		}

		// Make an empty argument list.
		arguments := []reflect.Value{}

		// Parse headers, if wanted by the request
		if request.Headers != nil {
			headers := make(map[string]string)
			for _, name := range request.Headers {
				headers[name] = r.Header.Get(name)
			}
			arguments = append(arguments, reflect.ValueOf(headers))
		}

		// Parse URL parameters if any available
		if request.Params != nil {
			params, e := parseURLParams(r, request.Params)
			if e != nil {
				c.internalError(w, e)
				return
			}
			arguments = append(arguments, params...)
		}

		// Parse query parameters if any available
		if request.Query != nil {
			var queryParams map[string]string
			for _, key := range request.Query {
				queryParams[key] = r.URL.Query().Get(key)
			}
			arguments = append(arguments, reflect.ValueOf(queryParams))
		}

		// Parse JSON Body
		if request.Body.IsJSON {
			v, e := parseBodyJSON(r, request.Body.JSONStructName)
			if e != nil {
				c.internalError(w, e)
				return
			}
			arguments = append(arguments, v)
		}

		// Parse Multipart Form Body
		if request.Body.IsMultipart {
			for _, form := range request.Body.Forms {
				v, e := parseBodyMultipart(r, form)
				if e != nil {
					c.internalError(w, e)
					return
				}
				arguments = append(arguments, v)
			}
		}

		// Call module function
		fnResults := executeModuleCall(fnValue, arguments)

		// Check response from module.
		// Expect to have to values, response and error
		if len(fnResults) != 2 {
			c.internalError(w, fmt.Errorf("module function reponse does not contain two arguments"))
			return
		}

		// Get results
		result := fnResults[0].Interface()
		resultError := fnResults[1].Interface()

		// Respond to the HTTP request
		if resultError != nil {
			var e error
			var ok bool
			if e, ok = resultError.(error); !ok {
				e = fmt.Errorf("%s", resultError)
			}

			c.logger.Printf("module error response to '%s': %v\n", request.Name, e)

			c.rw.WriteError(w, e)
			return
		}

		c.rw.Write(w, result)
	}
}

// executeModuleCall wraps the module method call in a function to be able to recover from a panic.
// This is due to the implementation of the go reflect package.
func executeModuleCall(v reflect.Value, args []reflect.Value) []reflect.Value {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	return v.Call(args)
}
