package markdown

import (
	"testing"
	"regexp"
)

func TestStr(t *testing.T) {
	var s = "([["
	t.Log(Str(s).replace(regexp.MustCompile(`\\([\[\]])`), "$1"))
}
