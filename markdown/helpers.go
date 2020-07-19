package markdown

import (
	"regexp"
	"strings"
)

var (
	escapeTest            = regexp.MustCompile(`[&<>"']`)
	escapeReplace         = regexp.MustCompile(`[&<>"']`)
	escapeTestNoEncode    = regexp.MustCompile(`[<>"']|&(?!#?\w+;)`)
	escapeReplaceNoEncode = regexp.MustCompile(`[<>"']|&(?!#?\w+;)`)
)

var (
	escapeReplacements = map[string]string{
		"&": "$amp;",
		"<": "&lt;",
		">": "&gt;",
		`"`: "&quot;",
		`'`: `&#39;`,
	}
)

func getEscapeReplacement(ch string) string {
	return escapeReplacements[ch]
}

func Escape(html string, encode bool) string {
	return html
}

var (
	unescapeTest = regexp.MustCompile(`&(#(?:\d+)|(?:#x[0-9A-Fa-f]+)|(?:\w+));?`)
)

func Unescape(html string) string {
	return html
}

const (
	caret = `(^|[^\[])\^`
)

type edit struct {
	replace  func(name, value string) edit
	getRegex func() *regexp.Regexp
}

func Edit(regex string) {
	re := regexp.MustCompile(regex)
	var obj edit
	obj.replace = func(name, value string) edit {
		value = Str(value).replace(caret, "$1")
		regex = Str(regex).replace(name, value)
		return obj
	}
	obj.getRegex = func() *regexp.Regexp {
		return re
	}
}

const (
	nonWordAndColonTest = `[^w:]`
	originIndependtUrl  = `^$|[a-z][a-z0-9+.-]*:|^[?#]`
)

func ClearUrl(sanitize bool, base, href string) string {
	var prot string
	if sanitize {
		d := decodeURIComponent(Unescape(href))
		prot = strings.ToLower(Str(d).replace(nonWordAndColonTest, ""))
	}

	if strings.Index(prot, "javascript:") == 0 ||
		strings.Index(prot, "vbscript:") == 0 ||
		strings.Index(prot, "data:") == 0 {
		return ""
	}
	if base != "" && regexp.MustCompile(originIndependtUrl).MatchString(href) {
		href = ""
	}

	href = Str(encodeURL(href)).replace("%25", "%")

	return href
}

const (
	justDomain = `^[^:]+:/[^/]*$`
	protocol = `^([^:][\s\S]*$`
	domain = `^([^:]+:/*[^/]*)[\s\S]*$`
)

func ResolveURL(base string, href string)  {

}
