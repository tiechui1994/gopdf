package markdown

import (
	"regexp"
)

type link struct {
	href  string
	title string
}

type Tokens struct {
	tokens []Token
	link   map[string]link
}

type Lex struct {
	tokens    Tokens
	options   *Options
	tokenizer *Tokenizer
}

func New(opt *Options) *Lex {
	lex := new(Lex)
	if opt == nil {
		opt = GetDefaults()
	}

	lex.options = opt
	lex.tokens.link = make(map[string]link)
	if lex.options.tokenizer == nil {
		lex.options.tokenizer = new(Tokenizer)
	}
	lex.tokenizer = lex.options.tokenizer
	lex.tokenizer.options = lex.options

	rules := GetDefaultRules()
	if lex.options.pedantic {

		rules.block = struct{}{}  // block$1.pedantic;
		rules.inline = struct{}{} // inline$1.pedantic

	} else if lex.options.gfm {

		rules.block = struct{}{} // block$1.gfm

		if lex.options.breaks {
			rules.inline = struct{}{} // inline$1.breaks
		} else {
			rules.inline = struct{}{} // inline$1.gfm
		}
	}

	lex.tokenizer.rules = rules

	return lex
}

func (lex *Lex) lex(src string) Tokens {
	rebreak := regexp.MustCompile(`\r\n|\r`)
	retab := regexp.MustCompile(`\t`)
	src = retab.ReplaceAllString(rebreak.ReplaceAllString(src, "\n"), "    ")
	lex.blockTokens(src, lex.tokens, true)
	lex.inline(lex.tokens)
	return lex.tokens
}

func (lex *Lex) blockTokens(src string, tokens Tokens, top bool) {
	rereplace := regexp.MustCompile(`^ +$`)
	src = rereplace.ReplaceAllString(src, "")

	for src != "" {
		token := lex.tokenizer.space(src)
		if token.Type != "" {
			tokens.tokens = append(tokens.tokens, token)
		}
	}

}

func (lex *Lex) inline(tokens Tokens) {

}
