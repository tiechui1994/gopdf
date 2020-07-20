package markdown

import "regexp"

type Token struct {
	Depth  int
	Raw    string
	Text   string
	Type   string
	Href   string
	Title  string
	Tokens []Token
}

func outputLink(cap []string, link struct {
	href  string
	title string
}, raw string) Token {
	href := link.href
	title := link.title
	if title != "" {
		title = Escape(title, false)
	} else {
		title = ""
	}

	text := Str(cap[1]).replace(regexp.MustCompile(`\\([\[\]])`), "$1")
	if cap[0][0] != '!' {
		return Token{
			Type:  "image",
			Raw:   raw,
			Href:  href,
			Title: title,
			Text:  text,
		}
	} else {
		return Token{
			Type:  "link",
			Raw:   raw,
			Href:  href,
			Title: title,
			Text:  Escape(text, false),
		}
	}
}

type Tokenizer struct {
	options *Options
	rules   *Rule
}

func NewTokenizer(opt *Options) *Tokenizer {
	var option *Options
	if opt == nil {
		option = GetDefaults()
	} else {
		option = opt
	}
	return &Tokenizer{
		options: option,
	}
}
