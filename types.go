package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"reflect"
	"strings"

	"github.com/docker/docker/api/client/bundlefile"
	"github.com/docker/docker/api/types/swarm"
	"github.com/fatih/color"
)

var yellow = color.New(color.FgYellow)

type Stack struct {
	Name   string
	Bundle *bundlefile.Bundlefile
}

type ServicePrinter struct {
	w           io.Writer
	detail      bool
	isDifferent bool
}

func NewServicePrinter(w io.Writer, detail bool) *ServicePrinter {
	return &ServicePrinter{w: w, detail: detail}
}

func (sp *ServicePrinter) PrintServiceSpec(spec swarm.ServiceSpec) {
	sp.isDifferent = false
	sp._printServiceSpec("", spec)
}

func (sp *ServicePrinter) _printServiceSpec(namespace string, current interface{}) {
	currentType := reflect.TypeOf(current)
	currentValue := reflect.ValueOf(current)
	switch currentType.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < currentValue.Len(); i++ {
			newNamespace := fmt.Sprintf("%s[%d]", namespace, i)
			sp._printServiceSpec(newNamespace, currentValue.Index(i).Interface())
		}
	case reflect.Map:
		for _, k := range currentValue.MapKeys() {
			newNamespace := fmt.Sprintf("%s.%s", namespace, k.Interface())
			sp._printServiceSpec(newNamespace, currentValue.MapIndex(k).Interface())
		}
	case reflect.Ptr:
		if !currentValue.IsNil() {
			sp._printServiceSpec(namespace, reflect.Indirect(currentValue).Interface())
		}
	case reflect.Struct:
		for i := 0; i < currentValue.NumField(); i++ {
			f := currentValue.Type().Field(i)
			if f.PkgPath == "" {
				newNamespace := fmt.Sprintf("%s.%s", namespace, f.Name)
				sp._printServiceSpec(newNamespace, currentValue.Field(i).Interface())
			}
		}
	default:
		sc := fmt.Sprint(current)
		sp.println(nil, namespace, sc)
	}
}

func (sp *ServicePrinter) PrintServiceSpecDiff(current, expected swarm.ServiceSpec) bool {
	sp.isDifferent = false
	sp._printServiceSpecDiff("", current, expected)
	return sp.isDifferent
}

func (sp *ServicePrinter) _printServiceSpecDiff(namespace string, current, expected interface{}) {
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
				sp._printServiceSpecDiff(newNamespace, reflect.Zero(expectedValue.Index(i).Type()).Interface(), expectedValue.Index(i).Interface())
			} else if i >= expectedValue.Len() {
				sp._printServiceSpecDiff(newNamespace, currentValue.Index(i).Interface(), reflect.Zero(currentValue.Index(i).Type()).Interface())
			} else if i >= expectedValue.Len() {
				sp._printServiceSpecDiff(newNamespace, currentValue.Index(i).Interface(), reflect.Indirect(reflect.New(currentValue.Index(i).Type())).Interface())
			} else {
				sp._printServiceSpecDiff(newNamespace, currentValue.Index(i).Interface(), expectedValue.Index(i).Interface())
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
			sp._printServiceSpecDiff(newNamespace, currentValue.MapIndex(k).Interface(), expectedKeyValue)
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
			sp._printServiceSpecDiff(newNamespace, currentKeyValue, expectedValue.MapIndex(k).Interface())
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

		sp._printServiceSpecDiff(namespace, dcv, dev)

	case reflect.Struct:
		for i := 0; i < currentValue.NumField(); i++ {
			f := currentValue.Type().Field(i)
			if f.PkgPath == "" {
				newNamespace := fmt.Sprintf("%s.%s", namespace, f.Name)
				sp._printServiceSpecDiff(newNamespace, currentValue.Field(i).Interface(), expectedValue.Field(i).Interface())
			}
		}
	default:
		sc := fmt.Sprint(current)
		se := fmt.Sprint(expected)

		if sc == se {
			if sp.detail {
				sp.printDiffln(nil, namespace, sc, se)
			}
		} else {
			var c *color.Color
			if sp.detail {
				c = yellow
			}

			sp.isDifferent = true
			sp.printDiffln(c, namespace, sc, se)
		}
	}
}

func (sp *ServicePrinter) println(c *color.Color, namespace, current string) {
	spaces := 50 - len(namespace)
	spaceString := strings.Repeat(" ", spaces)
	if c != nil {
		namespace = c.SprintFunc()(namespace)
		current = c.SprintFunc()(current)
	}
	fmt.Fprintf(sp.w, "   %s:%s\"%s\"\n", namespace, spaceString, current)
}
func (sp *ServicePrinter) printDiffln(c *color.Color, namespace, current, expected string) {
	action := "=>"
	spaces := 50 - len(namespace)
	spaceString := strings.Repeat(" ", spaces)
	if c != nil {
		namespace = c.SprintFunc()(namespace)
		current = c.SprintFunc()(current)
		expected = c.SprintFunc()(expected)
		action = c.SprintFunc()(action)
	}
	fmt.Fprintf(sp.w, "   %s:%s\"%s\" %s \"%s\"\n", namespace, spaceString, current, action, expected)
}
