package gopdf

import (
	"math"
	"strings"

	"github.com/tiechui1994/gopdf/core"
)

const (
	FRAME_STRAIGHT = 1 // 实线边框
	FRAME_DASHED   = 2 // 虚线边框
	FRAME_DOTTED   = 3 // 点状线的边框
	FRAME_NONE     = 4 // 无边框
)

// 边框
type Frame struct {
	pdf       *core.Report
	font      Font
	frameType int // 边框类型

	width, height float64
	lineHeight    float64
	lineSpace     float64

	fontColor string
	backColor float64 // 灰度颜色, 0-1.0

	margin Scope
	border Scope

	contents           []string
	horizontalCentered bool
	verticalCentered   bool
	rightAlign         bool
}

func NewFrame(lineHeight, lineSpce float64, pdf *core.Report) *Frame {
	currX, _ := pdf.GetXY()
	endX := pdf.GetPageEndX()
	if endX-currX <= 0 {
		panic("please modify current X")
	}

	f := &Frame{
		pdf:        pdf,
		frameType:  FRAME_NONE,
		width:      endX - currX,
		height:     lineHeight,
		lineHeight: lineHeight,
		lineSpace:  lineSpce,
	}

	return f
}

func NewFrameWithWidth(width float64, lineHeight, lineSpce float64, pdf *core.Report) *Frame {
	currX, _ := pdf.GetXY()
	endX := pdf.GetPageEndX()
	if endX-currX <= 0 {
		panic("please modify current X")
	}

	if endX-currX <= width {
		width = endX - currX
	}

	f := &Frame{
		pdf:        pdf,
		frameType:  FRAME_NONE,
		width:      width,
		height:     lineHeight,
		lineHeight: lineHeight,
		lineSpace:  lineSpce,
	}

	return f
}

func (frame *Frame) CopyWithContent(content string) *Frame {
	f := &Frame{
		pdf:        frame.pdf,
		frameType:  frame.frameType,
		width:      frame.width,
		lineHeight: frame.lineHeight,
		lineSpace:  frame.lineSpace,
		fontColor:  frame.fontColor,
		backColor:  frame.backColor,
	}

	f.SetBorder(frame.border)
	f.SetFont(frame.font)
	f.SetContent(content)

	return f
}

func (frame *Frame) SetMarign(margin Scope) *Frame {
	replaceMarign(&margin)
	currX, _ := frame.pdf.GetXY()
	endX := frame.pdf.GetPageEndX()

	if endX-(currX+margin.Left) <= 0 {
		panic("the marign out of page boundary")
	}

	// 宽度检测
	if endX-(currX+margin.Left) <= frame.width {
		frame.width = endX - (currX + margin.Left)
	}

	frame.margin = margin

	return frame
}

func (frame *Frame) SetBorder(border Scope) *Frame {
	replaceBorder(&border)
	currX, _ := frame.pdf.GetXY()
	endX := frame.pdf.GetPageEndX()

	// 最大宽度检测
	if endX-(currX+frame.margin.Left) >= frame.width+border.Left+border.Right {
		frame.width += border.Left + border.Right
	}

	frame.border = border

	return frame
}

func (frame *Frame) SetFrameType(frameType int) *Frame {
	if frameType < FRAME_STRAIGHT || frameType > FRAME_NONE {
		return frame
	}

	frame.frameType = frameType

	return frame
}

func (frame *Frame) GetHeight() (height float64) {
	return frame.height
}

func (frame *Frame) GetWidth() (width float64) {
	return frame.width
}

// 水平居中
func (frame *Frame) SetHorizontalCentered() *Frame {
	frame.horizontalCentered = true
	frame.rightAlign = false
	return frame
}

// 垂直居中
func (frame *Frame) SetVerticalCentered() *Frame {
	frame.verticalCentered = true
	return frame
}

// 居右
func (frame *Frame) SetRightAlign() *Frame {
	frame.rightAlign = true
	frame.horizontalCentered = false
	return frame
}

func (frame *Frame) SetFont(font Font) *Frame {
	frame.font = font
	// 注册, 启动
	frame.pdf.Font(font.Family, font.Size, font.Style)
	frame.pdf.SetFontWithStyle(font.Family, font.Style, font.Size)

	return frame
}

func (frame *Frame) SetContent(content string) *Frame {
	convertStr := strings.Replace(content, "\t", "    ", -1)

	var (
		unit         = frame.pdf.GetUnit()
		blocks       = strings.Split(convertStr, "\n") // 分行
		contentWidth = frame.width - math.Abs(frame.border.Left) - math.Abs(frame.border.Right)
	)

	// 必须检查字体
	if isEmpty(frame.font) {
		panic("there no avliable font")
	}

	// 必须先进行注册, 才能设置
	frame.pdf.Font(frame.font.Family, frame.font.Size, frame.font.Style)
	frame.pdf.SetFontWithStyle(frame.font.Family, frame.font.Style, frame.font.Size)
	if len(blocks) == 1 {
		if frame.pdf.MeasureTextWidth(convertStr)/unit < contentWidth {
			frame.contents = []string{convertStr}
			frame.height = math.Abs(frame.border.Top) + math.Abs(frame.border.Bottom) + frame.lineHeight
			return frame
		}
	}

	for i := range blocks {
		// 单独的一行
		if frame.pdf.MeasureTextWidth(convertStr)/unit < contentWidth {
			frame.contents = append(frame.contents, blocks[i])
			continue
		}

		var (
			line []rune
		)
		// 单独的一行需要拆分
		for _, r := range []rune(blocks[i]) {
			line = append(line, r)
			lineLength := frame.pdf.MeasureTextWidth(string(line))
			if lineLength/unit >= contentWidth {
				if lineLength-contentWidth/unit > unit*2 {
					frame.contents = append(frame.contents, string(line[0:len(line)-1]))
					line = line[len(line)-1:]
				} else {
					frame.contents = append(frame.contents, string(line))
					line = []rune{}
				}
			}
		}

		// 剩余单独成行
		if len(line) > 0 {
			frame.contents = append(frame.contents, string(line))
		}
	}

	// 重新计算 frame 的高度
	length := float64(len(frame.contents))
	frame.height = frame.border.Top + frame.lineHeight*length + frame.lineSpace*(length-1)

	return frame
}

// 自动分页
func (frame *Frame) GenerateAtomicCellWithAutoPage() error {
	var (
		sx, sy                     = frame.pdf.GetXY()
		x, y                       float64
		isFirstSetVerticalCentered bool
		pageEndY                   = frame.pdf.GetPageEndY()
	)

	if isEmpty(frame.font) {
		panic("no font")
	}

	switch frame.frameType {
	case FRAME_STRAIGHT:
		frame.pdf.LineType("straight", 0.01)
	case FRAME_DASHED:
		frame.pdf.LineType("dashed", 0.01)
	case FRAME_DOTTED:
		frame.pdf.LineType("dotted", 0.01)
	}

	for i := 0; i < len(frame.contents); i++ {
		var (
			hOriginBorder Scope
			vOriginBorder Scope
		)
		// todo: 水平居中, 只是对当前的行设置新的 Border
		if frame.horizontalCentered {
			frame.pdf.SetFontWithStyle(frame.font.Family, frame.font.Style, frame.font.Size)
			hOriginBorder = frame.border
			width := frame.pdf.MeasureTextWidth(frame.contents[i]) / frame.pdf.GetUnit()
			if width < frame.width {
				m := (frame.width - width) / 2
				frame.border = Scope{m, hOriginBorder.Top, 0, hOriginBorder.Right}
			}
		}

		// todo: 水平居右, 只是对当前的行设置新的 Border
		if frame.rightAlign {
			frame.pdf.SetFontWithStyle(frame.font.Family, frame.font.Style, frame.font.Size)
			hOriginBorder = frame.border
			width := frame.pdf.MeasureTextWidth(frame.contents[i]) / frame.pdf.GetUnit()
			m := frame.width - width
			frame.border = Scope{m, hOriginBorder.Top, 0, hOriginBorder.Right}
		}

		// todo: 垂直居中, 只能操作一次
		if i == 0 && frame.verticalCentered {
			isFirstSetVerticalCentered = true
			frame.verticalCentered = false
			vOriginBorder = frame.border
			length := float64(len(frame.contents))
			height := (length-1)*frame.lineSpace + length*frame.lineHeight + frame.border.Top
			if height < frame.height {
				m := (frame.height - height) / 2
				frame.border = Scope{vOriginBorder.Left, m, vOriginBorder.Right, 0}
			}
		}

		x, y = frame.getContentPosition(sx, sy, i)

		// todo: 换页的依据
		if (y < pageEndY || y >= pageEndY) && y+frame.lineHeight > pageEndY {
			frame.SetMarign(Scope{frame.margin.Left, 0, frame.margin.Right, 0})
			frame.SetBorder(Scope{frame.border.Left, 0, frame.border.Right, 0})
			frame.contents = frame.contents[i:]
			frame.replaceHeight()

			// 两条竖线 + 一条横线
			if frame.frameType != FRAME_NONE {
				frame.pdf.LineV(sx+frame.margin.Left, y, y+frame.lineHeight+frame.lineSpace)
				frame.pdf.LineV(sx+frame.margin.Left+frame.width, y, y+frame.lineHeight+frame.lineSpace)
				frame.pdf.LineH(sx+frame.margin.Left, y+frame.lineHeight+frame.lineSpace,
					sx+frame.margin.Left+frame.width)
			}

			frame.pdf.AddNewPage(false)
			frame.pdf.SetXY(frame.pdf.GetPageStartXY())
			return frame.GenerateAtomicCellWithAutoPage()
		}

		// todo: 不需要换页, 只需要增加数据
		if !isEmpty(frame.fontColor) {
			frame.pdf.TextColor(getColorRGB(frame.fontColor))
		}

		if !isEmpty(frame.backColor) {
			x1 := x - frame.border.Left
			y1 := y
			frame.pdf.GrayColor(x1, y1, frame.width, frame.lineHeight+frame.lineSpace, frame.backColor)
		}

		frame.pdf.Font(frame.font.Family, frame.font.Size, frame.font.Style) // 添加设置
		frame.pdf.Cell(x, y, frame.contents[i])

		if frame.horizontalCentered || frame.rightAlign {
			frame.border = hOriginBorder
		}

		if isFirstSetVerticalCentered {
			isFirstSetVerticalCentered = false
			frame.border = vOriginBorder
		}

		if i == 0 && frame.frameType != FRAME_NONE {
			// 两条竖线 + 一条横线
			frame.pdf.LineH(sx+frame.margin.Left, y-frame.margin.Top, sx+frame.margin.Left+frame.width)
			frame.pdf.LineV(sx+frame.margin.Left, y-frame.margin.Top, y+frame.lineHeight+frame.lineSpace)
			frame.pdf.LineV(sx+frame.margin.Left+frame.width, y-frame.margin.Top, y+frame.lineHeight+frame.lineSpace)
			continue
		}

		// 两条竖线
		if frame.frameType != FRAME_NONE {
			frame.pdf.LineV(sx+frame.margin.Left, y, y+frame.lineHeight+frame.lineSpace)
			frame.pdf.LineV(sx+frame.margin.Left+frame.width, y, y+frame.lineHeight+frame.lineSpace)
		}

		// 一条横线
		if i == len(frame.contents)-1 && frame.frameType != FRAME_NONE {
			frame.pdf.LineH(sx+frame.margin.Left, y+frame.lineHeight+frame.lineSpace,
				sx+frame.margin.Left+frame.width)
		}
	}

	if !isEmpty(frame.fontColor) {
		frame.pdf.TextColor(getColorRGB(frame.fontColor))
	}
	x, _ = frame.pdf.GetPageStartXY()
	frame.pdf.SetXY(x, y+frame.lineHeight+frame.margin.Bottom) // 定格最终的位置

	return nil
}

// 非自动分页
func (frame *Frame) GenerateAtomicCell() error {
	var (
		sx, sy                     = frame.pdf.GetXY()
		x, y                       float64
		isFirstSetVerticalCentered bool
	)

	if isEmpty(frame.font) {
		panic("no font")
	}

	switch frame.frameType {
	case FRAME_STRAIGHT:
		frame.pdf.LineType("straight", 0.01)
	case FRAME_DASHED:
		frame.pdf.LineType("dashed", 0.01)
	case FRAME_DOTTED:
		frame.pdf.LineType("dotted", 0.01)
	}

	for i := 0; i < len(frame.contents); i++ {
		var (
			hOriginBorder Scope
			vOriginBorder Scope
		)
		// todo: 水平居中, 只是对当前的行设置新的 Border
		if frame.horizontalCentered {
			frame.pdf.SetFontWithStyle(frame.font.Family, frame.font.Style, frame.font.Size)
			hOriginBorder = frame.border
			width := frame.pdf.MeasureTextWidth(frame.contents[i]) / frame.pdf.GetUnit()
			if width < frame.width {
				m := (frame.width - width) / 2
				frame.border = Scope{m, hOriginBorder.Top, 0, hOriginBorder.Right}
			}
		}

		// todo: 水平居右, 只是对当前的行设置新的 Border
		if frame.rightAlign {
			frame.pdf.SetFontWithStyle(frame.font.Family, frame.font.Style, frame.font.Size)
			hOriginBorder = frame.border
			width := frame.pdf.MeasureTextWidth(frame.contents[i]) / frame.pdf.GetUnit()
			m := frame.width - width
			frame.border = Scope{m, hOriginBorder.Top, 0, hOriginBorder.Right}
		}

		// todo: 垂直居中, 只能操作一次
		if i == 0 && frame.verticalCentered {
			isFirstSetVerticalCentered = true
			frame.verticalCentered = false
			vOriginBorder = frame.border
			length := float64(len(frame.contents))
			height := (length-1)*frame.lineSpace + length*frame.lineHeight + frame.border.Top
			if height < frame.height {
				m := (frame.height - height) / 2
				frame.border = Scope{vOriginBorder.Left, m, vOriginBorder.Right, 0}
			}
		}

		x, y = frame.getContentPosition(sx, sy, i)

		// todo: 不需要换页, 只需要增加数据
		if !isEmpty(frame.fontColor) {
			frame.pdf.TextColor(getColorRGB(frame.fontColor))
		}

		if !isEmpty(frame.backColor) {
			x1 := x - frame.border.Left
			y1 := y
			frame.pdf.GrayColor(x1, y1, frame.width, frame.lineHeight+frame.lineSpace, frame.backColor)
		}

		frame.pdf.Font(frame.font.Family, frame.font.Size, frame.font.Style) // 添加设置
		frame.pdf.Cell(x, y, frame.contents[i])

		if frame.horizontalCentered || frame.rightAlign {
			frame.border = hOriginBorder
		}

		if isFirstSetVerticalCentered {
			isFirstSetVerticalCentered = false
			frame.border = vOriginBorder
		}

		if i == 0 && frame.frameType != FRAME_NONE {
			// 两条竖线 + 一条横线
			frame.pdf.LineH(sx+frame.margin.Left, y-frame.margin.Top, sx+frame.margin.Left+frame.width)
			frame.pdf.LineV(sx+frame.margin.Left, y-frame.margin.Top, y+frame.lineHeight+frame.lineSpace)
			frame.pdf.LineV(sx+frame.margin.Left+frame.width, y-frame.margin.Top, y+frame.lineHeight+frame.lineSpace)
			continue
		}

		// 两条竖线
		if frame.frameType != FRAME_NONE {
			frame.pdf.LineV(sx+frame.margin.Left, y, y+frame.lineHeight+frame.lineSpace)
			frame.pdf.LineV(sx+frame.margin.Left+frame.width, y, y+frame.lineHeight+frame.lineSpace)
		}

		// 一条横线
		if i == len(frame.contents)-1 && frame.frameType != FRAME_NONE {
			frame.pdf.LineH(sx+frame.margin.Left, y+frame.lineHeight+frame.lineSpace,
				sx+frame.margin.Left+frame.width)
		}
	}

	if !isEmpty(frame.fontColor) {
		frame.pdf.TextColor(getColorRGB(frame.fontColor))
	}
	x, _ = frame.pdf.GetPageStartXY()
	frame.pdf.SetXY(x, y+frame.lineHeight+frame.margin.Bottom) // 定格最终的位置

	return nil

}

func (frame *Frame) replaceHeight() {
	if len(frame.contents) == 0 {
		frame.height = 0
	}
	length := float64(len(frame.contents))
	frame.height = frame.lineHeight*length + frame.lineSpace*(length-1) + frame.border.Top
}

func (frame *Frame) getContentPosition(sx, sy float64, index int) (x, y float64) {
	x = sx + frame.margin.Left + frame.border.Left
	y = sy + frame.margin.Top + frame.border.Top

	y += float64(index) * (frame.lineHeight + frame.lineSpace)

	return x, y
}
