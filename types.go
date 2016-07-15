package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"text/tabwriter"

	"github.com/docker/engine-api/types/swarm"
	"github.com/fatih/color"
)

var red *color.Color
var green *color.Color
var yellow *color.Color
var normal *color.Color

var detail bool

var w *tabwriter.Writer

func init() {
	red = color.New(color.FgRed)
	green = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	normal = color.New(color.FgWhite)
	w = tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
}

func PrintServiceSpecDiff(current, expected swarm.ServiceSpec) {
	_printServiceSpecDiff("", current, expected)
	w.Flush()
}

func _printServiceSpecDiff(namespace string, current, expected interface{}) {
	currentType := reflect.TypeOf(current)
	expectedType := reflect.TypeOf(expected)
	currentValue := reflect.ValueOf(current)
	expectedValue := reflect.ValueOf(expected)

	if currentType != expectedType {
		log.Fatal("Types are different ", currentType, expectedType)
	}

	switch currentType.Kind() {
	case reflect.Array, reflect.Slice:
		c := int(math.Max(float64(currentValue.Len()), float64(expectedValue.Len())))

		for i := 0; i < c; i++ {
			newNamespace := fmt.Sprintf("%s[%d]", namespace, i)
			if i >= currentValue.Len() {
				_printServiceSpecDiff(newNamespace, reflect.Zero(expectedValue.Index(i).Type()).Interface(), expectedValue.Index(i).Interface())
			} else if i >= expectedValue.Len() {
				_printServiceSpecDiff(newNamespace, currentValue.Index(i).Interface(), reflect.Zero(currentValue.Index(i).Type()).Interface())
			} else {
				_printServiceSpecDiff(newNamespace, currentValue.Index(i).Interface(), expectedValue.Index(i).Interface())
			}
		}

	case reflect.Map:
		for _, k := range currentValue.MapKeys() {
			ev := expectedValue.MapIndex(k)
			var expectedKeyValue interface{}
			if ev.IsValid() {
				expectedKeyValue = ev.Interface()
			} else {
				expectedKeyValue = reflect.Zero(currentValue.MapIndex(k).Type()).Interface()
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
				currentKeyValue = reflect.Zero(expectedValue.MapIndex(k).Type()).Interface()
			}
			newNamespace := fmt.Sprintf("%s.%s", namespace, k.Interface())
			_printServiceSpecDiff(newNamespace, currentKeyValue, expectedValue.MapIndex(k).Interface())
		}

	case reflect.Ptr:
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
			if detail {
				fmt.Fprintf(w, "   %s:\t\"%s\" => \"%s\".\n", namespace, sc, se)
			}
		} else if sc == "" {
			fmt.Fprintf(w, "   %s:\t\"\" => \"%s\"\n", namespace, se)
		} else if se == "" {
			fmt.Fprintf(w, "   %s:\t\"%s\" => \"\"\n", namespace, sc)
		} else {
			fmt.Fprintf(w, "   %s:\t\"%s\" => \"%s\"\n", namespace, sc, se)
		}
	}
}
