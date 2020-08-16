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

func NewLex() *Lexer {
	lex := Lexer{}
	lex.tokens.tokens = []Token{}
	lex.tokens.links = make(map[string]Link)
	return &lex
}

func (l *Lexer) lex(text string) []Token {
	re_break := MustCompile(`\r\n|\r`, RE2)
	text, _ = re_break.Replace(text, "", 0, -1)
	re_blank := MustCompile(`\t`, RE2)
	text, _ = re_blank.Replace(text, "    ", 0, -1)

	l.blockTokens(text, &l.tokens.tokens, true)
	l.inline(&l.tokens.tokens)

	return l.tokens.tokens
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
		if token, _ = space(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			if token.Type != "" {
				*tokens = append(*tokens, token)
			}
			continue
		}

		// code
		if token, _ = code(src, l.tokens.tokens); !IsEmpty(token) {
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
		if token, _ = fences(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// heading
		if token, _ = heading(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// TODO: table no leading pipe (gfm)

		// hr
		if token, _ = hr(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// blockquote
		if token, _ = blockquote(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			token.Tokens = l.blockTokens(token.Text, &[]Token{}, top)
			*tokens = append(*tokens, token)
			continue
		}

		// list
		if token, _ = list(src); !IsEmpty(token) {
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
		if token, _ = lheading(src); !IsEmpty(token) {
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
		if token, _ = paragraph(src); !IsEmpty(token) {
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
			l.inlineTokens(token.Text, &token.Tokens, false, false, "")
		case "table":
			// TODO

		case "blockquote":
			l.inline(&token.Tokens)
		case "list":
			in := len(token.Items)
			for k := 0; k < in; k++ {
				l.inline(&token.Items[k].Tokens)
			}
		default:
			// do nothing
		}
	}

	return *tokens
}

func (l *Lexer) inlineTokens(text string, tokens *[]Token, inLink, inRawBlock bool, prevChar string) []Token {
	maskedSrc := []rune(text)
	src := []rune(text)
	if len(l.tokens.links) > 0 {
		links := make([]string, 0, len(l.tokens.links))
		for key := range l.tokens.links {
			links = append(links, key)
		}
		sort.Strings(links)

		if len(links) > 0 {
			linkSearch := inline["reflinkSearch"]
			match, _ := linkSearch.Exec(maskedSrc)
			for match != nil {
				zero := match.GroupByNumber(0).Runes()
				search := str(zero).slice(str(zero).lastIndexOf("[")+1, -1).string()
				if includes(links, search, 0) {
					maskedSrc = []rune(str(maskedSrc).slice(0, match.Index).string() + "[" +
						strings.Repeat("a", len(zero)-2) + "]" + str(maskedSrc).slice(linkSearch.LastIndex).string())
				}
			}
		}
	}

	// Mask out other blocks
	blockSkip := inline["blockSkip"]
	match, _ := blockSkip.Exec(maskedSrc)
	if match != nil {
		zero := match.GroupByNumber(0).Runes()
		maskedSrc = []rune(str(maskedSrc).slice(0, match.Index).string() + "[" +
			strings.Repeat("a", len(zero)-2) + "]" + str(maskedSrc).slice(blockSkip.LastIndex).string())
	}

	for len(src) > 0 {
		// escape
		if token, _ := escape(src); !IsEmpty(token) {
			src = src[len(token.Raw):]
			*tokens = append(*tokens, token)
			continue
		}

		// TODO: tag

		// link
		if token, _ := link(src); !IsEmpty(token) {
			src = src[len(token.Raw):]
			if token.Type == "link" {
				token.Tokens = l.inlineTokens(token.Text, &[]Token{}, true, inRawBlock, "")
			}
			*tokens = append(*tokens, token)
			continue
		}

		// relink, nolink
		if token, _ := reflink(src, l.tokens.links); !IsEmpty(token) {
			src = src[len(token.Raw):]
			if token.Type == "link" {
				token.Tokens = l.inlineTokens(token.Text, &[]Token{}, true, inRawBlock, "")
			}
			*tokens = append(*tokens, token)
			continue
		}

		// strong
		if token, _ := strong(src, maskedSrc, prevChar); !IsEmpty(token) {
			src = src[len(token.Raw):]
			token.Tokens = l.inlineTokens(token.Text, &[]Token{}, inLink, inRawBlock, "")
			*tokens = append(*tokens, token)
			continue
		}

		// em
		if token, _ := em(src, maskedSrc, prevChar); !IsEmpty(token) {
			src = src[len(token.Raw):]
			token.Tokens = l.inlineTokens(token.Text, &[]Token{}, inLink, inRawBlock, "")
			*tokens = append(*tokens, token)
			continue
		}

		// code
		if token, _ := codespan(src); !IsEmpty(token) {
			src = src[len(token.Raw):]
			*tokens = append(*tokens, token)
			continue
		}

		// br
		if token, _ := br(src); !IsEmpty(token) {
			src = src[len(token.Raw):]
			*tokens = append(*tokens, token)
			continue
		}

		// del
		//if token, _ := del(src); !IsEmpty(token) {
		//	src = src[len(token.Raw):]
		//	token.Tokens = l.inlineTokens(token.Text, &[]Token{}, inLink, inRawBlock, "")
		//	*tokens = append(*tokens, token)
		//	continue
		//}

		// autolink
		if token, _ := autoLink(src); !IsEmpty(token) {
			src = src[len(token.Raw):]
			*tokens = append(*tokens, token)
			continue
		}

		// url
		//token, _ := url(src)
		//if !inLink && !IsEmpty(token) {
		//	src = src[len(token.Raw):]
		//	*tokens = append(*tokens, token)
		//	continue
		//}

		// text
		if token, _ := inlineText(src, inRawBlock); !IsEmpty(token) {
			src = src[len(token.Raw):]
			prevChar = str([]rune(token.Raw)).slice(-1).string()
			*tokens = append(*tokens, token)
			continue
		}

		if len(src) > 0 {
			errMsg := "Infinite loop on byte: " + string(src[0])
			panic(errMsg)
		}
	}

	return *tokens
}
