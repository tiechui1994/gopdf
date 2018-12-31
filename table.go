package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
)

// 构建表格
type Table struct {
	pdf *core.Report

	nextRow, nextCol int
	rows, cols       int
	width, height    float64

	// 列宽百分比: 应加起来为1
	colWidths []float64

	// 保存行高
	rowHeights []float64

	// 默认行高
	lineHeight float64
	lineSpace  float64

	// table的位置调整
	margin Scope

	cells [][]*TableCell

	// 辅助作用
	isFirstCalTableCellHeight bool
}

// row行, col列
func NewTable(cols, rows int, width, lineHeight float64, pdf *core.Report) *Table {
	contentWidth, _ := pdf.GetContentWidthAndHeight()
	if width > contentWidth {
		width = contentWidth
	}

	pdf.LineType("straight", 0.1)
	pdf.GrayStroke(0)

	t := &Table{
		pdf: pdf,

		rows:    rows,
		cols:    cols,
		width:   width,
		height:  0,
		nextCol: 0,
		nextRow: 0,

		colWidths:  []float64{},
		lineHeight: lineHeight,

		isFirstCalTableCellHeight: true,
	}
	(*t).pdf = pdf
	// Initialize column widths as all equal.
	colWidth := float64(1.0) / float64(cols)
	for i := 0; i < cols; i++ {
		t.colWidths = append(t.colWidths, colWidth)
	}

	cells := make([][]*TableCell, rows)
	for i := range cells {
		cells[i] = make([]*TableCell, cols)
	}
	t.cells = cells

	return t
}

// 获取某列的宽度
func (table *Table) GetColWithByIndex(row, col int) float64 {
	if row < 0 || row > len(table.cells) || col < 0 || col > len(table.cells[row]) {
		panic("the index out range")
	}

	if table.cells[row][col] == nil {
		panic("there has no row")
	}

	count := 0.0
	for i := 0; i < table.cells[row][col].colspan; i++ {
		count += table.colWidths[i+col] * table.width
	}

	return count
}

// 设置表的行高, 行高必须大于当前使用字体的行高
func (table *Table) SetLineHeight(lineHeight float64) {
	table.lineHeight = lineHeight
}

// 设置表的外
func (table *Table) SetMargin(margin Scope) {
	table.margin = margin
}

func (table *Table) checkWritedRowAndCol() {
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
	table.checkWritedRowAndCol()
	cells := table.cells

	// 重新获取高度, 第一次, 而且只有一次
	if table.isFirstCalTableCellHeight {
		for i := 0; i < table.rows; i++ {
			for j := 0; j < table.cols; j++ {
				if cells[i][j] != nil && cells[i][j].element != nil && cells[i][j].colspan > 0 {
					cells[i][j].height = cells[i][j].element.GetHeight()
				}
			}
		}
	}

	// 第一遍计算行高度
	for i := 0; i < table.rows; i++ {
		var maxHeight float64 // 当前行的最大高度
		for j := 0; j < table.cols; j++ {
			if cells[i][j] != nil && maxHeight < cells[i][j].height {
				maxHeight = cells[i][j].height
			}
		}

		for j := 0; j < table.cols; j++ {
			if cells[i][j] != nil {
				cells[i][j].height = maxHeight
			}

			// 同步操作
			if table.isFirstCalTableCellHeight && cells[i][j].element != nil {
				cells[i][j].element.SetHeight(maxHeight)
			}
		}
	}

	// 只能同步操作一次
	if table.isFirstCalTableCellHeight {
		table.isFirstCalTableCellHeight = false
	}

	// 第二遍计算rowsapn非1的行高度
	for i := 0; i < table.rows; i++ {
		for j := 0; j < table.cols; j++ {
			if cells[i][j] != nil {
				rowspan := cells[i][j].rowspan
				colspan := cells[i][j].colspan

				if rowspan+colspan > 2 {
					var totalHeight float64
					for v := 0; v < rowspan; v++ {
						totalHeight += cells[i+v][j].height
					}
					cells[i][j].height = totalHeight
				}
			}
		}
	}

	table.cells = cells
}

// 垂直线
func (table *Table) getVLinePosition(sx, sy float64, col, row int) (x1, y1 float64, x2, y2 float64) {
	var (
		x, y float64
		cell = table.cells[row][col]
	)

	for i := 0; i < col; i++ {
		x += table.colWidths[i]
	}
	x = sx + x*table.width + table.margin.Left

	for i := 0; i < row; i++ {
		if table.cells[i][0].colspan > 0 && table.cells[i][0].rowspan > 0 {
			y += table.cells[i][0].height
		}
	}
	y = sy + y + table.margin.Top

	return x, y, x, y + cell.height
}

// 水平线
func (table *Table) getHLinePosition(sx, sy float64, col, row int) (x1, y1 float64, x2, y2 float64) {
	var (
		x, y float64
	)

	for i := 0; i < col; i++ {
		x += table.colWidths[i]
	}
	x = sx + x*table.width + table.margin.Left

	for i := 0; i < row; i++ {
		y += table.cells[i][0].height
	}
	y = sy + y + table.margin.Top

	return x, y, x + table.colWidths[col]*table.width, y
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
		count += table.cells[i][0].height
	}
	return count
}

// 自动换行生成
func (table *Table) GenerateAtomicCell() (error) {
	var (
		sx, sy         = table.pdf.GetXY()
		pageEndY       = table.pdf.GetPageEndY()
		x1, y1, x2, y2 float64
	)

	// 重新计算行高
	table.replaceCellHeight()

	for i := 0; i < table.rows; i++ {
		x1, y1, x2, y2 = table.getVLinePosition(sx, sy, 0, i)

		// todo: 需要换页
		if y1 < pageEndY && y2 > pageEndY {
			// todo: 1) 写入部分数据, 坐标系必须变换
			var needSetHLine bool
			for k := 0; k < table.cols; k++ {
				cell := table.cells[i][k]
				cellOriginHeight := cell.height

				x1, y1, _, _ := table.getHLinePosition(sx, sy, k, i)
				cell.table.pdf.SetXY(x1, y1)
				cell.element.GenerateAtomicCell()      // 会修改element的高度
				cell.height = cell.element.GetHeight() // 将修改后的高度同步到本地的Cell当中
				cell.table.pdf.SetXY(sx, sy)

				if cellOriginHeight-cell.element.GetHeight() > 0 {
					needSetHLine = true
				}

				// todo: 2) 垂直线
				if table.hasVLine(k, i) {
					x1, y1, x2, y2 = table.getVLinePosition(sx, sy, k, i)
					table.pdf.Line(x1, y1, x2, pageEndY)
				}
			}

			// todo: 3) 只有当一个有写入则必须有水平线
			if needSetHLine {
				for k := 0; k < table.cols; k++ {
					x1, y1, x2, y2 = table.getHLinePosition(sx, sy, k, i)
					table.pdf.Line(x1, y1, x2, y2)
				}
			}

			// todo: 4) 补全右侧垂直线 和 底层水平线
			x1, y1, x2, y2 = table.getVLinePosition(sx, sy, 0, 0)
			table.pdf.LineH(x1, pageEndY, x1+table.width)
			table.pdf.LineV(x1+table.width, y1, pageEndY)

			// todo: 5) 增加新页面
			table.pdf.AddNewPage(false)
			table.margin.Top = 0
			table.cells = table.cells[i:]
			table.rows = len(table.cells)

			table.pdf.LineType("straight", 0.1)
			table.pdf.GrayStroke(0)
			table.pdf.SetXY(table.pdf.GetPageStartXY())

			if table.rows == 0 {
				return nil
			}

			// todo: 6) 剩下页面
			return table.GenerateAtomicCell()
		}

		// todo: 不需要换页
		for j := 0; j < table.cols; j++ {
			// todo: 1.水平线
			if table.hasHLine(j, i) {
				x1, y1, x2, y2 = table.getHLinePosition(sx, sy, j, i)
				table.pdf.Line(x1, y1, x2, y2)
			}

			// todo: 2. 垂直线
			if table.hasVLine(j, i) {
				x1, y1, x2, y2 = table.getVLinePosition(sx, sy, j, i)
				table.pdf.Line(x1, y1, x2, y2)
			}

			// todo: 3. 写入数据, 坐标系必须变换
			cell := table.cells[i][j]
			if cell.element == nil {
				continue
			}

			x1, y1, _, _ := table.getHLinePosition(sx, sy, j, i)
			cell.table.pdf.SetXY(x1, y1)
			cell.element.GenerateAtomicCell()
			cell.table.pdf.SetXY(sx, sy)
		}
	}

	// todo: 最后一个页面的最后部分
	height := table.getTableHeight()
	x1, y1, x2, y2 = table.getVLinePosition(sx, sy, 0, 0)
	table.pdf.LineH(x1, y1+height+table.margin.Top, x1+table.width)
	table.pdf.LineV(x1+table.width, y1, y1+height+table.margin.Top)

	table.pdf.SetXY(0, y1+height+table.margin.Top) // 定格最终的位置

	return nil
}

// 根据当前的节点位置, 获取下一个节点的位置
func (table *Table) getNextRowAndCol(row, col int) (nexRow int, nextCol int) {
	for i := row; i < table.rows; i++ {
		j := 0
		if i == row {
			j = col
		}

		for ; j < table.cols; j++ {
			if isEmpty(table.cells[i][j]) {
				return i, j
			}
		}
	}

	// 最后一个
	return -1, -1
}

// 创建长宽为1的单元格
func (table *Table) NewCell() *TableCell {
	curRow, curCol := table.nextRow, table.nextCol

	cell := &TableCell{
		row:     curRow,
		col:     curCol,
		rowspan: 1,
		colspan: 1,
		table:   table,
		height:  table.lineHeight,
	}

	table.cells[curRow][curCol] = cell
	table.nextRow, table.nextCol = table.getNextRowAndCol(curRow, curCol)
	return cell
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

// 创建固定长度的单元格
func (table *Table) NewCellByRange(w, h int) *TableCell {
	colspan, rowspan := w, h
	if colspan == 1 && rowspan == 1 {
		return table.NewCell()
	}

	curRow, curCol := table.nextRow, table.nextCol
	if w >= table.cols-curCol {
		w = table.cols - curCol
	}

	cell := &TableCell{
		row:     curRow,
		col:     curCol,
		rowspan: rowspan,
		colspan: colspan,
		table:   table,
		height:  table.lineHeight,
	}

	table.cells[curRow][curCol] = cell

	// 构建空白单元格
	for i := 0; i < rowspan; i++ {
		j := 0
		if i == 0 {
			j = 1
		}
		for ; j < colspan; j++ {
			table.newSpaceCell(curCol+j, curRow+i, -curRow, -curCol)
		}
	}
	table.nextRow, table.nextCol = table.getNextRowAndCol(curRow, curCol)
	return cell
}

type TableCell struct {
	// 位置
	row, col int
	// 单元格大小
	rowspan, colspan int

	// 单元格内容
	element Element

	height float64 // 单元格的实际高度, 计算: border.top + border.bottom + (lines-1) * lineSpace + lines * lineHeight
	table  *Table
}

// Element: 创建,并且设置了字体, 偏移量
func (cell *TableCell) SetElement(e Element) *TableCell {
	cell.element = e
	cell.height = cell.element.GetHeight()
	return cell
}
