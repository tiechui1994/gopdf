package gopdf

import (
	"math"
	"strings"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/util"
)

// 边框
type Span struct {
	pdf  *core.Report
	font core.Font

	width, height float64
	lineHeight    float64
	lineSpace     float64

	fontColor string
	margin    core.Scope
	border    core.Scope

	contents           []string
	horizontalCentered bool
	verticalCentered   bool
	rightAlign         bool
}

func NewSpan(lineHeight, lineSpce float64, pdf *core.Report) *Span {
	currX, _ := pdf.GetXY()
	endX := pdf.GetPageEndX()
	if endX-currX <= 0 {
		panic("please modify current X")
	}

	f := &Span{
		pdf:        pdf,
		width:      endX - currX,
		height:     lineHeight,
		lineHeight: lineHeight,
		lineSpace:  lineSpce,
	}

	return f
}

func NewSpanWithWidth(width float64, lineHeight, lineSpce float64, pdf *core.Report) *Span {
	currX, _ := pdf.GetXY()
	endX := pdf.GetPageEndX()
	if endX-currX <= 0 {
		panic("please modify current X")
	}

	if endX-currX <= width {
		width = endX - currX
	}

	f := &Span{
		pdf:        pdf,
		width:      width,
		height:     lineHeight,
		lineHeight: lineHeight,
		lineSpace:  lineSpce,
	}

	return f
}

func (span *Span) Copy(content string) *Span {
	f := &Span{
		pdf:        span.pdf,
		width:      span.width,
		lineHeight: span.lineHeight,
		lineSpace:  span.lineSpace,
		fontColor:  span.fontColor,
	}

	f.SetBorder(span.border)
	f.SetFont(span.font)
	f.SetContent(content)

	return f
}

func (span *Span) SetHeight(height float64) *Span {
	span.height = height
	return span
}
func (span *Span) SetMarign(margin core.Scope) *Span {
	margin.ReplaceMarign()
	currX, _ := span.pdf.GetXY()
	endX := span.pdf.GetPageEndX()

	if endX-(currX+margin.Left) <= 0 {
		panic("the marign out of page boundary")
	}

	// 宽度检测
	if endX-(currX+margin.Left) <= span.width {
		span.width = endX - (currX + margin.Left)
	}

	span.margin = margin

	return span
}
func (span *Span) SetBorder(border core.Scope) *Span {
	border.ReplaceBorder()
	currX, _ := span.pdf.GetXY()
	endX := span.pdf.GetPageEndX()

	// 最大宽度检测
	if endX-(currX+span.margin.Left) >= span.width+border.Left+border.Right {
		span.width += border.Left + border.Right
	}

	span.border = border

	return span
}

func (span *Span) GetHeight() (height float64) {
	return span.height
}
func (span *Span) GetWidth() (width float64) {
	return span.width
}

func (span *Span) HorizontalCentered() *Span {
	span.horizontalCentered = true
	span.rightAlign = false
	return span
}
func (span *Span) VerticalCentered() *Span {
	span.verticalCentered = true
	return span
}
func (span *Span) RightAlign() *Span {
	span.rightAlign = true
	span.horizontalCentered = false
	return span
}

func (span *Span) SetFontColor(color string) *Span {
	util.CheckColor(color)
	span.fontColor = color
	return span
}
func (span *Span) SetFont(font core.Font) *Span {
	span.font = font
	// 注册, 启动
	span.pdf.Font(font.Family, font.Size, font.Style)
	span.pdf.SetFontWithStyle(font.Family, font.Style, font.Size)

	return span
}
func (span *Span) SetFontWithColor(font core.Font, color string) *Span {
	span.SetFont(font)
	span.SetFontColor(color)
	return span
}

func (span *Span) SetContent(content string) *Span {
	convertStr := strings.Replace(content, "\t", "    ", -1)

	var (
		unit         = span.pdf.GetUnit()
		blocks       = strings.Split(convertStr, "\n") // 分行
		contentWidth = span.width - math.Abs(span.border.Left) - math.Abs(span.border.Right)
	)

	// 必须检查字体
	if util.IsEmpty(span.font) {
		panic("there no avliable font")
	}

	// 必须先进行注册, 才能设置
	span.pdf.Font(span.font.Family, span.font.Size, span.font.Style)
	span.pdf.SetFontWithStyle(span.font.Family, span.font.Style, span.font.Size)

	if len(blocks) == 1 {
		if span.pdf.MeasureTextWidth(convertStr)/unit < contentWidth {
			span.contents = []string{convertStr}
			span.height = math.Abs(span.border.Top) + math.Abs(span.border.Bottom) + span.lineHeight
			return span
		}
	}

	for i := range blocks {
		// 单独的一行
		if span.pdf.MeasureTextWidth(convertStr)/unit < contentWidth {
			span.contents = append(span.contents, blocks[i])
			continue
		}

		var (
			line []rune
		)
		// 单独的一行需要拆分
		for _, r := range []rune(blocks[i]) {
			line = append(line, r)
			lineLength := span.pdf.MeasureTextWidth(string(line))
			if lineLength/unit >= contentWidth {
				if lineLength-contentWidth/unit > unit*2 {
					span.contents = append(span.contents, string(line[0:len(line)-1]))
					line = line[len(line)-1:]
				} else {
					span.contents = append(span.contents, string(line))
					line = []rune{}
				}
			}
		}

		// 剩余单独成行
		if len(line) > 0 {
			span.contents = append(span.contents, string(line))
		}
	}

	// 重新计算 span 的高度
	length := float64(len(span.contents))
	height := span.border.Top + span.lineHeight*length + span.lineSpace*(length-1)
	span.height = math.Max(height, span.height)

	return span
}

// 非自动分页
func (span *Span) GenerateAtomicCell() error {
	var (
		sx, sy = span.pdf.GetXY()
		x, y   float64
		border core.Scope
	)

	if util.IsEmpty(span.font) {
		panic("no font")
	}

	border = span.border
	span.pdf.Font(span.font.Family, span.font.Size, span.font.Style)
	span.pdf.SetFontWithStyle(span.font.Family, span.font.Style, span.font.Size)

	// todo: 垂直居中
	if span.verticalCentered {
		length := float64(len(span.contents))
		height := (length-1)*span.lineSpace + length*span.lineHeight + span.border.Top
		if height < span.height {
			top := (span.height - height) / 2
			span.border = core.NewScope(border.Left, top, border.Right, 0)
		}
	}

	if !util.IsEmpty(span.fontColor) {
		span.pdf.TextColor(util.GetColorRGB(span.fontColor))
	}

	for i := 0; i < len(span.contents); i++ {
		// todo: 水平居中, 只是对当前的行设置新的 Border
		if span.horizontalCentered {
			width := span.pdf.MeasureTextWidth(span.contents[i]) / span.pdf.GetUnit()
			if width < span.width {
				left := (span.width - width) / 2
				span.border = core.NewScope(left, border.Top, 0, border.Right)
			}
		}

		// todo: 水平居右, 只是对当前的行设置新的 Border
		if span.rightAlign {
			width := span.pdf.MeasureTextWidth(span.contents[i]) / span.pdf.GetUnit()
			left := span.width - width
			span.border = core.NewScope(left, border.Top, 0, border.Right)
		}

		x, y = span.getContentPosition(sx, sy, i)

		span.pdf.Cell(x, y, span.contents[i])
	}

	// todo: 颜色恢复
	if !util.IsEmpty(span.fontColor) {
		span.pdf.TextDefaultColor()
	}

	x, _ = span.pdf.GetPageStartXY()
	span.pdf.SetXY(x, y+span.lineHeight+span.margin.Bottom) // 定格最终的位置

	return nil
}

func (span *Span) getContentPosition(sx, sy float64, index int) (x, y float64) {
	x = sx + span.margin.Left + span.border.Left
	y = sy + span.margin.Top + span.border.Top

	y += float64(index) * (span.lineHeight + span.lineSpace)

	return x, y
}
