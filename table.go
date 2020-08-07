package gopdf

import (
	"fmt"
	"math"
	"strconv"

	"github.com/tiechui1994/gopdf/core"
)

/**
Table写入的实现思路:
	先构建一个标准table(n*m),
	然后在标准的table基础上构建不规则的table.

标准table: (4*5)
+-----+-----+-----+-----+-----+
|(0,0)|(0,1)|(0,2)|(0,3)|(0,4)|
+-----+-----+-----+-----+-----+
|(1,0)|(1,1)|(1,2)|(1,3)|(1,4)|
+-----+-----+-----+-----+-----+
|(2,0)|(2,1)|(2,2)|(2,3)|(2,4)|
+-----+-----+-----+-----+-----+
|(3,0)|(3,1)|(3,2)|(3,3)|(3,4)|
+-----+-----+-----+-----+-----+

不规则table: 在 (4*5) 基础上构建
+-----+-----+-----+-----+-----+
|(0,0)      |(0,2)      |(0,4)|
+	        +-----+-----+     +
|           |(1,2)      |     |
+-----+-----+-----+-----+     +
|(2,0)            |(2,3)|     |
+                 +     +-----+
|                 |     |(3,4)|
+-----+-----+-----+-----+-----+

在创建不规则table的时候, 先构建标准table, 然后再 "描述" 不规则table,
一旦不规则 table 描述完毕之后, 其余的工作交由程序, 自动分页, 换行去生成
描述的不规则table.

Table主要负载的是生成最终表格.
core.Cell接口的实现类可以生成自定义的单元格. 默认的一个实现是TextCell, 基于纯文本的Cell

空白cell: 比如(0,1), (1,0) 等, 非用户写入单元格内容的开始位置, 辅助作用.
非空白cell: 比如(0,0), (0,2), (3,4)等, 用户要写入单元格内容的开始位置(table当中)

注: 目前在table的分页当中, 背景颜色和线条存在bug.
**/

//Table 构建表格
type Table struct {
	pdf           *core.Report
	rows, cols    int // 行数,列数
	width, height float64
	colwidths     []SizedColumn  // 列宽百分比: 应加起来为1
	rowheights    []float64      // 保存行高
	cells         [][]*TableCell // 单元格

	lineHeight float64    // 默认行高
	margin     core.Scope // 位置调整

	nextrow, nextcol int // 下一个位置

	tableCheck  bool                // table 完整性检查
	cachedRow   []float64           // 缓存行
	cachedCol   []float64           // 缓存列
	sizedColumn map[int]SizedColumn // 缓存列
}

//SizedColumn 列的宽度
type SizedColumn struct {
	col   int     //列序号
	width float64 //列宽，是一个小于1的浮点数，且所有列宽的总数为1
}

//TableCell 表格的单元格
type TableCell struct {
	table            *Table // table元素
	row, col         int    // 位置
	rowspan, colspan int    // 单元格大小

	element    core.Cell // 单元格元素
	minheight  float64   // 当前最小单元格的高度, rowspan=1, 辅助计算
	height     float64   // 当前表格单元真实高度, rowspan >= 1, 实际计算垂直线高度的时候使用
	cellwrited int       // 写入的行数
}

//SetElement 设置单元格的元素
func (cell *TableCell) SetElement(e core.Cell) *TableCell {
	cell.element = e
	return cell
}

//NewTable 新建表格
func NewTable(cols, rows int, width, lineHeight float64, pdf *core.Report) *Table {
	contentWidth, _ := pdf.GetContentWidthAndHeight()
	if width > contentWidth {
		width = contentWidth
	}

	t := &Table{
		pdf:    pdf,
		rows:   rows,
		cols:   cols,
		width:  width,
		height: 0,

		nextcol: 0,
		nextrow: 0,

		lineHeight:  lineHeight,
		colwidths:   []SizedColumn{},
		rowheights:  []float64{},
		sizedColumn: map[int]SizedColumn{},
	}
	//TODO:列宽处理
	for i := 0; i < cols; i++ {
		t.colwidths = append(t.colwidths, SizedColumn{col: i, width: float64(1.0) / float64(cols)})
	}

	cells := make([][]*TableCell, rows)
	for i := range cells {
		cells[i] = make([]*TableCell, cols)
	}

	t.cells = cells

	return t
}

//NewCell 创建长宽为1的单元格
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

//NewCellByRange 创建固定长度的单元格
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

// SetSpan 设置合并单元格
func (table *Table) SetSpan(row, col int, rowspan, colspan int) {
	// 防止非法的宽度
	if !table.checkSpan(row, col, rowspan, colspan) {
		panic("inlivid layout, please check w and h")
	}
	cell := table.cells[row][col]
	cell.rowspan = rowspan
	cell.colspan = colspan
	cell.height = table.lineHeight * float64(rowspan)
	cell.element.SetWidth(table.GetColWidth(row, col))
	if rowspan > 1 {
		for i := 1; i < rowspan; i++ {
			nextCell := table.cells[row+i][col]
			nextCell.element = nil
			nextCell.rowspan = 0
			nextCell.colspan = 0
		}
	}
	if colspan > 1 {
		for i := 1; i < colspan; i++ {
			nextCell := table.cells[row][col+i]
			nextCell.element = nil
			nextCell.rowspan = 0
			nextCell.colspan = 0
		}
	}
}

//checkSpan 检测当前cell的宽和高是否合法
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

	if rowspan == 1 && colspan <= maxcol || colspan == 1 && rowspan <= maxrow {
		return true
	}

	// 检测合法性
	if colspan <= maxcol && rowspan <= maxrow {
		for i := row; i < row+rowspan; i++ {
			for j := col; j < col+colspan; j++ {
				if cells[i][j] != nil {
					return false
				}
			}
		}

		return true
	}

	return false
}

//setNext 设置下一个单元格开始坐标
func (table *Table) setNext(colspan, rowspan int) {
	table.nextcol += colspan
	if table.nextcol == table.cols {
		table.nextcol = 0
		table.nextrow++
	}

	// 获取最近行的空白Cell的坐标
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

//SetColumnWidth 设置指定列的列宽
func (table *Table) SetColumnWidth(col int, width float64) {
	// 查找所有设置过列宽的列，然后计算出设置后的列宽,假设列总宽为x，剩余的列平均分配1-x
	var totalSizedWidth float64
	sizedColumnCount := len(table.sizedColumn)
	for i := 0; i < sizedColumnCount; i++ {
		totalSizedWidth += table.sizedColumn[i].width
	}
	newTotalSizedWidth := totalSizedWidth + width
	if newTotalSizedWidth < 1 {
		table.sizedColumn[col] = SizedColumn{col: col, width: width}
		table.colwidths[col].width = width
	}
	sizedColumnCount++
	if table.cols == sizedColumnCount {
		return
	}
	existNotSizedWidth := floatRound((1-newTotalSizedWidth)/float64(table.cols-sizedColumnCount), 2)
	for i := 0; i < table.cols; i++ {
		if _, ok := table.sizedColumn[i]; ok == false {
			table.colwidths[i].width = existNotSizedWidth
		}
	}
	// 验证所有的宽度是否超过了1
	var total float64
	for i := 0; i < table.cols; i++ {
		total += table.colwidths[i].width
	}
	table.colwidths[table.cols-1].width = table.colwidths[table.cols-1].width + 1.0 - total
}

// 截取小数位数
func floatRound(f float64, n int) float64 {
	format := "%." + strconv.Itoa(n) + "f"
	res, _ := strconv.ParseFloat(fmt.Sprintf(format, f), 64)
	return res
}

//GetColWidth 获取某列的宽度
func (table *Table) GetColWidth(row, col int) float64 {
	if row < 0 || row > len(table.cells) || col < 0 || col > len(table.cells[row]) {
		panic("the index out range")
	}

	count := 0.0
	for i := 0; i < table.cells[row][col].colspan; i++ {
		count += table.colwidths[i+col].width * table.width
	}

	return count
}

//SetLineHeight 设置表的行高, 行高必须大于当前使用字体的行高
func (table *Table) SetLineHeight(lineHeight float64) {
	table.lineHeight = lineHeight
}

//SetMargin 设置表的外
func (table *Table) SetMargin(margin core.Scope) {
	margin.ReplaceMarign()
	table.margin = margin
}

//GenerateAtomicCell 自动新建单元格
func (table *Table) GenerateAtomicCell() error {
	var (
		sx, sy        = table.pdf.GetXY() // 基准坐标
		_, pageEndY   = table.pdf.GetPageEndXY()
		x1, y1, _, y2 float64 // 当前位置
	)

	// 重新计算行高, 并且缓存每个位置的开始坐标
	table.resetCellHeight()
	table.cachedPoints(sx, sy)

	for i := 0; i < table.rows; i++ {
		for j := 0; j < table.cols; j++ {
			// 这是遍历的每一个cell的rowspan是1
			_, y1, _, y2 = table.getVLinePosition(sx, sy, j, i)
			if table.cells[i][j].rowspan > 1 {
				y2 = y1 + table.cells[i][j].minheight
			}

			// 换页
			if y1 < pageEndY && y2 > pageEndY {
				cell := table.cells[i][j]
				if cell.row == 0 && cell.col == 0 && !table.checkFirstRowCanWrite(sx, sy) {
					table.pdf.AddNewPage(false)
					table.margin.Top = 0
					table.pdf.SetXY(table.pdf.GetPageStartXY())
					return table.GenerateAtomicCell()
				}
				// 写完剩余的内容
				table.writeCurrentPageRestCells(i, j, sx, sy)

				// 画当前页面边框线
				table.drawPageLines(sx, sy)

				// 重置tableCells
				table.resetTableCells()

				// 相关动态变量重置
				table.pdf.AddNewPage(false)
				table.margin.Top = 0
				table.rows = len(table.cells)
				table.pdf.SetXY(table.pdf.GetPageStartXY())

				table.pdf.LineType("straight", 0.1)

				if table.rows == 0 {
					return nil
				}

				return table.GenerateAtomicCell()
			}

			if table.cells[i][j].element == nil {
				continue
			}

			x1, y1, _, y2 = table.getVLinePosition(sx, sy, j, i) // 真实的垂直线

			// 当前cell高度跨页
			if y1 < pageEndY && y2 > pageEndY {
				table.writePartialPageCell(i, j, sx, sy) // 部分写入
			}

			// 当前celll没有跨页
			if y1 < pageEndY && y2 < pageEndY {
				table.writeCurrentPageCell(i, j, sx, sy)
			}
		}
	}

	// 最后一个页面的最后部分
	table.drawLastPageLines(sx, sy)

	// 重置当前的坐标(非常重要)
	height := table.getLastPageHeight()
	_, y1, _, y2 = table.getVLinePosition(sx, sy, 0, 0)
	x1, _ = table.pdf.GetPageStartXY()
	table.pdf.SetXY(x1, y1+height+table.margin.Top+table.margin.Bottom)

	return nil
}

func (table *Table) checkFirstRowCanWrite(sx, sy float64) (ok bool) {
	var (
		_, pageEndY = table.pdf.GetPageEndXY()
	)

	row := 0
	for col := 0; col < table.cols; col++ {
		cell := table.cells[row][col]
		_, y, _, _ := table.getHLinePosition(sx, sy, col, 0)

		if cell.rowspan < 1 {
			continue
		}

		if cell.rowspan >= 1 {
			wn, _ := cell.element.TryGenerateAtomicCell(pageEndY - y)
			if wn > 0 {
				ok = true
				return ok
			}
		}
	}

	return ok
}

// row,col 定位cell, sx,sy是table基准坐标
func (table *Table) writeCurrentPageCell(row, col int, sx, sy float64) {
	var (
		x1, y1, _, y2 float64
		_, pageEndY   = table.pdf.GetPageEndXY()
		cell          = table.cells[row][col]
	)

	x1, y1, _, y2 = table.getVLinePosition(sx, sy, col, row)
	cell.table.pdf.SetXY(x1, y1)

	if cell.element != nil {
		// 检查当前Cell下面的Cell能否写入(下一个Cell跨页), 如果不能写入, 需要修正写入的高度值[颜色控制]
		i, j := cell.row+cell.rowspan-table.cells[0][0].row, cell.col-table.cells[0][0].col
		if i < len(table.cells) {
			_, y3, _, y4 := table.getVLinePosition(sx, sy, j, i)
			if y3 < pageEndY && y4 >= pageEndY {
				if !table.checkNextCellCanWrite(sx, sy, row, col) {
					y2 = pageEndY
				}
			}
		}

		if cell.element.GetHeight() == 0 {
			cell.element.GenerateAtomicCell(y2 - y1)
			cell.cellwrited = cell.rowspan
			return
		}

		cell.element.GenerateAtomicCell(y2 - y1)
		cell.cellwrited = cell.rowspan
	}
}
func (table *Table) writePartialPageCell(row, col int, sx, sy float64) {
	var (
		x1, y1      float64
		_, pageEndY = table.pdf.GetPageEndXY()
		cell        = table.cells[row][col]
	)

	x1, y1, _, _ = table.getVLinePosition(sx, sy, col, row) // 垂直线
	cell.table.pdf.SetXY(x1, y1)

	if cell.element != nil {
		if cell.element.GetHeight() == 0 {
			cell.element.GenerateAtomicCell(pageEndY - y1)
			cell.cellwrited = cell.rowspan
			return
		}

		// 尝试写入(跨页的Cell), 写不进去就不再写
		wn, _ := cell.element.TryGenerateAtomicCell(pageEndY - y1)
		if wn == 0 {
			cell.cellwrited = 0
			return
		}

		// 真正的写入
		wn, _, _ = cell.element.GenerateAtomicCell(pageEndY - y1)

		// 设置 cellwrited 的值
		if wn > 0 && cell.element.GetHeight() == 0 {
			cell.cellwrited = cell.rowspan
		}

		if wn > 0 && cell.element.GetHeight() != 0 {
			if cell.rowspan == 1 {
				cell.cellwrited = 0
			}

			if cell.rowspan > 1 {
				count := 0
				for i := row; i < row+cell.rowspan; i++ {
					_, y1, _, y2 := table.getVLinePosition(sx, sy, col, i)
					if table.cells[i][col].element != nil {
						y2 = y1 + table.cells[i][col].minheight
					}

					if y1 < pageEndY && y2 <= pageEndY {
						count++
					}
					if y1 > pageEndY || y2 > pageEndY {
						break
					}
				}

				cell.cellwrited = count
			}
		}
	}
}

// 当前页面的剩余内容
func (table *Table) writeCurrentPageRestCells(row, col int, sx, sy float64) {
	for i := col; i < table.cols; i++ {
		table.writePartialPageCell(row, i, sx, sy)
	}
}

// 检查下一个Cell是否可以写入(当前的Cell必须是非空格Cell)
func (table *Table) checkNextCellCanWrite(sx, sy float64, row, col int) bool {
	var (
		canwrite    bool
		cells       = table.cells
		_, pageEndY = table.pdf.GetPageEndXY()
	)

	if cells[row][col].rowspan <= 0 {
		return canwrite
	}

	// 当前cell的下一行
	nextrow := cells[row][col].row + cells[row][col].rowspan - cells[0][0].row
	for k := col; k < table.cols; k++ {
		cell := cells[nextrow][col]
		_, y, _, _ := table.getHLinePosition(sx, sy, col, nextrow)

		// 空格Cell -> 寻找非空格Cell
		if cell.rowspan <= 0 {
			i, j := -cell.rowspan-cells[0][0].row, -cell.colspan-cells[0][0].col
			wn, _ := cells[i][j].element.TryGenerateAtomicCell(pageEndY - y)
			if wn > 0 {
				canwrite = true
				return canwrite
			}
		}

		// 非空格Cell
		if cell.rowspan >= 1 {
			wn, _ := cell.element.TryGenerateAtomicCell(pageEndY - y)
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
		_, pageEndY         = table.pdf.GetPageEndXY()
		x, y, x1, y1, _, y2 float64
	)

	// 计算当前页面最大的rows
	_, y1 = table.pdf.GetPageStartXY()
	_, y2 = table.pdf.GetPageEndXY()
	if rows > int((y2-y1)/table.lineHeight)+1 {
		rows = int((y2-y1)/table.lineHeight) + 1
	}

	table.pdf.LineType("straight", 0.1)

	// 两条水平线
	x, y, _, _ = table.getHLinePosition(sx, sy, 0, 0)
	table.pdf.LineH(x, y, x+table.width)
	table.pdf.LineH(x, pageEndY, x+table.width)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := table.cells[row][col]

			if cell.element == nil {
				continue
			}

			// 坐标化
			x, y, x1, y1 = table.getHLinePosition(sx, sy, col, row)
			x, y, _, y2 = table.getVLinePosition(sx, sy, col, row)

			// cell没有跨页
			if y1 < pageEndY && y2 < pageEndY {
				// cell的下一个cell跨页, 需要判断下一个cell是否写入, 当写入了, 需要底线, 否则, 不需要底线
				i, j := cell.row+cell.rowspan-table.cells[0][0].row, cell.col-table.cells[0][0].col
				_, y3, _, y4 := table.getVLinePosition(sx, sy, j, i)
				if y3 < pageEndY && y4 >= pageEndY {
					if !table.checkNextCellWrited(row, col) {
						y2 = pageEndY
						table.pdf.LineV(x1, y1, y2)
						continue
					}
				}

				table.pdf.LineV(x1, y1, y2)
				table.pdf.LineH(x, y2, x1)
			}

			// cell跨页, 需要先判断是否需要竖线
			if y1 < pageEndY && y2 >= pageEndY {
				if table.checkNeedVline(row, col) {
					table.pdf.LineV(x1, y1, pageEndY)
				}

				table.pdf.LineH(x, pageEndY, x1)
			}
		}
	}

	// 两条垂直线
	x, y, _, _ = table.getHLinePosition(sx, sy, 0, 0)
	table.pdf.LineV(x, y, pageEndY)
	table.pdf.LineV(x+table.width, y, pageEndY)
}

// 最后一页画线(基本参考了drawPageLines)
func (table *Table) drawLastPageLines(sx, sy float64) {
	var (
		rows, cols          = table.rows, table.cols
		pageEndY            = table.getLastPageHeight()
		x, y, x1, y1, _, y2 float64
	)

	table.pdf.LineType("straight", 0.1)

	x, y, _, _ = table.getHLinePosition(sx, sy, 0, 0)
	pageEndY = y + pageEndY

	table.pdf.LineH(x, y, x+table.width)
	table.pdf.LineH(x, pageEndY, x+table.width)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := table.cells[row][col]

			if cell.element == nil {
				continue
			}

			x, y, x1, y1 = table.getHLinePosition(sx, sy, col, row)
			x, y, _, y2 = table.getVLinePosition(sx, sy, col, row)

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

func (table *Table) checkNextCellWrited(row, col int) bool {
	var (
		cells      = table.cells
		cellwrited bool
	)

	if cells[row][col].rowspan <= 0 {
		return cellwrited
	}

	// row,col所定位的cell必须是非空白cell, 需要定位到下一个非空白cell
	nextrow := (cells[row][col].row - cells[0][0].row) + cells[row][col].rowspan

	// 如果nexrow行, 从col开始, 一直到col+colspan, 有内容写入, 则说明下一个单元格有内容写入,
	// 否则, 没有内容写入
	for k := col; k < col+cells[row][col].colspan; k++ {
		cell := cells[nextrow][col]

		// 空白cell
		if cell.rowspan <= 0 {
			// rowspan = 1 和 rowsapn != 1 需要区别对待, 原因是rowsapn=1, 即使写入了, cellwrited还是可能为0
			i, j := -cell.rowspan-cells[0][0].row, -cell.colspan-cells[0][0].col
			if cells[i][j].rowspan == 1 {
				height := cells[i][j].element.GetHeight()
				lastheight := cells[i][j].element.GetLastHeight()
				if math.Abs(lastheight-height) > 0.1 {
					return true
				}
			}

			if cells[i][j].rowspan > 1 {
				if cells[i][j].cellwrited >= cell.row-cells[i][j].row+1 {
					return true
				}
			}
		}

		// 非空白cell
		if cell.rowspan >= 1 {
			height := cell.element.GetHeight()
			lastheight := cell.element.GetLastHeight()
			if math.Abs(lastheight-height) > 0.1 {
				return true
			}
		}
	}

	return cellwrited
}

func (table *Table) checkNeedVline(row, col int) bool {
	var (
		negwrited bool
		curwrited bool
		cells     = table.cells
		origin    = cells[0][0]
	)

	if cells[row][col].rowspan <= 0 {
		return negwrited || curwrited
	}

	// row,col 所确定的cell必须是非空白cell. 只有当前非空白cell没有写入 && 邻居cell没有写入 => false, 否则, 返回 true

	// 当前的cell
	if cells[row][col].rowspan >= 1 {
		height := cells[row][col].element.GetHeight()
		lastheight := cells[row][col].element.GetLastHeight()
		if math.Abs(lastheight-height) > 0.1 || cells[row][col].cellwrited == cells[row][col].rowspan {
			curwrited = true
		}
	}

	// 邻居cell
	nextcol := cells[row][col].col + cells[row][col].colspan - cells[0][0].col
	if nextcol == table.cols {
		return true
	}

	if cells[row][nextcol].rowspan <= 0 {
		row, nextcol = -cells[row][nextcol].rowspan-origin.row, -cells[row][nextcol].colspan-origin.col
	}
	if cells[row][nextcol].rowspan >= 1 || cells[row][nextcol].cellwrited == cells[row][nextcol].rowspan {
		height := cells[row][nextcol].element.GetHeight()
		lastheight := cells[row][nextcol].element.GetLastHeight()
		if math.Abs(lastheight-height) > 0.1 {
			return true
		}
	}

	return negwrited || curwrited
}

// 重新计算 tablecell 的高度(精确)
func (table *Table) resetCellHeight() {
	table.checkTableConstraint()

	// 计算当前页面最大的rows
	_, y1 := table.pdf.GetPageStartXY()
	_, y2 := table.pdf.GetPageEndXY()
	rows := table.rows
	if rows > int((y2-y1)/table.lineHeight)+1 {
		rows = int((y2-y1)/table.lineHeight) + 1
	}
	cells := table.cells

	// 对于cells的元素重新赋值height和minheight
	for i := 0; i < rows; i++ {
		for j := 0; j < table.cols; j++ {
			cells[i][j].minheight = table.lineHeight
			cells[i][j].height = table.lineHeight

			if cells[i][j].element != nil {
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
			if max < cells[i][j].minheight {
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

						// 更新height, 当rowspan=1
						if cells[row][col].rowspan == 1 {
							cells[row][col].height += h
						}

						// 更新height, 当rowspan<0, 空格,需要更新非当前的实体(前面的)
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

func (table *Table) getMaxWriteLineNo() int {
	var (
		cells    = table.cells
		writenum int
	)

	_, y1 := table.pdf.GetPageStartXY()
	_, y2 := table.pdf.GetPageEndXY()
	rows := table.rows
	if rows > int((y2-y1)/table.lineHeight)+1 {
		rows = int((y2-y1)/table.lineHeight) + 1
	}

	// todo: 计算当前的已经当前行全部写入的最大的行数.
	for i := 0; i < rows; i++ {
		for j := 0; j < table.cols; j++ {
			cell := cells[i][j]
			if cell.rowspan < 1 {
				row, col := -cell.rowspan-cells[0][0].row, -cell.colspan-cells[0][0].col
				if cells[row][col].row+cells[row][col].cellwrited <= cell.row {
					return writenum
				}
			}

			if cell.rowspan == 1 {
				if cell.element.GetHeight() != 0 {
					return writenum
				}
			}

			if cell.rowspan > 1 {
				if cell.cellwrited < 1 {
					return writenum
				}
			}
		}

		writenum = i + 1
	}

	return writenum
}

// 重置cells
func (table *Table) resetTableCells() {
	var (
		cells  = table.cells
		origin = cells[0][0]
	)

	writenum := table.getMaxWriteLineNo()
	// TODO: fix writenum is out of table.cells
	if writenum >= len(table.cells) {
		table.cells = table.cells[writenum:]
		return
	}

	// cell重置,需要修正空格Cell
	row := writenum
	for col := 0; col < table.cols; {
		cell := table.cells[row][col]

		if cell.rowspan < 1 {
			var ox, oy int

			orow, ocol := -cell.rowspan-origin.row, -cell.colspan-origin.col
			for x := row; x < cells[orow][ocol].row+cells[orow][ocol].rowspan-origin.row; x++ {
				for y := col; y < col+cells[orow][ocol].colspan; y++ {
					if x == row && y == col {
						ox, oy = cells[x][y].row, cells[x][y].col

						cells[x][y].element = cells[orow][ocol].element
						cells[x][y].rowspan = cells[orow][ocol].rowspan - (ox - cells[orow][ocol].row)
						cells[x][y].colspan = cells[orow][ocol].colspan
						cells[x][y].cellwrited = 0

						continue
					}

					cells[x][y].rowspan = -ox
					cells[x][y].colspan = -oy
				}
			}

			col += cells[orow][ocol].colspan
			continue
		}

		if cell.rowspan >= 1 {
			col += cell.colspan
			cell.cellwrited = 0
			continue
		}
	}

	table.cells = table.cells[writenum:]
}

func (table *Table) cachedPoints(sx, sy float64) {
	// 只计算当前页面最大的rows
	_, y1 := table.pdf.GetPageStartXY()
	_, y2 := table.pdf.GetPageEndXY()
	rows := table.rows
	if rows > int((y2-y1)/table.lineHeight)+1 {
		rows = int((y2-y1)/table.lineHeight) + 1
	}

	var (
		x, y = sx + table.margin.Left, sy + table.margin.Top
	)

	// 只会缓存一次
	if table.cachedCol == nil {
		table.cachedCol = make([]float64, table.cols)

		for col := 0; col < table.cols; col++ {
			table.cachedCol[col] = x
			x += table.colwidths[col].width * table.width
		}
	}
	table.cachedRow = make([]float64, rows)

	for row := 0; row < rows; row++ {
		table.cachedRow[row] = y
		y += table.cells[row][0].minheight
	}
}

// 垂直线, table单元格的垂直线
func (table *Table) getVLinePosition(sx, sy float64, col, row int) (x1, y1 float64, x2, y2 float64) {
	var (
		x, y float64
		cell = table.cells[row][col]
	)

	x = table.cachedCol[col]
	y = table.cachedRow[row]

	return x, y, x, y + cell.height
}

// 水平线, table单元格的水平线
func (table *Table) getHLinePosition(sx, sy float64, col, row int) (x1, y1 float64, x2, y2 float64) {
	var (
		x, y float64
	)

	x = table.cachedCol[col]
	y = table.cachedRow[row]

	cell := table.cells[row][col]
	if cell.colspan > 1 {
		if cell.col+cell.colspan == table.cols {
			x1 = table.cachedCol[0] + table.width
		} else {
			x1 = table.cachedCol[cell.col+cell.colspan]
		}
	} else {
		x1 = x + table.colwidths[col].width*table.width
	}

	return x, y, x1, y
}

// 获取表的垂直高度
func (table *Table) getLastPageHeight() float64 {
	var count float64
	for i := 0; i < table.rows; i++ {
		count += table.cells[i][0].minheight
	}
	return count
}

// 校验table是否合法(只做一次)
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
				cells++
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
