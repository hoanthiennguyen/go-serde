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
	destVal := reflect.ValueOf(dest).Elem()
	switch valType.Kind() {
	case reflect.Int:
		val, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}

		destVal.Set(reflect.ValueOf(val))
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}

		if valType.Kind() == reflect.Float32 {
			val32 := float32(val)
			destVal.Set(reflect.ValueOf(val32))
		} else {
			destVal.Set(reflect.ValueOf(val))
		}
	case reflect.String:
		raw = removeQuote(raw)
		destVal.Set(reflect.ValueOf(raw))

	case reflect.Bool:
		val, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(val))

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
		destVal.Set(slice)

	case reflect.Struct:
		raw = removeQuote(raw)
		arrRaw := strings.Split(raw, ",")
		keyValMap := make(map[string]string)
		for _, e := range arrRaw {
			tmp := strings.Split(e, ":")
			key, val := tmp[0], tmp[1]

			key = removeQuote(key)
			keyValMap[key] = val
		}

		for _, f := range reflect.VisibleFields(valType) {
			fName := f.Name
			fTag := f.Tag.Get("json")

			fNameSerialized := fTag
			if fTag == "" {
				fNameSerialized = fName
			}

			fieldValRaw, ok := keyValMap[fNameSerialized]
			if !ok {
				continue
			}

			elemPtr := reflect.New(f.Type).Interface()
			if err := Deserailize(fieldValRaw, elemPtr); err != nil {
				return err
			}

			elemVal := reflect.ValueOf(elemPtr).Elem()
			destVal.FieldByName(fName).Set(elemVal)
		}

	}

	return nil
}

func removeQuote(raw string) string {
	val := raw[1:]
	return val[:len(val)-1]
}
