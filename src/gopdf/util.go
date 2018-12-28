package gopdf

import (
	"reflect"
	"strings"
	"strconv"
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

func checkColor(color string) {
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
}

func getRGB(color string) (r, g, b string) {
	checkColor(color)
	rgb := strings.Split(color, ",")
	return rgb[0], rgb[1], rgb[2]
}

func replactBorder(border *scope) {
	if border.left < 0 {
		border.left = 0
	}
	if border.right < 0 {
		border.right = 0
	}
	if border.top < 0 {
		border.bottom = 0
	}
	if border.bottom < 0 {
		border.bottom = 0
	}
}
