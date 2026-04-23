package gopdf

import (
	"os"
	"testing"

	"github.com/tiechui1994/gopdf/core"
)

func ComplexImageReport() {
	r := core.CreateReport()
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(ImageReportExecutor), core.Detail)

	r.Execute("image_test.pdf")
	r.SaveAtomicCellText("image_test.txt")
}
func ImageReportExecutor(report *core.Report) {
	report.Font(core.FontSansBold, 10, "")
	report.SetFont(core.FontSansBold, 10)
	cat := "example//pictures/cat.jpg"
	rand := "example//pictures/rand.jpeg"
	i1 := NewImage(rand, report)
	i1.GenerateAtomicCell()

	report.SetMargin(0, 5)

	i2 := NewImage(cat, report)
	i2.GenerateAtomicCell()

	report.SetMargin(0, 5)

	i3 := NewImageWithWidthAndHeight(cat, 20, 40, report)
	i3.GenerateAtomicCell()

	x, y := report.GetXY()
	report.LineH(x, y+5, x+100)

	x, y = x+0, y+40
	report.Oval(x, y, x+40, y+20)
}

func TestImage(t *testing.T) {
	for _, p := range []string{"example//pictures/rand.jpeg", "example//pictures/cat.jpg"} {
		if _, err := os.Stat(p); err != nil {
			t.Skip("image fixtures not in tree:", p)
		}
	}
	ComplexImageReport()
}
