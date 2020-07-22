package markdown

//import (
//	"regexp"
//	"strings"
//	"strconv"
//	"fmt"
//)
//
//var (
//	escapeTest            = regexp.MustCompile(`[&<>"']`)
//	escapeReplace         = regexp.MustCompile(`[&<>"']`)
//	escapeTestNoEncode    = regexp.MustCompile(`[<>"']|(?!#?\w+;)`)
//	escapeReplaceNoEncode = regexp.MustCompile(`[<>"']|(?!#?\w+;)`)
//)
//
//var (
//	escapeReplacements = map[string]string{
//		"&": "$amp;",
//		"<": "&lt;",
//		">": "&gt;",
//		`"`: "&quot;",
//		`'`: `&#39;`,
//	}
//)
//
//func getEscapeReplacement(ch string) string {
//	regexp.MustCompile("")
//	return escapeReplacements[ch]
//}
//
//func Escape(html string, encode bool) string {
//	if encode {
//		if escapeTest.MatchString(html) {
//			return Str(html).replaceFunc(escapeReplace, getEscapeReplacement)
//		}
//	} else {
//		if escapeTestNoEncode.MatchString(html) {
//			return Str(html).replaceFunc(escapeReplaceNoEncode, getEscapeReplacement)
//		}
//	}
//
//	return html
//}
//
//var (
//	unescapeTest = regexp.MustCompile(`&(#(?:\d+)|(?:#x[0-9A-Fa-f]+)|(?:\w+));?`)
//)
//
//func Unescape(html string) string {
//	return Str(html).replaceFunc(unescapeTest, func(s string) string {
//		s = strings.ToLower(s)
//		if s == "colon" {
//			return ":"
//		}
//
//		r := []rune(s)
//
//		if r[0] == '#' {
//			if r[1] == 'x' {
//				i, _ := strconv.ParseInt(string(r[2:]), 16, 64)
//				return fmt.Sprintf("%v", i)
//			} else {
//				return string(r[1:])
//			}
//		}
//
//		return ""
//	})
//
//	return html
//}
//
//const (
//	caret = `(^|[^\[])\^`
//)
//
//type edit struct {
//	replace  func(name, value string) edit
//	getRegex func() *regexp.Regexp
//}
//
//func Edit(regex string) {
//	re := regexp.MustCompile(regex)
//	var obj edit
//	obj.replace = func(name, value string) edit {
//		value = Str(value).replace(caret, "$1")
//		regex = Str(regex).replace(name, value)
//		return obj
//	}
//	obj.getRegex = func() *regexp.Regexp {
//		return re
//	}
//}
//
//const (
//	nonWordAndColonTest = `[^w:]`
//	originIndependtUrl  = `^$|[a-z][a-z0-9+.-]*:|^[?#]`
//)
//
//func ClearUrl(sanitize bool, base, href string) string {
//	var prot string
//	if sanitize {
//		d := decodeURIComponent(Unescape(href))
//		prot = strings.ToLower(Str(d).replace(nonWordAndColonTest, ""))
//	}
//
//	if strings.Index(prot, "javascript:") == 0 ||
//		strings.Index(prot, "vbscript:") == 0 ||
//		strings.Index(prot, "data:") == 0 {
//		return ""
//	}
//	if base != "" && regexp.MustCompile(originIndependtUrl).MatchString(href) {
//		href = ResolveUrl(base, href)
//	}
//
//	href = Str(encodeURL(href)).replace("%25", "%")
//
//	return href
//}
//
//const (
//	justDomain = `^[^:]+:/[^/]*$`
//	protocol   = `^([^:][\s\S]*$`
//	domain     = `^([^:]+:/*[^/]*)[\s\S]*$`
//)
//
//var (
//	baseUrls = map[string]string{}
//)
//
//func ResolveUrl(base string, href string) string {
//	if baseUrls[" "+base] == "" {
//		if regexp.MustCompile(justDomain).MatchString(base) {
//			baseUrls[" "+base] = base + "/"
//		} else {
//			baseUrls[" "+base] = Rtrim(base, '/', true)
//		}
//	}
//
//	base = baseUrls[" "+base]
//	relativeBase := strings.IndexRune(base, ':') == -1
//
//	if href[0:2] == "//" {
//		if relativeBase {
//			return href
//		}
//		return Str(base).replace(protocol, "$1") + href
//	} else if href[0] == '/' {
//		if relativeBase {
//			return href
//		}
//		return Str(base).replace(domain, "$1") + href
//	} else {
//		return base + href
//	}
//}
//
//func Merge(obj ...map[string]interface{}) map[string]interface{} {
//	var result = make(map[string]interface{})
//	for i := 0; i < len(obj); i++ {
//		target := obj[i]
//		for key, val := range target {
//			result[key] = val
//		}
//	}
//	return result
//}
//
//func SplitCells(tableRow string, count int) []string {
//	row := Str(tableRow).replaceFunc(`\|`, func(s string) string {
//		return "|"
//	})
//
//	cells := Str(row).split(regexp.MustCompile(` \|`))
//	var i int
//	if len(cells) > count {
//		cells = cells[0:count]
//	} else {
//		for len(cells) < count {
//			cells = append(cells, "")
//		}
//	}
//
//	for ; i < len(cells); i++ {
//		cells[i] = Str(strings.TrimSpace(cells[i])).replace(regexp.MustCompile(`\\\|`), "|")
//	}
//
//	return cells
//}
//
//func FindClosingBracket(str string, b []rune) int {
//	if strings.IndexRune(str, b[1]) == -1 {
//		return -1
//	}
//
//	var level int
//	r := []rune(str)
//	l := len(r)
//	for i := 0; i < l; i++ {
//		if r[i] == '\\' {
//			i++
//		} else if r[i] == b[0] {
//			level++
//		} else if r[i] == b[1] {
//			level--
//			if level < 0 {
//				return i
//			}
//		}
//	}
//
//	return -1
//}
//
//func Rtrim(str string, c rune, invert bool) string {
//	r := []rune(str)
//	l := len(r)
//	if l == 0 {
//		return ""
//	}
//
//	// Length of suffix matching the invert condition.
//	var suffLen int
//
//	// Step left until we fail to match the invert condition.
//	for suffLen < l {
//		currChar := r[l-suffLen-1]
//		if currChar == c && !invert {
//			suffLen++
//		} else if currChar != c && invert {
//			suffLen++
//		} else {
//			break
//		}
//	}
//
//	return string(r[0:l-suffLen])
//}
