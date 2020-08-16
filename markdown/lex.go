package markdown

import (
	"sort"
	"strings"
)

type Link struct {
	Href  string `json:"href"`
	Title string `json:"title"`
}

type Tokens struct {
	links  map[string]Link
	tokens []Token
}

type Lexer struct {
	tokens Tokens
}

func NewLex() {

}

func (l *Lexer) lex(text string) {
	re_break := MustCompile(`\r\n|\r`, RE2)
	text, _ = re_break.Replace(text, "", 0, -1)
	re_blank := MustCompile(`\t`, RE2)
	text, _ = re_blank.Replace(text, "    ", 0, -1)
}

func (l *Lexer) blockTokens(text string, tokens *[]Token, top bool) []Token {
	re_blank := MustCompile(`^ +$`, Multiline)
	text, _ = re_blank.Replace(text, "", 0, -1)

	var (
		token Token
	)
	src := []rune(text)
	for len(src) > 0 {
		// newline
		token, _ = space(src)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			if token.Type != "" {
				*tokens = append(*tokens, token)
			}
			continue
		}

		// code
		token, _ = code(src, l.tokens.tokens)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			if token.Type != "" {
				*tokens = append(*tokens, token)
			} else {
				n := len(*tokens)
				lastToken := (*tokens)[n-1]
				lastToken.Raw += "\n" + token.Raw
				lastToken.Text += "\n" + token.Text
				(*tokens)[n-1] = lastToken
			}
			continue
		}

		// fences
		token, _ = fences(src)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// heading
		token, _ = heading(src)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// TODO: table no leading pipe (gfm)

		// hr
		token, _ = hr(src)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// blockquote
		token, _ = blockquote(src)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			token.Tokens = l.blockTokens(token.Text, &[]Token{}, top)
			*tokens = append(*tokens, token)
			continue
		}

		// list
		token, _ = list(src)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			n := len(token.Items)
			for i := 0; i < n; i++ {
				token.Items[i].Tokens = l.blockTokens(token.Items[i].Text, &[]Token{}, false)
			}
			*tokens = append(*tokens, token)
			continue
		}

		// TODO: html

		// def
		token, _ = def(src)
		if top && !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			if IsEmpty(l.tokens.links[token.Tag]) {
				l.tokens.links[token.Tag] = Link{
					Href:  token.Href,
					Title: token.Title,
				}
			}
			continue
		}

		// TODO: table (gfm)

		// lheading
		token, _ = lheading(src)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// top-level paragraph
		token, _ = paragraph(src)
		if top && !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// text
		token, _ = paragraph(src)
		if !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			if token.Type != "" {
				*tokens = append(*tokens, token)
			} else {
				lastToken := (*tokens)[len(*tokens)-1]
				lastToken.Raw += "\n" + token.Raw
				lastToken.Text += "\n" + token.Text
				(*tokens)[len(*tokens)-1] = lastToken
			}
			continue
		}

		if len(src) > 0 {
			errMsg := "Infinite loop on byte: " + string(src[0])
			panic(errMsg)
		}
	}

	return *tokens
}

func (l *Lexer) inline(tokens *[]Token) []Token {
	n := len(*tokens)

	for i := 0; i < n; i++ {
		token := (*tokens)[i]
		switch token.Type {
		case "paragraph", "text", "heading":
			token.Tokens = []Token{}

		}
	}

	return *tokens
}

func (l *Lexer) inlineTokens(text string, tokens *[]Token, inLink, inRawBlock bool, prevChar string) {
	maskedSrc := []rune(text)
	if len(l.tokens.links) > 0 {
		links := make([]string, 0, len(l.tokens.links))
		for key := range l.tokens.links {
			links = append(links, key)
		}
		sort.Strings(links)

		if len(links) > 0 {
			match, _ := inline["reflinkSearch"].FindRunesMatch(maskedSrc)
			for match != nil {
				zero := match.GroupByNumber(0).Runes()
				index := str(zero).lastIndexOf("[") + 1
				search := str(zero).slice(index, -1).string()
				if includes(links, search, 0) {
					maskedSrc = []rune(str(maskedSrc).slice(0, match.Index).string() + "[" +
						strings.Repeat("a", len(zero)-2) + "]")
				}
			}
		}

		if match, _ := inline["blockSkip"].FindRunesMatch(maskedSrc); match != nil {
			zero := match.GroupByNumber(0).Runes()
			//reflink := inline["reflinkSearch"].MatchTimeout
			maskedSrc = []rune(str(maskedSrc).slice(0, match.Index).string() + "[" +
				strings.Repeat("a", len(zero)-2) + "]" + str(maskedSrc).slice(0, 0).string())
		}

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
