package gopdf

import (
	"testing"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"fmt"
	"log"

	"github.com/tiechui1994/gopdf/core"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}

const (
	MD_IG = "IPAexG"
	MD_MC = "Microsoft"
	MD_MB = "Microsoft Bold"
)

func MdReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: MD_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: MD_MC,
		FileName: "example//ttf/microsoft.ttf",
	}
	font3 := core.FontMap{
		FontName: MD_MB,
		FileName: "example//ttf/microsoft-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(MdReportExecutor), core.Detail)

	r.Execute("markdown_test.pdf")
	r.SaveAtomicCellText("markdown_test.txt")
}

func MdReportExecutor(report *core.Report) {
	data, _ := ioutil.ReadFile("./lex.json")
	var list []Token
	err := json.Unmarshal(data, &list)
	if err != nil {
		return
	}
	var fonts = map[string]string{
		FONT_BOLD:   MD_MB,
		FONT_NORMAL: MD_MC,
		FONT_IALIC:  MD_MC,
	}
	md, _ := NewMarkdownText(report, 0, fonts)
	md.SetTokens(list)
	md.GenerateAtomicCell()
}

func TestMd(t *testing.T) {
	MdReport()
}

func TestTokens(t *testing.T) {
	data, _ := ioutil.ReadFile("./lex.json")
	var list []Token
	err := json.Unmarshal(data, &list)
	if err != nil {
		t.Log(err)
		return
	}

	var buf bytes.Buffer
	encode := json.NewEncoder(&buf)
	encode.SetIndent("", " ")

	for _, val := range list {
		buf.Reset()
		encode.Encode(val)
		fmt.Printf("%v\n", buf.String())
		fmt.Printf("\n")
	}
}
