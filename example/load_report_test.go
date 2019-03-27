package example

import (
	"fmt"
	"github.com/tiechui1994/gopdf/core"
	"path/filepath"
	"testing"
)

func LoadReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: FONT_MY,
		FileName: "ttf//microsoft.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1})
	r.SetPage("A4", "mm", "P")

	dir, _ := filepath.Abs("pictures")
	txtpath := fmt.Sprintf("%v/load.txt", dir)

	r.LoadCellsFromText(txtpath)
	r.Execute("load_report_test.pdf")
}

func TestLoadReport(t *testing.T) {
	LoadReport()
}
