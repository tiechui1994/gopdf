package lex

import (
	"strings"
)

type Link struct {
	Href  string `json:"href"`
	Title string `json:"title"`
}

type Lexer struct {
	tokens []Token
	links  map[string]Link
}

func NewLex() *Lexer {
	block = initBlockGfm()
	inline = initInlineGfm(false)
	lex := Lexer{}
	lex.tokens = []Token{}
	lex.links = make(map[string]Link)
	return &lex
}

func (l *Lexer) Lex(text string) []Token {
	re_break := MustCompile(`\r\n|\r`, RE2)
	text, _ = re_break.Replace(text, "", 0, -1)
	re_blank := MustCompile(`\t`, RE2)
	text, _ = re_blank.Replace(text, "    ", 0, -1)

	l.blockTokens(text, &l.tokens, true)
	l.inline(&l.tokens)

	return l.tokens
}

func (l *Lexer) blockTokens(content string, tokens *[]Token, top bool) []Token {
	re_blank := MustCompile(`^ +$`, Multiline)
	content, _ = re_blank.Replace(content, "", 0, -1)

	src := []rune(content)
	for len(src) > 0 {
		// newline
		if token, _ := space(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			if token.Type != "" {
				*tokens = append(*tokens, token)
			}
			continue
		}

		// code
		if token, _ := code(src, &l.tokens); !IsEmpty(token) {
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
		if token, _ := fences(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// heading
		if token, _ := heading(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// table no leading pipe (gfm)
		if token, _ := nptable(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// hr
		if token, _ := hr(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// blockquote
		if token, _ := blockquote(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			token.Tokens = l.blockTokens(token.Text, &[]Token{}, top)
			*tokens = append(*tokens, token)
			continue
		}

		// list
		if token, _ := list(src); !IsEmpty(token) {
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
		if top {
			if token, _ := def(src); !IsEmpty(token) {
				src = src[len([]rune(token.Raw)):]
				if IsEmpty(l.links[token.Tag]) {
					l.links[token.Tag] = Link{
						Href:  token.Href,
						Title: token.Title,
					}
				}
				continue
			}
		}

		// table
		if token, _ := table(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// lheading
		if token, _ := lheading(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// top-level paragraph
		if top {
			if token, _ := paragraph(src); !IsEmpty(token) {
				src = src[len([]rune(token.Raw)):]
				*tokens = append(*tokens, token)
				continue
			}
		}

		// text
		if token, _ := text(src, tokens); !IsEmpty(token) {
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
			hn := len(token.Header)
			token.Elements.Header = make([][]Token, hn)
			for j := 0; j < hn; j++ {
				token.Elements.Header[j] = []Token{}
				l.inlineTokens(token.Header[j], &token.Elements.Header[j], false, false, "")
			}

			cn := len(token.Cells)
			token.Elements.Cells = make([][][]Token, cn)
			for j := 0; j < cn; j++ {
				row := token.Cells[j]
				token.Elements.Cells[j] = make([][]Token, len(row))
				for k := 0; k < len(row); k++ {
					token.Elements.Cells[j][k] = []Token{}
					l.inlineTokens(row[k], &token.Elements.Cells[j][k], false, false, "")
				}
			}

		case "blockquote":
			l.inline(&token.Tokens)
		case "list":
			in := len(token.Items)
			for k := 0; k < in; k++ {
				l.inline(&token.Items[k].Tokens)
				(*tokens)[i] = token
			}
		default:
			// do nothing
		}
		(*tokens)[i] = token
	}

	return *tokens
}

func (l *Lexer) inlineTokens(text string, tokens *[]Token, inLink, inRawBlock bool, prevChar string) []Token {
	maskedSrc := []rune(text)
	src := []rune(text)
	if len(l.links) > 0 {
		links := make([]string, 0, len(l.links))
		for key := range l.links {
			links = append(links, key)
		}

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
				match, _ = linkSearch.Exec(maskedSrc)
			}
		}
	}

	// Mask out other blocks
	blockSkip := inline["blockSkip"]
	match, _ := blockSkip.Exec(maskedSrc)
	for match != nil {
		zero := match.GroupByNumber(0).Runes()
		maskedSrc = []rune(str(maskedSrc).slice(0, match.Index).string() + "[" +
			strings.Repeat("a", len(zero)-2) + "]" + str(maskedSrc).slice(blockSkip.LastIndex).string())
		match, _ = blockSkip.Exec(maskedSrc)
	}

	for len(src) > 0 {
		// escape
		if token, _ := escape(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// TODO: tag

		// link
		if token, _ := link(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			if token.Type == "link" {
				token.Tokens = l.inlineTokens(token.Text, &[]Token{}, true, inRawBlock, "")
			}
			*tokens = append(*tokens, token)
			continue
		}

		// relink, nolink
		if token, _ := reflink(src, l.links); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			if token.Type == "link" {
				token.Tokens = l.inlineTokens(token.Text, &[]Token{}, true, inRawBlock, "")
			}
			*tokens = append(*tokens, token)
			continue
		}

		// strong
		if token, _ := strong(src, maskedSrc, prevChar); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			token.Tokens = l.inlineTokens(token.Text, &[]Token{}, inLink, inRawBlock, "")
			*tokens = append(*tokens, token)
			continue
		}

		// em
		if token, _ := em(src, maskedSrc, prevChar); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			token.Tokens = l.inlineTokens(token.Text, &[]Token{}, inLink, inRawBlock, "")
			*tokens = append(*tokens, token)
			continue
		}

		// code
		if token, _ := codespan(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// br
		if token, _ := br(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// del
		if token, _ := del(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			token.Tokens = l.inlineTokens(token.Text, &[]Token{}, inLink, inRawBlock, "")
			*tokens = append(*tokens, token)
			continue
		}

		// autolink
		if token, _ := autoLink(src); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			*tokens = append(*tokens, token)
			continue
		}

		// url
		if !inLink {
			if token, _ := url(src); !IsEmpty(token) {
				src = src[len([]rune(token.Raw)):]
				*tokens = append(*tokens, token)
				continue
			}
		}

		// text
		if token, _ := inlineText(src, inRawBlock); !IsEmpty(token) {
			src = src[len([]rune(token.Raw)):]
			prevChar = str(token.Raw).slice(-1).string()
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
