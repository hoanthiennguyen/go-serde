package main

import (
	"fmt"
	"reflect"
	"strconv"
)

func Deserailize(raw string, dest any) error {
	valType := reflect.TypeOf(dest)
	if valType.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	valType = valType.Elem()
	destVal := reflect.ValueOf(dest).Elem()
	switch valType.Kind() {
	case reflect.Pointer:
		if err := Deserailize(raw, destVal.Interface()); err != nil {
			return err
		}

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
		arrRaw, err := separateElements(raw)
		if err != nil {
			return err
		}

		numElem := len(arrRaw)
		elementType := valType.Elem()
		elementType, layers := unwrapPointer(elementType)

		slice := reflect.MakeSlice(valType, numElem, numElem)
		for j := 0; j < numElem; j++ {
			destElementPtr := reflect.New(elementType).Interface()
			if err := Deserailize(arrRaw[j], destElementPtr); err != nil {
				return err
			}

			elementVal := reflect.ValueOf(destElementPtr).Elem()
			elementVal = wrapPointer(elementVal, layers)
			slice.Index(j).Set(elementVal)

		}
		destVal.Set(slice)

	case reflect.Struct:
		keyValMap, err := separateKeyVal(raw)
		if err != nil {
			return err
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

			elementType := f.Type
			elementType, layers := unwrapPointer(elementType)

			elemPtr := reflect.New(elementType).Interface()
			if err := Deserailize(fieldValRaw, elemPtr); err != nil {
				return err
			}

			elementVal := reflect.ValueOf(elemPtr).Elem()
			elementVal = wrapPointer(elementVal, layers)

			destVal.FieldByName(fName).Set(elementVal)
		}

	}

	return nil
}

func removeQuote(raw string) string {
	val := raw[1:]
	return val[:len(val)-1]
}

func unwrapPointer(elementType reflect.Type) (reflect.Type, int) {
	layers := 0
	for elementType.Kind() == reflect.Pointer {
		elementType = elementType.Elem()
		layers++
	}

	return elementType, layers
}

func wrapPointer(val reflect.Value, layers int) reflect.Value {
	ptrVal := val
	for index := 0; index < layers; index++ {
		ptrVal = reflect.New(val.Type()).Elem()
		ptrVal.Set(val)
		ptrVal = ptrVal.Addr()
		val = ptrVal
	}

	return ptrVal
}
