package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
	"testing"
	"fmt"
)

const (
	TABLE_IG = "IPAexG"
	TABLE_MD = "MPBOLD"
	TABLE_MY = "微软雅黑"
)

func ComplexTableReportWithData() {
	r := core.CreateReport()
	r.IsMutiPage = true
	font1 := core.FontMap{
		FontName: TABLE_IG,
		FileName: "ttf//ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: TABLE_MD,
		FileName: "ttf//mplus-1p-bold.ttf",
	}
	font3 := core.FontMap{
		FontName: TABLE_MY,
		FileName: "ttf//microsoft.ttf",
	}
	fonts := []*core.FontMap{&font1, &font2, &font3}
	r.SetFonts(fonts)
	d := new(TableDetailWithData)
	r.RegisterBand(core.Band(*d), core.Detail)
	r.SetPage("A4", "mm", "P")
	conPt := r.GetUnit()
	r.SetFooterY(265)
	r.SetPageEndY(285.0)
	r.SetPageStartXY(0, 4*conPt)
	r.Execute("table_test_data.pdf")
	r.SaveAtomicCellText("table_test_data.txt")
}

type TableDetailWithData struct {
}

func (h TableDetailWithData) GetHeight(report *core.Report) float64 {
	return 0
}

func (h TableDetailWithData) Execute(report *core.Report) {
	conPt := report.GetUnit()
	report.SetXY(conPt, conPt)

	lineSpace := 0.01 * conPt
	lineHeight := 4 * conPt

	table := NewTable(5, 100, 30*conPt, lineHeight, report)
	table.SetMargin(Scope{2 * conPt, 5 * conPt, 0, 0})

	// 先把当前的行设置完毕, 然后才能添加单元格内容.
	c00 := table.NewCellByRange(1, 1)
	c01 := table.NewCellByRange(2, 1)
	c03 := table.NewCellByRange(2, 2)
	c10 := table.NewCellByRange(3, 1)

	f1 := Font{Family: TABLE_MY, Size: 15}
	c00.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 0), lineHeight, lineSpace, report).SetFont(f1).SetHorizontalCentered().SetContent("0-0"))
	c01.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 1), lineHeight, lineSpace, report).SetFont(f1).SetHorizontalCentered().SetContent("0-1"))
	c03.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 3), lineHeight, lineSpace, report).SetFont(f1).SetVerticalCentered().SetContent("0-3"))
	c10.SetElement(NewDivWithWidth(table.GetColWithByIndex(1, 0), lineHeight, lineSpace, report).SetFont(f1).SetContent("1-0"))

	f1 = Font{Family: TABLE_MY, Size: 10}
	border := Scope{0.5 * conPt, 0.5 * conPt, 0, 0}

	for i := 0; i < 98; i++ {
		cells := make([]*TableCell, 5)
		for j := 0; j < 5; j++ {
			cells[j] = table.NewCell()
		}

		for j := 0; j < 5; j++ {
			str := ``
			s := fmt.Sprintf("%v-%v", i+2, str)
			w := table.GetColWithByIndex(i+2, j)
			e := NewDivWithWidth(w, lineHeight, lineSpace, report)
			e.SetFont(f1)
			e.SetFontColor("178,34,34")
			e.SetBorder(border)
			e.SetContent(s)
			cells[j].SetElement(e)
		}
	}

	table.GenerateAtomicCell()
}

func ComplexTableReport() {
	r := core.CreateReport()
	r.SetPageEndY(281.0)
	//r.SetPageStartXY(2.83, 2.83)
	r.IsMutiPage = true
	font1 := core.FontMap{
		FontName: TABLE_IG,
		FileName: "ttf//ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: TABLE_MD,
		FileName: "ttf//mplus-1p-bold.ttf",
	}
	font3 := core.FontMap{
		FontName: TABLE_MY,
		FileName: "ttf//microsoft.ttf",
	}
	fonts := []*core.FontMap{&font1, &font2, &font3}
	r.SetFonts(fonts)
	d := new(TableDetail)
	r.RegisterBand(core.Band(*d), core.Detail)
	r.SetPage("A4", "mm", "P")
	r.SetFooterY(265)
	r.Execute("table_test.pdf")
	r.SaveAtomicCellText("table_test.txt")
}

type TableDetail struct {
}

func (h TableDetail) GetHeight(report *core.Report) float64 {
	return 6
}
func (h TableDetail) Execute(report *core.Report) {
	conPt := report.GetUnit()
	lineSpace := 0.01 * conPt
	lineHeight := 2 * conPt

	report.SetXY(conPt, conPt)
	table := NewTable(5, 100, 50*conPt, lineHeight, report)
	table.SetMargin(Scope{2 * conPt, 5 * conPt, 0, 0})

	// todo: 先把当前的行设置完毕, 然后才能添加单元格内容.
	c00 := table.NewCellByRange(1, 1)
	c01 := table.NewCellByRange(2, 1)
	c03 := table.NewCellByRange(2, 2)
	c10 := table.NewCellByRange(3, 1)

	f1 := Font{Family: TABLE_MY, Size: 15}
	c00.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 0), lineHeight, lineSpace, report).SetFont(f1).SetContent(""))
	c01.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 1), lineHeight, lineSpace, report).SetFont(f1).SetContent(""))
	c03.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 3), lineHeight, lineSpace, report).SetFont(f1).SetContent(""))
	c10.SetElement(NewDivWithWidth(table.GetColWithByIndex(1, 0), lineHeight, lineSpace, report).SetFont(f1).SetContent(""))

	f1 = Font{Family: TABLE_MY, Size: 10}
	border := Scope{0.5 * conPt, 0.5 * conPt, 0, 0}

	for i := 0; i < 98; i++ {
		cells := make([]*TableCell, 5)
		for j := 0; j < 5; j++ {
			cells[j] = table.NewCell()
		}

		for j := 0; j < 5; j++ {
			w := table.GetColWithByIndex(i+2, j)
			// todo: div执行的严格顺序
			e := NewDivWithWidth(w, lineHeight, lineSpace, report)
			e.SetFont(f1)
			e.SetBorder(border)
			e.SetContent("")
			cells[j].SetElement(e)
		}
	}

	table.GenerateAtomicCell()
}

func TestTableWithdata(t *testing.T) {
	ComplexTableReportWithData()
}

func TestTable(t *testing.T) {
	ComplexTableReport()
}
