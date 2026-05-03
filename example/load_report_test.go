package example

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/tiechui1994/gopdf/core"
)

func LoadReport() {
	r := core.CreateReport()
	if err := r.SetPage("A4", "P"); err != nil {
		panic(err)
	}

	dir, _ := filepath.Abs("pictures")
	txtpath := fmt.Sprintf("%v/load.txt", dir)

	r.LoadCellsFromText(txtpath)
	if err := r.Execute("load_report_test.pdf"); err != nil {
		panic(err)
	}
}

func TestLoadReport(t *testing.T) {
	LoadReport()
}
