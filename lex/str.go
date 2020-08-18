package lex

import (
	"strings"
)

type str []rune

func (s str) slice(start int, end ...int) str {
	n := len(s)
	i := start
	if start < 0 && start < -n {
		i = 0
	} else if start < 0 {
		i = n + start
	}

	j := n
	if len(end) > 0 {
		if end[0] >= 0 && end[0] < n {
			j = end[0]
		}

		if end[0] < 0 && end[0] >= -n {
			j = n + end[0]
		}

		if end[0] < -n {
			j = 0
		}
	}

	if j < i {
		return str("")
	}

	return str(s[i:j])
}

func (s str) lastIndexOf(r string) int {
	n := len(s)
	for i := n - 1; i >= 0; i-- {
		if strings.HasPrefix(string(s[i:]), r) {
			return i
		}
	}

	return -1
}

func (s str) indexOf(r string) int {
	n := len(s)
	for i := 0; i < n; i++ {
		if strings.HasPrefix(string(s[i:]), r) {
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
		if strings.HasPrefix(string(s[i:]), search) {
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

func (s str) replace(re *Regex, repl string) str {
	re.LastIndex = 0
	var result string
	if re.Global {
		result = re.ReplaceStr(string(s), repl, 0, -1)
	} else {
		result = re.ReplaceStr(string(s), repl, 0, 1)
	}

	return str(result)
}

func (s str) replaceFunc(re *Regex, f func(match *Match, offset int, s str) str) str {
	result := str{}
	if re.Global {
		re.LastIndex = 0

		last := 0
		match, _ := re.Exec(s)
		for match != nil {
			res := f(match, match.Index, s)
			result = append(result, s[last:match.Index]...)
			result = append(result, res...)
			last = re.LastIndex
			match, _ = re.Exec(s)
		}

		result = append(result, s[last:]...)
	} else {
		re.LastIndex = 0
		match, _ := re.Exec(s)
		for match != nil {
			res := f(match, match.Index, s)
			result = append(result, s[0:match.Index]...)
			result = append(result, res...)
			result = append(result, s[re.LastIndex:]...)
		}
	}

	return result
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

func strmap(arr []string, f func(value string, index int, array []string) string) []string {
	newarr := make([]string, len(arr))
	for i := 0; i < len(arr); i++ {
		newarr[i] = f(arr[i], i, arr)
	}
	return newarr
}
