package lex

import (
	"testing"
	"io/ioutil"
	"log"
	"encoding/json"
	"bytes"
)

var (
	data []byte
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	InitFunc()
	data, _ = ioutil.ReadFile("./lex.md")
}

func TestList(t *testing.T) {
	text := string(data)
	tokens, err := list([]rune(text))
	if err != nil {
		t.Log(err)
		return
	}

	var buf bytes.Buffer
	encode := json.NewEncoder(&buf)
	encode.SetIndent("", " ")
	encode.Encode(tokens)

	log.Printf("\n%v", buf.String())
}

func TestBlockquote(t *testing.T) {

	text := string(data)
	tokens, err := blockquote([]rune(text))
	if err != nil {
		t.Log(err)
		return
	}

	var buf bytes.Buffer
	encode := json.NewEncoder(&buf)
	encode.SetIndent("", " ")
	encode.Encode(tokens)

	log.Printf("\n%v", buf.String())
}

func TestBlock(t *testing.T) {
	var in = make(map[string]string)
	for k, v := range inline {
		in[k] = v.String()
	}
	data, _ := json.Marshal(in)
	ioutil.WriteFile("./inline.json", data, 0666)

	var bl = make(map[string]string)
	for k, v := range block {
		bl[k] = v.String()
	}
	data, _ = json.Marshal(bl)
	ioutil.WriteFile("./block.json", data, 0666)
}

func TestStr(t *testing.T) {
	s := str("01234567")
	t.Log("[-1]", s.slice(-1).string() == "7")
	t.Log("[-3]", s.slice(-3).string() == "567")
	t.Log("[-8]", s.slice(-8).string() == "01234567")
	t.Log("[-12]", s.slice(-12).string() == "01234567")

	t.Log("[-1,-7]", s.slice(-1, -7).string() == "")
	t.Log("[-1,-12]", s.slice(-1, -12).string() == "")
	t.Log("[-3,-9]", s.slice(-3, -9).string() == "")
	t.Log("[-3,-2]", s.slice(-3, -2).string() == "5")
	t.Log("[-3,-3]", s.slice(-3, -3).string() == "")
	t.Log("[-8,-2]", s.slice(-8, -2).string() == "012345")
	t.Log("[-8,-12]", s.slice(-8, -12).string() == "")
	t.Log("[-9,-12]", s.slice(-9, -12).string() == "")
	t.Log("[-9,-5]", s.slice(-9, -5).string() == "012")
}

func TestReplace(t *testing.T) {
	tableRow := "AA | BB | CC"
	re := MustCompile(`\|`, Global)
	row := str(tableRow).replaceFunc(re, func(match *Match, offset int, s str) str {
		t.Log(match, "-", offset, "-", s.string(), "-", len(s))
		var escaped bool
		curr := offset
		curr--
		for curr >= 0 && s[curr] == '\\' {
			escaped = !escaped
			curr--
		}
		if escaped {
			return str("|")
		} else {
			return str(" |")
		}
	})

	t.Log(string(row))
}
