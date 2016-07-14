package main

import (
	"fmt"
	"log"
	"math"
	"reflect"

	"github.com/fatih/color"
)

var red *color.Color
var green *color.Color
var yellow *color.Color
var normal *color.Color

func init() {
	red = color.New(color.FgRed)
	green = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	normal = color.New(color.FgWhite)
}

func PrintServiceSpecDiff(current, expected interface{}) {
	_printServiceSpecDiff("", current, expected)
}

func _printServiceSpecDiff(namespace string, current, expected interface{}) {
	currentType := reflect.TypeOf(current)
	expectedType := reflect.TypeOf(expected)

	if currentType != expectedType {
		log.Fatal("Types are different ", currentType, expectedType)
	}

	switch currentType.Kind() {
	case reflect.Array, reflect.Slice:
		currentValue := reflect.ValueOf(current)
		expectedValue := reflect.ValueOf(expected)

		c := int(math.Max(float64(currentValue.Len()), float64(expectedValue.Len())))

		for i := 0; i < c; i++ {
			newNamespace := fmt.Sprintf("%s[%d]", namespace, i)
			if i >= currentValue.Len() {
				_printServiceSpecDiff(newNamespace, reflect.Indirect(reflect.New(expectedValue.Index(i).Type())).Interface(), expectedValue.Index(i).Interface())
			} else if i >= expectedValue.Len() {
				_printServiceSpecDiff(newNamespace, currentValue.Index(i).Interface(), reflect.Indirect(reflect.New(currentValue.Index(i).Type())).Interface())
			} else {
				_printServiceSpecDiff(newNamespace, currentValue.Index(i).Interface(), expectedValue.Index(i).Interface())
			}
		}
	case reflect.Map:
		currentValue := reflect.ValueOf(current)
		expectedValue := reflect.ValueOf(expected)

		for _, k := range currentValue.MapKeys() {
			ev := expectedValue.MapIndex(k)
			var expectedKeyValue interface{}
			if ev.IsValid() {
				expectedKeyValue = ev.Interface()
			} else {
				expectedKeyValue = reflect.Indirect(reflect.New(currentValue.MapIndex(k).Type())).Interface()
			}
			newNamespace := fmt.Sprintf("%s.%s", namespace, k.Interface())
			_printServiceSpecDiff(newNamespace, currentValue.MapIndex(k).Interface(), expectedKeyValue)
		}

		for _, k := range expectedValue.MapKeys() {
			cv := currentValue.MapIndex(k)
			var currentKeyValue interface{}
			if cv.IsValid() {
				continue
			} else {
				currentKeyValue = reflect.Indirect(reflect.New(expectedValue.MapIndex(k).Type())).Interface()
			}
			newNamespace := fmt.Sprintf("%s.%s", namespace, k.Interface())
			_printServiceSpecDiff(newNamespace, currentKeyValue, expectedValue.MapIndex(k).Interface())
		}
	case reflect.Ptr:
		currentValue := reflect.ValueOf(current)
		expectedValue := reflect.ValueOf(expected)

		var dcv interface{}
		var dev interface{}

		if currentValue.IsNil() {
			dcv = reflect.Zero(currentType.Elem()).Interface()
		} else {
			dcv = reflect.Indirect(currentValue).Interface()
		}

		if expectedValue.IsNil() {
			dev = reflect.Zero(expectedType.Elem()).Interface()
		} else {
			dev = reflect.Indirect(expectedValue).Interface()
		}

		_printServiceSpecDiff(namespace, dcv, dev)
	case reflect.Struct:
		currentValue := reflect.ValueOf(current)
		expectedValue := reflect.ValueOf(expected)

		for i := 0; i < currentValue.NumField(); i++ {
			f := currentValue.Type().Field(i)
			if f.PkgPath == "" {
				newNamespace := fmt.Sprintf("%s.%s", namespace, f.Name)
				_printServiceSpecDiff(newNamespace, currentValue.Field(i).Interface(), expectedValue.Field(i).Interface())
			}
		}
	default:
		sc := fmt.Sprint(current)
		se := fmt.Sprint(expected)

		if sc == se {
			normal.Printf("  %s:\t\t\t\t\"%s\" => \"%s\"\n", namespace, sc, se)
		} else if sc == "" {
			green.Printf("  %s:\t\t\t\t\"\" => \"%s\"\n", namespace, se)
		} else if se == "" {
			red.Printf("  %s:\t\t\t\t\"%s\" => \"\"\n", namespace, sc)
		} else {
			yellow.Printf("  %s:\t\t\t\t\"%s\" => \"%s\"\n", namespace, sc, se)
		}
	}
}
