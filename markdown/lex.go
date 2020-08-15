package markdown

import (
	"github.com/dlclark/regexp2"
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
	re_break := regexp2.MustCompile(`\r\n|\r`, regexp2.RE2)
	text, _ = re_break.Replace(text, "", 0, -1)
	re_blank := regexp2.MustCompile(`\t`, regexp2.RE2)
	text, _ = re_blank.Replace(text, "    ", 0, -1)

}

func (l *Lexer) blockTokens(text string, tokens *[]Token, top bool) []Token {
	re_blank := regexp2.MustCompile(`^ +$`, regexp2.Multiline)
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
				//source := match.GroupByNumber(0).Runes()
				//source = source[strings.LastIndex(source, "[")+1:]

			}
		}
	}
}

func includes(source []rune, search string, index int) bool {
	if index < 0 || index > len(source) {
		return false
	}

	for i := index; i < len(source); i++ {
		if strings.HasPrefix(string(source[i:]), search) {
			return true
		}
	}

	return false
}


