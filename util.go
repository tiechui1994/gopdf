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
	//fmt.Println(len(rgb), color)
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

func getRGB(color string) (r, g, b int) {
	checkColor(color)
	rgb := strings.Split(color, ",")
	return Atoi(rgb[0]), Atoi(rgb[1]), Atoi(rgb[2])
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
	if border.Bottom < 0 {
		border.Bottom = 0
	}
}

func Atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
