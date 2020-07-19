package markdown

import (
	"regexp"
	"net/url"
	"strings"
)

type String struct {
	data []rune
}

func Str(text string) *String {
	return &str{
		data: []rune(text),
	}
}

func (s *String) replace(regex string, repl string) string {
	re := regexp.MustCompile(regex)
	return re.ReplaceAllString(string(s.data), repl)
}

func (s *String) replaceFunc(regex string, repl string) string {
	re := regexp.MustCompile(regex)
	return re.ReplaceAllString(string(s.data), repl)
}

func (s *String) charAt(i int) string {
	if i < 0 || i >= len(s.data) {
		return ""
	}

	return string(s.data[i])
}

func decodeURIComponent(str string) string {
	r := strings.Replace(str, "%20", "+", -1)
	r, _ = url.QueryUnescape(r)
	return r
}

func encodeURL(str string) string {
	return url.QueryEscape(str)
}
