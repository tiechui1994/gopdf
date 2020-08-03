package example

import (
	"fmt"
	"github.com/tiechui1994/gopdf"
	"github.com/tiechui1994/gopdf/core"
	"testing"
)

const (
	TABLE_IG = "IPAexG"
	TABLE_MD = "MPBOLD"
	TABLE_MY = "微软雅黑"
)

func SimpleTable() {
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
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(SimpleTableExecutor), core.Detail)

	r.Execute("simple_table.pdf")
	fmt.Println(r.GetCurrentPageNo())
}
func SimpleTableExecutor(report *core.Report) {
	lineSpace := 1.0
	lineHeight := 18.0

	table := gopdf.NewTable(5, 100, 415, lineHeight, report)
	table.SetMargin(core.Scope{})

	// 先把当前的行设置完毕, 然后才能添加单元格内容.
	c00 := table.NewCellByRange(1, 1)
	c01 := table.NewCellByRange(2, 1)
	c03 := table.NewCellByRange(2, 2)
	c10 := table.NewCellByRange(3, 1)

	f1 := core.Font{Family: TABLE_MY, Size: 15, Style: ""}
	border := core.NewScope(4.0, 4.0, 4.0, 3.0)
	c00.SetElement(gopdf.NewTextCell(table.GetColWidth(0, 0), lineHeight, lineSpace, report).SetFont(f1).SetBorder(border).HorizontalCentered().SetContent("0-0"))
	c01.SetElement(gopdf.NewTextCell(table.GetColWidth(0, 1), lineHeight, lineSpace, report).SetFont(f1).SetBorder(border).RightAlign().SetContent("0-1"))
	c03.SetElement(gopdf.NewTextCell(table.GetColWidth(0, 3), lineHeight, lineSpace, report).SetFont(f1).SetBorder(border).RightAlign().SetContent("0-3近日，江苏苏州市姑苏区市场监督管理局向苏州定园旅游服务有限公司送达行政处罚决定书，对定园进行处罚，吊销其营业执照，此举开创了我国旅游景点因虚假宣传被吊销营业执照的先河"))
	c10.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 0), lineHeight, lineSpace, report).SetFont(f1).SetBorder(border).VerticalCentered().SetContent("1-0近日，江苏苏州市姑苏区市场监督管理局向苏州定园旅游服务有限公司送达行政处罚决定书了我国旅游景点因虚假宣传被吊销营业执照的先河"))

	f1 = core.Font{Family: TABLE_MY, Size: 10}
	border = core.NewScope(4.0, 4.0, 0, 0)

	for i := 0; i < 98; i++ {
		cells := make([]*gopdf.TableCell, 5)
		for j := 0; j < 5; j++ {
			cells[j] = table.NewCell()
		}

		for j := 0; j < 5; j++ {
			str := `有限公司送达行政处罚决定书`
			s := fmt.Sprintf("%v-%v", i+2, str)
			w := table.GetColWidth(i+2, j)
			e := gopdf.NewTextCell(w, lineHeight, lineSpace, report)
			e.SetFont(f1)
			if i%2 == 0 {
				e.SetBackColor("255,192,203")
			}
			e.SetBorder(border)
			e.SetContent(s)
			cells[j].SetElement(e)
		}
	}

	table.GenerateAtomicCell()
}

func TestSimpleTable(t *testing.T) {
	SimpleTable()
}
