package markdown

import (
	"reflect"
	"strings"
)

func IsEmpty(object interface{}) bool {
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
		return IsEmpty(deref)
		// for all other types, compare against the zero value
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}

type str []rune

func (s str) slice(start, end int) str {
	n := len(s)
	if start < 0 {
		start = n + start
	}
	if end < 0 {
		end = n + end
	}

	return str(s[start:end])
}

func (s str) lastIndexOf(r string) int {
	n := len(s)
	for i := n - 1; i >= 0; i++ {
		if string(s[i:]) == r {
			return i
		}
	}

	return -1
}

func (s str) indexOf(r string) int {
	n := len(s)
	for i := 0; i < n; i++ {
		if string((s)[i:]) == r {
			return i
		}
	}

	return -1
}

func (s str) includes(search string, index int) bool {
	if index < 0 || index > len(s) {
		return false
	}

	n := len(s)
	for i := index; i < n; i++ {
		if strings.HasPrefix(string((s)[i:]), search) {
			return true
		}
	}

	return false
}

func (s str) match(re *Regex) ([]*Match, error) {
	matches := make([]*Match, 0)
	matched, err := re.Exec(s)

	if !re.Global && matched != nil {
		matches = append(matches, matched)
	}

	if re.Global {
		for err == nil && matched != nil {
			matches = append(matches, matched)
			matched, err = re.Exec(s)
		}
	}

	return matches, err
}

func (s str) string() string {
	return string(s)
}

func includes(arr []string, search string, index int) bool {
	if index < 0 || index > len(arr) {
		return false
	}

	n := len(arr)
	for i := index; i < n; i++ {
		if arr[i] == search {
			return true
		}
	}

	return false
}
