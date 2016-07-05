package main

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"strings"

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

func print(c *color.Color, depth int, field, msg string) {
	if field != "" {
		c.Printf("%s%s: %s", strings.Repeat(" ", depth*2), field, msg)
	} else {
		c.Printf("%s%s", strings.Repeat(" ", depth*2), msg)
	}
}
func PrintServiceSpecDiff(current, expected interface{}) {
	_printServiceSpecDiff(0, "", current, expected)
}

func _printServiceSpecDiff(depth int, field string, current, expected interface{}) {
	depth++
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

		print(normal, depth, field, "[\n")
		for i := 0; i < c; i++ {
			if i >= currentValue.Len() {
				_printServiceSpecDiff(depth, "", reflect.Indirect(reflect.New(expectedValue.Index(i).Type())).Interface(), expectedValue.Index(i).Interface())
			} else if i >= expectedValue.Len() {
				_printServiceSpecDiff(depth, "", currentValue.Index(i).Interface(), reflect.Indirect(reflect.New(currentValue.Index(i).Type())).Interface())
			} else {
				_printServiceSpecDiff(depth, "", currentValue.Index(i).Interface(), expectedValue.Index(i).Interface())
			}
		}
		print(normal, depth, "", "]\n")
	case reflect.Map:
	case reflect.Ptr:
	case reflect.Struct:
		currentValue := reflect.ValueOf(current)
		expectedValue := reflect.ValueOf(expected)

		print(normal, depth, field, "{\n")
		for i := 0; i < currentValue.NumField(); i++ {
			field = currentValue.Type().Field(i).Name
			_printServiceSpecDiff(depth, field, currentValue.Field(i).Interface(), expectedValue.Field(i).Interface())
		}
		print(normal, depth, "", "}\n")
	default:
		sc := fmt.Sprintf("%s", current)
		se := fmt.Sprintf("%s", expected)

		if sc == se {
			print(normal, depth, field, fmt.Sprintf(`    "%s" => "%s"`, sc, se))
		} else if sc == "" {
			print(green, depth, field, fmt.Sprintf(`+   "%s"`, se))
		} else if se == "" {
			print(red, depth, field, fmt.Sprintf(`-   "%s"`, sc))
		} else {
			print(yellow, depth, field, fmt.Sprintf(`+/- "%s" => "%s"`, sc, se))
		}
		fmt.Print("\n")
	}
	/*
		Image string
		Labels map[string]string
		Command string
		Args []string
		Env             []string          `json:",omitempty"`
		Dir             string            `json:",omitempty"`
		User            string            `json:",omitempty"`
		Mounts          []Mount           `json:",omitempty"`
		StopGracePeriod *time.Duration    `json:",omitempty"`
	*/
}

/*
db
 +/- port: 5000 => 6000
 - port: 5000
 + port: 5000



db
 +/-  port: 5000 => 6000
 +/-  env : ["a","b","c","d"] => ["a","c","b","d"]


*/
