package rest

import (
	"reflect"
)

var TypeRegistry = make(map[string]reflect.Type)

// AddTypeToRegistry adds the provided object to the types map.
// Map key will be the name of the type, e.g. 'package.Object'
func AddTypeToRegistry(object interface{}) {
	TypeRegistry[reflect.TypeOf(object).String()] = reflect.TypeOf(object)
}

// AddTypeAndNameToRegistry adds the provided object to the types map.
// Map key will be the name provided.
func AddTypeAndNameToRegistry(object interface{}, name string) {
	TypeRegistry[name] = reflect.TypeOf(object)
}
