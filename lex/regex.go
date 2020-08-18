package lex

import (
	"github.com/dlclark/regexp2"
)

type Options int32

const (
	None                    = Options(0x0000)
	IgnoreCase              = Options(0x0001) // "i"
	Multiline               = Options(0x0002) // "m"
	ExplicitCapture         = Options(0x0004) // "n"
	Compiled                = Options(0x0008) // "c"
	Singleline              = Options(0x0010) // "s"
	IgnorePatternWhitespace = Options(0x0020) // "x"
	RightToLeft             = Options(0x0040) // "r"
	Debug                   = Options(0x0080) // "d"
	ECMAScript              = Options(0x0100) // "e"
	RE2                     = Options(0x0200) // RE2 (regexp package) compatibility mode
	Global                  = Options(0x0400) // "g"
)

type Regex struct {
	*regexp2.Regexp
	option     Options
	LastIndex  int
	Global     bool
	Multiline  bool
	IgnoreCase bool
	Source     string
}

type Match = regexp2.Match

func MustCompile(source string, opt Options) *Regex {
	reg := Regex{
		option:     opt,
		Global:     opt&Global == Global,
		IgnoreCase: opt&IgnoreCase == IgnoreCase,
		Multiline:  opt&Multiline == Multiline,
	}

	if reg.Global {
		opt = opt - Global
	}

	reg.Regexp = regexp2.MustCompile(source, regexp2.RegexOptions(opt))
	reg.Source = reg.Regexp.String()
	return &reg
}

func (r *Regex) Exec(src []rune) (*Match, error) {
	match, err := r.FindRunesMatchStartingAt(src, r.LastIndex)
	if match == nil || err != nil {
		r.LastIndex = 0
		return nil, err
	}

	if r.Global {
		r.LastIndex = match.Index + match.Length
	}

	return match, nil
}

func (r *Regex) Test(src []rune) bool {
	match, err := r.Exec(src)
	return match != nil && err == nil
}

func (r *Regex) ReplaceRune(input []rune, repl string, startAt, count int) []rune {
	result, err := r.Replace(string(input), repl, startAt, count)
	if err != nil {
		return []rune(input)
	}

	return []rune(result)
}

func (r *Regex) ReplaceStr(input string, repl string, startAt, count int) string {
	result, err := r.Replace(input, repl, startAt, count)
	if err != nil {
		return input
	}

	return result
}
