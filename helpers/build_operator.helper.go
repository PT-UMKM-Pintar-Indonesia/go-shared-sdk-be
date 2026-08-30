package sdk_helper

import "reflect"

func Equal[T comparable](a, b T) bool {
	return a == b
}

func NotEqual[T comparable](a, b T) bool {
	return a != b
}

func GetTypeOfStringJSStyle(v any) string {
	if v == nil {
		return "null"
	}

	val := reflect.ValueOf(v)
	kind := val.Kind()

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return "number"

	case reflect.Bool:
		return "boolean"

	case reflect.String:
		return "string"

	case reflect.Func:
		return "function"

	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct, reflect.Ptr, reflect.Chan, reflect.UnsafePointer:
		return "object"

	case reflect.Interface:
		if val.IsNil() {
			return "null"
		}
		return GetTypeOfStringJSStyle(val.Elem().Interface())

	default:
		return "unknown"
	}
}

func IsTypeOf(v any, expectedType string) bool {
	actualType := GetTypeOfStringJSStyle(v)
	return actualType == expectedType
}

func IsPointer(v any) bool {
	if v == nil {
		return false
	}

	return reflect.ValueOf(v).Kind() == reflect.Ptr
}

func IsNil(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {

	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return val.IsNil()

	default:
		return false
	}
}
