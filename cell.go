package gopdf

import (
	"math"
	"strings"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/util"
)

type TextCell struct {
	pdf      *core.Report
	width    float64 // 宽度, 必须
	height   float64
	contents []string // 内容
	origin   int

	lineHeight float64 // 行高
	lineSpace  float64 // 行间距

	border core.Scope // 内边距调整, left, top, right

	// 颜色控制
	font      core.Font
	fontColor string
	backColor string

	verticalCentered   bool // 垂直居中
	horizontalCentered bool // 水平居中
	rightAlign         bool // 水平居左
}

func NewTextCell(width, lineHeight, lineSpace float64, pdf *core.Report) *TextCell {
	endX := pdf.GetPageEndX()
	curX, _ := pdf.GetXY()
	if width > endX-curX {
		width = endX - curX
	}

	cell := &TextCell{
		width:      width,
		height:     0,
		pdf:        pdf,
		lineHeight: lineHeight,
		lineSpace:  lineSpace,
	}

	return cell
}

func (cell *TextCell) Copy(content string) *TextCell {
	text := &TextCell{
		pdf:        cell.pdf,
		width:      cell.width,
		height:     0,
		lineHeight: cell.lineHeight,
		lineSpace:  cell.lineSpace,
		border:     cell.border,
		fontColor:  cell.fontColor,
	}

	text.SetFont(cell.font)

	text.SetContent(content)

	return text
}

func (cell *TextCell) VerticalCentered() *TextCell {
	cell.verticalCentered = true
	return cell
}
func (cell *TextCell) HorizontalCentered() *TextCell {
	cell.rightAlign = false
	cell.horizontalCentered = true
	return cell
}
func (cell *TextCell) RightAlign() *TextCell {
	cell.horizontalCentered = false
	cell.rightAlign = true
	return cell
}

func (cell *TextCell) SetFontColor(color string) *TextCell {
	util.CheckColor(color)
	cell.fontColor = color
	return cell
}
func (cell *TextCell) SetBackColor(color string) *TextCell {
	util.CheckColor(color)
	cell.backColor = color
	return cell
}

func (cell *TextCell) SetFont(font core.Font) *TextCell {
	cell.font = font
	// 注册, 启动
	cell.pdf.Font(font.Family, font.Size, font.Style)
	cell.pdf.SetFontWithStyle(font.Family, font.Style, font.Size)
	return cell
}
func (cell *TextCell) SetFontWithColor(font core.Font, color string) *TextCell {
	cell.SetFont(font)
	cell.SetFontColor(color)
	return cell
}

func (cell *TextCell) SetBorder(border core.Scope) *TextCell {
	border.ReplaceBorder()
	cell.border = border
	return cell
}

func (cell *TextCell) SetContent(s string) *TextCell {
	convertStr := strings.Replace(s, "\t", "    ", -1)
	var (
		unit         = cell.pdf.GetUnit()
		blocks       = strings.Split(convertStr, "\n") // 分行
		contentWidth = cell.width - math.Abs(cell.border.Left) - math.Abs(cell.border.Right)
	)

	// 必须检查字体
	if util.IsEmpty(cell.font) {
		panic("there no avliable font")
	}

	// 必须先进行注册, 才能设置
	cell.pdf.Font(cell.font.Family, cell.font.Size, cell.font.Style)
	cell.pdf.SetFontWithStyle(cell.font.Family, cell.font.Style, cell.font.Size)

	if len(blocks) == 1 {
		if cell.pdf.MeasureTextWidth(convertStr)/unit < contentWidth {
			cell.contents = []string{convertStr}
			cell.height = math.Abs(cell.border.Top) + math.Abs(cell.border.Bottom) + cell.lineHeight
			return cell
		}
	}

	for i := range blocks {
		// 不需要拆分
		if cell.pdf.MeasureTextWidth(convertStr)/unit < contentWidth {
			cell.contents = append(cell.contents, blocks[i])
			continue
		}

		// 需要拆分
		var line []rune
		for _, r := range []rune(blocks[i]) {
			line = append(line, r)
			lineLength := cell.pdf.MeasureTextWidth(string(line))
			if lineLength/unit >= contentWidth {
				if lineLength-contentWidth/unit > unit*2 {
					cell.contents = append(cell.contents, string(line[0:len(line)-1]))
					line = line[len(line)-1:]
				} else {
					cell.contents = append(cell.contents, string(line))
					line = []rune{}
				}
			}
		}

		if len(line) > 0 {
			cell.contents = append(cell.contents, string(line))
		}
	}
	length := float64(len(cell.contents))
	cell.height = cell.border.Top + math.Abs(cell.border.Bottom) + cell.lineHeight*length + cell.lineSpace*(length-1)
	cell.origin = len(cell.contents)
	return cell
}

func (cell *TextCell) GenerateAtomicCell(maxheight float64) (int, int, error) {
	var (
		sx, sy = cell.pdf.GetXY() // 基准坐标
		lines  int                // 可以写入的行数
		x, y   float64            //  实际开始的坐标
	)

	cell.pdf.Font(cell.font.Family, cell.font.Size, cell.font.Style)
	cell.pdf.SetFontWithStyle(cell.font.Family, cell.font.Style, cell.font.Size)

	// 计算需要打印的行数
	if maxheight > cell.height || math.Abs(maxheight-cell.height) < 0.01 {
		lines = len(cell.contents)

		if maxheight > cell.height && cell.verticalCentered { // 垂直居中
			sy += (maxheight - cell.height) / 2
		}
	} else {
		lines = int((maxheight + cell.lineSpace) / (cell.lineHeight + cell.lineSpace))
	}

	// 背景颜色
	if !util.IsEmpty(cell.backColor) {
		cell.pdf.BackgroundColor(sx, sy, cell.width, maxheight, cell.backColor, "0000")
	}

	// 写入cell数据
	for i := 0; i < lines; i++ {
		width := cell.pdf.MeasureTextWidth(cell.contents[i]) / cell.pdf.GetUnit()
		// 水平居左
		x = sx + cell.border.Left
		// 水平居右
		if cell.rightAlign {
			x = sx + (cell.width - width - cell.border.Right)
		}
		// 水平居中
		if cell.horizontalCentered {
			x = sx + (cell.width-width)/2
		}

		y = sy + float64(i)*(cell.lineHeight+cell.lineSpace) + cell.border.Top

		// 字体颜色控制
		if !util.IsEmpty(cell.fontColor) {
			cell.pdf.TextColor(util.GetColorRGB(cell.fontColor))
		}

		cell.pdf.Cell(x, y, cell.contents[i])

		if !util.IsEmpty(cell.fontColor) {
			cell.pdf.TextDefaultColor()
		}
	}

	// cell的height和contents重置
	if lines >= len(cell.contents) {
		cell.contents = nil
	} else {
		cell.contents = cell.contents[lines:]
	}

	if len(cell.contents) == 0 {
		cell.height = 0
	} else {
		length := float64(len(cell.contents))
		cell.height = cell.border.Top + math.Abs(cell.border.Bottom) + cell.lineHeight*length + cell.lineSpace*(length-1)
	}
	return lines, len(cell.contents), nil
}

func (cell *TextCell) TryGenerateAtomicCell(maxheight float64) (int, int) {
	var (
		lines int // 可以写入的行数
	)

	cell.pdf.Font(cell.font.Family, cell.font.Size, cell.font.Style)
	cell.pdf.SetFontWithStyle(cell.font.Family, cell.font.Style, cell.font.Size)

	// 计算需要打印的行数
	if maxheight > cell.height || math.Abs(maxheight-cell.height) < 0.01 {
		lines = len(cell.contents)
	} else {
		lines = int((maxheight + cell.lineSpace) / (cell.lineHeight + cell.lineSpace))
	}

	remain := 0
	if lines < len(cell.contents) {
		remain = len(cell.contents) - lines
	}

	return lines, remain
}

func (cell *TextCell) GetHeight() float64 {
	return cell.height
}

func (cell *TextCell) GetLines() (origin, current int) {
	return cell.origin, len(cell.contents)
}
