package markdown

import (
	"testing"
	"io/ioutil"
	"log"
	"encoding/json"
	"bytes"
	"github.com/dlclark/regexp2"
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

func TestRe(t *testing.T) {
	re := regexp2.MustCompile(`^<([\s\S]*)>$`, regexp2.RE2)
	res, err := re.Replace("<Java,Vivo>", "$1", 0, -1)
	t.Log(res, err)
}
