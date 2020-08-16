package markdown

import (
	"testing"
	"strings"
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
Markdown Quick Reference
========================

This guide is a very brief overview, with examples, of the syntax that [Markdown] supports. It is itself written in Markdown and you can copy the samples over to the left-hand pane for experimentation. It is shown as *text* and not *rendered HTML*.

[Markdown]: http://daringfireball.net/projects/markdown/


Simple Text Formatting
======================

First thing is first. You can use *stars* or _underscores_ for italics. **Double stars** and __double underscores__ for bold. ***Three together*** for ___both___.

Paragraphs are pretty easy too. Just have a blank line between chunks of text.
`
	lex := NewLex()
	tokens := lex.lex(str)
	for _, token := range tokens {
		t.Log("\n", token.String())
	}
}
