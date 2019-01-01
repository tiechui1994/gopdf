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
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "mm", "P")

	r.RegisterExecutor(core.Executor(TableReportWithDataExecutor), core.Detail)

	r.Execute("table_test_data.pdf")
	r.SaveAtomicCellText("table_test_data.txt")
}
func TableReportWithDataExecutor(report *core.Report) {
	unit := report.GetUnit()

	lineSpace := 0.01 * unit
	lineHeight := 2 * unit

	table := NewTable(5, 100, 80*unit, lineHeight, report)
	table.SetMargin(Scope{0, 0, 0, 0})

	// 先把当前的行设置完毕, 然后才能添加单元格内容.
	c00 := table.NewCellByRange(1, 1)
	c01 := table.NewCellByRange(2, 1)
	c03 := table.NewCellByRange(2, 2)
	c10 := table.NewCellByRange(3, 1)

	f1 := Font{Family: TABLE_MY, Size: 15, Style: ""}
	c00.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 0), lineHeight, lineSpace, report).SetFont(f1).SetHorizontalCentered().SetContent("0-0"))
	c01.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 1), lineHeight, lineSpace, report).SetFont(f1).SetRightAlign().SetContent("0-1"))
	c03.SetElement(NewDivWithWidth(table.GetColWithByIndex(0, 3), lineHeight, lineSpace, report).SetFont(f1).SetRightAlign().SetContent("0-3-近日，江苏苏州市姑苏区市场监督管理局向苏州定园旅游服务有限公司送达行政处罚决定书，对定园进行处罚，吊销其营业执照，此举开创了我国旅游景点因虚假宣传被吊销营业执照的先河"))
	c10.SetElement(NewDivWithWidth(table.GetColWithByIndex(1, 0), lineHeight, lineSpace, report).SetFont(f1).SetContent("1-0-近日，江苏苏州市姑苏区市场监督管理局向苏州定园旅游服务有限公司送达行政处罚决定书了我国旅游景点因虚假宣传被吊销营业执照的先河"))

	f1 = Font{Family: TABLE_MY, Size: 10}
	border := Scope{0.5 * unit, 0.5 * unit, 0, 0}

	for i := 0; i < 98; i++ {
		cells := make([]*TableCell, 5)
		for j := 0; j < 5; j++ {
			cells[j] = table.NewCell()
		}

		for j := 0; j < 5; j++ {
			str := `有限公司送达行政处罚决定书`
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
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "mm", "P")

	r.RegisterExecutor(core.Executor(TableReportExecutor), core.Detail)

	r.Execute("table_test.pdf")
	r.SaveAtomicCellText("table_test.txt")
}
func TableReportExecutor(report *core.Report) {
	unit := report.GetUnit()
	lineSpace := 0.01 * unit
	lineHeight := 2 * unit

	table := NewTable(5, 100, 50*unit, lineHeight, report)
	table.SetMargin(Scope{0 * unit, 0 * unit, 0, 0})

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
	border := Scope{0.5 * unit, 0.5 * unit, 0, 0}

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
