package lex

import (
	"reflect"
	"strings"
	"regexp"
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

type editor struct {
	getRegex func() *Regex
	replace  func(name, value string) *editor
}

func edit(re interface{}, options ...Options) *editor {
	e := new(editor)
	var regex string
	switch re.(type) {
	case *Regex:
		regex = re.(*Regex).Source
	case string:
		regex = re.(string)
	}

	var opt = Options(RE2)
	if len(options) > 0 {
		for _, val := range options {
			opt = opt | val
		}
	}

	e.getRegex = func() *Regex {
		return MustCompile(regex, opt)
	}

	e.replace = func(name, value string) *editor {
		caret := MustCompile(`(^|[^\[])\^`, RE2)
		value, _ = caret.Replace(value, "$1", 0, -1)
		regex = strings.ReplaceAll(regex, name, value)
		return e
	}

	return e
}

func merge(args ...map[string]*Regex) map[string]*Regex {
	result := make(map[string]*Regex)
	for i := 0; i < len(args); i++ {
		if args[i] != nil {
			for k, v := range args[i] {
				result[k] = v
			}
		}
	}

	return result
}

func splitCells(tableRow string, count int) []string {
	re := MustCompile(`\|`, Global)
	row := str(tableRow).replaceFunc(re, func(match *Match, offset int, s str) str {
		var escaped bool
		curr := offset
		curr--
		for curr >= 0 && s[curr] == '\\' {
			escaped = !escaped
			curr--
		}
		if escaped {
			return str("|")
		} else {
			return str(" |")
		}
	})

	cells := regexp.MustCompile(` \|`).Split(row.string(), -1)
	if len(cells) > count && count != 0 {
		cells = cells[0:count]
	} else {
		for len(cells) < count {
			cells = append(cells, "")
		}
	}
	for i := 0; i < len(cells); i++ {
		re := MustCompile(`\|`, Global)
		cells[i] = str(strings.TrimSpace(cells[i])).replace(re, "|").string()
	}

	return cells
}

// token

func findClosingBracket(str []rune, b []rune) int {
	if strings.Index(string(str), string(b)) == -1 {
		return -1
	}

	l := len(str)
	level := 0
	for i := 0; i < l; i++ {
		if str[i] == '\\' {
			i++
		} else if str[i] == b[0] {
			level++
		} else if str[i] == b[1] {
			level--
			if level < 0 {
				return i
			}
		}
	}

	return -1
}

func outputLink(match *Match, link Link, raw string) Token {
	href := link.Href
	title := link.Title

	re := MustCompile(`\\([\[\]])`, RE2)
	text, _ := re.Replace(match.GroupByNumber(1).String(), "$1", 0, -1)

	zero := match.GroupByNumber(0).Runes()
	if zero[0] != '!' {
		return Token{
			Type:   "link",
			Raw:    raw,
			Href:   href,
			Title:  title,
			Text:   text,
			Tokens: []Token{},
		}
	}

	return Token{
		Type:   "image",
		Raw:    raw,
		Href:   href,
		Title:  title,
		Text:   text,
		Tokens: []Token{},
	}
}

func indentCodeCompensation(raw, text string) string {
	// ^(\s+)(?:```)
	restr := strings.Replace(`^(\s+)(?:$1$1$1)`, "$1", "`", -1)
	match, err := str(raw).match(MustCompile(restr, RE2))
	if len(match) == 0 || err != nil {
		return text
	}

	indentToCode := match[0].GroupByNumber(1).Runes()
	nodes := strings.Split(text, "\n")
	renode := MustCompile(`^\s+`, RE2)
	nodes = strmap(nodes, func(node string, index int, array []string) string {
		matchIndentInNode, err := str(node).match(renode)
		if err != nil || matchIndentInNode == nil {
			return node
		}

		indentInNode := matchIndentInNode[0].Runes()
		if len(indentInNode) > len(indentToCode) {
			return str(node).slice(len(indentToCode)).string()
		}

		return node
	})

	return strings.Join(nodes, "\n")
}
