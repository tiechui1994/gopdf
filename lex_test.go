package gopdf

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
