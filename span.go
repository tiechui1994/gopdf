package gopdf

import (
	"math"
	"strings"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/util"
)

// 不会进行自动分页, 可以用于页眉, 页脚的内容.
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
	x, _ := pdf.GetXY()
	endX, _ := pdf.GetPageEndXY()
	if endX-x <= 0 {
		panic("please modify current X")
	}

	f := &Span{
		pdf:        pdf,
		width:      endX - x,
		height:     lineHeight,
		lineHeight: lineHeight,
		lineSpace:  lineSpce,
	}

	return f
}

func NewSpanWithWidth(width float64, lineHeight, lineSpce float64, pdf *core.Report) *Span {
	currX, _ := pdf.GetXY()
	endX, _ := pdf.GetPageEndXY()
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
	config := span.pdf.GetConfig()

	_, height := config.GetWidthAndHeight()
	x1, _ := config.GetStart()
	x2, _ := config.GetEnd()
	x, y := span.pdf.GetXY()

	// X
	if x+margin.Left > x2 || x+margin.Left < x1 {
		return span
	}

	// Y
	if y+margin.Top > height || y+margin.Top < 0 {
		return span
	}

	// width
	if x+margin.Left+span.border.Left+span.width+span.border.Right > x2 {
		width := x2 - (x + margin.Left + span.border.Left + span.border.Right)
		if width <= 0 {
			return span
		}

		span.width = width
	}

	span.margin = margin

	return span
}
func (span *Span) SetBorder(border core.Scope) *Span {
	border.ReplaceBorder()

	config := span.pdf.GetConfig()

	x2, _ := config.GetEnd()
	x, y := span.pdf.GetXY()

	_, height := config.GetWidthAndHeight()
	if y+span.margin.Top+border.Top+border.Bottom > height {
		return span
	}

	// width
	if x+span.margin.Left+border.Left+span.width+border.Right > x2 {
		width := x2 - (x + span.margin.Left + border.Left + border.Right)
		if width <= 0 {
			return span
		}

		span.width = width
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
		blocks       = strings.Split(convertStr, "\n") // 分行
		contentWidth = span.width
	)

	// 必须检查字体
	if util.IsEmpty(span.font) {
		panic("there no avliable font")
	}

	// 必须先进行注册, 才能设置
	span.pdf.Font(span.font.Family, span.font.Size, span.font.Style)
	span.pdf.SetFontWithStyle(span.font.Family, span.font.Style, span.font.Size)

	if len(blocks) == 1 {
		if span.pdf.MeasureTextWidth(convertStr) < contentWidth {
			span.contents = []string{convertStr}
			height := span.border.Top + span.border.Bottom + span.lineHeight
			span.height = math.Max(height, span.height)

			return span
		}
	}

	for i := range blocks {
		// 单独的一行
		if span.pdf.MeasureTextWidth(convertStr) < contentWidth {
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
			if lineLength >= contentWidth {
				if lineLength-contentWidth > 2 {
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
	height := span.border.Top + span.border.Bottom + span.lineHeight*length + span.lineSpace*(length-1)
	span.height = math.Max(height, span.height)

	return span
}

func (span *Span) GenerateAtomicCell() error {
	var (
		sx, sy = span.pdf.GetXY()
		x, y   float64
		border core.Scope
	)

	if util.IsEmpty(span.font) {
		panic("no font")
	}

	span.pdf.Font(span.font.Family, span.font.Size, span.font.Style)
	span.pdf.SetFontWithStyle(span.font.Family, span.font.Style, span.font.Size)

	// 垂直居中
	if span.verticalCentered {
		length := float64(len(span.contents))
		height := (length-1)*span.lineSpace + length*span.lineHeight + span.border.Top + span.border.Bottom
		if height < span.height {
			top := (span.height - height) / 2
			span.border = core.NewScope(border.Left, top, border.Right, 0)
		}
	}

	border = span.border

	if !util.IsEmpty(span.fontColor) {
		span.pdf.TextColor(util.GetColorRGB(span.fontColor))
	}

	for i := 0; i < len(span.contents); i++ {
		// 水平居中, 只是对当前的行设置新的 Border
		if span.horizontalCentered {
			width := span.pdf.MeasureTextWidth(span.contents[i])
			if width < span.width {
				left := (border.Left + border.Right + span.width - width) / 2
				span.border = core.NewScope(left, border.Top, 0, left)
			}
		}

		// 水平居右, 只是对当前的行设置新的 Border
		if span.rightAlign {
			width := span.pdf.MeasureTextWidth(span.contents[i])
			if width < span.width {
				left := span.width - width + span.border.Right + span.border.Left
				span.border = core.NewScope(left, border.Top, 0, 0)
			}
		}

		x, y = span.getContentPosition(sx, sy, i)

		span.pdf.Cell(x, y, span.contents[i])
	}

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
