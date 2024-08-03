package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Deserailize(raw string, dest any) error {
	valType := reflect.TypeOf(dest)
	if valType.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	valType = valType.Elem()
	switch valType.Kind() {
	case reflect.Int:
		val, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}

		dest = &val
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}

		if valType.Kind() == reflect.Float32 {
			val32 := float32(val)
			dest = &val32
		} else {
			dest = &val
		}
	case reflect.String:
		raw = removeQuote(raw)
		dest = &raw

	case reflect.Bool:
		val, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		dest = &val

	case reflect.Array, reflect.Slice:
		raw = removeQuote(raw)
		arrRaw := strings.Split(raw, ",")
		numElem := len(arrRaw)
		elementType := valType.Elem()
		slice := reflect.MakeSlice(valType, numElem, numElem)
		for j := 0; j < numElem; j++ {
			destElementPtr := reflect.New(elementType).Interface()
			if err := Deserailize(arrRaw[j], destElementPtr); err != nil {
				return err
			}

			slice.Index(j).Set(reflect.ValueOf(destElementPtr).Elem())
		}

	}

	return nil
}

func removeQuote(raw string) string {
	val := raw[1:]
	return val[:len(val)-1]
}
