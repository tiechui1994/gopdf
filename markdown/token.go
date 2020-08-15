package markdown

import (
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
	"fmt"
	"encoding/json"
)

func matchAll(re *regexp2.Regexp, src []rune) ([]*regexp2.Match, error) {
	matches := make([]*regexp2.Match, 0)
	matched, err := re.FindRunesMatch(src)

	for err == nil && matched != nil {
		matches = append(matches, matched)
		matched, err = re.FindNextMatch(matched)
	}

	return matches, err
}

var (
	// block
	block_newline = `^\n+`
	block_code    = `^( {4}[^\n]+\n*)+`

	//^ {0,3}(`{3,}(?=[^`\n]*\n)|~{3,})([^\n]*)\n(?:|([\s\S]*?)\n)(?: {0,3}\1[~`]* *(?:\n+|$)|$)
	block_fences = `^ {0,3}($1{3,}(?=[^$1\n]*\n)|~{3,})([^\n]*)\n(?:|([\s\S]*?)\n)(?: {0,3}\1[~$1]* *(?:\n+|$)|$)`

	block_hr         = `^ {0,3}((?:- *){3,}|(?:_ *){3,}|(?:\* *){3,})(?:\n+|$)`
	block_heading    = `^ {0,3}(#{1,6}) +([^\n]*?)(?: +#+)? *(?:\n+|$)`
	block_blockquote = `^( {0,3}> ?(paragraph|[^\n]*)(?:\n|$))+`
	block_list       = `^( {0,3})(bull) [\s\S]+?(?:hr|def|\n{2,}(?! )(?!\1bull )\n*|\s*$)`
	block_def        = `^ {0,3}\[(label)\]: *\n? *<?([^\s>]+)>?(?:(?: +\n? *| *\n *)(title))? *(?:\n+|$)`
	block_lheading   = `^([^\n]+)\n {0,3}(=+|-+) *(?:\n+|$)`
	block__paragraph = `^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list)[^\n]+)*)`
	block_text       = `^[^\n]+`

	block__label = `(?!\s*\])(?:\\[\[\]]|[^\[\]])+`
	block__title = `(?:"(?:\\"?|[^"\\])*"|'[^'\n]*(?:\n[^'\n]+)*\n?'|\([^()]*\))`

	block_bullet  = `(?:[*+-]|\d{1,9}[.)])`
	block_item    = `^( *)(bull) ?[^\n]*(?:\n(?!\1bull ?)[^\n]*)*`
	block_comment = `<!--(?!-?>)[\s\S]*?-->`

	// inline
	// ^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])
	inline_escape   = `^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_$1{|}~])`
	inline_autolink = `^<(scheme:[^\s\x00-\x1f<>]*|email)>`
	inline_tag      = `^comment` +
		`|^</[a-zA-Z][\\w:-]*\\s*>` +
		`|^<[a-zA-Z][\\w-]*(?:attribute)*?\\s*/?>` +
		`|^<\\?[\\s\\S]*?\\?>` +
		`|^<![a-zA-Z]+\\s[\\s\\S]*?>` +
		`|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>`

	inline_link          = `^!?\[(label)\]\(\s*(href)(?:\s+(title))?\s*\)`
	inline_reflink       = `^!?\[(label)\]\[(?!\s*\])((?:\\[\[\]]?|[^\[\]\\])+)\]`
	inline_nolink        = `^!?\[(?!\s*\])((?:\[[^\[\]]*\]|\\[\[\]]|[^\[\]])*)\](?:\[\])?`
	inline_reflinkSearch = `reflink|nolink(?!\\()`

	inline_strong_start  = `^(?:(\*\*(?=[*punctuation]))|\*\*)(?![\s])|__`
	inline_strong_middle = `^\*\*(?:(?:(?!overlapSkip)(?:[^*]|\\\*)|overlapSkip)|\*(?:(?!overlapSkip)(?:[^*]|\\\*)|overlapSkip)*?\*)+?\*\*$|^__(?![\s])((?:(?:(?!overlapSkip)(?:[^_]|\\_)|overlapSkip)|_(?:(?!overlapSkip)(?:[^_]|\\_)|overlapSkip)*?_)+?)__$`
	inline_strong_endast = `[^punctuation\s]\*\*(?!\*)|[punctuation]\*\*(?!\*)(?:(?=[punctuation\s]|$))`
	inline_strong_endund = `[^\s]__(?!_)(?:(?=[punctuation\s])|$)`

	inline_em_start  = `^(?:(\*(?=[punctuation]))|\*)(?![*\s])|_`
	inline_em_middle = `^\*(?:(?:(?!overlapSkip)(?:[^*]|\\\*)|overlapSkip)|\*(?:(?!overlapSkip)(?:[^*]|\\\*)|overlapSkip)*?\*)+?\*$|^_(?![_\s])(?:(?:(?!overlapSkip)(?:[^_]|\\_)|overlapSkip)|_(?:(?!overlapSkip)(?:[^_]|\\_)|overlapSkip)*?_)+?_$`
	inline_em_endast = `[^punctuation\s]\*(?!\*)|[punctuation]\*(?!\*)(?:(?=[punctuation\s]|$))`
	inline_em_endund = `[^\s]_(?!_)(?:(?=[punctuation\s])|$)`

	// ^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)
	inline_code = `^($1+)([^$1]|[^$1][\s\S]*?[^$1])\1(?!$1)`
	inline_br   = `^( {2,}|\\)\n(?!\s*$)`

	// ^(`+|[^`])(?:[\s\S]*?(?:(?=[\\<!\[`*]|\b_|$)|[^ ](?= {2,}\n))|(?= {2,}\n))
	inline_text        = `^($1+|[^$1])(?:[\s\S]*?(?:(?=[\\<!\[$1*]|\b_|$)|[^ ](?= {2,}\n))|(?= {2,}\n))`
	inline_punctuation = `^([\s*punctuation])`

	// !"#$%&\'()+\\-.,/:;<=>?@\\[\\]`^{|}~
	inline__punctuation = `!"#$%&\'()+\\-.,/:;<=>?@\\[\\]$1^{|}~`

	// \\[[^\\]]*?\\]\\([^\\)]*?\\)|`[^`]*?`|<[^>]*?>
	inline_blockSkip   = `\\[[^\\]]*?\\]\\([^\\)]*?\\)|$1[^$1]*?$1|<[^>]*?>`
	inline_overlapSkip = `__[^_]*?__|\\*\\*\\[^\\*\\]*?\\*\\*`

	// \\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])
	inline__escape = `\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_$1{|}~])`
	inline_scheme  = `[a-zA-Z][a-zA-Z0-9+.-]{1,31}`

	// [a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])
	inline_email = `[a-zA-Z0-9.!#$%&'*+/=?^_$1{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])`

	inline_attribute = `\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>$1]+)?`

	// (?:\[(?:\\.|[^\[\]\\])*\]|\\.|`[^`]*`|[^\[\]\\`])*?
	inline_label = `(?:\[(?:\\.|[^\[\]\\])*\]|\\.|$1[^$1]*$1|[^\[\]\\$1])*?`
	inline_href  = `<(?:\\[<>]?|[^\s<>\\])*>|[^\s\x00-\x1f]*`
	inline_title = `"(?:\\"?|[^"\\])*"|'(?:\\'?|[^'\\])*'|\((?:\\\)?|[^)\\])*\)`
)

//gfm: true,
//headerIds: true,
//mangle: true,

type editor struct {
	getRegex func() *regexp2.Regexp
	replace  func(name, value string) *editor
}

func edit(re interface{}, options ...regexp2.RegexOptions) *editor {
	e := new(editor)
	var regex string
	switch re.(type) {
	case *regexp2.Regexp:
		regex = re.(*regexp2.Regexp).String()
	case string:
		regex = re.(string)
	}

	var opt = regexp2.None
	if len(options) > 0 {
		opt = options[0]
	}

	e.getRegex = func() *regexp2.Regexp {
		return regexp2.MustCompile(regex, opt)
	}

	e.replace = func(name, value string) *editor {
		caret := regexp.MustCompile(`(^|[^\[])\^`)
		if caret.MatchString(value) {
			value = caret.ReplaceAllString(value, "$1")
		}
		regex = strings.Replace(regex, name, value, -1)
		return e
	}

	return e
}

var (
	block  map[string]*regexp2.Regexp
	inline map[string]*regexp2.Regexp
)

func InitFunc() {
	block_fences = strings.Replace(block_fences, "$1", "`", -1)
	block = make(map[string]*regexp2.Regexp)
	inline = make(map[string]*regexp2.Regexp)

	option := regexp2.None
	block["newline"] = regexp2.MustCompile(block_newline, option)
	block["code"] = regexp2.MustCompile(block_code, option)
	block["fences"] = regexp2.MustCompile(block_fences, option)
	block["hr"] = regexp2.MustCompile(block_hr, option)
	block["heading"] = regexp2.MustCompile(block_heading, option)
	block["list"] = regexp2.MustCompile(block_list, option)
	block["def"] = regexp2.MustCompile(block_def, option)
	block["lheading"] = regexp2.MustCompile(block_lheading, option)
	block["text"] = regexp2.MustCompile(block_text, option)

	block["_label"] = regexp2.MustCompile(block__label, option)
	block["_title"] = regexp2.MustCompile(block__title, option)
	block["def"] = edit(block["def"]).
		replace("label", block["_label"].String()).
		replace("title", block["_title"].String()).
		getRegex()

	block["item"] = regexp2.MustCompile(block_item, option)
	block["bullet"] = regexp2.MustCompile(block_bullet, option|regexp2.ECMAScript)
	block["item"] = edit(block["item"], regexp2.Multiline|regexp2.RE2).
		replace("bull", block["bullet"].String()).
		getRegex()

	block["list"] = regexp2.MustCompile(block_list, option)
	block["list"] = edit(block["list"]).
		replace("bull", block["bullet"].String()).
		replace("hr", "\\n+(?=\\1?(?:(?:- *){3,}|(?:_ *){3,}|(?:\\* *){3,})(?:\\n+|$))").
		replace("def", "\\n+(?="+block["def"].String()+")").
		getRegex()

	block["_paragraph"] = regexp2.MustCompile(block__paragraph, option)
	block["paragraph"] = edit(block["_paragraph"]).
		replace("hr", block["hr"].String()).
		replace("heading", " {0,3}#{1,6} ").
		replace("|lheading", ""). // setex headings don't interrupt commonmark paragraphs
		replace("blockquote", " {0,3}>").
		replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").
		replace("list", " {0,3}(?:[*+-]|1[.)]) "). // only lists starting from 1 can interrupt
		getRegex()

	block["blockquote"] = regexp2.MustCompile(block_blockquote, option)
	block["blockquote"] = edit(block["blockquote"]).
		replace("paragraph", block["paragraph"].String()).
		getRegex()

}

type Token struct {
	Type string `json:"type"`
	Raw  string `json:"raw"`
	Text string `json:"text"`

	// list
	Ordered bool            `json:"ordered"`
	Start   json.RawMessage `json:"start"`
	Loose   bool            `json:"loose"`
	Task    bool            `json:"task"`
	Items   []Token         `json:"items"`
	Checked json.RawMessage `json:"checked"`

	// heading
	Depth int `json:"depth"`

	// link
	Href  string `json:"href"`
	Title string `json:"title"`

	Tokens []Token `json:"tokens"`
}

func space(src []rune) (token Token, err error) {
	match, err := block["newline"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).Runes()
	if len(raw) > 1 {
		return Token{
			Type: "space",
			Raw:  string(raw),
		}, nil
	}

	return Token{
		Raw: "\n",
	}, nil
}

func code(src []rune, tokens []Token) (token Token, err error) {
	match, err := block["code"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	last := tokens[len(tokens)-1]
	raw := match.GroupByNumber(0).String()
	if last.Type == "paragraph" {
		return Token{
			Raw:  raw,
			Text: strings.TrimRight(raw, " "),
		}, nil
	}

	text, _ := regexp2.MustCompile(`^ {4}`, regexp2.Multiline).Replace(raw, "", 0, -1)
	return Token{
		Type: "code",
		Raw:  raw,
		Text: text,
	}, nil
}

func fences(src []rune) (token Token, err error) {
	match, err := block["fences"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	return Token{
		Type: "code",
		Raw:  raw,
		Text: raw,
	}, nil
}

func heading(src []rune) (token Token, err error) {
	match, err := block["heading"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := match.GroupByNumber(2).String()
	return Token{
		Type:  "heading",
		Raw:   raw,
		Depth: len(match.Captures[1].String()),
		Text:  text,
	}, nil
}

func hr(src []rune) (token Token, err error) {
	match, err := block["hr"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}
	text := match.GroupByNumber(0).String()
	return Token{
		Type: "hr",
		Raw:  text,
	}, nil
}

func blockquote(src []rune) (token Token, err error) {
	match, err := block["blockquote"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	regex := regexp2.MustCompile(`^ *> ?`, regexp2.Multiline)
	text, _ := regex.Replace(raw, "", 0, -1)

	return Token{
		Type: "blockquote",
		Raw:  raw,
		Text: text,
	}, nil
}

func list(src []rune) (token Token, err error) {
	match, err := block["list"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}
	var raw = match.GroupByNumber(0).Runes()
	var bull = match.GroupByNumber(2).Runes()
	isordered := len(bull) > 1
	isparen := bull[len(bull)-1] == ')'

	var start string
	if isordered {
		start = string(bull[0:len(bull)-1])
	}

	list := Token{
		Type:    "list",
		Raw:     string(raw),
		Ordered: isordered,
		Start:   json.RawMessage([]byte(start)),
		Loose:   false,
		Tokens:  []Token{},
	}

	itemmatch, err := matchAll(block["item"], raw)
	if err != nil {
		return token, err
	}

	var next bool
	length := len(itemmatch)
	replace := regexp.MustCompile(`^ *([*+-]|\d+[.)]) *`)
	reloose := regexp2.MustCompile(`\n\n(?!\s*$)`, regexp2.RE2)
	retask := regexp.MustCompile(`^\[[ xX]\] `)
	for i := 0; i < length; i++ {
		item := itemmatch[i].Runes()
		raw := item
		space := len(item)
		item = []rune(replace.ReplaceAllString(string(item), ""))

		index := strings.Index(string(item), "\n ")
		if index != -1 {
			space -= len(item)
			temp, _ := regexp2.MustCompile(`'^ {1,`+string(space)+`}`, regexp2.Multiline).Replace(string(item), "", 0, -1)
			item = []rune(temp)
		}

		if i != length-1 {
			t, _ := block["bullet"].FindRunesMatch(itemmatch[i+1].Runes())
			b := t.GroupByNumber(0).Runes()

			var condiotion bool
			if isordered {
				condiotion = len(b) == 1 || (!isparen && b[len(b)-1] == ')')
			} else {
				condiotion = len(b) > 1
			}

			// todo: do nothing
			if condiotion {
			}
		}

		temp, _ := reloose.MatchRunes(item)
		loose := next || temp
		if i != length-1 {
			next = item[len(item)-1] == '\n'
			if !loose {
				loose = next
			}
		}

		if loose {
			list.Loose = true
		}

		ischecked := "undefined"
		istask := retask.MatchString(string(item))
		if istask {
			ischecked = fmt.Sprintf("%v", item[1] != ' ')
			item = []rune(regexp.MustCompile(`^\[[ xX]\] +`).ReplaceAllString(string(item), ""))
		}

		token := Token{
			Type:  "list_item",
			Raw:   string(raw),
			Text:  string(item),
			Task:  istask,
			Loose: loose,
		}

		if ischecked != "undefined" {
			token.Checked = []byte(fmt.Sprintf("%v", []byte("null")))
		} else {
			token.Checked = json.RawMessage{}
		}

		list.Tokens = append(list.Tokens, token)
	}

	return list, nil
}

func def(src []rune) (token Token, err error) {
	match, err := block["def"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	g3 := match.GroupByNumber(3).Runes()
	if string(g3) != "" {
		g3 = g3[1:len(g3)-1]
	}

	return Token{
		Raw:   match.GroupByNumber(0).String(),
		Href:  match.GroupByNumber(2).String(),
		Title: string(g3),
	}, nil
}

func lheading(src []rune) (token Token, err error) {
	match, err := block["lheading"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	var depth = 2
	text := match.GroupByNumber(2).Runes()
	if text[0] == '=' {
		depth = 1
	}
	return Token{
		Type:  "heading",
		Raw:   match.GroupByNumber(0).String(),
		Text:  string(text),
		Depth: depth,
	}, nil
}

func paragraph(src []rune) (token Token, err error) {
	match, err := block["paragraph"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	text := match.GroupByNumber(1).Runes()
	if text[len(text)-1] == '\n' {
		text = text[0:len(text)-1]
	}

	return Token{
		Type: "paragraph",
		Raw:  match.GroupByNumber(0).String(),
		Text: string(text),
	}, nil
}

func text(src []rune, tokens []Token) (token Token, err error) {
	match, err := block["text"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	last := tokens[len(tokens)-1]
	if last.Type == "text" {
		return Token{
			Raw:  raw,
			Text: raw,
		}, nil
	}

	return Token{
		Type: "text",
		Raw:  raw,
		Text: raw,
	}, nil
}

func escape(src []rune) (token Token, err error) {
	match, err := inline["escape"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := match.GroupByNumber(1).String()

	return Token{
		Type: "escape",
		Raw:  raw,
		Text: text,
	}, nil
}

func link(src []rune) (token Token, err error) {
	match, err := inline["link"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	zero := match.GroupByNumber(0).Runes()
	first := match.GroupByNumber(1).Runes()
	second := match.GroupByNumber(2).Runes()
	third := match.GroupByNumber(3).Runes()
	lastParentIndex := findClosingBracket(second, []rune("()"))
	if lastParentIndex > -1 {
		start := 4
		if strings.IndexRune(string(zero), '!') == 0 {
			start = 5
		}
		linkLen := start + len(first) + lastParentIndex
		second = second[:lastParentIndex]
		zero = []rune(strings.TrimSpace(string(zero[:linkLen])))
		third = []rune("")
	}

	href := strings.TrimSpace(string(second))
	title := ""
	if len(third) > 0 {
		title = string(third[1:])
	}
	re := regexp2.MustCompile(`^<([\s\S]*)>$`, regexp2.RE2)
	href, _ = re.Replace(href, "$1", 0, -1)

	if href != "" {
		href, _ = inline["_escapes"].Replace(href, "$1", 0, -1)
	}
	if title != "" {
		title, _ = inline["_escapes"].Replace(title, "$1", 0, -1)
	}
	token = Token{
		Href:  href,
		Title: title,
	}
	return outputLink(match, token, string(zero)), nil
}

func reflink(src []rune, links map[string]Token) (token Token, err error) {
	match, err := inline["reflink"].FindRunesMatch(src)
	if err != nil {
		return token, err
	}
	if match == nil {
		match, err = inline["nolink"].FindRunesMatch(src)
	}

	if err != nil || match == nil {
		return token, err
	}

	zero := match.GroupByNumber(0).Runes()
	first := match.GroupByNumber(1).Runes()
	second := match.GroupByNumber(2).Runes()

	var link string
	if len(second) > 0 {
		link = string(second)
	} else {
		link = string(first)
	}

	re := regexp2.MustCompile(`\s+`, regexp2.RE2)
	link, _ = re.Replace(link, " ", 0, -1)
	link = strings.ToLower(link)
	ltoken, ok := links[link]
	if !ok || ltoken.Href == "" {
		text := string(zero[0])
		return Token{
			Type: "text",
			Raw:  text,
			Text: text,
		}, nil
	}

	return outputLink(match, ltoken, string(zero)), nil
}

func strong(src []rune, markedSrc []rune, preChar string) (token Token, err error) {
	match, err := inline["strong"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	punctaute, _ := inline["punctuation"].FindRunesMatch([]rune(preChar))

	first := match.GroupByNumber(1).Runes()
	if len(first) == 0 || len(first) > 0 && preChar == "" || punctaute != nil {
		index := len(markedSrc) - len(src)
		markedSrc = markedSrc[index:]

		zero := match.GroupByNumber(0).String()
		var endReg *regexp2.Regexp
		if zero == "**" {
			endReg = inline["strong_endAst"] // endAst
		} else {
			endReg = inline["strong_endUnd"] // endUnd
		}

		match, err = endReg.FindRunesMatch(markedSrc)
		for match != nil {
			text := markedSrc[0:match.Index+3]
			strongMatch, _ := inline["strong_middle"].FindRunesMatch(text)
			if strongMatch != nil {
				zero := strongMatch.GroupByNumber(0).Runes()
				return Token{
					Type: "strong",
					Raw:  string(src[0:len(zero)]),
					Text: string(src[2:len(zero)-2]),
				}, nil
			}

			match, err = endReg.FindRunesMatch(markedSrc)
		}
	}

	return token, nil
}

func em(src []rune, markedSrc []rune, preChar string) (token Token, err error) {
	match, err := inline["em"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	punctaute, _ := inline["punctuation"].FindRunesMatch([]rune(preChar))
	first := match.GroupByNumber(1).Runes()
	if len(first) == 0 || len(first) > 0 && preChar == "" || punctaute != nil {
		index := len(markedSrc) - len(src)
		markedSrc = markedSrc[index:]

		zero := match.GroupByNumber(0).String()
		var endReg *regexp2.Regexp
		if zero == "*" {
			endReg = inline["em_endAst"] // endAst
		} else {
			endReg = inline["em_endUnd"] // endUnd
		}

		match, err = endReg.FindRunesMatch(markedSrc)
		for match != nil {
			text := markedSrc[0:match.Index+2]
			strongMatch, _ := inline["em_middle"].FindRunesMatch(text)
			if strongMatch != nil {
				zero := strongMatch.GroupByNumber(0).Runes()
				return Token{
					Type: "em",
					Raw:  string(src[0:len(zero)]),
					Text: string(src[1:len(zero)-1]),
				}, nil
			}

			match, err = endReg.FindRunesMatch(markedSrc)
		}
	}

	return token, nil
}

func codespan(src []rune) (token Token, err error) {
	match, err := inline["code"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := match.GroupByNumber(2).String()

	re := regexp2.MustCompile(`\n`, regexp2.RE2)
	text, _ = re.Replace(text, " ", 0, -1)

	reHasNonSpaceChars := regexp2.MustCompile(`[^ ]`, regexp2.RE2)
	hasNonSpaceChars, _ := reHasNonSpaceChars.MatchString(text)
	hasSpaceCharsOnBothEnds := strings.HasPrefix(text, " ") && strings.HasSuffix(text, " ")

	if hasNonSpaceChars && hasSpaceCharsOnBothEnds {
		text = text[1:len(text)-1]
	}

	return Token{
		Type: "codespan",
		Raw:  raw,
		Text: text,
	}, nil
}

func br(src []rune) (token Token, err error) {
	match, err := inline["br"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	return Token{
		Type: "br",
		Raw:  raw,
	}, nil
}

func del(src []rune) (token Token, err error) {
	match, err := inline["del"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := match.GroupByNumber(1).String()
	return Token{
		Type: "del",
		Raw:  raw,
		Text: text,
	}, nil
}

func autoLink(src []rune) (token Token, err error) {
	match, err := inline["autolink"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	second := match.GroupByNumber(2).String()

	text := match.GroupByNumber(1).String()
	var href string
	if second == "@" {
		href = "mailto:" + text
	} else {
		href = text
	}

	return Token{
		Type: "link",
		Raw:  match.GroupByNumber(0).String(),
		Text: text,
		Href: href,
		Tokens: []Token{
			{
				Type: "text",
				Raw:  text,
				Text: text,
			},
		},
	}, nil
}

func url(src []rune) (token Token, err error) {
	match, err := inline["url"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	second := match.GroupByNumber(2).String()
	text := match.GroupByNumber(1).String()
	var href string
	if second == "@" {
		href = "mailto:" + text
	} else {
		var preCapZero, zeroCap string
		preCapZero = match.GroupByNumber(0).String()
		zeroMatch, _ := inline["_backpedal"].FindRunesMatch([]rune(preCapZero))
		if zeroMatch != nil {
			zeroCap = zeroMatch.GroupByNumber(0).String()
		}
		for preCapZero != zeroCap {
			preCapZero = zeroCap
			zeroMatch, _ = inline["_backpedal"].FindRunesMatch([]rune(preCapZero))
			if zeroMatch != nil {
				zeroCap = zeroMatch.GroupByNumber(0).String()
			} else {
				zeroCap = ""
			}
		}

		text = zeroCap
		if match.GroupByNumber(1).String() == "www." {
			href = "https://" + text
		} else {
			href = text
		}
	}

	return Token{
		Type: "link",
		Text: text,
		Href: href,
		Tokens: []Token{
			{
				Type: "text",
				Raw:  text,
				Text: text,
			},
		},
	}, nil
}

func inlineText(src []rune) (token Token, err error) {
	match, err := inline["text"].FindRunesMatch(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	return Token{
		Type: "text",
		Raw:  raw,
		Text: raw,
	}, nil
}
func findClosingBracket(str []rune, b []rune) int {
	if strings.Index(string(str), string(b)) == -1 {
		return -1
	}

	l := len(str)
	level := 0
	for i := 0; i < l; i++ {
		if str[i] == '\\' {
			i++
		} else if str[i] == b[0] {
			level++
		} else if str[i] == b[1] {
			level--
			if level < 0 {
				return i
			}
		}
	}

	return -1
}

func outputLink(match *regexp2.Match, link Token, raw string) Token {
	href := link.Href
	title := link.Title

	re := regexp2.MustCompile(`\\([\[\]])`, regexp2.RE2)
	text, _ := re.Replace(match.GroupByNumber(1).String(), "$1", 0, -1)

	zero := match.GroupByNumber(0).Runes()
	if zero[0] != '!' {
		return Token{
			Type:  "link",
			Raw:   raw,
			Href:  href,
			Title: title,
			Text:  text,
		}
	}

	return Token{
		Type:  "image",
		Raw:   raw,
		Href:  href,
		Title: title,
		Text:  text,
	}
}

func PreProccesText(text string) {
	re_break := regexp2.MustCompile(`\r\n|\r`, regexp2.None)
	text, _ = re_break.Replace(text, "", 0, -1)
	re_blank := regexp2.MustCompile(`\t`, regexp2.None)
	text, _ = re_blank.Replace(text, "    ", 0, -1)

	re_blank = regexp2.MustCompile(`^ +$`, regexp2.Multiline)
	text, _ = re_blank.Replace(text, "", 0, -1)

	src := []rune(text)
	for len(src) != 0 {
		_ = src
	}
}
