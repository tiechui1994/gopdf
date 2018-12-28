package gopdf

import (
	"gopdf/core"
	"fmt"
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

	// table的位置调整
	margin scope

	cells [][]*TableCell

	// 辅助作用
	isFirstCalTableCellHeight bool
}

// row行, col列
func NewTable(cols, rows int, width float64, pdf *core.Report, lineHeight ...float64) *Table {
	pdf.LineType("straight", 0.1)
	pdf.GrayStroke(0)
	if isEmpty(lineHeight) {
		lineHeight = append(lineHeight, 2.0*pdf.GetConvPt())
	}

	t := &Table{
		pdf: pdf,

		rows:    rows,
		cols:    cols,
		width:   width,
		height:  0,
		nextCol: 0,
		nextRow: 0,

		colWidths:  []float64{},
		lineHeight: lineHeight[0],

		isFirstCalTableCellHeight: true,
	}

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

// 获取某行的高度
func (table *Table) GetRowHeightByIndex(row, col int) float64 {
	if row < 0 || row > len(table.cells) || col < 0 || col > len(table.cells[row]) {
		panic("the index out range")
	}

	if table.cells[row][col] == nil {
		panic("there has no row")
	}
	table.replaceCellHeight()

	return table.cells[row][col].height
}

// 设置表的行高, 行高必须大于当前使用字体的行高
func (table *Table) SetLineHeight(lineHeight float64) {
	table.lineHeight = lineHeight
}

// 设置表的外
func (table *Table) SetMargin(left, right, top, bottom float64) {
	table.margin.left = left
	table.margin.right = right
	table.margin.top = top
	table.margin.bottom = bottom
}

// todo: 重新计算tablecell的高度
func (table *Table) replaceCellHeight() {
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
				fmt.Println(i, j, maxHeight, cells[i][j].element.GetHeight(), table.lineHeight)
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
	x = x*table.width + table.margin.left

	for i := 0; i < row; i++ {
		if table.cells[i][0].colspan > 0 && table.cells[i][0].rowspan > 0 {
			y += table.cells[i][0].height
		}
	}
	y = y + table.margin.top
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
	x = x*table.width + table.margin.left

	for i := 0; i < row; i++ {
		y += table.cells[i][0].height
	}
	y = y + table.margin.top

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
		sx = table.pdf.CurrX
		sy = table.pdf.CurrY
	)

	// 重新计算行高
	table.replaceCellHeight()

	// 划线
	var (
		x1, y1, x2, y2 float64
	)

	for i := 0; i < table.rows; i++ {
		x1, y1, x2, y2 = table.getVLinePosition(sx, sy, 0, i)
		if y1 < 281.0 && y2 > 281.0 {
			fmt.Println(i, y1, y2)
			// todo: 写入部分数据
			x1, y1, x2, y2 = table.getVLinePosition(sx, sy, 0, 0)
			table.pdf.LineH(x1, 280.0, x1+table.width)
			table.pdf.LineV(x1+table.width, y1, 280.0)

			// todo: 增加新页面
			table.pdf.AddNewPage(false)
			table.pdf.CurrX = 10
			table.pdf.CurrY = 10
			table.cells = table.cells[i+1:]
			table.rows = len(table.cells)
			if table.rows == 0 {
				fmt.Println(table.cells)
				return nil
			}
			table.pdf.LineType("straight", 0.1)
			table.pdf.GrayStroke(0)
			table.margin.top = 0

			// todo: 增加老的没有写完的数据
			return table.GenerateAtomicCell()
		}

		for j := 0; j < table.cols; j++ {
			// 水平线
			if table.hasHLine(j, i) {
				x1, y1, x2, y2 = table.getHLinePosition(sx, sy, j, i)
				table.pdf.Line(x1, y1, x2, y2)
			}

			// 垂直线
			if table.hasVLine(j, i) {
				x1, y1, x2, y2 = table.getVLinePosition(sx, sy, j, i)
				table.pdf.Line(x1, y1, x2, y2)
			}

			// todo: 写入数据
			cell := table.cells[i][j]
			// 空白
			if cell.element == nil {
				continue
			}

			x1, y1, _, _ := table.getHLinePosition(sx, sy, j, i)
			cell.table.pdf.CurrX = x1
			cell.table.pdf.CurrY = y1
			cell.element.GenerateAtomicCell()
			cell.table.pdf.CurrX = sx
			cell.table.pdf.CurrY = sy
		}
	}

	height := table.getTableHeight()
	x1, y1, x2, y2 = table.getVLinePosition(sx, sy, 0, 0)
	table.pdf.LineH(x1, y1+height+table.margin.top, x1+table.width)
	table.pdf.LineV(x1+table.width, y1, y1+height+table.margin.top)

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
	conPt := table.pdf.GetConvPt()
	curRow, curCol := table.nextRow, table.nextCol

	cell := &TableCell{
		row:         curRow,
		col:         curCol,
		borderWidth: 1 * conPt,
		rowspan:     1,
		colspan:     1,
		lineSpace:   0.1 * conPt,
		table:       table,
		height:      table.lineHeight,
	}

	table.cells[curRow][curCol] = cell
	table.nextRow, table.nextCol = table.getNextRowAndCol(curRow, curCol)
	return cell
}

// 创建长宽为1的空白单元格
func (table *Table) newSpaceCell(col, row int, pr, pc int) *TableCell {
	cell := &TableCell{
		row:         row,
		col:         col,
		borderWidth: 0,
		colspan:     pc,
		rowspan:     pr,
		lineSpace:   0.1 * table.pdf.GetConvPt(),
		table:       table,
		height:      table.lineHeight,
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
	cell := &TableCell{
		row:         curRow,
		col:         curCol,
		borderWidth: 1,
		rowspan:     rowspan,
		colspan:     colspan,
		lineSpace:   0.1 * table.pdf.GetConvPt(),
		table:       table,
		height:      table.lineHeight,
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
	// color: R,G,B
	font            Font
	backgroundColor string
	fontColor       string

	// 位置
	row, col int
	// 外边框
	borderWidth float64
	border      scope

	// 单元格大小
	rowspan, colspan int

	// 单元格内容
	element Element

	height    float64 // 单元格的实际高度, 计算: border.top + border.bottom + (lines-1) * lineSpace + lines * lineHeight
	lineSpace float64 // 行间距
	table     *Table
}

// 设置单元格的背景颜色
func (cell *TableCell) SetBackgroundColor(color string) {
	checkColor(color)
	cell.backgroundColor = color
}

// 设置单元格的字体颜色
func (cell *TableCell) SetFontColor(color string) {
	checkColor(color)
	cell.fontColor = color
}

// 设置单元格的内边距
func (cell *TableCell) SetBorder(left, top, right, bottom, borderWidth float64) {
	cell.border.left = left
	cell.border.top = top
	cell.border.right = right
	cell.border.bottom = bottom
	if borderWidth <= 0 {
		cell.borderWidth = 1
	} else {
		cell.borderWidth = borderWidth
	}
}

func (cell *TableCell) SetFont(font Font) {
	cell.font = font
	cell.table.pdf.Font(font.Family, font.Size, font.Style)
}

func (cell *TableCell) SetElement(e Element) {
	cell.element = e
	cell.height = cell.element.GetHeight()
}
