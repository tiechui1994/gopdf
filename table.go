package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
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

	tableCheck bool // table 完整性检查
}

type TableCell struct {
	table            *Table // table元素
	row, col         int    // 位置
	rowspan, colspan int    // 单元格大小

	element    core.Cell // 单元格元素
	minheight  float64   // 当前最小单元格的高度, rowspan=1, 辅助计算
	height     float64   // 当前表格单元真实高度, rowspan >= 1, 实际计算垂直线高度的时候使用
	cellwrited int
}

func (cell *TableCell) SetElement(e core.Cell) *TableCell {
	cell.element = e
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
		hasWrited:  2 ^ 32,
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
	if row == -1 && col == -1 {
		panic("there has no cell")
	}

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
	if row == -1 && col == -1 {
		panic("there has no cell")
	}

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
		height:    table.lineHeight * float64(h),
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

func (table *Table) GenerateAtomicCell() error {
	var (
		sx, sy        = table.pdf.GetXY() // 基准坐标
		pageEndY      = table.pdf.GetPageEndY()
		x1, y1, _, y2 float64 // 当前位置
	)

	// 重新计算行高
	table.resetCellHeight()

	for i := 0; i < table.rows; i++ {
		for j := 0; j < table.cols; j++ {
			// cell的rowspan是1的y1,y2
			_, y1, _, y2 = table.getVLinePosition(sx, sy, j, i)
			if table.cells[i][j].rowspan > 1 {
				y2 = y1 + table.cells[i][j].minheight
			}

			// 进入换页操作

			if y1 < pageEndY && y2 > pageEndY {
				if i == 0 {
					table.pdf.AddNewPage(false)
					table.hasWrited = 2 ^ 32
					table.margin.Top = 0
					table.pdf.SetXY(table.pdf.GetPageStartXY())
					return table.GenerateAtomicCell()
				}

				// 写完剩余的
				table.writeCurrentPageRestCells(i, j, sx, sy)

				// 当上半部分完整, 突然分页的状况
				if table.hasWrited > table.cells[i][j].row-table.cells[0][0].row {
					table.hasWrited = table.cells[i][j].row - table.cells[0][0].row
				}

				// 划线
				table.drawPageLines(sx, sy)

				// 分页
				table.resetTableCells()
				table.pdf.AddNewPage(false)
				table.margin.Top = 0
				table.rows = len(table.cells)
				table.hasWrited = 2 ^ 32

				table.pdf.LineType("straight", 0.1)
				table.pdf.GrayStroke(0)
				table.pdf.SetXY(table.pdf.GetPageStartXY())

				if table.rows == 0 {
					return nil
				}

				// 写入剩下页面
				return table.GenerateAtomicCell()
			}

			// 拦截空白cell
			if table.cells[i][j].element == nil {
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
	table.drawLastPageLines(sx, sy)

	height := table.getLastPageHeight()
	x1, y1, _, y2 = table.getVLinePosition(sx, sy, 0, 0)

	x1, _ = table.pdf.GetPageStartXY()
	table.pdf.SetXY(x1, y1+height+table.margin.Top+table.margin.Bottom) // 定格最终的位置

	return nil
}

// row,col 定位cell, sx,sy是table基准坐标
func (table *Table) writeCurrentPageCell(row, col int, sx, sy float64) {
	var (
		x1, y1, _, y2 float64
		pageEndY      = table.pdf.GetPageEndY()
	)

	// 写入数据前, 必须变换坐标系
	cell := table.cells[row][col]

	x1, y1, _, y2 = table.getVLinePosition(sx, sy, col, row) // 垂直线
	cell.table.pdf.SetXY(x1, y1)

	if cell.element != nil {
		// todo: 检查下一个Cell能否写入, 如果不能写入, 需要修正写入的高度值
		i, j := cell.row+cell.rowspan-table.cells[0][0].row, cell.col-table.cells[0][0].col
		if i < len(table.cells) {
			_, y3, _, y4 := table.getVLinePosition(sx, sy, j, i)
			if y3 < pageEndY && y4 >= pageEndY {
				if !table.checkNextCellCanWrite(sx, sy, row, col) {
					y2 = pageEndY
				}
			}
		}

		cell.element.GenerateAtomicCell(y2 - y1)
		cell.cellwrited = cell.rowspan
	}
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
		// todo:尝试写入, 写不进去就不再写
		wn, _ := cell.element.TryGenerateAtomicCell(pageEndY - y1)
		if wn == 0 {
			return
		}

		// 真正的写入
		n, _, _ := cell.element.GenerateAtomicCell(pageEndY - y1)

		if n > 0 && cell.element.GetHeight() == 0 {
			cell.cellwrited = cell.rowspan
		}

		// 统计写入的行数
		if n > 0 && cell.rowspan > 1 && cell.element.GetHeight() != 0 {
			count := 0
			for i := row; i < row+cell.rowspan; i++ {
				_, y1, _, y2 := table.getVLinePosition(sx, sy, col, i)
				if table.cells[i][col].element != nil {
					y2 = y1 + table.cells[i][col].minheight
				}
				if y2 < pageEndY {
					count++
				}
				if y1 > pageEndY {
					break
				}
			}

			cell.cellwrited = count
		}
	}
}

// 当前页面的剩余内容
func (table *Table) writeCurrentPageRestCells(row, col int, sx, sy float64) {
	var (
		x1, y1   float64
		pageEndY = table.pdf.GetPageEndY()
	)

	for i := col; i < table.cols; i++ {
		cell := table.cells[row][i]

		if cell.element == nil {
			continue
		}

		x1, y1, _, _ = table.getHLinePosition(sx, sy, i, row) // 计算初始点
		cell.table.pdf.SetXY(x1, y1)
		if y1 > pageEndY {
			continue
		}
		// todo: 只有不需要写入的才修改高度
		wn, _ := cell.element.TryGenerateAtomicCell(pageEndY - y1)
		if wn == 0 {
			return
		}

		// 真正的写入
		n, _, _ := cell.element.GenerateAtomicCell(pageEndY - y1) //写入内容

		if n > 0 && cell.element.GetHeight() == 0 {
			cell.cellwrited = cell.rowspan
		}

		// 统计写入的行数
		if n > 0 && cell.rowspan > 1 && cell.element.GetHeight() != 0 {
			count := 0
			for k := row; k < row+cell.rowspan; k++ {
				_, y1, _, y2 := table.getVLinePosition(sx, sy, i, k)
				if table.cells[k][i].element != nil {
					y2 = y1 + table.cells[k][i].minheight
				}
				if y2 < pageEndY {
					count++
				}
				if y1 > pageEndY {
					break
				}
			}

			cell.cellwrited = count
		}
	}
}

func (table *Table) checkNextCellCanWrite(sx, sy float64, row, col int) bool {
	var (
		canwrite bool
		cells    = table.cells
		pageEndY = table.pdf.GetPageEndY()
	)

	if cells[row][col].rowspan <= 0 {
		return canwrite
	}

	nextrow := cells[row][col].row + cells[row][col].rowspan - cells[0][0].row // 当前cell的下一行
	for k := col; k < table.cols; k++ {
		cell := cells[nextrow][col]
		_, y, _, _ := table.getHLinePosition(sx, sy, col, nextrow)

		if cell.rowspan <= 0 {
			i, j := -cell.rowspan-cells[0][0].row, -cell.colspan-cells[0][0].col
			wn, _ := cells[i][j].element.TryGenerateAtomicCell(pageEndY - y)
			if wn > 0 {
				canwrite = true
				return canwrite
			}
		}

		if cell.rowspan >= 1 {
			wn, _ := cells[nextrow][col].element.TryGenerateAtomicCell(pageEndY - y)
			if wn > 0 {
				canwrite = true
				return canwrite
			}
		}
	}

	return canwrite
}

// 对当前的Page进行画线
func (table *Table) drawPageLines(sx, sy float64) {
	var (
		rows, cols          = table.rows, table.cols
		pageEndY            = table.pdf.GetPageEndY()
		x, y, x1, y1, _, y2 float64
	)

	table.pdf.LineType("straight", 0.1)
	table.pdf.GrayStroke(0)

	x, y, _, _ = table.getHLinePosition(sx, sy, 0, 0)
	table.pdf.LineH(x, y, x+table.width)
	table.pdf.LineH(x, pageEndY, x+table.width)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := table.cells[row][col]

			x, y, x1, y1 = table.getHLinePosition(sx, sy, col, row)
			x, y, _, y2 = table.getVLinePosition(sx, sy, col, row)

			if cell.element == nil {
				continue
			}

			if y1 < pageEndY && y2 < pageEndY {
				// 判断当前的cell的下面的实体是否写入
				i, j := cell.row+cell.rowspan-table.cells[0][0].row, cell.col-table.cells[0][0].col
				_, y3, _, y4 := table.getVLinePosition(sx, sy, j, i)
				if y3 < pageEndY && y4 >= pageEndY {
					if !table.checkNextCellWrite(row, col) {
						y2 = pageEndY
						table.pdf.LineV(x1, y1, y2)
						table.pdf.LineH(x, y2, x1)
						continue
					}
				}

				table.pdf.LineV(x1, y1, y2) // 需要判断竖线否
				table.pdf.LineH(x, y2, x1)  // 需要判断底线否
			}

			// TODO: 需要先判断是否需要竖线
			if y1 < pageEndY && y2 >= pageEndY {
				if table.checkNeedVline(row, col) {
					table.pdf.LineV(x1, y1, pageEndY)
				}
				table.pdf.LineH(x, pageEndY, x1)
			}
		}
	}

	x, y, _, _ = table.getHLinePosition(sx, sy, 0, 0)
	table.pdf.LineV(x, y, pageEndY)
	table.pdf.LineV(x+table.width, y, pageEndY)
}

// 最后一页画线
func (table *Table) drawLastPageLines(sx, sy float64) {
	var (
		rows, cols          = table.rows, table.cols
		pageEndY            = table.getLastPageHeight()
		x, y, x1, y1, _, y2 float64
	)

	table.pdf.LineType("straight", 0.1)
	table.pdf.GrayStroke(0)

	x, y, _, _ = table.getHLinePosition(sx, sy, 0, 0)
	pageEndY = y + pageEndY

	table.pdf.LineH(x, y, x+table.width)
	table.pdf.LineH(x, pageEndY, x+table.width)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := table.cells[row][col]

			x, y, x1, y1 = table.getHLinePosition(sx, sy, col, row)
			x, y, _, y2 = table.getVLinePosition(sx, sy, col, row)

			if cell.element == nil {
				continue
			}

			if y1 < pageEndY && y2 < pageEndY {
				table.pdf.LineV(x1, y1, y2)
				table.pdf.LineH(x, y2, x1)
			}

			if y1 < pageEndY && y2 >= pageEndY {
				table.pdf.LineV(x1, y1, pageEndY)
				table.pdf.LineH(x, pageEndY, x1)
			}
		}
	}

	x, y, _, _ = table.getHLinePosition(sx, sy, 0, 0)
	table.pdf.LineV(x, y, pageEndY)
	table.pdf.LineV(x+table.width, y, pageEndY)
}

func (table *Table) checkNextCellWrite(row, col int) bool {
	var (
		cells      = table.cells
		cellwrited bool
	)

	if cells[row][col].rowspan <= 0 {
		return cellwrited
	}

	nextrow := cells[row][col].row + cells[row][col].rowspan - cells[0][0].row
	for k := col; k < table.cols; k++ {
		cell := cells[nextrow][col]

		if cell.rowspan <= 0 {
			i, j := -cell.rowspan-cells[0][0].row, -cell.colspan-cells[0][0].col
			if cells[i][j].cellwrited >= cell.row-cells[i][j].row+1 { // 当前的行被写入
				cellwrited = true
				return cellwrited
			}
		}

		if cell.rowspan == 1 {
			origin, remian := cell.element.GetLines()
			if cell.element.GetHeight() == 0 || origin-remian > 0 {
				cellwrited = true
				return cellwrited
			}
		}

		if cell.rowspan > 1 {
			if cell.cellwrited > 0 {
				cellwrited = true
				return cellwrited
			}
		}
	}

	return cellwrited
}

func (table *Table) checkNeedVline(row, col int) bool {
	var (
		needvline bool
		curwrited bool
		cells     = table.cells
		origin    = cells[0][0]
	)

	if cells[row][col].rowspan <= 0 {
		return needvline
	}

	// 当前的cell
	if cells[row][col].rowspan == 1 {
		origin, remian := cells[row][col].element.GetLines()
		if cells[row][col].element.GetHeight() == 0 || origin-remian > 0 {
			curwrited = true
		}
	}
	if cells[row][col].rowspan > 1 {
		if cells[row][col].cellwrited > 0 {
			curwrited = true
		}
	}

	// 邻居cell
	nextcol := cells[row][col].col + cells[row][col].colspan - cells[0][0].col
	if nextcol == table.cols {
		needvline = true
		return needvline && curwrited
	}

	if cells[row][nextcol].rowspan <= 0 {
		row, nextcol = -cells[row][nextcol].rowspan-origin.row, -cells[row][nextcol].colspan-origin.col
	}

	if cells[row][nextcol].rowspan == 1 {
		origin, remian := cells[row][nextcol].element.GetLines()
		if cells[row][nextcol].element.GetHeight() == 0 || origin-remian > 0 {
			needvline = true
			return needvline && curwrited
		}
	}

	if cells[row][nextcol].rowspan > 1 {
		if cells[row][nextcol].cellwrited > 0 {
			needvline = true
			return needvline && curwrited
		}
	}

	return needvline && curwrited
}

// 校验table是否合法
func (table *Table) checkTableConstraint() {
	if !table.tableCheck {
		return
	}

	table.tableCheck = false
	var (
		cells int
		area  int
	)
	for i := 0; i < table.rows; i++ {
		for j := 0; j < table.cols; j++ {
			cell := table.cells[i][j]
			if cell != nil {
				cells += 1
			}
			if cell != nil && cell.element != nil {
				area += cell.rowspan * cell.colspan
			}
		}
	}

	if cells != table.cols*table.rows || area != table.cols*table.rows {
		panic("please check setting rows, cols and writed cell")
	}
}

// TODO: 重新计算tablecell的高度(精确)
func (table *Table) resetCellHeight() {
	table.checkTableConstraint()

	x1, _ := table.pdf.GetPageStartXY()
	x2 := table.pdf.GetPageEndY()
	rows := table.rows
	if rows > int((x2-x1)/table.lineHeight)+1 {
		rows = int((x2-x1)/table.lineHeight) + 1
	}
	cells := table.cells

	// 对于cells的元素重新赋值height和minheight
	for i := 0; i < rows; i++ {
		for j := 0; j < table.cols; j++ {
			if cells[i][j] != nil && cells[i][j].element == nil {
				cells[i][j].minheight = table.lineHeight
				cells[i][j].height = table.lineHeight
			}

			if cells[i][j] != nil && cells[i][j].element != nil {
				cells[i][j].height = cells[i][j].element.GetHeight()
				if cells[i][j].rowspan == 1 {
					cells[i][j].minheight = cells[i][j].height
				}
			}
		}
	}

	// 第一遍计算rowspan是1的高度
	for i := 0; i < rows; i++ {
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
	for i := 0; i < rows; i++ {
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

						// todo: 更新height, 当rowspan=1
						if cells[row][col].rowspan == 1 {
							cells[row][col].height += h
						}

						// todo: 更新height, 当rowspan<0, 空格,需要更新非当前的实体(前面的)
						if cells[row][col].rowspan <= 0 {
							cells[row][col].height += h

							orow := -cells[row][col].rowspan - cells[0][0].row
							ocol := -cells[row][col].colspan - cells[0][0].col
							if orow == i && ocol < j || orow < i {
								cells[orow][ocol].height += h
							}
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

func (table *Table) resetTableCells() {
	var (
		min    = 2 ^ 32
		cells  = table.cells
		origin = cells[0][0]
	)

	// 获取写入的最小行数
	for col := 0; col < table.cols; col++ {
		count := 0
		cell := cells[table.hasWrited][col]

		if cell.rowspan <= 0 {
			i, j := -cell.rowspan-origin.row, -cell.colspan-origin.col

			// 从cells[table.hasWrite]算起已经写入的行数
			count += cells[i][j].cellwrited - (cell.row - cells[i][j].row)
			if cells[i][j].cellwrited == cells[i][j].rowspan {
				// TODO: 先计算当前的实体(空格->实体), 然后跳跃到下一个实体(条件性)
				// 后面第一个实体
				srow := cells[i][j].row - origin.row + cells[i][j].rowspan
				count += table.count(srow, col)
			}
		}

		if cell.rowspan >= 1 {
			count += cell.cellwrited
			if cell.cellwrited == cell.rowspan {
				// TODO: 先计算当前的实体(空格->实体), 然后跳跃到下一个实体(条件性)
				srow := cell.row - origin.row + cell.rowspan
				count += table.count(srow, col)
			}
		}

		if min > count {
			min = count
		}
	}

	// cell重置
	row := table.hasWrited + min
	if row >= len(table.cells) { // TODO: 这里的判断是修复未知的错误
		row = len(table.cells) - 1
	}
	for col := 0; col < table.cols; {
		cell := table.cells[row][col]
		if cell.rowspan <= 0 {
			i, j := -cell.rowspan-origin.row, -cell.colspan-origin.col
			var ox, oy int

			for x := row; x < cells[i][j].row+cells[i][j].rowspan-origin.row; x++ {
				for y := col; y < col+cells[i][j].colspan; y++ {
					if x == row && y == col {
						ox, oy = cells[x][y].row, cells[x][y].col

						cells[x][y].element = cells[i][j].element
						cells[x][y].rowspan = cells[i][j].rowspan - (ox - cells[i][j].row)
						cells[x][y].colspan = cells[i][j].colspan
						cells[x][y].cellwrited = 0

						continue
					}

					cells[x][y].rowspan = -ox
					cells[x][y].colspan = -oy
				}
			}

			col += cells[i][j].colspan
			continue
		}

		if cell.rowspan >= 1 {
			col += cell.colspan
			cell.cellwrited = 0
			continue
		}
	}

	table.cells = table.cells[table.hasWrited+min:]
}

func (table *Table) count(srow, scol int) int {
	var (
		origin = table.cells[0][0]
		count  int
	)
	for row := srow; row < table.rows; {
		cell := table.cells[row][scol]

		if cell.rowspan <= 0 {
			i, j := -cell.rowspan-origin.row, -cell.colspan-origin.col // 实体
			if table.cells[i][j].cellwrited == 0 { // 当前的实体未写
				break
			}

			if table.cells[i][j].cellwrited >= cell.row-table.cells[i][j].row+1 {
				count += table.cells[i][j].cellwrited - (cell.row - table.cells[i][j].row)
			}

			row = table.cells[i][j].row - origin.row + table.cells[i][j].rowspan // 实体的下一个"实体"
		}

		if cell.rowspan >= 1 {
			if cell.cellwrited == 0 { // 当前的实体未写
				break
			}

			count += cell.cellwrited

			if cell.rowspan > cell.cellwrited { // 当前的实体未写完
				break
			}

			row += cell.rowspan
		}
	}

	return count
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

// 获取表的垂直高度
func (table *Table) getLastPageHeight() float64 {
	var count float64
	for i := 0; i < table.rows; i++ {
		count += table.cells[i][0].minheight
	}
	return count
}
