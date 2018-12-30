package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
	"testing"
)

const (
	IMAGE_IG = "IPAexG"
	IMAGE_MD = "MPBOLD"
)

func ComplexImageReport() {
	r := core.CreateReport()
	r.SetPageEndY(281.0)
	r.SetPageStartXY(2.83, 2.83)
	r.IsMutiPage = true
	font1 := core.FontMap{
		FontName: IMAGE_IG,
		FileName: "ttf//ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: IMAGE_MD,
		FileName: "ttf//mplus-1p-bold.ttf",
	}
	fonts := []*core.FontMap{&font1, &font2}
	r.SetFonts(fonts)
	d := new(ImageDetail)
	r.RegisterBand(core.Band(*d), core.Detail)

	r.SetPage("A4", "mm", "P")
	r.SetFooterY(265)
	r.Execute("image_test.pdf")
	r.SaveText("image_test.txt")
}

type ImageDetail struct {
}

func (h ImageDetail) GetHeight(report *core.Report) float64 {
	return 6
}
func (h ImageDetail) Execute(report *core.Report) {
	//conPt := report.GetConvPt()
	report.SetXY(10, 10)

	report.Font(DIV_MD, 10, "")
	report.SetFont(DIV_MD, 10)
	cat := "example//pictures/cat.jpg"
	rand := "example//pictures/rand.png"
	i1 := NewImage(20, 40, rand, report)
	i1.GenerateAtomicCell()
	i2 := NewImageWithOutCompress(cat, report)
	i2.GenerateAtomicCell()
}

func TestImage(t *testing.T) {
	ComplexImageReport()
}
