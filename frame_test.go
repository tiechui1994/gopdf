package gopdf

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/tiechui1994/gopdf/core"
)

const (
	FRAME_IG = "IPAexG"
	FRAME_MD = "MPBOLD"
	FRAME_MY = "微软雅黑"
)

func ManyFrameReportWithData() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: TABLE_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: FRAME_MD,
		FileName: "example//ttf/mplus-1p-bold.ttf",
	}
	font3 := core.FontMap{
		FontName: FRAME_MY,
		FileName: "example//ttf/microsoft.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(ManyFrameReportWithDataExecutor), core.Detail)

	r.Execute("many_frame_data.pdf")
}
func ManyFrameReportWithDataExecutor(report *core.Report) {
	lineSpace := 1.0
	lineHeight := 18.0

	rows, cols := 100, 5
	table := NewFrame(cols, rows, 415, lineHeight, report)
	table.SetMargin(core.Scope{})

	for i := 0; i < rows; i += 5 {
		key := rand.Intn(3)
		//key := (i+1)%2 + 1
		f1 := core.Font{Family: FRAME_MY, Size: 10}
		border := core.NewScope(4.0, 4.0, 0, 0)

		switch key {
		case 0:
			for row := 0; row < 5; row++ {
				for col := 0; col < cols; col++ {
					conent := fmt.Sprintf("%v-(%v,%v)", 0, i+row, col)
					cell := table.NewCell()
					txt := NewTextCell(table.GetColWidth(i+row, col), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
					txt.SetFont(f1).SetBorder(border).SetContent(conent + GetRandStr(1))
					cell.SetElement(txt)
				}
			}

		case 1:
			c00 := table.NewCellByRange(1, 5)
			c01 := table.NewCellByRange(2, 2)
			c03 := table.NewCellByRange(2, 3)
			c21 := table.NewCellByRange(2, 1)
			c31 := table.NewCellByRange(4, 1)
			c41 := table.NewCellByRange(4, 1)

			t00 := NewTextCell(table.GetColWidth(i+0, 0), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t01 := NewTextCell(table.GetColWidth(i+0, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t03 := NewTextCell(table.GetColWidth(i+0, 3), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t21 := NewTextCell(table.GetColWidth(i+2, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t31 := NewTextCell(table.GetColWidth(i+3, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t41 := NewTextCell(table.GetColWidth(i+4, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())

			t00.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 1, i+0, 0) + GetRandStr(5))
			t01.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 1, i+0, 1) + GetRandStr(4))
			t03.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 1, i+0, 3) + GetRandStr(6))
			t21.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 1, i+2, 1) + GetRandStr(2))
			t31.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 1, i+3, 1) + GetRandStr(4))
			t41.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 1, i+4, 1) + GetRandStr(4))

			c00.SetElement(t00)
			c01.SetElement(t01)
			c03.SetElement(t03)
			c21.SetElement(t21)
			c31.SetElement(t31)
			c41.SetElement(t41)

		case 2:
			c00 := table.NewCellByRange(3, 2)
			c03 := table.NewCellByRange(2, 3)
			c20 := table.NewCellByRange(1, 2)
			c21 := table.NewCellByRange(2, 3)
			c33 := table.NewCellByRange(2, 2)
			c40 := table.NewCellByRange(1, 1)

			t00 := NewTextCell(table.GetColWidth(i+0, 0), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t03 := NewTextCell(table.GetColWidth(i+0, 3), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t20 := NewTextCell(table.GetColWidth(i+2, 0), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t21 := NewTextCell(table.GetColWidth(i+2, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t33 := NewTextCell(table.GetColWidth(i+3, 3), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t40 := NewTextCell(table.GetColWidth(i+4, 0), lineHeight, lineSpace, report).SetBackColor(GetRandColor())

			t00.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 2, i+0, 0) + GetRandStr(6))
			t03.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 2, i+0, 3) + GetRandStr(6))
			t20.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 2, i+2, 0) + GetRandStr(2))
			t21.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 2, i+2, 1) + GetRandStr(6))
			t33.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 2, i+3, 3) + GetRandStr(4))
			t40.SetFont(f1).SetBorder(border).SetContent(fmt.Sprintf("%v-(%v,%v)", 2, i+4, 0) + GetRandStr(1))

			c00.SetElement(t00)
			c03.SetElement(t03)
			c20.SetElement(t20)
			c21.SetElement(t21)
			c33.SetElement(t33)
			c40.SetElement(t40)
		}

	}

	table.GenerateAtomicCell()
}

func TestManyFrameReportWithData(t *testing.T) {
	start := time.Now().Unix()
	ManyFrameReportWithData()
	end := time.Now().Unix()
	fmt.Println("i", 1, end-start)
}
