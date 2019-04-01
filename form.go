package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
	"fmt"
)

// 构建表格
type Table struct {
	pdf           *core.Report
	rows, cols    int //
	width, height float64
	colwidths     []float64      // 列宽百分比: 应加起来为1
	rowheights    []float64      // 保存行高
	cells         [][]*TableCell // 单元格

	lineHeight float64    // 默认行高
	margin     core.Scope // 位置调整

	nextrow, nextcol int // 下一个位置
	hasWrited        int // 当前页面已经写入的行数

	states [][]int // 当前页面的table单元格状态, 0,表示写入内容, 1表示未写入内容
}

type TableCell struct {
	table            *Table // table元素
	row, col         int    // 位置
	rowspan, colspan int    // 单元格大小

	element   core.Cell // 单元格元素
	minheight float64   // 当前最小单元格的高度, rowspan=1, 辅助计算
	height    float64   // 当前表格单元真实高度, rowspan >= 1, 实际计算垂直线高度的时候使用
}

func (cell *TableCell) SetElement(e core.Cell) *TableCell {
	cell.element = e
	cell.height = cell.element.GetHeight()
	if cell.rowspan == 1 {
		cell.minheight = cell.height
	}

	return cell
}

func NewTable(cols, rows int, width, lineHeight float64, pdf *core.Report) *Table {
	contentWidth, _ := pdf.GetContentWidthAndHeight()
	if width > contentWidth {
		width = contentWidth
	}

	pdf.LineType("straight", 0.1)
	pdf.GrayStroke(0)

	t := &Table{
		pdf:    pdf,
		rows:   rows,
		cols:   cols,
		width:  width,
		height: 0,

		nextcol: 0,
		nextrow: 0,

		lineHeight: lineHeight,
		colwidths:  []float64{},
		rowheights: []float64{},
		hasWrited:  1000,

		states: make([][]int, 0),
	}

	for i := 0; i < cols; i++ {
		t.colwidths = append(t.colwidths, float64(1.0)/float64(cols))
	}

	cells := make([][]*TableCell, rows)
	for i := range cells {
		cells[i] = make([]*TableCell, cols)
	}

	t.cells = cells

	return t
}

// 创建长宽为1的单元格
func (table *Table) NewCell() *TableCell {
	row, col := table.nextrow, table.nextcol
	cell := &TableCell{
		row:       row,
		col:       col,
		rowspan:   1,
		colspan:   1,
		table:     table,
		height:    table.lineHeight,
		minheight: table.lineHeight,
	}

	table.cells[row][col] = cell

	// 计算nextcol, nextrow
	table.setNext(1, 1)

	return cell
}

// 创建固定长度的单元格
func (table *Table) NewCellByRange(w, h int) *TableCell {
	colspan, rowspan := w, h
	if colspan <= 0 || rowspan <= 0 {
		panic("w and h must more than 0")
	}

	if colspan == 1 && rowspan == 1 {
		return table.NewCell()
	}

	row, col := table.nextrow, table.nextcol

	// 防止非法的宽度
	if !table.checkSpan(row, col, rowspan, colspan) {
		panic("inlivid layout, please check w and h")
	}

	cell := &TableCell{
		row:       row,
		col:       col,
		rowspan:   rowspan,
		colspan:   colspan,
		table:     table,
		height:    table.lineHeight,
		minheight: table.lineHeight,
	}

	table.cells[row][col] = cell

	// 构建空白单元格
	for i := 0; i < rowspan; i++ {
		var j int
		if i == 0 {
			j = 1
		}

		for ; j < colspan; j++ {
			table.cells[row+i][col+j] = &TableCell{
				row:       row + i,
				col:       col + j,
				rowspan:   -row,
				colspan:   -col,
				table:     table,
				height:    table.lineHeight,
				minheight: table.lineHeight,
			}
		}
	}

	// 计算nextcol, nextrow, 需要遍历处理
	table.setNext(colspan, rowspan)

	return cell
}

// 检测当前cell的宽和高
func (table *Table) checkSpan(row, col int, rowspan, colspan int) bool {
	var (
		cells          = table.cells
		maxrow, maxcol int
	)

	// 获取单方面的最大maxrow和maxcol
	for i := col; i < table.cols; i++ {
		if cells[row][i] != nil {
			maxcol = table.cols - col + 1
		}

		if i == table.cols-1 {
			maxcol = table.cols
		}
	}

	for i := row; i < table.rows; i++ {
		if cells[i][col] != nil {
			maxrow = table.rows - row + 1
		}

		if i == table.rows-1 {
			maxrow = table.rows
		}
	}

	// 检测合法性
	if colspan <= maxcol && rowspan <= maxrow {
		for i := row; i < table.rows; i++ {
			for j := col; j < table.cols; j++ {

				// cells[i][j]不为nil, 可以继续向下搜索
				if cells[i][j] == nil {
					if rowspan == i-row+1 && colspan <= j-col+1 {
						return true
					}

					if colspan == j-col+1 && rowspan <= i-row+1 {
						return true
					}
					continue
				}

				// cells[i][j]为nil, 不能向下搜索, 需要改变行号
				if cells[i][j] != nil {
					if rowspan == i-row+1 && colspan <= j-col+1 {
						return true
					}

					if colspan == j-col+1 && rowspan <= i-row+1 {
						return true
					}

					break
				}
			}
		}
	}

	return false
}

// 设置下一个单元格开始坐标
func (table *Table) setNext(colspan, rowspan int) {
	table.nextcol += colspan
	if table.nextcol == table.cols {
		table.nextcol = 0
		table.nextrow += 1
	}

	for i := table.nextrow; i < table.rows; i++ {
		var j int
		if i == table.nextrow {
			j = table.nextcol
		}

		for ; j < table.cols; j++ {
			if table.cells[i][j] == nil {
				table.nextrow, table.nextcol = i, j
				return
			}
		}
	}

	if table.nextrow == table.rows {
		table.nextcol = -1
		table.nextrow = -1
	}
}

/********************************************************************************************************************/

// 获取某列的宽度
func (table *Table) GetColWidth(row, col int) float64 {
	if row < 0 || row > len(table.cells) || col < 0 || col > len(table.cells[row]) {
		panic("the index out range")
	}

	count := 0.0
	for i := 0; i < table.cells[row][col].colspan; i++ {
		count += table.colwidths[i+col] * table.width
	}

	return count
}

// 设置表的行高, 行高必须大于当前使用字体的行高
func (table *Table) SetLineHeight(lineHeight float64) {
	table.lineHeight = lineHeight
}

// 设置表的外
func (table *Table) SetMargin(margin core.Scope) {
	margin.ReplaceMarign()
	table.margin = margin
}

/********************************************************************************************************************/

func (table *Table) Generate() error {
	var (
		sx, sy        = table.pdf.GetXY() // 基准坐标
		pageEndY      = table.pdf.GetPageEndY()
		x1, y1, _, y2 float64 // 当前位置
	)

	// 重新计算行高
	table.replaceCellHeight()

	for i := 0; i < table.rows; i++ {

		table.states = append(table.states, make([]int, table.cols))

		for j := 0; j < table.cols; j++ {

			// cell的rowspan是1的y1,y2
			_, y1, _, y2 = table.getVLinePosition(sx, sy, j, i) // 真实的垂直线
			if table.cells[i][j].rowspan > 1 {
				y2 = y1 + table.cells[i][j].minheight
			}

			// 进入换页操作
			if y1 > pageEndY {
				fmt.Println("pos", table.cells[i][j].row, table.cells[i][j].col, y1-pageEndY)
			}

			if y1 < pageEndY && y2 > pageEndY || y1 > pageEndY {
				if i == 0 {
					table.pdf.AddNewPage(false)
					table.hasWrited = 10000
					table.margin.Top = 0
					table.pdf.SetXY(table.pdf.GetPageStartXY())
					return table.Generate()
				}

				table.writeCurrentPageRestCells(i, j, sx, sy)

				// 当上半部分完整, 突然分页的状况
				if table.hasWrited > table.cells[i][j].row-table.cells[0][0].row {
					table.hasWrited = table.cells[i][j].row - table.cells[0][0].row
				}

				// 划线
				table.drawPageLineByStates(sx, sy)

				table.pdf.AddNewPage(false)
				table.margin.Top = 0
				table.cells = table.cells[table.hasWrited:]
				table.rows = len(table.cells)
				table.states = nil
				table.hasWrited = 1000

				table.pdf.LineType("straight", 0.1)
				table.pdf.GrayStroke(0)
				table.pdf.SetXY(table.pdf.GetPageStartXY())

				if table.rows == 0 {
					return nil
				}

				// 写入剩下页面
				return table.Generate()
			}

			// 拦截空白cell
			if table.cells[i][j].element == nil {
				table.states[i][j] = 1
				continue
			}

			x1, y1, _, y2 = table.getVLinePosition(sx, sy, j, i) // 真实的垂直线

			// 当前cell高度跨页
			if y1 < pageEndY && y2 > pageEndY {
				if table.hasWrited > table.cells[i][j].row-table.cells[0][0].row {
					table.hasWrited = table.cells[i][j].row - table.cells[0][0].row
				}
				table.writePartialPageCell(i, j, sx, sy) // 部分写入
			}

			// 当前celll没有跨页
			if y1 < pageEndY && y2 < pageEndY {
				table.writeCurrentPageCell(i, j, sx, sy)
			}
		}
	}

	// todo: 最后一个页面的最后部分
	height := table.getTableHeight()
	x1, y1, _, y2 = table.getVLinePosition(sx, sy, 0, 0)
	table.pdf.LineH(x1, y1+height+table.margin.Top, x1+table.width)
	table.pdf.LineV(x1+table.width, y1, y1+height+table.margin.Top)

	x1, _ = table.pdf.GetPageStartXY()
	table.pdf.SetXY(x1, y1+height+table.margin.Top+table.margin.Bottom) // 定格最终的位置

	return nil
}

// 自动换行生成
func (table *Table) GenerateAtomicCell() error {
	var (
		sx, sy        = table.pdf.GetXY() // 基准坐标
		pageEndY      = table.pdf.GetPageEndY()
		x1, y1, _, y2 float64 // 当前位置
	)

	// 重新计算行高
	table.replaceCellHeight()

	for i := 0; i < table.rows; i++ {
		x1, y1, _, y2 = table.getVLinePosition(sx, sy, 0, i) // 真实的垂直线
		if table.cells[i][0].rowspan > 1 {
			y2 = y1 + table.cells[i][0].minheight
		}

		// todo: 换页
		if y1 < pageEndY && y2 > pageEndY {
			var (
				currentRowWriteAll = true // 条件: 当前的行不存在空白(即rowspan=1),且每个单元格全部写完
			)

			// table首行, 直接分页
			if i == 0 {
				table.pdf.AddNewPage(false)
				table.margin.Top = 0
				table.pdf.SetXY(table.pdf.GetPageStartXY())

				return table.GenerateAtomicCell()
			}

			currentRowWriteAll = table.writeCurrentPageRestCells(i, 0, sx, sy)

			// 增加新页面
			table.pdf.AddNewPage(false)
			table.margin.Top = 0
			if currentRowWriteAll {
				table.cells = table.cells[i+1:]
			} else {
				table.cells = table.cells[i:]
			}
			table.rows = len(table.cells)

			table.pdf.LineType("straight", 0.1)
			table.pdf.GrayStroke(0)
			table.pdf.SetXY(table.pdf.GetPageStartXY())

			if table.rows == 0 {
				return nil
			}

			// 写入剩下页面
			return table.GenerateAtomicCell()
		}

		// todo: 当前页
		for j := 0; j < table.cols; j++ {
			if table.cells[i][j].element == nil {
				continue
			}

			table.writeCurrentPageCell(i, j, sx, sy)
		}
	}

	// todo: 最后一个页面的最后部分
	height := table.getTableHeight()
	x1, y1, _, y2 = table.getVLinePosition(sx, sy, 0, 0)
	table.pdf.LineH(x1, y1+height+table.margin.Top, x1+table.width)
	table.pdf.LineV(x1+table.width, y1, y1+height+table.margin.Top)

	x1, _ = table.pdf.GetPageStartXY()
	table.pdf.SetXY(x1, y1+height+table.margin.Top+table.margin.Bottom) // 定格最终的位置

	return nil
}

// row,col 定位cell, sx,sy是table基准坐标
func (table *Table) writeCurrentPageCell(row, col int, sx, sy float64) {
	var (
		x1, y1, _, y2 float64
	)

	// 写入数据前, 必须变换坐标系
	cell := table.cells[row][col]

	x1, y1, _, y2 = table.getVLinePosition(sx, sy, col, row) // 垂直线
	cell.table.pdf.SetXY(x1, y1)

	if cell.element != nil {
		n, _ := cell.element.GenerateAtomicCell(y2 - y1)
		if n > 0 {
			table.states[row][col] = 1
		} else {
			table.states[row][col] = 0
		}
	}

	cell.table.pdf.SetXY(sx, sy)
}
func (table *Table) writePartialPageCell(row, col int, sx, sy float64) {
	var (
		x1, y1   float64
		pageEndY = table.pdf.GetPageEndY()
	)

	// 写入数据前, 必须变换坐标系
	cell := table.cells[row][col]

	x1, y1, _, _ = table.getVLinePosition(sx, sy, col, row) // 垂直线
	cell.table.pdf.SetXY(x1, y1)

	if cell.element != nil {
		n, _ := cell.element.GenerateAtomicCell(pageEndY - y1)
		if n > 0 {
			table.states[row][col] = 1
		} else {
			table.states[row][col] = 0
		}
	}

	cell.table.pdf.SetXY(sx, sy)
}

// 当前页面的剩余内容
func (table *Table) writeCurrentPageRestCells(row, col int, sx, sy float64) bool {
	var (
		x1, y1   float64
		pageEndY = table.pdf.GetPageEndY()
	)

	for i := col; i < table.cols; i++ {
		cell := table.cells[row][i]

		if cell.element == nil {
			table.states[row][i] = 1
			continue
		}

		x1, y1, _, _ = table.getHLinePosition(sx, sy, i, row) // 计算初始点
		cell.table.pdf.SetXY(x1, y1)
		pageEndY = table.pdf.GetPageEndY()
		n, _ := cell.element.GenerateAtomicCell(pageEndY - y1) // 写入内容, 写完之后会修正其高度

		if n > 0 {
			table.states[row][i] = 1
		} else {
			table.states[row][i] = 0
		}
	}

	return true
}

// 划线
func (table *Table) drawPageLineByStates(sx, sy float64) {
	var (
		rows, cols           = len(table.states), len(table.states[0])
		pageEndY             = table.pdf.GetPageEndY()
		x, y, x1, y1, x2, y2 float64
	)

	table.pdf.LineType("straight", 0.1)
	table.pdf.GrayStroke(0)

	x, y, _, _ = table.getHLinePosition(sx, sy, 0, 0)
	table.pdf.LineH(x, y, x+table.width)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := table.cells[row][col]
			if cell.element == nil {
				continue
			}

			x, y, x1, y1 = table.getHLinePosition(sx, sy, col, row)
			fmt.Println("HLine: ", x, y, x1, y1, pageEndY)
			x, y, x2, y2 = table.getVLinePosition(sx, sy, col, row)
			fmt.Println("VLine: ", x, y, x2, y2, pageEndY)

			if y2 < pageEndY {
				table.pdf.LineV(x1, y1, y2)
				table.pdf.LineH(x, y2, x1)
			}

			if y2 >= pageEndY {
				//table.pdf.LineV(x1, y1, pageEndY)
			}
		}
	}
}

// 校验table是否合法
func (table *Table) checkTable() {
	var count int
	for i := 0; i < table.rows; i++ {
		for j := 0; j < table.cols; j++ {
			if table.cells[i][j] != nil {
				count += 1
			}
		}
	}

	if count != table.cols*table.rows {
		panic("please check setting rows, cols and writed cell")
	}
}

// todo: 重新计算tablecell的高度
func (table *Table) replaceCellHeight() {
	table.checkTable()

	cells := table.cells

	// 对于cells的元素重新赋值height和minheight
	for i := 0; i < table.rows; i++ {
		for j := 0; j < table.cols; j++ {
			if cells[i][j] != nil && cells[i][j].element != nil {
				cells[i][j].height = cells[i][j].element.GetHeight()
				if cells[i][j].rowspan == 1 || cells[i][j].rowspan < 0 {
					cells[i][j].minheight = cells[i][j].height
				}
			}
		}
	}

	// 第一遍计算rowspan是1的高度
	for i := 0; i < table.rows; i++ {
		var max float64 // 当前行的最大高度
		for j := 0; j < table.cols; j++ {
			if cells[i][j] != nil && max < cells[i][j].minheight {
				max = cells[i][j].minheight
			}
		}

		for j := 0; j < table.cols; j++ {
			if cells[i][j] != nil {
				cells[i][j].minheight = max // todo: 当前行(包括空白)的自身高度

				if cells[i][j].rowspan == 1 || cells[i][j].rowspan < 0 {
					cells[i][j].height = max
				}
			}
		}
	}

	// 第二遍计算rowsapn非1的行高度
	for i := 0; i < table.rows; i++ {
		for j := 0; j < table.cols; j++ {
			if cells[i][j] != nil && cells[i][j].rowspan > 1 {

				var totalHeight float64
				for k := 0; k < cells[i][j].rowspan; k++ {
					totalHeight += cells[i+k][j].minheight // todo: 计算所有行的高度
				}

				if totalHeight < cells[i][j].height {
					h := cells[i][j].height - totalHeight

					row := (cells[i][j].row - cells[0][0].row) + cells[i][j].rowspan - 1 // 最后一行
					for col := 0; col < table.cols; col++ {
						// 更新minheight
						cells[row][col].minheight += h

						// 更新height
						if cells[row][col].rowspan == 1 || cells[i][j].rowspan < 0 {
							cells[row][col].height += h
						}
					}
				} else {
					cells[i][j].height = totalHeight
				}
			}
		}
	}

	table.cells = cells
}

// 垂直线, table单元格的垂直线
func (table *Table) getVLinePosition(sx, sy float64, col, row int) (x1, y1 float64, x2, y2 float64) {
	var (
		x, y float64
		cell = table.cells[row][col]
	)

	for i := 0; i < col; i++ {
		x += table.colwidths[i] * table.width
	}
	x += sx + table.margin.Left

	for i := 0; i < row; i++ {
		y += table.cells[i][0].minheight
	}
	y = sy + y + table.margin.Top

	return x, y, x, y + cell.height
}

// 水平线, table单元格的水平线
func (table *Table) getHLinePosition(sx, sy float64, col, row int) (x1, y1 float64, x2, y2 float64) {
	var (
		x, y float64
		w    float64
	)

	for i := 0; i < col; i++ {
		x += table.colwidths[i] * table.width
	}
	x += sx + table.margin.Left

	for i := 0; i < row; i++ {
		y += table.cells[i][0].minheight
	}
	y += sy + table.margin.Top

	if table.cells[row][col].colspan > 1 {
		for k := 0; k < table.cells[row][col].colspan; k++ {
			w += table.colwidths[col+k] * table.width
		}
	} else {
		w = table.colwidths[col] * table.width
	}

	return x, y, x + w, y
}

// 节点垂直平线
func (table *Table) hasVLine(col, row int) bool {
	if col == 0 {
		return true
	}

	cell := table.cells[row][col]
	// 单独或者多个, 肯定是第一个
	if cell.rowspan+cell.colspan >= 2 {
		return true
	}

	// 距离"原点"的高度
	h := cell.col + cell.colspan
	if h == 0 {
		return true
	}

	return false
}

// 节点水平线
func (table *Table) hasHLine(col, row int) bool {
	if row == 0 {
		return true
	}

	var (
		cell = table.cells[row][col]
	)

	// 单独或者多个, 肯定是第一个
	if cell.rowspan+cell.colspan >= 2 {
		return true
	}

	v := cell.row + cell.rowspan
	if v == 0 {
		return true
	}

	return false
}

// 获取表的垂直高度
func (table *Table) getTableHeight() float64 {
	var count float64
	for i := 0; i < table.rows; i++ {
		count += table.cells[i][0].minheight
	}
	return count
}
