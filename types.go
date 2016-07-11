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

func print(c *color.Color, depth int, field, action, msg string) {
	if field != "" {
		c.Printf("%s%s%s: %s", action, strings.Repeat(" ", depth*2), field, msg)
	} else {
		c.Printf("%s%s%s", action, strings.Repeat(" ", depth*2), msg)
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

		print(normal, depth, field, "", "[\n")
		for i := 0; i < c; i++ {
			if i >= currentValue.Len() {
				_printServiceSpecDiff(depth, "", reflect.Indirect(reflect.New(expectedValue.Index(i).Type())).Interface(), expectedValue.Index(i).Interface())
			} else if i >= expectedValue.Len() {
				_printServiceSpecDiff(depth, "", currentValue.Index(i).Interface(), reflect.Indirect(reflect.New(currentValue.Index(i).Type())).Interface())
			} else {
				_printServiceSpecDiff(depth, "", currentValue.Index(i).Interface(), expectedValue.Index(i).Interface())
			}
		}
		print(normal, depth, "", "", "]\n")
	case reflect.Map:
		currentValue := reflect.ValueOf(current)
		expectedValue := reflect.ValueOf(expected)
		print(normal, depth, field, "", "{\n")

		for _, k := range currentValue.MapKeys() {
			ev := expectedValue.MapIndex(k)
			var expectedKeyValue interface{}
			if ev.IsValid() {
				expectedKeyValue = ev.Interface()
			} else {
				expectedKeyValue = reflect.Indirect(reflect.New(currentValue.MapIndex(k).Type())).Interface()
			}
			_printServiceSpecDiff(depth, fmt.Sprintf("%s", k.Interface()), currentValue.MapIndex(k).Interface(), expectedKeyValue)
		}

		for _, k := range expectedValue.MapKeys() {
			cv := currentValue.MapIndex(k)
			var currentKeyValue interface{}
			if cv.IsValid() {
				continue
			} else {
				currentKeyValue = reflect.Indirect(reflect.New(expectedValue.MapIndex(k).Type())).Interface()
			}
			_printServiceSpecDiff(depth, fmt.Sprintf("%s", k.Interface()), currentKeyValue, expectedValue.MapIndex(k).Interface())
		}

		print(normal, depth, "", "", "}\n")
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

		_printServiceSpecDiff(depth, "", dcv, dev)

		/*
			                current          expected
					nil              nil
					*time.Duration   nil
					nil              *time.Duration
					*time.Duration   *time.Duration



			                *time.Duration
					time.Duration

					currentValue := reflect.ValueOf(current)
					expectedValue := reflect.ValueOf(expected)

					var curr, exp interface{}

					if !reflect.Indirect(currentValue).IsValid() {
						curr = nil
					} else {
						curr = currentValue.Interface()
					}

					if !reflect.Indirect(expectedValue).IsValid() {
						curr = nil
					} else {
						curr = currentValue.Interface()
					}

					_printServiceSpecDiff(depth, "", curr, exp)

		*/
	case reflect.Struct:
		currentValue := reflect.ValueOf(current)
		expectedValue := reflect.ValueOf(expected)

		first := true
		for i := 0; i < currentValue.NumField(); i++ {
			f := currentValue.Type().Field(i)
			if f.PkgPath == "" {
				field = f.Name
				if first {
					print(normal, depth, field, "", "{\n")
					first = false
				}
				_printServiceSpecDiff(depth, field, currentValue.Field(i).Interface(), expectedValue.Field(i).Interface())
			}

		}
		if !first {
			print(normal, depth, "", "", "}\n")
		}
	default:
		sc := fmt.Sprint(current)
		se := fmt.Sprint(expected)

		if sc == se {
			print(normal, depth, field, "", fmt.Sprintf(`"%s" => "%s"`, sc, se))
		} else if sc == "" {
			print(green, depth, field, "+", fmt.Sprintf(`"%s"`, se))
		} else if se == "" {
			print(red, depth, field, "-", fmt.Sprintf(`"%s"`, sc))
		} else {
			print(yellow, depth, field, "+/-", fmt.Sprintf(`"%s" => "%s"`, sc, se))
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
