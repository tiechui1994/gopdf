package example

import (
	"testing"
	"math/rand"
	"time"
	"fmt"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf"
)

const (
	TABLE_IG = "IPAexG"
	TABLE_MD = "MPBOLD"
	TABLE_MY = "微软雅黑"
)

func TableReport() {
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

	r.Execute("table_test_data.pdf")
	r.SaveAtomicCellText("table_test_data.txt")
}
func TableReportExecutor(report *core.Report) {
	var (
		cells      [][]*gopdf.TableCell
		rows, cols = 321, 20
	)

	unit := report.GetUnit()

	lineSpace := 0.01 * unit
	lineHeight := 2 * unit

	table := gopdf.NewTable(cols, rows, 82*unit, lineHeight, report)
	table.SetMargin(core.Scope{0, 0, 0, 0})

	cells = Cells(rows, cols, table)
	f1 := core.Font{Family: TABLE_MY, Size: 15, Style: ""}
	border := core.NewScope(0.8*unit, 0.8*unit, 0.8*unit, 0.8*unit)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			cell := cells[i][j]
			if cell != nil {
				content := fmt.Sprintf("%v - %v", i, j)
				textCell := gopdf.NewTextCell(
					table.GetColWidth(i, j), lineHeight, lineSpace, report)

				textCell.SetFont(f1).SetBorder(border).HorizontalCentered().SetContent(content)

				cell.SetElement(textCell)
			}
		}
	}

	table.GenerateAtomicCell()
}

type Cell struct {
	row, col         int // 位置
	rowspan, colspan int // 高度和宽度
}

func Cells(rows, cols int, table *gopdf.Table) [][]*gopdf.TableCell {
	var (
		cells      [][]*Cell
		tableCells [][]*gopdf.TableCell
	)

	// 初始化
	cells = make([][]*Cell, rows)
	tableCells = make([][]*gopdf.TableCell, rows)
	for i := 0; i < rows; i++ {
		cells[i] = make([]*Cell, cols)
		tableCells[i] = make([]*gopdf.TableCell, cols)
	}

	makeCell(cells, 0, 0)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			cell := cells[i][j]
			if cell.colspan >= 1 && cell.rowspan >= 1 {
				tableCells[i][j] = table.NewCellByRange(cell.colspan, cell.rowspan)
			}
		}
	}

	return tableCells
}

func getrowspan(max int, maxrow int, random *rand.Rand) int {
	rowspan := random.Intn(maxrow) + 1
	if rowspan > max {
		return getrowspan(max, maxrow, random)
	}
	return rowspan
}

func makeCell(cells [][]*Cell, row, col int) {
	var (
		random         = rand.New(rand.NewSource(time.Now().Unix()))
		maxrow, maxcol = getReminedMax(cells, row, col)
	)

	rowspan := getrowspan(len(cells[0]), maxrow, random)
	colspan := random.Intn(int(maxcol/2)+1) + 1
	cell := &Cell{
		row:     row,
		col:     col,
		rowspan: rowspan,
		colspan: colspan,
	}
	cells[row][col] = cell
	fmt.Printf("cell:%+v", cell)

	// 空白
	if rowspan > 1 || colspan > 1 {
		for i := 0; i < rowspan; i++ {
			var j int
			if i == 0 {
				j = 1
			}
			for ; j < colspan; j++ {
				cells[row+i][col+j] = &Cell{
					row:     row + i,
					col:     col + j,
					rowspan: -row,
					colspan: -col,
				}
			}
		}
	}

	row, col = getNextPostion(cells, row, col)
	if row == -1 || col == -1 {
		return
	}
	fmt.Printf("pos:%v,%v \n", row, col)
	makeCell(cells, row, col)
}

func getNextPostion(cells [][]*Cell, row, col int) (nextrow, nextcol int) {
	var (
		rows = len(cells)
		cols = len(cells[0])
		cell = cells[row][col]
	)

	if cell == nil {
		return row, col
	}

	if cell != nil && (cell.rowspan <= 0 || cell.colspan <= 0) {
		cell = cells[-cell.rowspan][-cell.colspan]
		nextcol = cell.col + cell.colspan

		if nextcol == cols {
			nextrow = row + 1
			if nextrow == rows {
				return -1, -1
			}

			if cells[nextrow][0] == nil {
				return nextrow, 0
			}

			return getNextPostion(cells, nextrow, 0)
		}

		return getNextPostion(cells, row, nextcol)
	}

	if cell != nil && (cell.rowspan >= 1 || cell.colspan >= 1) {
		nextcol = col + cells[row][col].colspan

		if nextcol == cols {
			nextrow = row + 1

			if nextrow == rows {
				return -1, -1
			}

			if cells[nextrow][0] == nil {
				return nextrow, 0
			}

			return getNextPostion(cells, nextrow, 0)
		}

		if cells[row][nextcol] == nil {
			return row, nextcol
		}

		return getNextPostion(cells, row, nextcol)
	}

	panic("error")
}

func getReminedMax(cells [][]*Cell, row, col int) (maxrow, maxcol int) {
	var (
		count int
		rows  = len(cells)
		cols  = len(cells[0])
	)

	if row == 0 && col == 0 {
		return rows, cols
	}

	// 计算最大行
	count = 0
	for i := 0; i < rows; i++ {
		if cells[i][col] == nil {
			continue
		}
		count += 1
	}
	maxrow = rows - count

	// 计算最大列
	count = 0
	for i := 0; i < cols; i++ {
		if cells[row][i] == nil {
			continue
		}
		count += 1
	}
	maxcol = cols - count

	return maxrow, maxcol
}

func TestTableReport(t *testing.T) {
	TableReport()
}
