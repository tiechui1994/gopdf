package lex

import (
	"testing"
	"log"
	"github.com/dlclark/regexp2"
)

func TestExec(t *testing.T) {
	src := []rune("215 中华人民工行 926 2315")
	reg := MustCompile(`[0-5]{2}`, RE2|Global)
	m, err := reg.Exec(src)
	log.Println(err, m.GroupByNumber(0), reg.LastIndex)
	m, err = reg.Exec(src)
	log.Println(err, m.GroupByNumber(0), reg.LastIndex)
	m, err = reg.Exec(src)
	log.Println(err, m.GroupByNumber(0), reg.LastIndex)
}

func TestMatch1(t *testing.T) {
	src := `123 234 12wx vivo12
google11 34得到 782cc
wwww2121 327qq sss为1`
	reg := MustCompile("([0-9]{1})([a-z]{2})", Global|Multiline|RE2)

	matches, err := str([]rune(src)).match(reg)
	if err == nil {
		for _, m := range matches {
			t.Log(m.String())
		}
	}

	t.Log(reg.LastIndex)
}

func TestMatch2(t *testing.T) {
	src := `123 234 12wx vivo12
google11 34得到 782cc
wwww2121 327qq sss为1`
	reg := MustCompile("([0-9]{1})([a-z]{2})", Multiline|RE2)

	matches, err := str([]rune(src)).match(reg)
	if err == nil {
		for _, m := range matches {
			t.Log(m.String(), m.GroupByNumber(1).String(), m.GroupByNumber(2).String(), m.Index)
		}
	}

	t.Log(reg.LastIndex)
}

func TestRegexGroupReplace(t *testing.T) {
	re := regexp2.MustCompile(`^<([\s\S]*)>$`, regexp2.RE2)
	res, err := re.Replace("<Java,Vivo>", "$1", 0, -1)
	t.Log(res, err)
}
