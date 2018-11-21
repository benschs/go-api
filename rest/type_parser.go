package rest

import (
	"fmt"
	"reflect"
)

// ParseType converts the src interface to the type t.
// The returned interface can then be asserted to the
// underlying type t:
//
// Example:
//	v, _ := ParseType(src, reflect.TypeOf(MyStruct{}))
//	myStructValue := v.(MyStruct)
//
// This is meant to be used to convert json to any type.
func ParseType(src interface{}, t reflect.Type) (interface{}, error) {
	v, e := parseTypeToValue(src, t)
	if e != nil {
		return nil, e
	}
	return v.Interface(), nil
}

func parseTypeToValue(src interface{}, t reflect.Type) (*reflect.Value, error) {

	switch t.Kind() {
	case reflect.Struct:
		srcMap, ok := src.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("t of type struct but src not of type map[string]interface{}")
		}
		value, e := parseStruct(srcMap, t)
		return value, e
	case reflect.Slice:
		typeValue := reflect.New(t).Elem()
		srcSlice, ok := src.([]interface{})
		if !ok {
			return nil, fmt.Errorf("type slice expected but got '%s'", reflect.TypeOf(src).Kind())
		}
		for _, v := range srcSlice {
			entryValue, e := parseTypeToValue(v, t.Elem())
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

func parseStruct(src map[string]interface{}, t reflect.Type) (*reflect.Value, error) {

	typeValue := reflect.New(t).Elem()

	// Loop through the struct fields
	for i := 0; i < t.NumField(); i++ {
		// Get the struct field
		field := t.Field(i)

		// Get the value from the source map and check if it is valid.
		srcValue := reflect.ValueOf(src[field.Tag.Get("json")])
		if !srcValue.IsValid() {
			// If it is not valid, ignore it.
			// Most likely the value was not set in the json
			// Could possibly add a check to see if omitempty or
			// some other tag is set the int src
			continue
		}

		// Check if the struct field is a complex field and call the parseType function.
		// => Recursion.
		switch srcValue.Kind() {
		// If struct, convert the src to map[string]interface anf call parseType
		case reflect.Struct:
			// No need for interface type assertion check,
			// as we knwo from teh switch case that conversion is ok
			respValue, e := parseTypeToValue(srcValue.Interface().(map[string]interface{}), field.Type)
			if e != nil {
				return nil, e
			}
			typeValue.Field(i).Set(*respValue)
			continue

		// If slice, convert the src to []interface and call parseType
		case reflect.Slice:
			// No need for interface type assertion check,
			// as we knwo from teh switch case that conversion is ok
			respValue, e := parseTypeToValue(srcValue.Interface().([]interface{}), field.Type)
			if e != nil {
				return nil, e
			}
			typeValue.Field(i).Set(*respValue)
			continue

		// Must be simple type, call parseType directly
		default:
			respValue, e := parseTypeToValue(srcValue.Interface(), field.Type)
			if e != nil {
				return nil, e
			}
			typeValue.Field(i).Set(*respValue)
			continue
		}
	}

	return &typeValue, nil
}
