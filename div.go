package gopdf

import (
	"math"
	"strings"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/util"
)

type Div struct {
	pdf           *core.Report
	width, height float64
	contents      []string // 内容

	lineHeight float64 // 行高
	lineSpace  float64 // 行间距

	// div的位置调整和内容调整
	margin core.Scope
	border core.Scope

	// 颜色控制
	font      core.Font
	fontColor string
	backColor string

	// 居中
	horizontalCentered bool
	verticalCentered   bool
	rightAlign         bool
}

func NewDiv(width, lineHeight, lineSpace float64, pdf *core.Report) *Div {
	// 最大宽度控制
	endX := pdf.GetPageEndX()
	curX, _ := pdf.GetXY()
	if width > endX-curX {
		width = endX - curX
	}

	div := &Div{
		width:      width,
		height:     0,
		pdf:        pdf,
		lineHeight: lineHeight,
		lineSpace:  lineSpace,
	}
	(*div).pdf = pdf

	return div
}

func (div *Div) Copy(content string) *Div {
	d := &Div{
		pdf:        div.pdf,
		width:      div.width,
		height:     0,
		lineHeight: div.lineHeight,
		lineSpace:  div.lineSpace,
		margin:     div.margin,
		border:     div.border,
		fontColor:  div.fontColor,
	}

	d.SetFont(div.font)
	d.SetContent(content)

	return d
}

func (div *Div) HorizontalCentered() *Div {
	div.horizontalCentered = true
	div.rightAlign = false
	return div
}
func (div *Div) VerticalCentered() *Div {
	div.verticalCentered = true
	return div
}
func (div *Div) RightAlign() *Div {
	div.rightAlign = true
	div.horizontalCentered = false
	return div
}

func (div *Div) SetLineSpace(lineSpace float64) *Div {
	div.lineSpace = lineSpace
	return div
}
func (div *Div) SetLineHeight(lineHeight float64) *Div {
	div.lineHeight = lineHeight
	return div
}

// 注: 当div使用为tablecell的时候, 禁止设置margin
// margin的Right无法生效,而且Bottom必须大于0
func (div *Div) SetMarign(marign core.Scope) *Div {
	marign.ReplaceMarign()
	div.margin = marign

	// 重新设置div的宽度
	endX := div.pdf.GetPageEndX()
	curX, _ := div.pdf.GetXY()
	if div.width > endX-(curX+div.margin.Left) {
		div.width = endX - (curX + div.margin.Left)
	}

	return div
}
func (div *Div) SetBorder(border core.Scope) {
	border.ReplaceBorder()
	div.border = border
}

func (div *Div) SetFontColor(color string) *Div {
	util.CheckColor(color)
	div.fontColor = color
	return div
}
func (div *Div) SetBackColor(color string) *Div {
	util.CheckColor(color)
	div.backColor = color
	return div
}

// 注册字体
func (div *Div) SetFont(font core.Font) *Div {
	div.font = font
	// 注册, 启动
	div.pdf.Font(font.Family, font.Size, font.Style)
	div.pdf.SetFontWithStyle(font.Family, font.Style, font.Size)
	return div
}
func (div *Div) SetFontWithColor(font core.Font, color string) *Div {
	div.SetFont(font)
	div.SetFontColor(color)

	return div
}

// todo: 使用注册的字体进行分行计算
func (div *Div) SetContent(s string) *Div {
	convertStr := strings.Replace(s, "\t", "    ", -1)

	var (
		unit         = div.pdf.GetUnit()
		blocks       = strings.Split(convertStr, "\n") // 分行
		contentWidth = div.width - math.Abs(div.border.Left) - math.Abs(div.border.Right)
	)

	// 必须检查字体
	if util.IsEmpty(div.font) {
		panic("there no avliable font")
	}

	// 必须先进行注册, 才能设置
	div.pdf.Font(div.font.Family, div.font.Size, div.font.Style)
	div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)
	if len(blocks) == 1 {
		if div.pdf.MeasureTextWidth(convertStr)/unit < contentWidth {
			div.contents = []string{convertStr}
			div.height = math.Abs(div.border.Top) + math.Abs(div.border.Bottom) + div.lineHeight
			return div
		}
	}

	for i := range blocks {
		// 单独的一行
		if div.pdf.MeasureTextWidth(convertStr)/unit < contentWidth {
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
			if lineLength/unit >= contentWidth {
				if lineLength-contentWidth/unit > unit*2 {
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
	div.height = div.border.Top + div.lineHeight*length + div.lineSpace*(length-1)
	return div
}

func (div *Div) GetHeight() float64 {
	return div.height
}

// 自动换行
func (div *Div) GenerateAtomicCellWithAutoPage() error {
	var (
		sx, sy                     = div.pdf.GetXY()
		x, y                       float64
		isFirstSetVerticalCentered bool
		pageEndY                   = div.pdf.GetPageEndY()
	)

	if util.IsEmpty(div.font) {
		panic("no font")
	}

	for i := 0; i < len(div.contents); i++ {
		var (
			hOriginBorder core.Scope
			vOriginBorder core.Scope
		)
		// todo: 水平居中, 只是对当前的行设置新的 Border
		if div.horizontalCentered {
			div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)
			hOriginBorder = div.border
			width := div.pdf.MeasureTextWidth(div.contents[i]) / div.pdf.GetUnit()
			if width < div.width {
				m := (div.width - width) / 2
				div.border = core.NewScope(m, hOriginBorder.Top, 0, hOriginBorder.Right)
			}
		}

		// todo: 水平居右, 只是对当前的行设置新的 Border
		if div.rightAlign {
			div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)
			hOriginBorder = div.border
			width := div.pdf.MeasureTextWidth(div.contents[i]) / div.pdf.GetUnit()
			m := div.width - width
			div.border = core.NewScope(m, hOriginBorder.Top, 0, hOriginBorder.Right)
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
				div.border = core.NewScope(vOriginBorder.Left, m, vOriginBorder.Right, 0)
			}
		}

		x, y = div.getContentPosition(sx, sy, i)

		// todo: 换页
		if y+div.lineHeight > pageEndY {
			var newX, newY float64 // 新页面的X,Y位置

			div.SetMarign(core.NewScope(div.margin.Left, 0, div.margin.Right, 0))
			div.SetBorder(core.NewScope(div.border.Left, 0, div.border.Right, 0))
			div.contents = div.contents[i:]

			_, newY = div.pdf.GetPageStartXY()
			if len(div.contents) > 0 {
				newX, _ = div.pdf.GetXY()
			} else {
				newX, _ = div.pdf.GetPageStartXY()
			}

			div.pdf.AddNewPage(false)
			div.pdf.SetXY(newX, newY)

			return div.GenerateAtomicCellWithAutoPage()
		}

		// todo: 当前页
		if !util.IsEmpty(div.fontColor) {
			div.pdf.TextColor(util.GetColorRGB(div.fontColor))
		}
		if !util.IsEmpty(div.backColor) {
			x1 := x - div.border.Left
			y1 := y
			div.pdf.BackgroundColor(x1, y1, div.width, div.lineHeight+div.lineSpace, div.backColor, "1110")
		}

		div.pdf.Font(div.font.Family, div.font.Size, div.font.Style) // 添加设置
		div.pdf.Cell(x, y, div.contents[i])

		// todo: 颜色恢复
		if !util.IsEmpty(div.fontColor) {
			div.pdf.TextDefaultColor()
		}

		if div.horizontalCentered || div.rightAlign {
			div.border = hOriginBorder
		}

		if isFirstSetVerticalCentered {
			isFirstSetVerticalCentered = false
			div.border = vOriginBorder
		}
	}

	x, _ = div.pdf.GetPageStartXY()
	div.pdf.SetXY(x, y+div.lineHeight+div.margin.Bottom) // 定格最终的位置

	return nil
}

// 非自动换行, 只写当前的页面, 不支持垂直居中??
func (div *Div) GenerateAtomicCell() error {
	var (
		x, y     float64
		sx, sy   = div.pdf.GetXY()
		pageEndY = div.pdf.GetPageEndY()
	)
	if util.IsEmpty(div.font) {
		panic("no font")
	}

	for i := 0; i < len(div.contents); i++ {
		var (
			hOriginBorder core.Scope
		)
		// todo: 水平居中, 只是对当前的行设置新的 Border
		if div.horizontalCentered {
			div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)
			hOriginBorder = div.border
			width := div.pdf.MeasureTextWidth(div.contents[i]) / div.pdf.GetUnit()
			if width < div.width {
				m := (div.width - width) / 2
				div.border = core.NewScope(m, hOriginBorder.Top, 0, hOriginBorder.Right)
			}
		}

		// todo: 水平居右, 只是对当前的行设置新的 Border
		if div.rightAlign {
			div.pdf.SetFontWithStyle(div.font.Family, div.font.Style, div.font.Size)
			hOriginBorder = div.border
			width := div.pdf.MeasureTextWidth(div.contents[i]) / div.pdf.GetUnit()
			m := div.width - width
			div.border = core.NewScope(m, hOriginBorder.Top, 0, hOriginBorder.Right)
		}

		x, y = div.getContentPosition(sx, sy, i)

		// todo: 换页
		if y+div.lineHeight >= pageEndY {
			div.contents = div.contents[i:]
			div.replaceHeight()

			if !util.IsEmpty(div.backColor) {
				x1 := x - div.border.Left
				y1 := y - div.border.Top
				h := pageEndY - y1
				div.pdf.BackgroundColor(x1, y1, div.width, h, div.backColor, "1010") // 不需要底线
			}

			div.margin.Top = 0
			return nil
		}

		// todo: 当前页,
		if !util.IsEmpty(div.fontColor) {
			div.pdf.TextColor(util.GetColorRGB(div.fontColor))
		}
		if !util.IsEmpty(div.backColor) {
			x1 := x - div.border.Left
			y1 := y - div.border.Top
			h := div.lineHeight + div.lineSpace
			// 最后一行
			if i == len(div.contents)-1 {
				h += div.border.Top
				originHeight := float64(len(div.contents))*div.lineHeight + div.border.Top + float64(len(div.contents)-1)*div.lineSpace
				h += div.height - originHeight
			}
			div.pdf.BackgroundColor(x1, y1, div.width, h, div.backColor, "1110")
		}

		div.pdf.Font(div.font.Family, div.font.Size, div.font.Style) // 添加设置
		div.pdf.Cell(x, y, div.contents[i])

		// todo:颜色恢复
		if !util.IsEmpty(div.fontColor) {
			div.pdf.TextDefaultColor()
		}

		if div.horizontalCentered || div.rightAlign {
			div.border = hOriginBorder
		}
	}

	x, _ = div.pdf.GetPageStartXY()
	div.pdf.SetXY(x, y+div.lineHeight+div.margin.Bottom) // 定格最终的位置

	return nil
}

// 重新设置div的高度
func (div *Div) replaceHeight() {
	if len(div.contents) == 0 {
		div.height = 0
	}
	length := float64(len(div.contents))
	div.height = div.lineHeight*length + div.lineSpace*(length-1) + div.border.Top
}

func (div *Div) getContentPosition(sx, sy float64, index int) (x, y float64) {
	x = sx + div.margin.Left + div.border.Left
	y = sy + div.margin.Top + div.border.Top

	y += float64(index) * (div.lineHeight + div.lineSpace)

	return x, y
}
