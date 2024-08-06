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
		arrRaw, err := separateElements(raw)
		if err != nil {
			return err
		}

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

func separateKeyVal(src string) (map[string]string, error) {
	src = removeQuote(src)
	result := make(map[string]string)
	srcLen := len(src)
	var fieldNameToken, fieldValToken *TokenPosition
	var currentState DeserailizeState = StateGetQuoteStartOfFieldName
	var openBracket int
	setEndOfFieldVal := func(index int) {
		fieldValToken.End = index
		fieldValToken.Completed = true
		currentState = StateExpectingComma
	}

	for index := 0; index < srcLen; index++ {
		char := string(src[index])
		switch currentState {
		case StateGetQuoteStartOfFieldName:
			if char == " " {
				continue
			}
			if char == `"` {
				currentState = StateGetEndOfFieldName
				fieldNameToken = &TokenPosition{
					Start: index,
				}
			} else {
				return nil, fmt.Errorf("invalid format state %v: char %s", currentState, char)
			}

		case StateGetEndOfFieldName:
			if char == `"` {
				fieldNameToken.End = index
				fieldNameToken.Completed = true
				currentState = StateWaitingColon
			}

		case StateWaitingColon:
			if char == " " {
				continue
			}
			if char == ":" {
				currentState = StateGettingFieldVal
			} else {
				return nil, fmt.Errorf("invalid format state %v: char %s", currentState, char)
			}

		case StateGettingFieldVal:
			fieldValToken = &TokenPosition{
				Start: index,
			}

			switch char {
			case "[":
				currentState = StateGetClosingSquareBracketFieldVal
				openBracket = 1
			case "{":
				currentState = StateGetClosingCurlyBracketFieldVal
				openBracket = 1
			case `"`:
				currentState = StateGetClosingQuoteFieldVal
			default:
				currentState = StateGetClosingCommaFieldVal
				if index == srcLen-1 {
					fieldValToken.End = index
					fieldValToken.Completed = true
				}
			}

		case StateGetClosingCommaFieldVal:
			if char == "," {
				fieldValToken.End = index - 1
				fieldValToken.Completed = true
				currentState = StateGetQuoteStartOfFieldName
			} else if index == srcLen-1 {
				fieldValToken.End = index
				fieldValToken.Completed = true
			}

		case StateGetClosingSquareBracketFieldVal:
			if char == `]` {
				openBracket--
			} else if char == `[` {
				openBracket++
			}

			if openBracket == 0 {
				setEndOfFieldVal(index)
			}

		case StateGetClosingCurlyBracketFieldVal:
			if char == `}` {
				openBracket--
			} else if char == `{` {
				openBracket++
			}

			if openBracket == 0 {
				setEndOfFieldVal(index)
			}

		case StateGetClosingQuoteFieldVal:
			if char == `"` && string(src[index-1]) != `\` {
				setEndOfFieldVal(index)
			}

		case StateExpectingComma:
			if char == " " {
				continue
			}

			if char == "," {
				currentState = StateGetQuoteStartOfFieldName
			} else {
				return nil, fmt.Errorf("invalid format state %v: char %s", currentState, char)
			}
		}

		if fieldNameToken.IsCompleted() && fieldValToken.IsCompleted() {
			fieldName := src[fieldNameToken.Start+1 : fieldNameToken.End]
			fieldVal := src[fieldValToken.Start : fieldValToken.End+1]
			fieldVal = strings.TrimSpace(fieldVal)
			result[fieldName] = fieldVal

			fieldNameToken = nil
			fieldValToken = nil
		}

	}

	return result, nil
}

type DeserailizeState int

const (
	StateGetEndOfFieldName DeserailizeState = iota
	StateWaitingColon
	StateGettingFieldVal
	StateGetClosingSquareBracketFieldVal
	StateGetClosingCurlyBracketFieldVal
	StateGetClosingQuoteFieldVal
	StateGetClosingCommaFieldVal
	StateGetQuoteStartOfFieldName
	StateExpectingComma
	StateEnd
)

func (s DeserailizeState) String() string {
	switch s {
	case StateGetEndOfFieldName:
		return "StateGetEndOfFieldName"
	case StateWaitingColon:
		return "StateWaitingColon"
	case StateGettingFieldVal:
		return "StateGettingFieldVal"
	case StateGetClosingSquareBracketFieldVal:
		return "StateGetClosingSquareBracketFieldVal"
	case StateGetClosingCurlyBracketFieldVal:
		return "StateGetClosingCurlyBracketFieldVal"
	case StateGetClosingQuoteFieldVal:
		return "StateGetClosingQuoteFieldVal"
	case StateGetClosingCommaFieldVal:
		return "StateGetClosingCommaFieldVal"
	case StateGetQuoteStartOfFieldName:
		return "StateGetQuoteStartOfFieldName"
	case StateExpectingComma:
		return "StateExpectingComma"
	case StateEnd:
		return "StateEnd"
	default:
		return "Unknown"
	}
}

type TokenPosition struct {
	Start     int
	End       int
	Completed bool
}

func (t *TokenPosition) IsCompleted() bool {
	return t != nil && t.Completed
}

func separateElements(src string) ([]string, error) {
	src = removeQuote(src)
	srcLen := len(src)
	var itemToken *TokenPosition

	openBracket := 0
	currentState := StateGettingFieldVal
	setEndOfFieldVal := func(index int) {
		itemToken.End = index
		itemToken.Completed = true
		currentState = StateExpectingComma
	}
	var result []string

	for index := 0; index < srcLen; index++ {
		char := string(src[index])
		switch currentState {
		case StateGettingFieldVal:
			if char == " " {
				continue
			}
			itemToken = &TokenPosition{
				Start: index,
			}
			switch char {
			case "[":
				currentState = StateGetClosingSquareBracketFieldVal
				openBracket = 1
			case "{":
				currentState = StateGetClosingCurlyBracketFieldVal
				openBracket = 1
			case `"`:
				currentState = StateGetClosingQuoteFieldVal
			default:
				currentState = StateGetClosingCommaFieldVal
				if index == srcLen-1 {
					itemToken.End = index
					itemToken.Completed = true
				}
			}
		case StateGetClosingSquareBracketFieldVal:
			if char == `]` {
				openBracket--
			} else if char == `[` {
				openBracket++
			}

			if openBracket == 0 {
				setEndOfFieldVal(index)
			}

		case StateGetClosingCurlyBracketFieldVal:
			if char == `}` {
				openBracket--
			} else if char == `{` {
				openBracket++
			}

			if openBracket == 0 {
				setEndOfFieldVal(index)
			}

		case StateGetClosingQuoteFieldVal:
			if char == `"` && string(src[index-1]) != `\` {
				setEndOfFieldVal(index)
			}

		case StateGetClosingCommaFieldVal:
			if char == "," {
				itemToken.End = index - 1
				itemToken.Completed = true
				currentState = StateGettingFieldVal
			} else if index == srcLen-1 {
				itemToken.End = index
				itemToken.Completed = true
			}
		case StateExpectingComma:
			if char == " " {
				continue
			}

			if char == "," {
				currentState = StateGettingFieldVal
			} else {
				return nil, fmt.Errorf("invalid format state %v: char %s", currentState, char)
			}
		}

		if itemToken.IsCompleted() {
			tokenVal := src[itemToken.Start : itemToken.End+1]
			tokenVal = strings.TrimSpace(tokenVal)
			result = append(result, tokenVal)
			itemToken = nil
		}

	}
	return result, nil
}
