package markdown

import (
	"testing"
	"strings"
	"github.com/dlclark/regexp2"
	"encoding/json"
	"io/ioutil"
)

func TestString(t *testing.T) {
	var x = "中华人民共和[国"
	idx1 := strings.LastIndex(x, "[")
	idx2 := strings.LastIndexFunc(x, func(r rune) bool {
		return r == '['
	})

	t.Log(idx1, idx2)
}

func TestRegex(t *testing.T) {
	str := `^([\s*!"#$%&'()+\-.,/:;<=>?@\[\]$1{|}~])`
	str = strings.ReplaceAll(str, "$1", "`", )
	re := regexp2.MustCompile(str, regexp2.RE2)
	t.Log(re.String())

	rex := `!?\[((?:\[(?:\\.|[^\[\]\\])*\]|\\.|$1[^$1]*$1|[^\[\]\\$1])*?)\]\[(?!\s*\])((?:\\[\[\]]?|[^\[\]\\])+)\]|!?\[(?!\s*\])((?:\[[^\[\]]*\]|\\[\[\]]|[^\[\]])*)\](?:\[\])?(?!\()`
	rex = strings.ReplaceAll(rex, "$1", "`")

	re = regexp2.MustCompile(rex, regexp2.RE2)
	t.Log(re.String())
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
