package core

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/signintech/gopdf"
)

type FontMap struct {
	FontName string
	FileName string
}

// 对接 gopdf
type Converter struct {
	GoPdf       *gopdf.GoPdf
	AtomicCells []string // 原子单元, 多个单元格最终汇总成PDF文件
	Fonts       []*FontMap
	Unit        float64
	LineW       float64
}

// var p.Unit float64 = 2.834645669

func (p *Converter) AddAtomicCell(cell string) {
	p.AtomicCells = append(p.AtomicCells, cell)
}

// 读取PDF文件(UTF-8), 解析到Text
func (p *Converter) ReadFile(fileName string) error {
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	text := strings.Replace(string(buf), "\r", "", -1)
	var UTF8_BOM = []byte{239, 187, 191}
	if text[0:3] == string(UTF8_BOM) {
		text = text[3:]
	}
	p.AtomicCells = strings.Split(text, "\n")
	return nil
}

// 将 Text(写入的, 而不是读入的) -> PDF文件
func (p *Converter) Execute() {
	lines := p.AtomicCells
	for _, line := range lines {
		elements := strings.Split(line, "|")
		switch elements[0] {
		case "P":
			p.Page(line, elements) // PDF开始, P是标准的格式, P1属于自定义的格式
		case "NP":
			p.NewPage(line, elements) // 新页面
		case "F":
			p.Font(line, elements) // 字体
		case "TC":
			p.TextColor(line, elements) // 颜色
		case "LC":
			p.LineColor(line, elements) //
		case "FC":
			p.FillColor(line, elements)
		case "GF", "GS":
			p.Grey(line, elements)
		case "C", "C1", "CR":
			p.Cell(line, elements) // 单元格
		case "M":
			p.Move(line, elements)
		case "L", "LV", "LH", "LT":
			p.Line(line, elements) // 新行
		case "R":
			p.Rect(line, elements) // 表
		case "O":
			p.Oval(line, elements) //
		case "I":
			p.Image(line, elements) // 图片
		case "Marign":
			p.Margin(line, elements)
		default:
			if len(line) > 0 && line[0:1] != "v" {
				fmt.Println("skip:" + line + ":")
			}
		}
	}
}

// 添加字体
func (p *Converter) AddFont() {
	for _, font := range p.Fonts {
		err := p.GoPdf.AddTTFFont(font.FontName, font.FileName)
		if err != nil {
			panic("font file:" + font.FileName + " not found")
		}
	}
}

// PDF文件页面的开始
// [P, mm|pt|in, A4, P|L]
// mm|pt|in 表示的尺寸单位, 毫米,像素,英尺
// P|L 表示Portait, Landscape, 表示布局
func (p *Converter) Page(line string, elements []string) {
	p.GoPdf = new(gopdf.GoPdf)

	CheckLength(line, elements, 4)
	switch elements[2] {
	/* A0 ~ A5 纸张像素表示
	'A0': [2383.94, 3370.39],
	'A1': [1683.78, 2383.94],
	'A2': [1190.55, 1683.78],
	'A3': [841.89, 1190.55],
	'A4': [595.28, 841.89],
	'A5': [419.53, 595.28],
	*/
	case "A3":
		config := defaultConfigs["A3"]
		p.setUnit(elements[1])
		if elements[3] == "P" {
			p.Start(config.width, config.height) // 像素
		} else if elements[3] == "L" {
			p.Start(config.height, config.width)
		} else {
			panic("Page Orientation accept P or L")
		}
	case "A4":
		config := defaultConfigs["A4"]
		p.setUnit(elements[1])
		if elements[3] == "P" {
			p.Start(config.width, config.height) // 像素
		} else if elements[3] == "L" {
			p.Start(config.height, config.width)
		} else {
			panic("Page Orientation accept P or L")
		}
	case "LTR":
		config := defaultConfigs["LTR"]
		p.setUnit(elements[1])
		if elements[3] == "P" {
			p.Start(config.width, config.height) // 像素
		} else if elements[3] == "L" {
			p.Start(config.height, config.width)
		} else {
			panic("Page Orientation accept P or L")
		}
	default:
		panic("This size not supported yet:" + elements[2])
	}
	p.AddFont()
	p.GoPdf.AddPage()
}

// 单位转换率设置, 基准的像素Pt
func (p *Converter) setUnit(unit string) {
	// 1mm ~ 2.8pt 1in ~ 72pt
	switch unit {
	case "mm":
		p.Unit = 2.834645669
	case "pt":
		p.Unit = 1
	case "in":
		p.Unit = 72
	default:
		panic("This unit is not specified :" + unit)
	}
}

// 构建新的页面
func (p *Converter) NewPage(line string, elements []string) {
	p.GoPdf.AddPage()
}

// 设置PDF文件基本信息(单位,页面大小)
func (p *Converter) Start(w float64, h float64) {
	p.GoPdf.Start(gopdf.Config{
		Unit:     "pt",
		PageSize: gopdf.Rect{W: w, H: h},
	}) // 595.28, 841.89 = A4
}

// 设置当前文本使用的字体
// ["", "family", "style", "size"]
// style: "" or "U", ("B", "I")(需要字体本身支持)
func (p *Converter) Font(line string, elements []string) {
	CheckLength(line, elements, 4)
	err := p.GoPdf.SetFont(elements[1], elements[2], AtoiPanic(elements[3], line))
	if err != nil {
		panic(err.Error() + " line;" + line)
	}
}

// 设置笔画的灰度 | 设置填充的灰度
// ["GF|GS", grayScale]
// grayScale: 0.0 到 1.0
func (p *Converter) Grey(line string, elements []string) {
	CheckLength(line, elements, 2)
	if elements[0] == "GF" {
		p.GoPdf.SetGrayFill(ParseFloatPanic(elements[1], line))
	}
	if elements[0] == "GS" {
		p.GoPdf.SetGrayStroke(ParseFloatPanic(elements[1], line))
	}
}

// 文本颜色
// ["", R, G, B] // RGB文本颜色
func (p *Converter) TextColor(line string, elements []string) {
	CheckLength(line, elements, 4)
	p.GoPdf.SetTextColor(uint8(AtoiPanic(elements[1], line)),
		uint8(AtoiPanic(elements[2], line)),
		uint8(AtoiPanic(elements[3], line)))
}

// 画笔颜色
// ["", R, G, B]
func (p *Converter) LineColor(line string, elements []string) {
	CheckLength(line, elements, 4)
	p.GoPdf.SetStrokeColor(uint8(AtoiPanic(elements[1], line)),
		uint8(AtoiPanic(elements[2], line)),
		uint8(AtoiPanic(elements[3], line)))
}

// 背景色
func (p *Converter) FillColor(line string, elements []string) {
	CheckLength(line, elements, 4)
	p.GoPdf.SetFillColor(uint8(AtoiPanic(elements[1], line)),
		uint8(AtoiPanic(elements[2], line)),
		uint8(AtoiPanic(elements[3], line)))
}

// 椭圆
// ["", x1, y1, x2, y2]
func (p *Converter) Oval(line string, elements []string) {
	CheckLength(line, elements, 5)
	p.GoPdf.Oval(ParseFloatPanic(elements[1], line)*p.Unit,
		ParseFloatPanic(elements[2], line)*p.Unit,
		ParseFloatPanic(elements[3], line)*p.Unit,
		ParseFloatPanic(elements[4], line)*p.Unit)
}

// 长方形
// ["R", x1, y1, x2, y2]
func (p *Converter) Rect(line string, eles []string) {
	CheckLength(line, eles, 5)
	adj := p.LineW * p.Unit * 0.5
	p.GoPdf.Line(
		ParseFloatPanic(eles[1], line)*p.Unit,
		ParseFloatPanic(eles[2], line)*p.Unit+adj,
		ParseFloatPanic(eles[3], line)*p.Unit+adj*2,
		ParseFloatPanic(eles[2], line)*p.Unit+adj)

	p.GoPdf.Line(
		ParseFloatPanic(eles[1], line)*p.Unit+adj,
		ParseFloatPanic(eles[2], line)*p.Unit,
		ParseFloatPanic(eles[1], line)*p.Unit+adj,
		ParseFloatPanic(eles[4], line)*p.Unit+adj*2)

	p.GoPdf.Line(
		ParseFloatPanic(eles[1], line)*p.Unit,
		ParseFloatPanic(eles[4], line)*p.Unit+adj,
		ParseFloatPanic(eles[3], line)*p.Unit+adj*2,
		ParseFloatPanic(eles[4], line)*p.Unit+adj)

	p.GoPdf.Line(
		ParseFloatPanic(eles[3], line)*p.Unit+adj,
		ParseFloatPanic(eles[2], line)*p.Unit,
		ParseFloatPanic(eles[3], line)*p.Unit+adj,
		ParseFloatPanic(eles[4], line)*p.Unit+adj*2)
}

// 图片
// ["I", path, x, y, x1, y2]
func (p *Converter) Image(line string, elements []string) {
	CheckLength(line, elements, 6)
	r := new(gopdf.Rect)
	r.W = ParseFloatPanic(elements[4], line)*p.Unit - ParseFloatPanic(elements[2], line)*p.Unit
	r.H = ParseFloatPanic(elements[5], line)*p.Unit - ParseFloatPanic(elements[3], line)*p.Unit

	p.GoPdf.Image(
		elements[1],
		ParseFloatPanic(elements[2], line)*p.Unit,
		ParseFloatPanic(elements[3], line)*p.Unit,
		r,
	)
}

// 线
// ["L", x1, y1, x2, y2] 两点之间的线
// ["LH", x1, y1, x2] 水平线
// ["LV", x1, y2, y2] 垂直线
// ["LT", "dashed|dotted|straight", w] 虚线,点,直线
func (p *Converter) Line(line string, elements []string) {
	switch elements[0] {
	case "L":
		CheckLength(line, elements, 5)
		p.GoPdf.Line(
			ParseFloatPanic(elements[1], line)*p.Unit,
			ParseFloatPanic(elements[2], line)*p.Unit,
			ParseFloatPanic(elements[3], line)*p.Unit,
			ParseFloatPanic(elements[4], line)*p.Unit,
		)
	case "LH":
		CheckLength(line, elements, 4)
		p.GoPdf.Line(
			ParseFloatPanic(elements[1], line)*p.Unit,
			ParseFloatPanic(elements[2], line)*p.Unit,
			ParseFloatPanic(elements[3], line)*p.Unit,
			ParseFloatPanic(elements[2], line)*p.Unit,
		)
	case "LV":
		CheckLength(line, elements, 4)
		p.GoPdf.Line(
			ParseFloatPanic(elements[1], line)*p.Unit,
			ParseFloatPanic(elements[2], line)*p.Unit,
			ParseFloatPanic(elements[1], line)*p.Unit,
			ParseFloatPanic(elements[3], line)*p.Unit,
		)
	case "LT":
		CheckLength(line, elements, 3)
		lineType := elements[1]
		if lineType == "" {
			lineType = "straight"
		}
		p.GoPdf.SetLineType(lineType)
		p.LineW = ParseFloatPanic(elements[2], line)
		p.GoPdf.SetLineWidth(p.LineW * p.Unit)
	}
}

// 单元格
// ["C", family, size, x, y, content] // 从(x,y) 位置开始写入content
// ["C1", x, y, content] // 从(x,y) 位置开始写入content
// ["CR", x, y, w, content] // 从右往左写入w长度的内容
func (p *Converter) Cell(line string, elements []string) {
	switch elements[0] {
	case "C":
		CheckLength(line, elements, 6)
		err := p.GoPdf.SetFont(elements[1], "", AtoiPanic(elements[2], line))
		if err != nil {
			panic(err.Error() + " line;" + line)
		}
		p.setPosition(elements[3], elements[4], line)
		p.GoPdf.Cell(nil, elements[5])
	case "C1":
		CheckLength(line, elements, 4)
		p.setPosition(elements[1], elements[2], line)
		p.GoPdf.Cell(nil, elements[3])
	case "CR":
		CheckLength(line, elements, 5)
		tw, err := p.GoPdf.MeasureTextWidth(elements[4])
		if err != nil {
			panic(err.Error() + " line;" + line)
		}
		x := ParseFloatPanic(elements[1], line) * p.Unit
		y := ParseFloatPanic(elements[2], line) * p.Unit
		w := ParseFloatPanic(elements[3], line) * p.Unit
		finalx := x + w - tw
		p.GoPdf.SetX(finalx)
		p.GoPdf.SetY(y)
		p.GoPdf.Cell(nil, elements[4])
	}
}

// 设置新的位置
func (p *Converter) Move(line string, eles []string) {
	CheckLength(line, eles, 3)
	p.setPosition(eles[1], eles[2], line)
}

func (p *Converter) Margin(line string, eles []string) {
	CheckLength(line, eles, 3)
	top := ParseFloatPanic(eles[1], line)
	left := ParseFloatPanic(eles[2], line)
	if top != 0.0 {
		p.GoPdf.SetTopMargin(top)
	}

	if left != 0.0 {
		p.GoPdf.SetLeftMargin(left)
	}
}

func (p *Converter) setPosition(x string, y string, line string) {
	p.GoPdf.SetX(ParseFloatPanic(x, line) * p.Unit)
	p.GoPdf.SetY(ParseFloatPanic(y, line) * p.Unit)
}

func CheckLength(line string, eles []string, no int) {
	if len(eles) < no {
		panic("Column short:" + line)
	}
}

func AtoiPanic(s string, line string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(s + " not Integer :" + line)
	}
	return i
}

func ParseFloatPanic(num string, line string) float64 {
	if num == "" {
		return 0
	}
	f, err := strconv.ParseFloat(num, 64)
	if err != nil {
		panic(num + " not Numeric :" + line)
	}
	return f
}
