package example

import (
	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf"
	"time"
	"math/rand"
	"fmt"
	"strings"
	"testing"
)

var (
	seed = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func MutilTable() {
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

	r.RegisterExecutor(core.Executor(MutilTableExecutor), core.Detail)

	r.Execute("mutil_table.pdf")
}
func MutilTableExecutor(report *core.Report) {
	lineSpace := 1.0
	lineHeight := 18.0

	rows, cols := 100, 5
	table := gopdf.NewTable(cols, rows, 415, lineHeight, report)
	table.SetMargin(core.Scope{})

	for i := 0; i < rows; i += 5 {
		key := rand.Intn(3)
		//key := (i+1)%2 + 1
		f1 := core.Font{Family: TABLE_MY, Size: 10}
		border := core.NewScope(4.0, 4.0, 0, 0)

		switch key {
		case 0:
			for row := 0; row < 5; row++ {
				for col := 0; col < cols; col++ {
					conent := fmt.Sprintf("%v-(%v,%v)", 0, i+row, col)
					cell := table.NewCell()
					txt := gopdf.NewTextCell(table.GetColWidth(i+row, col), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
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

			t00 := gopdf.NewTextCell(table.GetColWidth(i+0, 0), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t01 := gopdf.NewTextCell(table.GetColWidth(i+0, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t03 := gopdf.NewTextCell(table.GetColWidth(i+0, 3), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t21 := gopdf.NewTextCell(table.GetColWidth(i+2, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t31 := gopdf.NewTextCell(table.GetColWidth(i+3, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t41 := gopdf.NewTextCell(table.GetColWidth(i+4, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())

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

			t00 := gopdf.NewTextCell(table.GetColWidth(i+0, 0), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t03 := gopdf.NewTextCell(table.GetColWidth(i+0, 3), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t20 := gopdf.NewTextCell(table.GetColWidth(i+2, 0), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t21 := gopdf.NewTextCell(table.GetColWidth(i+2, 1), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t33 := gopdf.NewTextCell(table.GetColWidth(i+3, 3), lineHeight, lineSpace, report).SetBackColor(GetRandColor())
			t40 := gopdf.NewTextCell(table.GetColWidth(i+4, 0), lineHeight, lineSpace, report).SetBackColor(GetRandColor())

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

func GetRandStr(l ...int) string {
	str := "0123456789ABCDEFGHIGKLMNOPQRSTUVWXYZ"
	l = append(l, 8)
	r := seed.Intn(l[0]*11) + 8
	//r := l[0] * 13
	//r = 1200
	data := strings.Repeat(str, r/36+1)
	return data[:r] + "---"
}

func GetRandColor() (color string) {
	r, g, b := seed.Intn(256), seed.Intn(256), seed.Intn(256)
	if float64(r)*0.299+float64(g)*0.578+float64(b)*0.114 >= 192 {
		color = fmt.Sprintf("%v,%v,%v", r, g, b)
		return color
	}

	return GetRandColor()
}

func TestMutilTable(t *testing.T) {
	MutilTable()
}
