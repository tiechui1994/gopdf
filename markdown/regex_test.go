package markdown

import (
	"testing"
	"log"
)

func TestRegex_Exec(t *testing.T) {
	src := []rune("215 中华人民工行 926 2315")
	reg := MustCompile(`[0-5]{2}`, RE2|Global)
	m, err := reg.Exec(src)
	log.Println(err, m.GroupByNumber(0), reg.LastIndex)
	m, err = reg.Exec(src)
	log.Println(err, m.GroupByNumber(0), reg.LastIndex)
	m, err = reg.Exec(src)
	log.Println(err, m.GroupByNumber(0), reg.LastIndex)
}
