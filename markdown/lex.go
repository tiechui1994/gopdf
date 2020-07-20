package markdown

type token struct {
	Depth  int
	Raw    string
	Text   string
	Tokens []token
	Type   string
}

type Lex struct {
	tokens    []token
	options   *Options
	tokenizer *Tokenizer
}

func New(opt *Options) *Lex {
	lex := new(Lex)
	if opt == nil {
		opt = GetDefaults()
	}

	lex.options = opt
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

func (lex *Lex) lex(src string) []token {
	lex.blockTokens(src, lex.tokens, true)
	lex.inline(lex.tokens)
	return lex.tokens
}

func (lex *Lex) blockTokens() {

}

func (lex *Lex) inline() {

}
