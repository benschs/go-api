package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/go-chi/chi"
)

func (c *Controller) internalError(w http.ResponseWriter, e error) {
	c.logger.Println(e.Error())
	fmt.Fprintf(w, "INTERNAL ERROR: %v", e)
	// http.Error(w, "internal server error", http.StatusInternalServerError)
}

func parseURLParams(r *http.Request, params map[string]string) ([]reflect.Value, error) {
	arguments := []reflect.Value{}
	for i, v := range params {
		p := chi.URLParam(r, i)
		if v == "string" {
			arguments = append(arguments, reflect.ValueOf(p))
		}
		if v == "int" {
			pInt, e := strconv.Atoi(p)
			if e != nil {
				return nil, fmt.Errorf("could not parse URL paramter '%s': %v", i, e)
			}
			arguments = append(arguments, reflect.ValueOf(pInt))
		}
	}

	return arguments, nil
}

func parseBodyJSON(r *http.Request, bodyTypeName string) (bodyValue reflect.Value, e error) {
	if TypeRegistry[bodyTypeName] == nil {
		e = fmt.Errorf("json body type '%s' not found in type registry", bodyTypeName)
		return
	}

	// Create value and type of body struct
	bodyValue = reflect.New(TypeRegistry[bodyTypeName]).Elem()

	// Unmarshal any json body to a map
	var body interface{}
	e = unmarshalBodyToJSON(r, &body)
	if e != nil {
		return bodyValue, fmt.Errorf("could not parse json body: %v", e)
	}

	// bodyType := bodyValue.Type()
	valuePointer, e := parseTypeToValue(body, TypeRegistry[bodyTypeName])
	if e != nil {
		return bodyValue, fmt.Errorf("could not parse json body: %v", e)
	}

	return *valuePointer, nil
}

func parseBodyMultipart(r *http.Request, form MultipartForm) (reflect.Value, error) {
	if !form.IsFile {
		// Parse request form
		e := r.ParseMultipartForm(32 << 20)
		if e != nil {
			return reflect.Value{}, fmt.Errorf("could not parse multipart form: %v", e)
		}

		// Get file from request
		v := r.FormValue(form.Name)
		return reflect.ValueOf(v), nil
	}

	fi, e := ParseRequestFormFile(r, form.Name)
	if e != nil {
		return reflect.Value{}, fmt.Errorf("could not parse request form file: %v", e)
	}

	return reflect.ValueOf(fi), nil
}

func unmarshalBodyToJSON(r *http.Request, model interface{}) error {
	buf, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	temp := ioutil.NopCloser(bytes.NewBuffer(buf))
	defer temp.Close()

	err = json.Unmarshal(buf, &model)
	if err != nil {
		return err
	}
	return nil
}

func readFileData(filePath string) []byte {
	var err error

	input := io.ReadCloser(os.Stdin)
	if input, err = os.Open(filePath); err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	data, err := ioutil.ReadAll(input)
	input.Close()
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
	return data
}
