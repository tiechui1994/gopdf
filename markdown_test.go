package gopdf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/lex"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}

func MarkdownReport() {
	r := core.CreateReport()
	if err := r.SetPage("A4", "P"); err != nil {
		panic(err)
	}

	r.RegisterExecutor(core.Executor(MarkdownReportExecutor), core.Detail)

	if err := r.Execute("markdown_test.pdf"); err != nil {
		panic(err)
	}
	if err := r.SaveAtomicCellText("markdown_test.txt"); err != nil {
		panic(err)
	}
}

func writeMarkdownParseTreeJSON(path string, tokens []Token) error {
	b, err := FormatParseTreeJSON(tokens, "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

func MarkdownReportExecutor(report *core.Report) {
	data, _ := ioutil.ReadFile("./markdown.md")
	var lexer = lex.NewLex()
	tokens := lexer.Lex(string(data))
	if err := writeMarkdownParseTreeJSON("markdown_test_ast.json", tokens); err != nil {
		panic(err)
	}
	var fonts = map[string]string{
		FONT_BOLD:   core.FontSansBold,
		FONT_NORMAL: core.FontSans,
		FONT_ITALIC: core.FontSans,
		FONT_MONO:   core.FontSans,
	}
	md, _ := NewMarkdownText(report, 0, fonts)
	md.SetTokens(tokens)
	md.GenerateAtomicCell()
}

// MarkdownReportComplex renders markdown_complex.md to markdown_complex_test.{pdf,txt}
func MarkdownReportComplex() {
	r := core.CreateReport()
	if err := r.SetPage("A4", "P"); err != nil {
		panic(err)
	}
	r.RegisterExecutor(core.Executor(MarkdownReportExecutorComplex), core.Detail)
	if err := r.Execute("markdown_complex_test.pdf"); err != nil {
		panic(err)
	}
	if err := r.SaveAtomicCellText("markdown_complex_test.txt"); err != nil {
		panic(err)
	}
}

func MarkdownReportExecutorComplex(report *core.Report) {
	data, _ := ioutil.ReadFile("./markdown_complex.md")
	var lexer = lex.NewLex()
	tokens := lexer.Lex(string(data))
	if err := writeMarkdownParseTreeJSON("markdown_complex_test_ast.json", tokens); err != nil {
		panic(err)
	}
	var fonts = map[string]string{
		FONT_BOLD:   core.FontSansBold,
		FONT_NORMAL: core.FontSans,
		FONT_ITALIC: core.FontSans,
		FONT_MONO:   core.FontSans,
	}
	md, _ := NewMarkdownText(report, 0, fonts)
	md.SetTokens(tokens)
	md.GenerateAtomicCell()
}

func TestMarkdown(t *testing.T) {
	MarkdownReport()
}

func TestMarkdownComplex(t *testing.T) {
	MarkdownReportComplex()
}

func TestTokens(t *testing.T) {
	data, _ := ioutil.ReadFile("./markdown/src/mark.json")
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

func TestDrawPNG(t *testing.T) {
	DrawPNG("./test.png")
	DrawSunLine("./sunline.png")
	DrawFiveCycle("./fivecycle.png")
}
