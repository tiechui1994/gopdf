package gopdf

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"testing"

	"github.com/tiechui1994/gopdf/core"
)

const (
	TABLE_IG = "IPAexG"
	TABLE_MD = "MPBOLD"
	TABLE_MY = "微软雅黑"
)

var (
	seed = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func ComplexTableReportWithData() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: TABLE_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: TABLE_MD,
		FileName: "example//ttf/mplus-1p-bold.ttf",
	}
	font3 := core.FontMap{
		FontName: TABLE_MY,
		FileName: "example//ttf/microsoft.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(ComplexTableReportWithDataExecutor), core.Detail)

	r.Execute("table_test_data.pdf")
	r.SaveAtomicCellText("table_test_data.txt")
	fmt.Println(r.GetCurrentPageNo())
}
func ComplexTableReportWithDataExecutor(report *core.Report) {
	lineSpace := 1.0
	lineHeight := 18.0

	table := NewTable(5, 100, 415, lineHeight, report)
	table.SetMargin(core.Scope{})

	// 先把当前的行设置完毕, 然后才能添加单元格内容.
	c00 := table.NewCellByRange(1, 1)
	c01 := table.NewCellByRange(2, 1)
	c03 := table.NewCellByRange(2, 2)
	c10 := table.NewCellByRange(3, 1)

	f1 := core.Font{Family: TABLE_MY, Size: 15, Style: ""}
	border := core.NewScope(4.0, 4.0, 4.0, 3.0)
	c00.SetElement(NewTextCell(table.GetColWidth(0, 0), lineHeight, lineSpace, report).SetFont(f1).SetBorder(border).HorizontalCentered().SetContent("0-0"))
	c01.SetElement(NewTextCell(table.GetColWidth(0, 1), lineHeight, lineSpace, report).SetFont(f1).SetBorder(border).RightAlign().SetContent("0-1"))
	c03.SetElement(NewTextCell(table.GetColWidth(0, 3), lineHeight, lineSpace, report).SetFont(f1).SetBorder(border).RightAlign().SetContent("0-3近日，江苏苏州市姑苏区市场监督管理局向苏州定园旅游服务有限公司送达行政处罚决定书，对定园进行处罚，吊销其营业执照，此举开创了我国旅游景点因虚假宣传被吊销营业执照的先河"))
	c10.SetElement(NewTextCell(table.GetColWidth(1, 0), lineHeight, lineSpace, report).SetFont(f1).SetBorder(border).VerticalCentered().SetContent("1-0近日，江苏苏州市姑苏区市场监督管理局向苏州定园旅游服务有限公司送达行政处罚决定书了我国旅游景点因虚假宣传被吊销营业执照的先河"))

	f1 = core.Font{Family: TABLE_MY, Size: 10}
	border = core.NewScope(4.0, 4.0, 0, 0)

	for i := 0; i < 98; i++ {
		cells := make([]*TableCell, 5)
		for j := 0; j < 5; j++ {
			cells[j] = table.NewCell()
		}

		for j := 0; j < 5; j++ {
			str := `有限公司送达行政处罚决定书`
			s := fmt.Sprintf("%v-%v", i+2, str)
			w := table.GetColWidth(i+2, j)
			e := NewTextCell(w, lineHeight, lineSpace, report)
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

func ComplexTableReport() {
	r := core.CreateReport()

	font1 := core.FontMap{
		FontName: TABLE_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: TABLE_MD,
		FileName: "example//ttf/mplus-1p-bold.ttf",
	}
	font3 := core.FontMap{
		FontName: TABLE_MY,
		FileName: "example//ttf/microsoft.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(ComplexTableReportExecutor), core.Detail)

	r.Execute("complex_table_test.pdf")
	r.SaveAtomicCellText("complex_table_test.txt")
}
func ComplexTableReportExecutor(report *core.Report) {
	lineSpace := 1.0
	lineHeight := 18.0

	form := NewTable(5, 100, 415, lineHeight, report)
	form.SetMargin(core.Scope{})

	// todo: 先把当前的行设置完毕, 然后才能添加单元格内容.
	c00 := form.NewCellByRange(1, 1)
	c01 := form.NewCellByRange(2, 1)
	c03 := form.NewCellByRange(2, 2)
	c10 := form.NewCellByRange(3, 1)

	f1 := core.Font{Family: TABLE_MY, Size: 15}
	c00.SetElement(NewTextCell(form.GetColWidth(0, 0), lineHeight, lineSpace, report).SetFont(f1).SetContent(GetRandStr()))
	c01.SetElement(NewTextCell(form.GetColWidth(0, 1), lineHeight, lineSpace, report).SetFont(f1).SetContent(GetRandStr()))
	c03.SetElement(NewTextCell(form.GetColWidth(0, 3), lineHeight, lineSpace, report).SetFont(f1).SetContent(GetRandStr()))
	c10.SetElement(NewTextCell(form.GetColWidth(1, 0), lineHeight, lineSpace, report).SetFont(f1).SetContent(GetRandStr()))

	f1 = core.Font{Family: TABLE_MY, Size: 10}
	border := core.NewScope(4.0, 4.0, 0, 0)

	for i := 0; i < 98; i++ {
		cells := make([]*TableCell, 5)
		for j := 0; j < 5; j++ {
			cells[j] = form.NewCell()
		}

		for j := 0; j < 5; j++ {
			w := form.GetColWidth(i+2, j)
			// todo: div执行的严格顺序
			e := NewTextCell(w, lineHeight, lineSpace, report)
			e.SetFont(f1)
			if i%2 == 0 {
				e.SetBackColor("255,192,203")
			}
			e.SetBorder(border)
			e.SetContent(GetRandStr())
			cells[j].SetElement(e)
		}
	}

	form.GenerateAtomicCell()
}

func ManyTableReportWithData() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: TABLE_IG,
		FileName: "example//ttf/ipaexg.ttf",
	}
	font2 := core.FontMap{
		FontName: TABLE_MD,
		FileName: "example//ttf/mplus-1p-bold.ttf",
	}
	font3 := core.FontMap{
		FontName: TABLE_MY,
		FileName: "example//ttf/microsoft.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2, &font3})
	r.SetPage("A4", "P")

	r.RegisterExecutor(core.Executor(ManyTableReportWithDataExecutor), core.Detail)

	r.Execute("many_table_data.pdf")
}
func ManyTableReportWithDataExecutor(report *core.Report) {
	lineSpace := 1.0
	lineHeight := 18.0

	rows, cols := 100, 5
	table := NewTable(cols, rows, 415, lineHeight, report)
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

func GetRandStr(l ...int) string {
	str := "0123456789ABCDEFGHIGKLMNOPQRSTUVWXYZ"
	l = append(l, 8)
	r := seed.Intn(l[0]*11) + 8
	//r := l[0] * 13
	r = 455
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

func TestComplexTableReport(t *testing.T) {
	ComplexTableReport()
}

func TestComplexTableReportWithData(t *testing.T) {
	ComplexTableReportWithData()
}

func TestManyTableReportWithData(t *testing.T) {
	start := time.Now().Unix()
	ManyTableReportWithData()
	end := time.Now().Unix()
	fmt.Println("i", 1, end-start)
}
