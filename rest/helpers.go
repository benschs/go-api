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
	valuePointer, e := parseType(body, TypeRegistry[bodyTypeName])
	if e != nil {
		return bodyValue, fmt.Errorf("could not parse json body: %v", e)
	}

	return *valuePointer, nil
}

func parseType(src interface{}, t reflect.Type) (*reflect.Value, error) {

	switch t.Kind() {
	case reflect.Struct:
		value, e := parseStruct(src, t)
		return value, e
	case reflect.Slice:
		typeValue := reflect.New(t).Elem()
		srcSlice, ok := src.([]interface{})
		if !ok {
			return nil, fmt.Errorf("type slice expected but got '%s'", reflect.TypeOf(src).Kind())
		}
		for _, v := range srcSlice {
			entryValue, e := parseType(v, t.Elem())
			if e != nil {
				return nil, e
			}
			typeValue = reflect.Append(typeValue, *entryValue)
		}
		return &typeValue, nil
	default:
		// JSON marshaller assumes all numbers to be float64 if not defined otherwise.
		// Here we check if the value is float64 and if so, we convert it to the correct type.
		//
		// Currently only support for Int and Float types
		srcValueFloat, ok := src.(float64)
		if ok {
			switch t.Kind() {
			case reflect.Float64:
				srcValue := reflect.ValueOf(srcValueFloat)
				return &srcValue, nil
			case reflect.Float32:
				srcValue := reflect.ValueOf(float32(srcValueFloat))
				return &srcValue, nil
			case reflect.Int64:
				srcValue := reflect.ValueOf(int64(srcValueFloat))
				return &srcValue, nil
			case reflect.Int32:
				srcValue := reflect.ValueOf(int32(srcValueFloat))
				return &srcValue, nil
			case reflect.Int16:
				srcValue := reflect.ValueOf(int16(srcValueFloat))
				return &srcValue, nil
			case reflect.Int:
				srcValue := reflect.ValueOf(int(srcValueFloat))
				return &srcValue, nil
			default:
				return nil, fmt.Errorf("Could not parse value %v to type %s", src, t.Kind())
			}
		}

		// Check that the src kind matches the type kind
		if t.Kind() != reflect.Interface && reflect.TypeOf(src).Kind() != t.Kind() {
			return nil, fmt.Errorf("Could not parse value %v (type: %s) to type %s", src, reflect.TypeOf(src).Kind(), t.Kind())
		}

		v := reflect.ValueOf(src)
		return &v, nil
	}
}

func parseStruct(src interface{}, t reflect.Type) (*reflect.Value, error) {
	srcMap, ok := src.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("src not of type map[string]interface{}")
	}

	typeValue := reflect.New(t).Elem()

	// Loop through the struct fields
	for i := 0; i < t.NumField(); i++ {
		// Get the struct field
		field := t.Field(i)

		// Get the value from the source map and check if it is valid.
		srcValue := reflect.ValueOf(srcMap[field.Tag.Get("json")])
		if !srcValue.IsValid() {
			// If it is not valid, ignore it.
			// Most likely the value was not set in the json
			// Could possibly add a check to see if omitempty or
			// some other tag is set the int srcMap
			continue
		}

		// Check if the struct field is a complex field and call the parseType function.
		// => Recursion.
		switch srcValue.Kind() {
		// If struct, convert the src to map[string]interface anf call parseType
		case reflect.Struct:
			// No need for interface type assertion check,
			// as we knwo from teh switch case that conversion is ok
			respValue, e := parseType(srcValue.Interface().(map[string]interface{}), field.Type)
			if e != nil {
				return nil, e
			}
			typeValue.Field(i).Set(*respValue)
			continue

		// If slice, convert the src to []interface and call parseType
		case reflect.Slice:
			// No need for interface type assertion check,
			// as we knwo from teh switch case that conversion is ok
			respValue, e := parseType(srcValue.Interface().([]interface{}), field.Type)
			if e != nil {
				return nil, e
			}
			typeValue.Field(i).Set(*respValue)
			continue

		// Must be simple type, call parseType directly
		default:
			respValue, e := parseType(srcValue.Interface(), field.Type)
			if e != nil {
				return nil, e
			}
			typeValue.Field(i).Set(*respValue)
			continue
		}
	}

	return &typeValue, nil
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
