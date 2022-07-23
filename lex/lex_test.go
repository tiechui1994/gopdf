package lex

import (
	"testing"
	"strings"
	"io/ioutil"
)

func TestString(t *testing.T) {
	var x = "中华人民共和[国"
	idx1 := strings.LastIndex(x, "[")
	idx2 := strings.LastIndexFunc(x, func(r rune) bool {
		return r == '['
	})

	t.Log(idx1, idx2)
}

func TestNewLex(t *testing.T) {
	data, _ := ioutil.ReadFile("./src/mark.md")
	lex := NewLex()
	tokens := lex.Lex(string(data))
	for _, token := range tokens {
		_ = token
		t.Log("\n", token.String())
	}
}

func TestLex(t *testing.T) {
	NewLex()
	text := `| AA | BB | CC |
| -- | -- | -- |
| 1 | 2 | 3, **oo** | 
| - 4, x | 5 | 6 *ss*  |`
	t.Log(block["table"].Exec([]rune(text)))
}
