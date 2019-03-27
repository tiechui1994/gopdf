package gopdf

import (
	"reflect"
	"strconv"
	"strings"
)

func isEmpty(object interface{}) bool {
	if object == nil {
		return true
	}

	objValue := reflect.ValueOf(object)
	switch objValue.Kind() {
	// collection types are empty when they have no element
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return isEmpty(deref)
		// for all other types, compare against the zero value
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}

func checkColor(color string) string {
	color = strings.Replace(color, " ", "", -1)
	rgb := strings.Split(color, ",")
	if len(rgb) != 3 {
		panic("the color err")
	}

	for i := range rgb {
		value, err := strconv.Atoi(rgb[i])
		if err != nil {
			panic(err)
		}
		if value < 0 || value > 255 {
			panic("the R,G,B value error")
		}
	}

	return color
}

func getColorRGB(color string) (r, g, b int) {
	color = checkColor(color)
	rgb := strings.Split(color, ",")
	return atoi(rgb[0]), atoi(rgb[1]), atoi(rgb[2])
}

func replaceBorder(border *Scope) {
	if border.Left < 0 {
		border.Left = 0
	}
	if border.Right < 0 {
		border.Right = 0
	}
	if border.Top < 0 {
		border.Top = 0
	}

	border.Bottom = 0
}

func replaceMarign(margin *Scope) {
	margin.Right = 0
	if margin.Bottom < 0 {
		margin.Bottom = 0
	}
}

func atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
