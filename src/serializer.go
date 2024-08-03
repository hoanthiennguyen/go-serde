package main

import (
	"fmt"
	"reflect"
	"strings"
)

func Serialize(data any) string {
	valType := reflect.TypeOf(data)
	switch valType.Kind() {
	case reflect.Int, reflect.Uint, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprint(data)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprint(data)
	case reflect.String:
		return fmt.Sprintf("\"%s\"", data)
	case reflect.Bool:
		return fmt.Sprint(data)
	case reflect.Array, reflect.Slice:
		arr := []string{}
		for i := 0; i < reflect.ValueOf(data).Len(); i++ {
			val := reflect.ValueOf(data).Index(i).Interface()
			arr = append(arr, Serialize(val))
		}
		return "[" + strings.Join(arr, ",") + "]"

	case reflect.Struct:
		arr := []string{}
		for _, f := range reflect.VisibleFields(valType) {
			fName := f.Name
			fTag := f.Tag.Get("json")

			fNameSerialized := fTag
			if fTag == "" {
				fNameSerialized = fName
			}

			val := reflect.ValueOf(data).FieldByName(fName).Interface()
			arr = append(arr, fmt.Sprintf("\"%s\":%s", fNameSerialized, Serialize(val)))
		}
		return "{" + strings.Join(arr, ",") + "}"

	case reflect.Pointer:
		val := reflect.ValueOf(data).Elem().Interface()
		return Serialize(val)
	}

	return "unknown"
}
