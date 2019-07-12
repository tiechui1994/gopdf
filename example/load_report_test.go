package example

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/tiechui1994/gopdf/core"
)

func LoadReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: FONT_MY,
		FileName: "ttf//microsoft.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1})
	r.SetPage("A4", "P")

	dir, _ := filepath.Abs("pictures")
	txtpath := fmt.Sprintf("%v/load.txt", dir)

	r.LoadCellsFromText(txtpath)
	r.Execute("load_report_test.pdf")
}

func TestLoadReport(t *testing.T) {
	LoadReport()
}
