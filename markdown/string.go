package markdown

import (
	"regexp"
	"net/url"
	"strings"
	"log"
)

type String struct {
	data []rune
}

func Str(text string) *String {
	return &String{
		data: []rune(text),
	}
}

func (s *String) replace(regex interface{}, repl string) string {
	var re *regexp.Regexp
	switch regex.(type) {
	case string:
		re = regexp.MustCompile(regex.(string))
	case *regexp.Regexp:
		re, _ = regex.(*regexp.Regexp)
	default:
		panic("invalid regex")
	}
	log.Println(string(s.data), repl, re.MatchString(string(s.data)))
	return re.ReplaceAllString(string(s.data), repl)
}

func (s *String) replaceFunc(regex interface{}, fn func(string) string) string {
	var re *regexp.Regexp
	switch regex.(type) {
	case string:
		re = regexp.MustCompile(regex.(string))
	case *regexp.Regexp:
		re, _ = regex.(*regexp.Regexp)
	default:
		panic("invalid regex")
	}

	return re.ReplaceAllStringFunc(string(s.data), fn)
}

func (s *String) charAt(i int) string {
	if i < 0 || i >= len(s.data) {
		return ""
	}

	return string(s.data[i])
}

func (s *String) split(c interface{}) []string {
	switch c.(type) {
	case string:
		return strings.Split(string(s.data), c.(string))
	case *regexp.Regexp:
		re, _ := c.(*regexp.Regexp)
		return re.Split(string(s.data), -1)
	default:
		return strings.Split(string(s.data), " ")
	}
}

func decodeURIComponent(str string) string {
	r := strings.Replace(str, "%20", "+", -1)
	r, _ = url.QueryUnescape(r)
	return r
}

func encodeURL(str string) string {
	return url.QueryEscape(str)
}
