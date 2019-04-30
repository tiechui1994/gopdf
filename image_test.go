package gopdf

import (
	"testing"

	"github.com/tiechui1994/gopdf/core"
)

const (
	IMAGE_IG = "IPAexG"
	IMAGE_MD = "MPBOLD"
)

func ComplexImageReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: IMAGE_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: IMAGE_MD,
		FileName: "example//ttf/mplus-1p-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2})
	r.SetPage("A4", "mm", "P")

	r.RegisterExecutor(core.Executor(ImageReportExecutor), core.Detail)

	r.Execute("image_test.pdf")
	r.SaveAtomicCellText("image_test.txt")
}
func ImageReportExecutor(report *core.Report) {
	report.Font(DIV_MD, 10, "")
	report.SetFont(DIV_MD, 10)
	cat := "example//pictures/cat.jpg"
	rand := "example//pictures/rand.jpeg"
	i1 := NewImage(rand, report)
	i1.GenerateAtomicCell()
	i2 := NewImage(cat, report)
	i2.GenerateAtomicCell()
	i3 := NewImageWithWidthAndHeight(cat, 20, 40, report)
	i3.GenerateAtomicCell()
	//DrawPNG("example//pictures/random.png")

	report.Font(DIV_MD, 30, "")
	report.SetFont(DIV_MD, 30)
	x, y := report.GetXY()
	report.CellGray(x, y+10, `ABCDEFGHIJKLMNOPQRS`, 0.85)
}

func TestImage(t *testing.T) {
	ComplexImageReport()
}
