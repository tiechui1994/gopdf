package gopdf

import (
	"gopdf/core"
	"testing"
	"fmt"
)

const (
	TABLE_IG = "IPAexG"
	TABLE_MD = "MPBOLD"
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
	fonts := []*core.FontMap{&font1, &font2}
	r.SetFonts(fonts)
	d := new(TableDetailWithData)
	r.RegisterBand(core.Band(*d), core.Detail)
	r.SetPage("A4", "mm", "P")
	r.SetFooterY(265)
	r.Execute("table_test_data.pdf")
	r.SaveText("table_test_data.txt")
}

type TableDetailWithData struct {
}

func (h TableDetailWithData) GetHeight(report *core.Report) float64 {
	return 6
}
func (h TableDetailWithData) Execute(report *core.Report) {
	conPt := report.GetConvPt()
	report.CurrX = 1 * conPt
	report.CurrY = 1 * conPt

	table := NewTable(5, 100, 50*conPt, report)
	f1 := Font{Family: TABLE_IG, Size: 9}
	table.SetMargin(2*conPt, 0, 0, 0)

	// 先把当前的行设置完毕, 然后才能添加单元格内容.
	c00 := table.NewCellByRange(1, 1)
	c01 := table.NewCellByRange(2, 1)
	c03 := table.NewCellByRange(2, 2)

	c10 := table.NewCellByRange(3, 1)

	f1 = Font{Family: TABLE_MD, Size: 15}
	c00.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 0), report).SetFont(f1).SetContent("0-0"))
	c01.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 1), report).SetFont(f1).SetContent("0-1"))
	c03.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 3), report).SetFont(f1).SetContent("0-3"))
	c10.SetElement(NewDivWithWidth(table.GetColWithByIndex(1, 0), report).SetFont(f1).SetContent("1-0"))

	f1 = Font{Family: TABLE_IG, Size: 10}
	for i := 0; i < 98; i++ {
		cells := make([]*TableCell, 5)
		for j := 0; j < 5; j++ {
			cells[j] = table.NewCell()
		}

		for j := 0; j < 5; j++ {
			s := fmt.Sprintf("%v-%v", i+2, j)
			w := table.GetColWithByIndex(i+2, j)
			e := NewDivWithWidth(w, report).SetFont(f1).SetContent(s)
			cells[j].SetElement(e)
		}
	}

	fmt.Println(table.cells)
	table.GenerateAtomicCell()
}

func ComplexTableReport() {
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
	fonts := []*core.FontMap{&font1, &font2}
	r.SetFonts(fonts)
	d := new(TableDetail)
	r.RegisterBand(core.Band(*d), core.Detail)
	r.SetPage("A4", "mm", "P")
	r.SetFooterY(265)
	r.Execute("table_test.pdf")
	r.SaveText("table_test.txt")
}

type TableDetail struct {
}

func (h TableDetail) GetHeight(report *core.Report) float64 {
	return 6
}
func (h TableDetail) Execute(report *core.Report) {
	conPt := report.GetConvPt()
	report.Font(TABLE_IG, 9, "")
	report.LineType("straight", 0.01)
	report.GrayStroke(0)
	report.CurrX = 1 * conPt
	report.CurrY = 1 * conPt
	table := NewTable(5, 100, 50*conPt, report)
	table.SetMargin(2*conPt, 0, 5*conPt, 0)

	// 先把当前的行设置完毕, 然后才能添加单元格内容.
	table.NewCellByRange(1, 1)
	table.NewCellByRange(2, 1)
	table.NewCellByRange(2, 2)

	table.NewCellByRange(3, 1)

	for i := 0; i < 98; i++ {
		for j := 0; j < 5; j++ {
			table.NewCell()
		}
	}

	fmt.Println(table.cells)
	table.GenerateAtomicCell()
}

func TestTableWithdata(t *testing.T) {
	ComplexTableReportWithData()
}

//func TestTable(t *testing.T) {
//	ComplexTableReport()
//}
