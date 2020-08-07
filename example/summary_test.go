package example

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/tiechui1994/gopdf"
	"github.com/tiechui1994/gopdf/core"
)

const (
	ErrSummaryFile    = 1
	SummaryFONT_MY    = "微软雅黑"
	SummaryFONT_MD    = "MPBOLD"
	SummaryDateFormat = "2006-01-02 15:04:05"
)

var (
	largeSummaryFont = core.Font{Family: SummaryFONT_MY, Size: 15}
	headSummaryFont  = core.Font{Family: SummaryFONT_MY, Size: 12}
	textSummaryFont  = core.Font{Family: SummaryFONT_MY, Size: 10}
	titleSummaryFont = core.Font{Family: SummaryFONT_MY, Size: 18}
)

func SummaryReport() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: SummaryFONT_MY,
		FileName: "ttf//microsoft.ttf",
	}
	font2 := core.FontMap{
		FontName: SummaryFONT_MD,
		FileName: "ttf//mplus-1p-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2})
	r.SetPage("A4", "L")
	// r.SetPage("A4", "L", core.DefaultPageMargin["Narrow"])
	r.FisrtPageNeedHeader = true
	r.FisrtPageNeedFooter = true

	r.RegisterExecutor(core.Executor(SummaryReportExecutor), core.Detail)
	r.RegisterExecutor(core.Executor(SummaryReportFooterExecutor), core.Footer)
	r.RegisterExecutor(core.Executor(SummaryReportHeaderExecutor), core.Header)

	r.Execute(fmt.Sprintf("summary_report_test.pdf"))
}

func SummaryReportExecutor(report *core.Report) {
	var (
		lineSpace = 5.0
		lineHight = 16.0
	)

	// dir, _ := filepath.Abs("pictures")
	width, _ := report.GetContentWidthAndHeight()
	div := gopdf.NewDivWithWidth(100, lineHight, lineSpace, report)
	div.SetFont(textSummaryFont)
	line := gopdf.NewHLine(report).SetMargin(core.Scope{Top: 15.0, Bottom: 6.8}).SetWidth(0.1).SetColor(1)
	report.SetMargin(0, -25)
	nameDiv := gopdf.NewDivWithWidth(width, 15, 0, report)
	nameDiv.SetFont(headSummaryFont)
	nameDiv.HorizontalCentered().SetContent("XXXXX人民医院").GenerateAtomicCell()
	title := gopdf.NewDivWithWidth(width, 15, 0, report)
	title.SetFont(titleSummaryFont)
	title.HorizontalCentered().SetContent("系统打印报表").GenerateAtomicCell()
	line.GenerateAtomicCell()
	//生成表格
	cells := make([][]string, 100)
	headerCell := []string{"月份", "重量", "", "", "", "", "总计", "数量", "", "", "", "", "总计"}
	headerCell2 := []string{"", "dddd", "eeee", "ffff", "gggg", "hhhh", "", "dddd", "eeee", "ffff", "gggg", "hhhh", ""}
	dataCell := []string{"1000", "1000", "1000", "1000", "1000", "5000", "1", "1", "2", "1", "2", "7"}
	cols := len(headerCell)
	rows := len(cells)
	cells[0] = make([]string, cols)
	isSpan := false
	spanCount := 0
	for i := 0; i < len(headerCell); i++ {
		cells[0][i] = headerCell[i]
		if headerCell[i] == "" {
			isSpan = true
			if isSpan {
				spanCount++
			} else {
				spanCount = 1
			}

		} else {
			spanCount = 0
		}
	}
	fmt.Println(len(headerCell), len(headerCell2))
	cells[1] = make([]string, cols)
	for i := 1; i < len(headerCell2); i++ {
		cells[1][i] = headerCell2[i]
	}
	for i := 2; i < rows; i++ {
		cells[i] = make([]string, cols)
		cells[i][0] = time.Now().UTC().Format("2006-01")
		for j := 1; j < cols; j++ {
			cells[i][j] = dataCell[j-1]
		}
	}
	report.SetMargin(0, 10)
	table := gopdf.NewTable(cols, rows, width, lineHight, report)
	table.SetMargin(core.Scope{})
	// c00 := table.NewCellByRange(1, 2)
	// c00.SetElement(gopdf.NewTextCell(table.GetColWidth(0, 0), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell[0]))
	// c01 := table.NewCellByRange(5, 1)
	// c01.SetElement(gopdf.NewTextCell(table.GetColWidth(0, 1), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell[1]))
	// c02 := table.NewCellByRange(1, 2)
	// c02.SetElement(gopdf.NewTextCell(table.GetColWidth(0, 6), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell[6]))
	// c03 := table.NewCellByRange(5, 1)
	// c03.SetElement(gopdf.NewTextCell(table.GetColWidth(0, 7), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell[7]))
	// c04 := table.NewCellByRange(1, 2)
	// c04.SetElement(gopdf.NewTextCell(table.GetColWidth(0, 12), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell[12]))
	// c05 := table.NewCellByRange(1, 1)
	// c05.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 1), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[1]))
	// c06 := table.NewCellByRange(1, 1)
	// c06.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 2), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[2]))
	// c07 := table.NewCellByRange(1, 1)
	// c07.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 3), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[3]))
	// c08 := table.NewCellByRange(1, 1)
	// c08.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 4), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[4]))
	// c09 := table.NewCellByRange(1, 1)
	// c09.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 5), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[5]))
	// c10 := table.NewCellByRange(1, 1)
	// c10.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 7), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[7]))
	// c11 := table.NewCellByRange(1, 1)
	// c11.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 8), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[8]))
	// c12 := table.NewCellByRange(1, 1)
	// c12.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 9), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[9]))
	// c13 := table.NewCellByRange(1, 1)
	// c13.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 10), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[10]))
	// c14 := table.NewCellByRange(1, 1)
	// c14.SetElement(gopdf.NewTextCell(table.GetColWidth(1, 11), lineHight, lineSpace, report).SetFont(textSummaryFont).HorizontalCentered().SetContent(headerCell2[11]))
	//设置列宽
	table.SetColumnWidth(0, 0.06)
	// table.SetColumnWidth(3, 0.1)
	// table.SetColumnWidth(6, 0.1)
	for i := 0; i < 100; i++ {
		for j := 0; j < cols; j++ {
			cell := table.NewCell()
			cellWidth := table.GetColWidth(i, j)
			// if i == 0 || j == 0 {
			// 	cellWidth = cellWidth / 4
			// }
			// fmt.Println(cellWidth)
			element := gopdf.NewTextCell(cellWidth, lineHight, lineSpace, report)
			element.SetFont(textSummaryFont)
			element.SetBorder(core.Scope{Left: 2, Top: 2})
			element.HorizontalCentered()
			element.VerticalCentered()
			// if i == 0 || j == 0 && hasRowName {
			// 	element.HorizontalCentered()
			// }
			element.SetContent(cells[i][j])
			cell.SetElement(element)
		}
	}
	// 合并单元格
	table.SetSpan(0, 0, 2, 1)
	table.SetSpan(0, 1, 1, 5)
	table.SetSpan(0, 6, 2, 1)
	table.SetSpan(0, 7, 1, 5)
	table.SetSpan(0, 12, 2, 1)

	table.GenerateAtomicCell()
	report.SetMargin(0, 4)
}

//NewTextCell new a text cell
func NewTextCell(width, lineHeight, lineSpace float64, pdf *core.Report) *gopdf.TextCell {
	endX, _ := pdf.GetPageEndXY()
	curX, _ := pdf.GetXY()
	if width > endX-curX {
		width = endX - curX
	}
	cell := gopdf.NewTextCell(width, lineHeight, lineSpace, pdf)

	return cell
}

func Summary2Report() {
	r := core.CreateReport()
	font1 := core.FontMap{
		FontName: SummaryFONT_MY,
		FileName: "ttf//microsoft.ttf",
	}
	font2 := core.FontMap{
		FontName: SummaryFONT_MD,
		FileName: "ttf//mplus-1p-bold.ttf",
	}
	r.SetFonts([]*core.FontMap{&font1, &font2})
	r.SetPage("A4", "L")
	// r.SetPage("A4", "L", core.DefaultPageMargin["Narrow"])
	r.FisrtPageNeedHeader = true
	r.FisrtPageNeedFooter = true

	r.RegisterExecutor(core.Executor(Summary2ReportExecutor), core.Detail)
	r.RegisterExecutor(core.Executor(SummaryReportFooterExecutor), core.Footer)
	r.RegisterExecutor(core.Executor(SummaryReportHeaderExecutor), core.Header)

	r.Execute(fmt.Sprintf("summary_report2_test.pdf"))
}

func Summary2ReportExecutor(report *core.Report) {
	var (
		lineSpace = 1.0
		lineHight = 16.0
	)

	// dir, _ := filepath.Abs("pictures")
	width, _ := report.GetContentWidthAndHeight()
	div := gopdf.NewDivWithWidth(100, lineHight, lineSpace, report)
	div.SetFont(textSummaryFont)
	line := gopdf.NewHLine(report).SetMargin(core.Scope{Top: 15.0, Bottom: 6.8}).SetWidth(0.1).SetColor(0)
	report.SetMargin(0, -25)
	nameDiv := gopdf.NewDivWithWidth(width, 15, 0, report)
	nameDiv.SetFont(headSummaryFont)
	nameDiv.HorizontalCentered().SetContent("XXXX人民医院").GenerateAtomicCell()
	//标题
	title := gopdf.NewDivWithWidth(width, 15, 0, report)
	title.SetFont(titleSummaryFont)
	title.HorizontalCentered().SetContent("系统打印报表").GenerateAtomicCell()
	line.GenerateAtomicCell()
	// 基本信息
	report.SetMargin(0, 5)
	baseInfoDiv := gopdf.NewDivWithWidth(width, lineHight, lineSpace, report)
	baseInfoDiv.SetFont(headSummaryFont)
	baseInfoDiv.HorizontalCentered().SetContent("出库人  __________________  接收人  ________________").GenerateAtomicCell()
	report.SetMargin(0, 5)
	outboundDiv := gopdf.NewDivWithWidth(width, lineHight, lineSpace, report)
	outboundDiv.SetFont(headSummaryFont)
	outboundDiv.HorizontalCentered().SetContent(fmt.Sprintf("出库时间：%s     总重量：%.4f公斤", time.Now().Format("2006-01-02 15:04:05"), 70.700)).GenerateAtomicCell()
	//生成表格
	cells := make([][]string, 100)
	headerCell := []string{"序号", "识别码", "科室", "类别", "重量", "收集时间", "回收员", "入库时间", "管理员"}
	dataCell := []string{"B85C11412356111", "内科", "eeee", "6300g", "2020-01-20 17:39:02", "张三", "2020-01-20 17:39:02", "李四"}
	cols := len(headerCell)
	rows := len(cells)
	cells[0] = make([]string, cols)
	for i := 0; i < len(headerCell); i++ {
		cells[0][i] = headerCell[i]
	}
	for i := 1; i < rows; i++ {
		cells[i] = make([]string, cols)
		cells[i][0] = strconv.FormatInt(int64(i), 10)
		for j := 1; j < cols; j++ {
			cells[i][j] = dataCell[j-1]
		}
	}
	report.SetMargin(0, 10)
	table := gopdf.NewTable(cols, rows, width, lineHight, report)
	//设置列宽
	table.SetColumnWidth(0, 0.05)
	table.SetColumnWidth(1, 0.15)
	table.SetColumnWidth(3, 0.08)
	table.SetColumnWidth(4, 0.08)
	table.SetColumnWidth(6, 0.08)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			cell := table.NewCell()
			cellWidth := table.GetColWidth(i, j)
			// if i == 0 || j == 0 {
			// 	cellWidth = cellWidth / 4
			// }
			// fmt.Println(cellWidth)
			element := gopdf.NewTextCell(cellWidth, lineHight, lineSpace, report)
			element.SetFont(textSummaryFont)
			element.SetBorder(core.Scope{Left: 2, Top: 2})
			element.HorizontalCentered()
			element.VerticalCentered()
			// if i == 0 || j == 0 && hasRowName {
			// 	element.HorizontalCentered()
			// }
			element.SetContent(cells[i][j])
			cell.SetElement(element)
		}
	}

	table.GenerateAtomicCell()
	report.SetMargin(0, 4)
}
func SummaryReportFooterExecutor(report *core.Report) {
	content := fmt.Sprintf("%s                                           第 %v / {#TotalPage#} 页", time.Now().Format("2006-01-02 15:04:05"), report.GetCurrentPageNo())
	footer := gopdf.NewSpan(30, 0, report).SetMarign(core.Scope{Top: 10})
	footer.SetFont(textSummaryFont)
	footer.SetFontColor("192,192,192")
	footer.HorizontalCentered().SetContent(content).GenerateAtomicCell()
}
func SummaryReportHeaderExecutor(report *core.Report) {
	width, _ := report.GetContentWidthAndHeight()
	// line := gopdf.NewHLine(report).SetMargin(core.Scope{Top: 10.0, Bottom: 6.8}).SetWidth(0.1)
	report.SetMargin(10, 20)
	nameDiv := gopdf.NewDivWithWidth(width/2, 15, 0, report)
	nameDiv.SetFontWithColor(headSummaryFont, "128,128,128")
	nameDiv.SetContent(time.Now().Format("2006-01-02 15:04:05")).GenerateAtomicCell()
	report.SetMargin(10+width/2, -15)
	nameDiv = gopdf.NewDivWithWidth(width/2, 15, 0, report)
	nameDiv.SetFontWithColor(headSummaryFont, "128,128,128")
	nameDiv.RightAlign().SetContent(fmt.Sprintf("第 %v / {#TotalPage#} 页", report.GetCurrentPageNo())).GenerateAtomicCell()
	// line.GenerateAtomicCell()
}
func TestSummaryReport(t *testing.T) {
	SummaryReport()
	Summary2Report()
}
