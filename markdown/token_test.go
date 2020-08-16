package markdown

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
