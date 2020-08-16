package markdown

import (
	"testing"
	"strings"
	"bytes"
	"encoding/json"
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
	str := `
- 1233
- 2222

**aa**
`
	lex := NewLex()
	tokens := lex.lex(str)
	var buf bytes.Buffer
	encode := json.NewEncoder(&buf)
	encode.SetIndent("", "  ")
	for _, token := range tokens {
		buf.Reset()
		encode.Encode(token)
		t.Log("\n", buf.String())
	}
}
