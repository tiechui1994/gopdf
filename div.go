package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
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
	margin Scope
	border Scope

	// 内容
	contents []string

	// 颜色控制
	font      Font
	fontColor string

	// 辅助作用
	isFirstSetHeight bool

	// 居中
	horizontalCentered bool
	verticalCentered   bool
}

func NewDivWithWidth(width, lineHeight, lineSpace float64, pdf *core.Report) *Div {
	div := &Div{
		width:            width,
		height:           0,
		pdf:              pdf,
		lineHeight:       lineHeight,
		lineSpace:        lineSpace,
		isFirstSetHeight: true,
	}
	(*div).pdf = pdf

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
func (div *Div) SetMarign(marign Scope) *Div {
	div.margin = marign

	if marign.Left != 0 {
		div.margin.Right = 0
	}

	if marign.Top != 0 {
		div.margin.Bottom = 0
	}
	return div
}

func (div *Div) SetFontColor(color string) *Div {
	checkColor(color)
	div.fontColor = color
	return div
}

// 注册字体
func (div *Div) SetFont(font Font) *Div {
	div.font = font
	div.pdf.Font(font.Family, font.Size, font.Style) // 可以多次注册
	return div
}

func (div *Div) SetBorder(border Scope) {
	if isEmpty(div.font) {
		panic("must set font")
	}

	div.border = border
	replaceBorder(&div.border)

	if border.Top != 0 {
		div.border.Bottom = 0
	}
}

// 水平居中
func (div *Div) SetHorizontalCentered() *Div {
	if isEmpty(div.font) {
		panic("must set font")
	}

	div.horizontalCentered = true
	return div
}

// 垂直居中
func (div *Div) SetVerticalCentered() *Div {
	if isEmpty(div.font) {
		panic("must set font")
	}

	div.verticalCentered = true
	return div
}

// todo: 使用注册的字体进行分行计算
func (div *Div) SetContent(s string) *Div {
	var (
		convertStr   = strings.Replace(s, "|", `\|`, -1)
		conPt        = div.pdf.GetConvPt()
		blocks       = strings.Split(convertStr, "\n") // 分行
		contentWidth = div.width - math.Abs(div.border.Left) - math.Abs(div.border.Right)
	)

	if isEmpty(div.font) {
		panic("there no avliable font")
	}

	// 必须先进行注册, 才能设置
	div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)
	if len(blocks) == 1 {
		if div.pdf.MeasureTextWidth(convertStr)/conPt < contentWidth {
			div.contents = []string{convertStr}
			div.height = math.Abs(div.border.Top) + math.Abs(div.border.Bottom) + div.lineHeight
			return div
		}
	}

	for i := range blocks {
		// 单独的一行
		if div.pdf.MeasureTextWidth(convertStr)/conPt < contentWidth {
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
			if lineLength/conPt >= contentWidth {
				if lineLength-contentWidth/conPt > conPt*2 {
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
	div.height = div.border.Top + div.border.Bottom + div.lineHeight*length + div.lineSpace*(length-1)
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
		sx, sy                     = div.pdf.GetXY()
		x, y                       float64
		isFirstSetVerticalCentered bool
		pageEndY                   = div.pdf.GetPageEndY()
	)

	if isEmpty(div.font) {
		panic("no font")
	}

	for i := 0; i < len(div.contents); i++ {
		var (
			hOriginBorder Scope
			vOriginBorder Scope
		)
		// todo: 水平居中, 只是对当前的行设置新的 Border
		if div.horizontalCentered {
			div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)
			hOriginBorder = div.border
			width := div.pdf.MeasureTextWidth(div.contents[i]) / div.pdf.GetConvPt()
			if width < div.width {
				m := (div.width - width) / 2
				div.border = Scope{m, hOriginBorder.Top, 0, hOriginBorder.Right}
			}
		}

		// todo: 垂直居中, 只能操作一次
		if i == 0 && div.verticalCentered {
			isFirstSetVerticalCentered = true
			div.verticalCentered = false
			vOriginBorder = div.border
			length := float64(len(div.contents))
			height := (length-1)*div.lineSpace + length*div.lineHeight + div.border.Top
			if height < div.height {
				m := (div.height - height) / 2
				div.border = Scope{vOriginBorder.Left, m, vOriginBorder.Right, 0}
			}
		}

		x, y = div.getContentPosition(sx, sy, i)

		// todo: 换页的依据
		if y < pageEndY && y+div.lineHeight > pageEndY {
			div.SetMarign(Scope{div.margin.Left, 0, div.margin.Right, 0})
			div.SetBorder(Scope{div.border.Left, 0, div.border.Right, 0})
			div.contents = div.contents[i:]
			div.pdf.AddNewPage(false)
			div.pdf.SetXY(div.pdf.GetPageStartXY())
			return div.GenerateAtomicCellWithAutoWarp()
		}

		// todo: 不需要换页, 只需要增加数据
		div.pdf.Cell(x, y, div.contents[i])

		if div.horizontalCentered {
			div.border = hOriginBorder
		}

		if isFirstSetVerticalCentered {
			isFirstSetVerticalCentered = false
			div.border = vOriginBorder
		}
	}

	if !isEmpty(div.fontColor) {
		div.pdf.TextColor(getColorRGB(div.fontColor))
	}
	div.pdf.SetXY(0, y+div.lineHeight) // 定格最终的位置
	return nil
}

// 非自动换行, 只写当前的页面
func (div *Div) GenerateAtomicCell() error {
	var (
		isFirstSetVerticalCentered bool
		x, y                       float64
		sx, sy                     = div.pdf.GetXY()
		pageEndY                   = div.pdf.GetPageEndY()
	)
	if isEmpty(div.font) {
		panic("no font")
	}

	for i := 0; i < len(div.contents); i++ {
		var (
			hOriginBorder Scope
			vOriginBorder Scope
		)
		// todo: 水平居中, 只是对当前的行设置新的 Border
		if div.horizontalCentered {
			div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)
			hOriginBorder = div.border
			width := div.pdf.MeasureTextWidth(div.contents[i]) / div.pdf.GetConvPt()
			if width < div.width {
				m := (div.width - width) / 2
				div.border = Scope{m, hOriginBorder.Top, 0, hOriginBorder.Right}
			}
		}

		// todo: 垂直居中, 只能操作一次, 只对第1行设置 Border
		if i == 0 && div.verticalCentered {
			isFirstSetVerticalCentered = true
			div.verticalCentered = false
			vOriginBorder = div.border
			length := float64(len(div.contents))
			height := (length-1)*div.lineSpace + length*div.lineHeight + div.border.Top
			if height < div.height {
				m := (div.height - height) / 2
				div.border = Scope{vOriginBorder.Left, m, vOriginBorder.Right, 0}
			}
		}

		x, y = div.getContentPosition(sx, sy, i)
		// 换页的依据
		if y < pageEndY && y+div.lineHeight > pageEndY {
			div.contents = div.contents[i:]
			div.replaceHeight()
			div.margin.Left = 0
			return nil
		}

		// 当前页
		if !isEmpty(div.fontColor) {
			div.pdf.TextColor(getColorRGB(div.fontColor))
		}
		div.pdf.Cell(x, y, div.contents[i])

		if div.horizontalCentered {
			div.border = hOriginBorder
		}

		if isFirstSetVerticalCentered {
			isFirstSetVerticalCentered = false
			div.border = vOriginBorder
		}
	}

	return nil
}

// 重新设置div的高度
func (div *Div) replaceHeight() {
	length := float64(len(div.contents))
	div.height = div.lineHeight*length + div.lineSpace*(length-1) + div.border.Top
}

func (div *Div) getContentPosition(sx, sy float64, index int) (x, y float64) {
	x = sx + div.margin.Left + div.border.Left
	y = sy + div.margin.Top + div.border.Top

	if index > 0 {
		y += float64(index) * (div.lineHeight + div.lineSpace)
	}
	return x, y
}
