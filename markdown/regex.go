package markdown

import "github.com/dlclark/regexp2"

type Option int32

const (
	None                    = Option(0x0000)
	IgnoreCase              = Option(0x0001) // "i"
	Multiline               = Option(0x0002) // "m"
	ExplicitCapture         = Option(0x0004) // "n"
	Compiled                = Option(0x0008) // "c"
	Singleline              = Option(0x0010) // "s"
	IgnorePatternWhitespace = Option(0x0020) // "x"
	RightToLeft             = Option(0x0040) // "r"
	Debug                   = Option(0x0080) // "d"
	ECMAScript              = Option(0x0100) // "e"
	RE2                     = Option(0x0200) // RE2 (regexp package) compatibility mode
	Global                  = Option(0x0400) // "g"
)

type Regex struct {
	*regexp2.Regexp
	option     Option
	LastIndex  int
	Global     bool
	Multiline  bool
	IgnoreCase bool
	Source     string
}
