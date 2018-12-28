package gopdf

import (
	"gopdf/core"
	"strings"
	"math"
)

type Div struct {
	pdf           *core.Report
	width, height float64

	// 行高
	lineHeight float64
	// 行间距
	lineSpace float64

	// div的位置调整和内容调整
	margin scope
	border scope

	// 内容
	contents []string

	// 颜色控制
	font            Font
	backgroundColor string
	fontColor       string

	// 辅助作用
	isFirstSetHeight bool
}

func NewDivWithWidth(width float64, pdf *core.Report) *Div {
	conPt := pdf.GetConvPt()
	div := &Div{
		width:            width,
		height:           0,
		pdf:              pdf,
		lineHeight:       2 * conPt,
		lineSpace:        0.1 * conPt,
		isFirstSetHeight: true,
	}

	return div
}

func (div *Div) CopyWithContent(content string) *Div {
	d := div.Copy()
	d.SetContent(content)
	return div
}

func (div *Div) Copy() *Div {
	d := &Div{
		pdf:              div.pdf,
		width:            div.width,
		height:           0,
		lineHeight:       div.lineHeight,
		lineSpace:        div.lineSpace,
		margin:           div.margin,
		border:           div.border,
		backgroundColor:  div.backgroundColor,
		fontColor:        div.fontColor,
		isFirstSetHeight: true,
	}

	d.SetFont(div.font)
	return d
}

func (div *Div) SetLineSpace(lineSpace float64) *Div {
	div.lineSpace = lineSpace
	return div
}

func (div *Div) SetLineHeight(lineHeight float64) *Div {
	div.lineHeight = lineHeight
	return div
}

// 注: left和right二者取其一, top和bottom二者取其一
func (div *Div) SetMarign(left, top, right, bottom float64) *Div {
	div.margin.left = left
	div.margin.top = top
	div.margin.right = right
	div.margin.bottom = bottom

	if left != 0 {
		div.margin.right = 0
	}

	if top != 0 {
		div.margin.bottom = 0
	}
	return div
}

func (div *Div) SetBorder(left, top, right, bottom float64) *Div {
	div.border.left = left
	div.border.top = top
	div.border.right = right
	div.border.bottom = bottom

	replactBorder(&div.border)

	if top != 0 {
		div.border.bottom = 0
	}

	return div
}

func (div *Div) SetFont(font Font) *Div {
	div.font = font
	div.pdf.Font(font.Family, font.Size, font.Style)
	return div
}

func (div *Div) SetBackgroundColor(color string) {
	checkColor(color)
	div.backgroundColor = color
}

func (div *Div) SetFontColor(color string) *Div {
	checkColor(color)
	div.fontColor = color
	return div
}

func (div *Div) SetContent(s string) *Div {
	var (
		conPt        = div.pdf.GetConvPt()
		blocks       = strings.Split(s, "\n") // 分行
		contentWidth = div.width - math.Abs(div.border.left) - math.Abs(div.border.right)
	)

	if isEmpty(div.font) {
		panic("there no avliable font")
	}

	div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)

	if len(blocks) == 1 {
		if div.pdf.MeasureTextWidth(s) < contentWidth {
			div.contents = []string{s}
			div.height = math.Abs(div.border.top) + math.Abs(div.border.bottom) + div.lineHeight
			return div
		}
	}

	for i := range blocks {
		// 单独的一行
		if div.pdf.MeasureTextWidth(s) < contentWidth {
			div.contents = append(div.contents, blocks[i])
			continue
		}

		var (
			line []rune
		)
		// 单独的一行需要拆分
		for _, r := range []rune(blocks[i]) {
			line = append(line, r)
			lineLength := div.pdf.MeasureTextWidth(string(line))
			if lineLength >= contentWidth {
				if lineLength-contentWidth > conPt*2 {
					div.contents = append(div.contents, string(line[0:len(line)-1]))
					line = line[len(line)-1:]
				} else {
					div.contents = append(div.contents, string(line))
					line = []rune{}
				}
			}
		}

		// 剩余单独成行
		if len(line) > 0 {
			div.contents = append(div.contents, string(line))
		}
	}

	// 重新计算 div 的高度
	length := float64(len(div.contents))
	div.height = math.Abs(div.border.top) + math.Abs(div.border.bottom) + div.lineHeight*length + div.lineSpace*(length-1)
	return div
}

// 同步操作, 只能操作一次
func (div *Div) SetHeight(height float64) {
	if div.isFirstSetHeight {
		div.height = height
		div.isFirstSetHeight = false
		return
	}

	panic("olny one time set height")
}

func (div *Div) GetHeight() float64 {
	if isEmpty(div.contents) {
		panic("has no content")
	}

	return div.height
}

// 自动换行
func (div *Div) GenerateAtomicCellWithAutoWarp() error {
	var (
		sx   = div.pdf.CurrX
		sy   = div.pdf.CurrY
		x, y float64
	)

	if isEmpty(div.font) {
		panic("no font")
	}

	for i := 0; i < len(div.contents); i++ {
		x, y = div.getContentPosition(sx, sy, i)
		// 换页的依据
		if y < 281.0 && y+div.lineHeight > 281.0 {
			div.SetMarign(div.margin.left, 0, div.margin.right, 0)
			div.SetBorder(div.border.left, 0, div.border.right, 0)
			div.contents = div.contents[i-1:]
			div.pdf.AddNewPage(false)
			div.pdf.CurrX = 10
			div.pdf.CurrY = 10
			return div.GenerateAtomicCellWithAutoWarp()
		}

		// 当前页
		div.pdf.Cell(x, y, div.contents[i])
	}

	return nil
}

// 非自动换行, 只写当前的页面
func (div *Div) GenerateAtomicCell() error {
	var (
		sx   = div.pdf.CurrX
		sy   = div.pdf.CurrY
		x, y float64
	)

	if isEmpty(div.font) {
		panic("no font")
	}

	for i := 0; i < len(div.contents); i++ {
		x, y = div.getContentPosition(sx, sy, i)
		// 换页的依据
		if y < 281.0 && y+div.lineHeight > 281.0 {
			return nil
		}

		// 当前页
		div.pdf.Cell(x, y, div.contents[i])
	}

	return nil
}

// 重新设置div的高度
func (div *Div) replaceHeight() {
	length := float64(len(div.contents))
	div.height = div.lineHeight*length + div.lineSpace*(length-1)
}

func (div *Div) getContentPosition(sx, sy float64, index int) (x, y float64) {
	x = sx + div.margin.left + div.border.left
	y = sy + div.margin.top + div.border.top
	if index > 0 {
		y += float64(index) * (div.lineHeight + div.lineSpace)
	}
	return x, y
}
