package lex

import (
	"strings"
	"fmt"
	"encoding/json"
	"bytes"
	"regexp"
)

type Token struct {
	Type string `json:"type"`
	Raw  string `json:"raw"`
	Text string `json:"text"`

	// list
	Ordered bool    `json:"ordered,omitempty"`
	Start   string  `json:"start,omitempty"`
	Loose   bool    `json:"loose,omitempty"`
	Task    bool    `json:"task,omitempty"`
	Items   []Token `json:"items,omitempty"`
	Checked string  `json:"checked,omitempty"`

	// heading
	Depth int `json:"depth,omitempty"`

	// link
	Href  string `json:"href,omitempty"`
	Title string `json:"title,omitempty"`

	// def
	Tag string `json:"tag,omitempty"`

	Tokens []Token `json:"tokens,omitempty"`

	// table
	Header []string   `json:"header,omitempty"`
	Align  []string   `json:"align,omitempty"`
	Cells  [][]string `json:"cells,omitempty"`
	Elements struct {
		Header [][]Token   `json:"header,omitempty"`
		Cells  [][][]Token `json:"cells,omitempty"`
	} `json:"elements,omitempty"`
}

func (t Token) String() string {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.Encode(t)
	return buf.String()
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

	block_bullet   = `(?:[*+-]|\d{1,9}[.)])`
	block_item     = `^( *)(bull) ?[^\n]*(?:\n(?!\1bull ?)[^\n]*)*`
	block__comment = `<!--(?!-?>)[\s\S]*?-->`

	block_gfm_nptable = `^ *([^|\n ].*\|.*)\n` + // Header
		` *([-:]+ *\|[-| :]*)` + // Algin
		`(?:\n((?:(?!\n|hr|heading|blockquote|code|fences|list).*(?:\n|$))*)\n*|$)` // Cells

	block_gfm_table = `^ *\|(.+)\n` + // Header
		` *\|?( *[-:]+[-| :]*)` + // Align
		`(?:\n *((?:(?!\n|hr|heading|blockquote|code|fences|list).*(?:\n|$))*)\n*|$)` // Cells

	// inline
	// ^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])
	inline_escape   = `^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_$1{|}~])`
	inline_autolink = `^<(scheme:[^\s\x00-\x1f<>]*|email)>`
	inline_tag      = `^comment` +
		`|^</[a-zA-Z][\w:-]*\s*>` +
		`|^<[a-zA-Z][\w-]*(?:attribute)*?\s*/?>` +
		`|^<\?[\s\S]*?\?>` +
		`|^<![a-zA-Z]+\s[\s\S]*?>` +
		`|^<!\[CDATA\[[\s\S]*?\]\]>`

	inline_link          = `^!?\[(label)\]\(\s*(href)(?:\s+(title))?\s*\)`
	inline_reflink       = `^!?\[(label)\]\[(?!\s*\])((?:\\[\[\]]?|[^\[\]\\])+)\]`
	inline_nolink        = `^!?\[(?!\s*\])((?:\[[^\[\]]*\]|\\[\[\]]|[^\[\]])*)\](?:\[\])?`
	inline_reflinkSearch = `reflink|nolink(?!\()`

	inline_strong_start  = `^(?:(\*\*(?=[*punctuation]))|\*\*)(?![\s])|__`
	inline_strong_middle = `^\*\*(?:(?:(?!overlapSkip)(?:[^*]|\\\*)|overlapSkip)|\*(?:(?!overlapSkip)(?:[^*]|\\\*)|overlapSkip)*?\*)+?\*\*$|^__(?![\s])((?:(?:(?!overlapSkip)(?:[^_]|\\_)|overlapSkip)|_(?:(?!overlapSkip)(?:[^_]|\\_)|overlapSkip)*?_)+?)__$`
	inline_strong_endAst = `[^punctuation\s]\*\*(?!\*)|[punctuation]\*\*(?!\*)(?:(?=[punctuation\s]|$))`
	inline_strong_endUnd = `[^\s]__(?!_)(?:(?=[punctuation\s])|$)`

	inline_em_start  = `^(?:(\*(?=[punctuation]))|\*)(?![*\s])|_`
	inline_em_middle = `^\*(?:(?:(?!overlapSkip)(?:[^*]|\\\*)|overlapSkip)|\*(?:(?!overlapSkip)(?:[^*]|\\\*)|overlapSkip)*?\*)+?\*$|^_(?![_\s])(?:(?:(?!overlapSkip)(?:[^_]|\\_)|overlapSkip)|_(?:(?!overlapSkip)(?:[^_]|\\_)|overlapSkip)*?_)+?_$`
	inline_em_endAst = `[^punctuation\s]\*(?!\*)|[punctuation]\*(?!\*)(?:(?=[punctuation\s]|$))`
	inline_em_endUnd = `[^\s]_(?!_)(?:(?=[punctuation\s])|$)`

	// ^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)
	inline_code = `^($1+)([^$1]|[^$1][\s\S]*?[^$1])\1(?!$1)`
	inline_br   = `^( {2,}|\\)\n(?!\s*$)`

	// ^(`+|[^`])(?:[\s\S]*?(?:(?=[\\<!\[`*]|\b_|$)|[^ ](?= {2,}\n))|(?= {2,}\n))
	inline_text        = `^($1+|[^$1])(?:[\s\S]*?(?:(?=[\\<!\[$1*]|\b_|$)|[^ ](?= {2,}\n))|(?= {2,}\n))`
	inline_punctuation = `^([\s*punctuation])`

	// !"#$%&\'()+\-.,/:;<=>?@\[\]`^{|}~
	inline__punctuation = `!"#$%&'()+\-.,/:;<=>?@\[\]$1^{|}~`

	// \\[[^\\]]*?\\]\\([^\\)]*?\\)|`[^`]*?`|<[^>]*?>
	inline__blockSkip   = `\[[^\]]*?\]\([^\)]*?\)|$1[^$1]*?$1|<[^>]*?>`
	inline__overlapSkip = `__[^_]*?__|\*\*\[^\*\]*?\*\*`

	// \\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])
	inline__escapes = `\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_$1{|}~])`
	inline__scheme  = `[a-zA-Z][a-zA-Z0-9+.-]{1,31}`

	// [a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])
	inline__email = `[a-zA-Z0-9.!#$%&'*+/=?^_$1{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])`

	// \s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?
	inline__attribute = `\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>$1]+)?`

	// (?:\[(?:\\.|[^\[\]\\])*\]|\\.|`[^`]*`|[^\[\]\\`])*?
	inline__label = `(?:\[(?:\\.|[^\[\]\\])*\]|\\.|$1[^$1]*$1|[^\[\]\\$1])*?`
	inline__href  = `<(?:\\[<>]?|[^\s<>\\])*>|[^\s\x00-\x1f]*`
	inline__title = `"(?:\\"?|[^"\\])*"|'(?:\\'?|[^'\\])*'|\((?:\\\)?|[^)\\])*\)`

	inline_gfm__extended_email = `[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![-_])`
	inline_gfm_url             = `^((?:ftp|https?):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email`
	inline_gfm__backpedal      = `(?:[^?!.,:;*_~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_~)]+(?!$))+`
	inline_gfm_del             = `^~+(?=\S)([\s\S]*?\S)~+`

	// ^(`+|[^`])(?:[\s\S]*?(?:(?=[\\<!\[`*~]|\b_|https?:\/\/|ftp:\/\/|www\.|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))|(?= {2,}\n|[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))
	inline_gfm_text = `^($1+|[^$1])(?:[\s\S]*?(?:(?=[\<!\[$1*~]|\b_|https?:\/\/|ftp:\/\/|www\.|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9.!#$%&'*+\/=?_$1{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_$1{\|}~-]+@))|(?= {2,}\n|[a-zA-Z0-9.!#$%&'*+\/=?_$1{\|}~-]+@))`
)

//gfm: true,
//headerIds: true,
//mangle: true,

var (
	block  map[string]*Regex
	inline map[string]*Regex
)

func initBlock() {
	block_fences = strings.Replace(block_fences, "$1", "`", -1)
	block = make(map[string]*Regex)

	option := Options(RE2)
	block["newline"] = MustCompile(block_newline, option)
	block["code"] = MustCompile(block_code, option)
	block["fences"] = MustCompile(block_fences, option)
	block["hr"] = MustCompile(block_hr, option)
	block["heading"] = MustCompile(block_heading, option)
	block["list"] = MustCompile(block_list, option)
	block["def"] = MustCompile(block_def, option)
	block["lheading"] = MustCompile(block_lheading, option)
	block["text"] = MustCompile(block_text, option)

	block["_label"] = MustCompile(block__label, option)
	block["_title"] = MustCompile(block__title, option)
	block["def"] = edit(block["def"]).
		replace("label", block["_label"].String()).
		replace("title", block["_title"].String()).
		getRegex()

	block["item"] = MustCompile(block_item, option)
	block["bullet"] = MustCompile(block_bullet, option)
	block["item"] = edit(block["item"], Global|Multiline).
		replace("bull", block["bullet"].String()).
		getRegex()

	block["list"] = MustCompile(block_list, option)
	block["list"] = edit(block["list"]).
		replace("bull", block["bullet"].String()).
		replace("hr", "\\n+(?=\\1?(?:(?:- *){3,}|(?:_ *){3,}|(?:\\* *){3,})(?:\\n+|$))").
		replace("def", "\\n+(?="+block["def"].String()+")").
		getRegex()

	block["_comment"] = MustCompile(block__comment, option)

	block["_paragraph"] = MustCompile(block__paragraph, option)
	block["paragraph"] = edit(block["_paragraph"]).
		replace("hr", block["hr"].String()).
		replace("heading", " {0,3}#{1,6} ").
		replace("|lheading", ""). // setex headings don't interrupt commonmark paragraphs
		replace("blockquote", " {0,3}>").
		replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").
		replace("list", " {0,3}(?:[*+-]|1[.)]) "). // only lists starting from 1 can interrupt
		getRegex()

	block["blockquote"] = MustCompile(block_blockquote, option)
	block["blockquote"] = edit(block["blockquote"]).
		replace("paragraph", block["paragraph"].String()).
		getRegex()
}

func initLine() {
	inline_escape = strings.ReplaceAll(inline_escape, "$1", "`")
	inline_code = strings.ReplaceAll(inline_code, "$1", "`")
	inline_text = strings.ReplaceAll(inline_text, "$1", "`")
	inline__punctuation = strings.ReplaceAll(inline__punctuation, "$1", "`")
	inline__blockSkip = strings.ReplaceAll(inline__blockSkip, "$1", "`")
	inline__escapes = strings.ReplaceAll(inline__escapes, "$1", "`")
	inline__email = strings.ReplaceAll(inline__email, "$1", "`")
	inline__attribute = strings.ReplaceAll(inline__attribute, "$1", "`")
	inline__label = strings.ReplaceAll(inline__label, "$1", "`")
	inline = make(map[string]*Regex)

	option := Options(RE2)
	inline["escape"] = MustCompile(inline_escape, option)
	inline["autolink"] = MustCompile(inline_autolink, option)
	inline["tag"] = MustCompile(inline_tag, option)
	inline["link"] = MustCompile(inline_link, option)
	inline["reflink"] = MustCompile(inline_reflink, option)
	inline["nolink"] = MustCompile(inline_nolink, option)

	inline["strong_start"] = MustCompile(inline_strong_start, option)
	inline["strong_middle"] = MustCompile(inline_strong_middle, option)
	inline["strong_endAst"] = MustCompile(inline_strong_endAst, option)
	inline["strong_endUnd"] = MustCompile(inline_strong_endUnd, option)

	inline["em_start"] = MustCompile(inline_em_start, option)
	inline["em_middle"] = MustCompile(inline_em_middle, option)
	inline["em_endAst"] = MustCompile(inline_em_endAst, option)
	inline["em_endUnd"] = MustCompile(inline_em_endUnd, option)

	inline["code"] = MustCompile(inline_code, option)
	inline["br"] = MustCompile(inline_br, option)
	inline["text"] = MustCompile(inline_text, option)
	inline["punctuation"] = MustCompile(inline_punctuation, option)

	inline["_punctuation"] = MustCompile(inline__punctuation, option)
	inline["punctuation"] = edit(inline["punctuation"]).
		replace("punctuation", inline__punctuation).
		getRegex()

	inline["_blockSkip"] = MustCompile(inline__blockSkip, option)
	inline["_overlapSkip"] = MustCompile(inline__overlapSkip, option)

	inline["em_start"] = edit(inline["em_start"]).
		replace("punctuation", inline["_punctuation"].String()).
		getRegex()
	inline["em_middle"] = edit(inline["em_middle"]).
		replace("punctuation", inline["_punctuation"].String()).
		replace("overlapSkip", inline["_overlapSkip"].String()).
		getRegex()
	inline["em_endAst"] = edit(inline["em_endAst"], Global).
		replace("punctuation", inline["_punctuation"].String()).
		getRegex()
	inline["em_endUnd"] = edit(inline["em_endUnd"], Global).
		replace("punctuation", inline["_punctuation"].String()).
		getRegex()

	inline["strong_start"] = edit(inline["strong_start"]).
		replace("punctuation", inline["_punctuation"].String()).
		getRegex()
	inline["strong_middle"] = edit(inline["strong_middle"]).
		replace("punctuation", inline["_punctuation"].String()).
		replace("blockSkip", inline["_blockSkip"].String()).
		getRegex()
	inline["strong_endAst"] = edit(inline["strong_endAst"], Global).
		replace("punctuation", inline["_punctuation"].String()).
		getRegex()
	inline["strong_endUnd"] = edit(inline["strong_endUnd"], Global).
		replace("punctuation", inline["_punctuation"].String()).
		getRegex()

	inline["blockSkip"] = edit(inline["_blockSkip"], Global).
		getRegex()
	inline["overlapSkip"] = edit(inline["_overlapSkip"], Global).
		getRegex()

	inline["_escapes"] = MustCompile(inline__escapes, option|Global)
	inline["_scheme"] = MustCompile(inline__scheme, option)
	inline["_email"] = MustCompile(inline__email, option)
	inline["autolink"] = edit(inline["autolink"]).
		replace("scheme", inline["_scheme"].String()).
		replace("email", inline["_email"].String()).
		getRegex()

	inline["_attribute"] = MustCompile(inline__attribute, option)
	inline["tag"] = edit(inline["tag"]).
		replace("comment", block["_comment"].String()).
		replace("attribute", inline["_attribute"].String()).
		getRegex()

	inline["_label"] = MustCompile(inline__label, option)
	inline["_href"] = MustCompile(inline__href, option)
	inline["_title"] = MustCompile(inline__title, option)
	inline["link"] = edit(inline["link"]).
		replace("label", inline["_label"].String()).
		replace("href", inline["_href"].String()).
		replace("title", inline["_title"].String()).
		getRegex()

	inline["reflink"] = edit(inline_reflink).
		replace("label", inline["_label"].String()).
		getRegex()

	inline["reflinkSearch"] = edit(inline_reflinkSearch, Global).
		replace("reflink", inline["reflink"].String()).
		replace("nolink", inline["nolink"].String()).
		getRegex()
}

func initBlockGfm() map[string]*Regex {
	normal := merge(block)
	option := Options(RE2)
	gfm := merge(normal, map[string]*Regex{
		"nptable": MustCompile(block_gfm_nptable, option),
		"table":   MustCompile(block_gfm_table, option),
	})
	gfm["nptable"] = edit(gfm["nptable"]).
		replace("hr", block["hr"].String()).
		replace("heading", " {0,3}#{1,6} ").
		replace("blockquote", " {0,3}>").
		replace("code", " {4}[^\n]").
		replace("fences", " {0,3}(?:`{3,}(?=[^`\n]*\n)|~{3,})[^\n]*\n").
		replace("list", " {0,3}(?:[*+-]|1[.)]) ").
		getRegex()

	gfm["table"] = edit(gfm["table"]).
		replace("hr", block["hr"].String()).
		replace("heading", " {0,3}#{1,6} ").
		replace("blockquote", " {0,3}>").
		replace("code", " {4}[^\n]").
		replace("fences", " {0,3}(?:`{3,}(?=[^`\n]*\n)|~{3,})[^\n]*\n").
		replace("list", " {0,3}(?:[*+-]|1[.)]) ").
		getRegex()

	return gfm
}

func initInlineGfm(breaks bool) map[string]*Regex {
	inline_gfm_text = strings.ReplaceAll(inline_gfm_text, "$1", "`")
	normal := merge(inline)
	option := Options(RE2)
	gfm := merge(normal, map[string]*Regex{
		"escape":          edit(inline["escape"]).replace("])", "~|])").getRegex(),
		"_extended_email": MustCompile(inline_gfm__extended_email, option),
		"url":             MustCompile(inline_gfm_url, option),
		"_backpedal":      MustCompile(inline_gfm__backpedal, option),
		"del":             MustCompile(inline_gfm_del, option),
		"text":            MustCompile(inline_gfm_text, option),
	})

	gfm["url"] = edit(gfm["url"], IgnoreCase).
		replace("email", gfm["_extended_email"].String()).
		getRegex()

	if breaks {
		gfm["br"] = edit(inline["br"]).replace("{2,}", "*").getRegex()
		gfm["text"] = edit(gfm["text"]).
			replace("\b_", "\b_| {2,}\n").
			replace(`\{2,\}`, "*").
			getRegex()
	}

	return gfm
}

func init() {
	initBlock()
	initLine()
}

func space(src []rune) (token Token, err error) {
	match, err := block["newline"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).Runes()
	if len(raw) > 1 {
		return Token{
			Type:   "space",
			Raw:    string(raw),
			Tokens: []Token{},
		}, nil
	}

	return Token{
		Raw:    "\n",
		Tokens: []Token{},
	}, nil
}

func code(src []rune, tokens *[]Token) (token Token, err error) {
	match, err := block["code"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	if len(*tokens) > 0 {
		last := (*tokens)[len(*tokens)-1]
		if last.Type == "paragraph" {
			return Token{
				Raw:    raw,
				Text:   strings.TrimRight(raw, " "),
				Tokens: []Token{},
			}, nil
		}
	}

	text := MustCompile(`^ {4}`, Multiline|Global).ReplaceStr(raw, "", 0, -1)
	return Token{
		Type:   "code",
		Raw:    raw,
		Text:   strings.TrimRight(text, "\n"),
		Tokens: []Token{},
	}, nil
}

func fences(src []rune) (token Token, err error) {
	match, err := block["fences"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	var text string
	if match.GroupCount() > 3 {
		text = match.GroupByNumber(3).String()
	}
	text = indentCodeCompensation(raw, text)
	return Token{
		Type:   "code",
		Raw:    raw,
		Text:   text,
		Tokens: []Token{},
	}, nil
}

func heading(src []rune) (token Token, err error) {
	match, err := block["heading"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := match.GroupByNumber(2).String()
	return Token{
		Type:   "heading",
		Raw:    raw,
		Depth:  len(match.GroupByNumber(1).Runes()),
		Text:   text,
		Tokens: []Token{},
	}, nil
}

func nptable(src []rune) (token Token, err error) {
	matched, err := block["nptable"].Exec(src)
	if err != nil || matched == nil {
		return token, err
	}

	first := matched.GroupByNumber(1).Runes()
	header := str(first).replace(MustCompile(`^ *| *\| *$`, Global), "").string()
	second := matched.GroupByNumber(2).Runes()
	align := str(second).replace(MustCompile(`^ *|\| *$`, Global), "").string()
	third := matched.GroupByNumber(3).Runes()
	var celles = make([]string, 0)
	if len(third) > 0 {
		temp := str(third).replace(MustCompile(`\n$`, RE2), "").string()
		celles = strings.Split(temp, "\n")
	}
	item := Token{
		Type:   "table",
		Header: splitCells(header, 0),
		Align:  regexp.MustCompile(` *\| *`).Split(align, -1),
		Cells:  make([][]string, len(celles)),
		Raw:    matched.GroupByNumber(0).String(),
	}

	if len(item.Header) == len(item.Align) {
		n := len(item.Align)
		for i := 0; i < n; i++ {
			if MustCompile(`^ *-+: *$`, RE2).Test([]rune(item.Align[i])) {
				item.Align[i] = "right"
			} else if MustCompile(`^ *:-+: *$`, RE2).Test([]rune(item.Align[i])) {
				item.Align[i] = "center"
			} else if MustCompile(`^ *:-+ *$`, RE2).Test([]rune(item.Align[i])) {
				item.Align[i] = "left"
			} else {
				item.Align[i] = ""
			}
		}

		n = len(item.Cells)
		for i := 0; i < n; i++ {
			cell := celles[i]
			item.Cells[i] = splitCells(cell, len(item.Header))
		}
		return item, nil
	}

	return token, nil
}

func hr(src []rune) (token Token, err error) {
	match, err := block["hr"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}
	return Token{
		Type:   "hr",
		Raw:    match.GroupByNumber(0).String(),
		Tokens: []Token{},
	}, nil
}

func blockquote(src []rune) (token Token, err error) {
	match, err := block["blockquote"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := MustCompile(`^ *> ?`, Multiline|Global).ReplaceStr(raw, "", 0, -1)
	return Token{
		Type:   "blockquote",
		Raw:    raw,
		Text:   text,
		Tokens: []Token{},
	}, nil
}

func list(src []rune) (token Token, err error) {
	match, err := block["list"].Exec(src)
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
		Start:   string(start),
		Loose:   false,
		Tokens:  []Token{},
		Items:   []Token{},
	}

	itemmatch, err := str(raw).match(block["item"])
	if err != nil {
		return token, err
	}

	var next bool
	length := len(itemmatch)
	replace := MustCompile(`^ *([*+-]|\d+[.)]) *`, RE2)
	reloose := MustCompile(`\n\n(?!\s*$)`, RE2)
	retask := MustCompile(`^\[[ xX]\] `, RE2)
	for i := 0; i < length; i++ {
		item := itemmatch[i].Runes()
		raw := item
		space := len(item)
		item = replace.ReplaceRune(item, "", 0, 1)

		index := str(item).indexOf("\n ")
		if index != -1 {
			space -= len(item)
			item = MustCompile(`'^ {1,`+string(space)+`}`, Multiline|Global).
				ReplaceRune(item, "", 0, -1)
		}

		if i != length-1 {
			t, _ := block["bullet"].Exec(itemmatch[i+1].Runes())
			b := t.GroupByNumber(0).Runes()

			var condiotion bool
			if isordered {
				condiotion = len(b) == 1 || (!isparen && b[len(b)-1] == ')')
			} else {
				condiotion = len(b) > 1
			}

			// todo: do nothing
			if condiotion {
				strs := make([]string, 0)
				for k := i + 1; k < length; k++ {
					strs = append(strs, itemmatch[k].String())
				}
				addBack := []rune(strings.Join(strs, "\n"))
				rawrune := []rune(list.Raw)
				list.Raw = string(rawrune[0:len(rawrune)-len(addBack)])
				i = length - 1
			}
		}

		loose := next || reloose.Test(item)
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
		istask := retask.Test(item)
		if istask {
			ischecked = fmt.Sprintf("%v", item[1] != ' ')
			item = MustCompile(`^\[[ xX]\] +`, RE2).ReplaceRune(item, "", 0, -1)
		}

		token := Token{
			Type:   "list_item",
			Raw:    string(raw),
			Text:   string(item),
			Task:   istask,
			Loose:  loose,
			Tokens: []Token{},
		}

		if ischecked != "undefined" {
			token.Checked = "null"
		} else {
			token.Checked = fmt.Sprintf("%v", ischecked)
		}

		list.Items = append(list.Items, token)
	}

	return list, nil
}

func def(src []rune) (token Token, err error) {
	match, err := block["def"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	third := match.GroupByNumber(3).Runes()
	if string(third) != "" {
		third = str(third).slice(1, len(third)-1)
	}
	first := match.GroupByNumber(1).String()
	tag := MustCompile(`\s+`, Global).ReplaceRune([]rune(strings.ToLower(first)), " ", 0, -1)

	return Token{
		Tag:    string(tag),
		Raw:    match.GroupByNumber(0).String(),
		Href:   match.GroupByNumber(2).String(),
		Title:  string(third),
		Tokens: []Token{},
	}, nil
}

func table(src []rune) (token Token, err error) {
	matched, err := block["table"].Exec(src)
	if err != nil || matched == nil {
		return token, err
	}

	first := matched.GroupByNumber(1).Runes()
	header := str(first).replace(MustCompile(`^ *| *\| *$`, Global), "").string()
	second := matched.GroupByNumber(2).Runes()
	align := str(second).replace(MustCompile(`^ *|\| *$`, Global), "").string()

	third := matched.GroupByNumber(3).Runes()
	var celles = make([]string, 0)
	if len(third) > 0 {
		temp := str(third).replace(MustCompile(`\n$`, RE2), "").string()
		celles = strings.Split(temp, "\n")
	}
	item := Token{
		Type:   "table",
		Header: splitCells(header, 0),
		Align:  regexp.MustCompile(` *\| *`).Split(align, -1),
		Cells:  make([][]string, len(celles)),
	}
	if len(item.Header) == len(item.Align) {
		item.Raw = matched.GroupByNumber(0).String()

		n := len(item.Align)
		for i := 0; i < n; i++ {
			if MustCompile(`^ *-+: *$`, RE2).Test([]rune(item.Align[i])) {
				item.Align[i] = "right"
			} else if MustCompile(`^ *:-+: *$`, RE2).Test([]rune(item.Align[i])) {
				item.Align[i] = "center"
			} else if MustCompile(`^ *:-+ *$`, RE2).Test([]rune(item.Align[i])) {
				item.Align[i] = "left"
			} else {
				item.Align[i] = ""
			}
		}

		n = len(item.Cells)
		for i := 0; i < n; i++ {
			cell := celles[i]
			cell = str(cell).replace(MustCompile(`^ *\| *| *\| *$`, Global), "").string()
			item.Cells[i] = splitCells(cell, len(item.Header))
		}
		return item, nil
	}

	return token, nil
}

func lheading(src []rune) (token Token, err error) {
	match, err := block["lheading"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	var depth = 2
	if match.GroupByNumber(2).Runes()[0] == '=' {
		depth = 1
	}
	return Token{
		Type:   "heading",
		Raw:    match.GroupByNumber(0).String(),
		Text:   match.GroupByNumber(1).String(),
		Depth:  depth,
		Tokens: []Token{},
	}, nil
}

func paragraph(src []rune) (token Token, err error) {
	match, err := block["paragraph"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	text := match.GroupByNumber(1).Runes()
	if text[len(text)-1] == '\n' {
		text = text[0:len(text)-1]
	}

	return Token{
		Type:   "paragraph",
		Raw:    match.GroupByNumber(0).String(),
		Text:   string(text),
		Tokens: []Token{},
	}, nil
}

func text(src []rune, tokens *[]Token) (token Token, err error) {
	match, err := block["text"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	n := len(*tokens)
	if n > 0 && (*tokens)[n-1].Type == "text" {
		return Token{
			Raw:    raw,
			Text:   raw,
			Tokens: []Token{},
		}, nil
	}

	return Token{
		Type:   "text",
		Raw:    raw,
		Text:   raw,
		Tokens: []Token{},
	}, nil
}

func escape(src []rune) (token Token, err error) {
	match, err := inline["escape"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := match.GroupByNumber(1).String()

	return Token{
		Type:   "escape",
		Raw:    raw,
		Text:   text,
		Tokens: []Token{},
	}, nil
}

func link(src []rune) (token Token, err error) {
	match, err := inline["link"].Exec(src)
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
		if str(zero).indexOf("!") == 0 {
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
		title = string(str(third).slice(1, -1))
	}
	href = MustCompile(`^<([\s\S]*)>$`, RE2).ReplaceStr(href, "$1", 0, 1)

	if href != "" {
		href = inline["_escapes"].ReplaceStr(href, "$1", 0, -1)
	}
	if title != "" {
		title = inline["_escapes"].ReplaceStr(title, "$1", 0, -1)
	}
	link := Link{
		Href:  href,
		Title: title,
	}
	return outputLink(match, link, string(zero)), nil
}

func reflink(src []rune, links map[string]Link) (token Token, err error) {
	match, err := inline["reflink"].Exec(src)
	if err != nil {
		return token, err
	}
	if match == nil {
		match, err = inline["nolink"].Exec(src)
	}

	if err != nil || match == nil {
		return token, err
	}

	zero := match.GroupByNumber(0).Runes()
	first := match.GroupByNumber(1).Runes()

	var link string
	if match.GroupCount() == 3 && match.GroupByNumber(2).Length > 0 {
		link = match.GroupByNumber(2).String()
	} else {
		link = string(first)
	}

	link = MustCompile(`\s+`, Global).ReplaceStr(link, " ", 0, -1)
	link = strings.ToLower(link)
	ltoken, ok := links[link]
	if !ok || ltoken.Href == "" {
		text := string(zero[0])
		return Token{
			Type:   "text",
			Raw:    text,
			Text:   text,
			Tokens: []Token{},
		}, nil
	}

	return outputLink(match, ltoken, string(zero)), nil
}

func strong(src []rune, markedSrc []rune, preChar string) (token Token, err error) {
	match, err := inline["strong_start"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	punctaute, _ := inline["punctuation"].Exec([]rune(preChar))

	first := match.GroupByNumber(1).Runes()
	if len(first) == 0 || (len(first) > 0 && (preChar == "" || punctaute != nil)) {
		markedSrc = str(markedSrc).slice(-1 * len(src))

		zero := match.GroupByNumber(0).String()
		var endReg *Regex
		if zero == "**" {
			endReg = inline["strong_endAst"] // endAst
		} else {
			endReg = inline["strong_endUnd"] // endUnd
		}

		endReg.LastIndex = 0
		match, err = endReg.Exec(markedSrc)
		for match != nil {
			text := str(markedSrc).slice(0, match.Index+3)
			strongMatch, _ := inline["strong_middle"].Exec(text)
			if strongMatch != nil {
				zero := strongMatch.GroupByNumber(0).Runes()
				return Token{
					Type:   "strong",
					Raw:    str(src).slice(0, len(zero)).string(),
					Text:   str(src).slice(2, len(zero)-2).string(),
					Tokens: []Token{},
				}, nil
			}

			match, err = endReg.Exec(markedSrc)
		}
	}

	return token, nil
}

func em(src []rune, markedSrc []rune, preChar string) (token Token, err error) {
	match, err := inline["em_start"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	punctaute, _ := inline["punctuation"].Exec([]rune(preChar))
	first := match.GroupByNumber(1).Runes()
	if len(first) == 0 || (len(first) > 0 && (preChar == "" || punctaute != nil)) {
		markedSrc = str(markedSrc).slice(-1 * len(src))

		zero := match.GroupByNumber(0).String()
		var endReg *Regex
		if zero == "*" {
			endReg = inline["em_endAst"] // endAst
		} else {
			endReg = inline["em_endUnd"] // endUnd
		}

		endReg.LastIndex = 0
		match, err = endReg.Exec(markedSrc)
		for match != nil {
			text := markedSrc[0:match.Index+2]
			strongMatch, _ := inline["em_middle"].Exec(text)
			if strongMatch != nil {
				zero := strongMatch.GroupByNumber(0).Runes()
				return Token{
					Type:   "em",
					Raw:    str(src).slice(0, len(zero)).string(),
					Text:   str(src).slice(1, len(zero)-1).string(),
					Tokens: []Token{},
				}, nil
			}

			match, err = endReg.Exec(markedSrc)
		}
	}

	return token, nil
}

func codespan(src []rune) (token Token, err error) {
	match, err := inline["code"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := match.GroupByNumber(2).String()

	re := MustCompile(`\n`, RE2)
	text, _ = re.Replace(text, " ", 0, -1)

	reHasNonSpaceChars := MustCompile(`[^ ]`, RE2)
	hasNonSpaceChars, _ := reHasNonSpaceChars.MatchString(text)
	hasSpaceCharsOnBothEnds := strings.HasPrefix(text, " ") && strings.HasSuffix(text, " ")

	if hasNonSpaceChars && hasSpaceCharsOnBothEnds {
		text = text[1:len(text)-1]
	}

	return Token{
		Type:   "codespan",
		Raw:    raw,
		Text:   text,
		Tokens: []Token{},
	}, nil
}

func br(src []rune) (token Token, err error) {
	match, err := inline["br"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	return Token{
		Type:   "br",
		Raw:    raw,
		Tokens: []Token{},
	}, nil
}

func del(src []rune) (token Token, err error) {
	match, err := inline["del"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	text := match.GroupByNumber(1).String()
	return Token{
		Type:   "del",
		Raw:    raw,
		Text:   text,
		Tokens: []Token{},
	}, nil
}

func autoLink(src []rune) (token Token, err error) {
	match, err := inline["autolink"].Exec(src)
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
	match, err := inline["url"].Exec(src)
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
		zeroMatch, _ := inline["_backpedal"].Exec([]rune(preCapZero))
		if zeroMatch != nil {
			zeroCap = zeroMatch.GroupByNumber(0).String()
		}
		for preCapZero != zeroCap {
			preCapZero = zeroCap
			zeroMatch, _ = inline["_backpedal"].Exec([]rune(preCapZero))
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

// TODO:
func inlineText(src []rune, inRawBlock bool) (token Token, err error) {
	match, err := inline["text"].Exec(src)
	if err != nil || match == nil {
		return token, err
	}

	raw := match.GroupByNumber(0).String()
	return Token{
		Type:   "text",
		Raw:    raw,
		Text:   raw,
		Tokens: []Token{},
	}, nil
}
