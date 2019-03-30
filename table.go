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
	table.nextcol += 1
	if table.nextcol == table.cols {
		table.nextcol = 0
		table.nextrow += 1
	}

	if table.nextrow == table.rows {
		table.nextcol = -1
		table.nextrow = -1
	}
	return cell
}

// 创建固定长度的单元格
func (table *Table) NewCellByRange(w, h int) *TableCell {
	colspan, rowspan := w, h
	if colspan == 1 && rowspan == 1 {
		return table.NewCell()
	}

	row, col := table.nextrow, table.nextcol

	// 防止非法的宽度
	if colspan >= table.cols-col {
		colspan = table.cols - col
	}
	if rowspan >= table.rows-row {
		rowspan = table.rows - row
	}

	if colspan <= 0 || rowspan <= 0 {
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
			table.newSpaceCell(col+j, row+i, -row, -col)
		}
	}

	// 计算nextcol, nextrow, 需要遍历处理
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
				return cell
			}
		}
	}

	if table.nextrow == table.rows {
		table.nextcol = -1
		table.nextrow = -1
	}

	return nil
}

// 创建长宽为1的空白单元格
func (table *Table) newSpaceCell(col, row int, pr, pc int) *TableCell {
	cell := &TableCell{
		row:     row,
		col:     col,
		colspan: pc,
		rowspan: pr,
		table:   table,
		height:  table.lineHeight,
	}

	table.cells[row][col] = cell
	return cell
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

			currentRowWriteAll = table.writeCurrentPageRestContent(i, sx, sy)

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

// 当前页面的剩余内容
func (table *Table) writeCurrentPageRestContent(row int, sx, sy float64) bool {
	var (
		x1, y1, x2, _ float64
		needHLine     bool // 是否需要设置水平线, 条件: 有内容写入
		pageEndY      = table.pdf.GetPageEndY()
		curPageEndY   float64 // 当前页面的endY值
	)

	for col := 0; col < table.cols; col++ {
		cell := table.cells[row][col]
		if cell.element == nil {
			continue
		}

		// 1) 变换坐标系, 计算新的初始点
		x1, y1, _, _ := table.getHLinePosition(sx, sy, col, row)
		cell.table.pdf.SetXY(x1, y1)
		pageEndY = table.pdf.GetPageEndY()
		n, _ := cell.element.GenerateAtomicCell(pageEndY - y1) // 写入内容, 写完之后会修正其高度

		curPageEndY = y1
		if n > 0 {
			curPageEndY = pageEndY
			needHLine = true
		}
	}

	// 3) 当有一个写入则必须有垂直线
	if needHLine {
		for k := 0; k < table.cols; k++ {
			if table.hasVLine(k, row) {
				x1, y1, x2, _ = table.getVLinePosition(sx, sy, k, row)
				table.pdf.Line(x1, y1, x2, curPageEndY)
			}
		}
	}

	// 4) 右侧垂直线
	x1, y1, x2, _ = table.getVLinePosition(sx, sy, 0, 0)
	table.pdf.LineV(x1+table.width, y1, curPageEndY)

	// 5) 底层水平线
	table.pdf.Line(x1, curPageEndY, x1+table.width, curPageEndY)

	// 6) 返回条件
	var currentRowWriteAll = true // 条件: 当前的行不存在空白(即rowspan=1),且每个单元格全部写完
	for col := 0; col < table.cols; col++ {
		cell := table.cells[row][col]
		if cell.rowspan != 1 || cell.element.GetHeight() != 0 {
			currentRowWriteAll = false
		}
	}

	return currentRowWriteAll
}

// row,col 定位cell, sx,sy是table基准坐标
func (table *Table) writeCurrentPageCell(row, col int, sx, sy float64) {
	var (
		x1, y1, x2, y2 float64
	)

	// 1. 写入数据, 坐标系必须变换
	cell := table.cells[row][col]

	x1, y1, x2, y2 = table.getVLinePosition(sx, sy, col, row) // 垂直线
	cell.table.pdf.SetXY(x1, y1)
	if cell.element != nil {
		cell.element.GenerateAtomicCell(y2 - y1)
	}

	// 2. 水平线
	if table.hasHLine(col, row) {
		x1, y1, x2, y2 = table.getHLinePosition(sx, sy, col, row)
		table.pdf.Line(x1, y1, x2, y2)
	}

	// 3. 垂直线
	if table.hasVLine(col, row) {
		x1, y1, x2, y2 = table.getVLinePosition(sx, sy, col, row)
		table.pdf.Line(x1, y1, x2, y2)
	}

	cell.table.pdf.SetXY(sx, sy)
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

					row := cells[i][j].row + cells[i][j].rowspan - 1 // 最后一行
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
