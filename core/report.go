package core

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// 需要解决的问题: currY的控制权, 用户 -> 程序 -> 自动化操作
// 页面的三部分: Header Page Footer
const (
	Header = "Header"
	Footer = "Footer"
	Detail = "Detail"

	// flags
	Flag_AutoAddNewPage = "AutoAddNewPage"
	Flag_ResetPageNo    = "ResetPageNo"
)

type pageMark struct {
	lineNo int
	pageNo int
}

type Executor func(report *Report)
type CallBack func(report *Report)

type Report struct {
	IsMutiPage      bool
	FisrtPageNeedFH bool // 首页需要执行页眉和页脚
	Vars            map[string]string

	currX     float64
	currY     float64
	executors map[string]*Executor
	flags     map[string]bool
	sumWork   map[string]float64
	unit      float64 // 转换单位
	pageNo    int     // 记录当前的 Page 的页数
	converter *Converter

	pageWidth, pageHeight       float64
	contentWidth, contentHeight float64
	pageStartX, pageStartY      float64
	pageEndX, pageEndY          float64

	callbacks []CallBack // 在PDF生成之后执行
	config    *Config
}

func CreateReport() *Report {
	report := new(Report)
	report.converter = new(Converter)

	report.Vars = make(map[string]string)
	report.executors = make(map[string]*Executor)
	report.sumWork = make(map[string]float64)
	report.callbacks = make([]CallBack, 0)
	report.flags = make(map[string]bool)

	report.IsMutiPage = true
	report.sumWork["__ft__"] = 0.0 // FooterY
	report.flags[Flag_AutoAddNewPage] = false
	report.flags[Flag_ResetPageNo] = false

	return report
}

func (r *Report) NoCompression() {
	r.converter.GoPdf.SetNoCompression()
}

/****************************************************************
压缩级别:
	-2 只使用哈夫曼压缩,
	-1 默认值, 压缩级别的6
	0  不进行压缩,
	1  最快的压缩, 但是压缩比率不是最好的
	9  最大限度的压缩, 但是执行效率也是最慢的
****************************************************************/
func (r *Report) CompressLevel(level int) {
	r.converter.GoPdf.SetCompressLevel(level)
}

// 写入PDF文件
func (r *Report) Execute(filename string) {
	if r.config == nil {
		panic("please set page config")
	}

	r.execute(true)
	r.converter.GoPdf.WritePdf(filename)

	for i := range r.callbacks {
		r.callbacks[i](r)
	}
}

// 获取PDF内容
func (r *Report) GetBytesPdf() (ret []byte) {
	if r.config == nil {
		panic("please set page config")
	}

	r.execute(true)
	ret = r.converter.GoPdf.GetBytesPdf()
	return
}

func (r *Report) LoadCellsFromText(fileName string) error {
	return r.converter.ReadFile(fileName)
}

// 转换, 内容 -> PDF文件
func (r *Report) execute(exec bool) {
	if exec {
		r.pageNo = 1
		r.currX, r.currY = r.GetPageStartXY()

		r.addAtomicCell("v|PAGE|" + strconv.Itoa(r.pageNo))
		r.ExecuteDetail()

		r.pagination() // 分页
	}
	r.converter.Execute()
}

// 分页, 只有一个页面的PDF没有此操作
func (r *Report) pagination() {
	if r.IsMutiPage == false {
		return
	}
	lines := r.converter.AtomicCells[:]
	list := new(List)

	// 第一次遍历单元格, 确定需要创建的PDF页
	for i, line := range lines {
		if len(line) < 8 {
			continue
		}
		if line[0:7] == "v|PAGE|" {
			h := new(pageMark)
			h.lineNo = i
			h.pageNo = AtoiPanic(line[7:], line)
			list.Add(h)
			//fmt.Printf("hist %v \n", h)
		}
	}

	// 第二次遍历单元格, 检查 TotalPage
	for i, line := range lines {
		if strings.Index(line, "{#TotalPage#}") > -1 {
			total := r.getpageNoBylineNo(i, list)
			//fmt.Printf("total :%v\n", total)
			lines[i] = strings.Replace(lines[i], "{#TotalPage#}", strconv.Itoa(total), -1)
		}
	}

	cells := make([]string, 0)
	for _, line := range lines {
		cells = append(cells, line)
	}
	r.converter.AtomicCells = cells
}

// 获取 lineNo 对应的 pageNo
func (r *Report) getpageNoBylineNo(lineNo int, list *List) int {
	count := 0
	page := 0

	// 遍历到当前的lineNo, 当前的count记录的是list的索引
	for i, l := range list.GetAsArray() {
		if l.(*pageMark).lineNo >= lineNo {
			count = i
			break
		}
	}

	// 从新的页面开始, 得到页面号码
	for i := count; i < list.Size(); i++ {
		pageNo := list.Get(i).(*pageMark).pageNo // 当前item的页号
		if pageNo <= page {
			return page
		}
		page = pageNo
		//fmt.Printf("page :%v\n", page)
	}

	return page
}

// 设置可用字体
func (r *Report) SetFonts(fmap []*FontMap) {
	r.converter.Fonts = fmap
}

// 获取当前页面编号
func (r *Report) GetCurrentPageNo() int {
	return r.pageNo
}

// 添加新的页面
func (r *Report) AddNewPage(resetpageNo bool) {
	r.ExecutePageFooter()

	r.addAtomicCell("NP") // 构建新的页面
	if resetpageNo {
		r.pageNo = 1
	} else {
		r.pageNo++
	}
	r.ExecutePageHeader()

	r.addAtomicCell("v|PAGE|" + strconv.Itoa(r.pageNo))
}

func (r *Report) ExecutePageFooter() {
	r.currY = r.config.endY / r.unit
	r.currX = r.config.startX / r.unit

	h := r.executors[Footer]
	if h != nil {
		(*h)(r)
	}
}
func (r *Report) ExecutePageHeader() {
	r.currX, r.currY = r.GetPageStartXY()
	h := r.executors[Header]
	if h != nil {
		(*h)(r)
	}
}
func (r *Report) ExecuteDetail() {
	h := r.executors[Detail]
	if h != nil {
		if r.flags[Flag_AutoAddNewPage] {
			r.AddNewPage(r.flags[Flag_ResetPageNo])
			r.flags[Flag_AutoAddNewPage] = false
			r.flags[Flag_ResetPageNo] = false
		}

		if r.FisrtPageNeedFH {
			r.ExecutePageHeader()
			currX, currY := r.currX, r.currY
			r.ExecutePageFooter()
			r.currX, r.currY = currX, currY
		}

		(*h)(r)
	}
}

func (r *Report) RegisterExecutor(execuror Executor, name string) {
	r.executors[name] = &execuror
}

// 换页坐标
func (r *Report) GetPageEndY() float64 {
	return r.pageEndY / r.unit
}

func (r *Report) GetPageEndX() float64 {
	return r.pageEndX / r.unit
}

// 页面开始坐标
func (r *Report) GetPageStartXY() (x, y float64) {
	return r.pageStartX / r.unit, r.pageStartY / r.unit
}

func (r *Report) GetContentWidthAndHeight() (width, height float64) {
	return r.contentWidth / r.unit, r.contentHeight / r.unit
}

// currX, currY, 坐标
func (r *Report) SetXY(currX, currY float64) {
	if currX > 0 {
		r.currX = currX
	}

	if currY > 0 {
		r.currY = currY
	}
}
func (r *Report) GetXY() (x, y float64) {
	return r.currX, r.currY
}

func (r *Report) SetMargin(dx, dy float64) {
	x, y := r.GetXY()
	r.SetXY(x+dx, y+dy)
}

// 设置页面的尺寸, unit: mm pt in  size: A4 LTR, 目前支持常用的两种方式
func (r *Report) SetPage(size string, unit string, orientation string) {
	r.setUnit(unit)
	config, ok := defaultConfigs[size]
	if !ok {
		panic("the config not exists, please add config")
	}

	switch size {
	case "A4":
		switch orientation {
		case "P":
			r.addAtomicCell("P|" + unit + "|A4|P")
			r.pageWidth = config.width / r.unit
			r.pageHeight = config.height / r.unit
		case "L":
			r.addAtomicCell("P|" + unit + "|A4|L")
			r.pageWidth = config.height / r.unit
			r.pageHeight = config.width / r.unit
		}
	case "LTR":
		switch orientation {
		case "P":
			r.pageWidth = config.width / r.unit
			r.pageHeight = config.height / r.unit
			r.addAtomicCell("P|" + unit + "|" + strconv.FormatFloat(r.pageWidth, 'f', 4, 64) +
				"|" + strconv.FormatFloat(r.pageHeight, 'f', 4, 64))
		case "L":
			r.pageWidth = config.height / r.unit
			r.pageHeight = config.width / r.unit
			r.addAtomicCell("P  |" + unit + "|" + strconv.FormatFloat(r.pageWidth, 'f', 4, 64) +
				"|" + strconv.FormatFloat(r.pageHeight, 'f', 4, 64))
		}
	}

	r.contentWidth = config.contentWidth
	r.contentHeight = config.contentHeight

	r.pageStartX = config.startX
	r.pageStartY = config.startY
	r.pageEndX = config.endX
	r.pageEndY = config.endY
	r.config = config

	r.execute(false)
}

func (r *Report) setUnit(unit string) {
	switch unit {
	case "mm":
		r.unit = 2.834645669
	case "pt":
		r.unit = 1
	case "in":
		r.unit = 72
	default:
		panic("This unit is not specified :" + unit)
	}
}
func (r *Report) GetUnit() float64 {
	if r.unit == 0.0 {
		panic("does not set unit")
	}
	return r.unit
}

// 获取底层的所有的原子单元内容
func (r *Report) GetAtomicCells() *[]string {
	return &r.converter.AtomicCells
}

// 保存原子操作单元
func (r *Report) SaveAtomicCellText(fileName string) {
	text := strings.Join(r.converter.AtomicCells, "\n")
	ioutil.WriteFile(fileName, []byte(text), os.ModePerm)
}

// 计算文本宽度, 必须先调用 SetFontWithStyle() 或者 SetFont()
func (r *Report) MeasureTextWidth(text string) float64 {
	w, err := r.converter.GoPdf.MeasureTextWidth(text)
	if err != nil {
		panic(err)
	}
	return w
}

// 设置当前文本字体, 先注册,后设置
func (r *Report) SetFontWithStyle(family, style string, size int) {
	r.converter.GoPdf.SetFont(family, style, size)
}
func (r *Report) SetFont(family string, size int) {
	r.SetFontWithStyle(family, "", size)
}

func (r *Report) AddCallBack(callback CallBack) {
	r.callbacks = append(r.callbacks, callback)
}

/********************************************
 将特定的字符串转换成底层可以识别的原子操作符
*********************************************/
func (r *Report) addAtomicCell(s string) {
	r.converter.AddAtomicCell(s)
}

// 注册当前字体
func (r *Report) Font(fontName string, size int, style string) {
	r.addAtomicCell("F|" + fontName + "|" + style + "|" + strconv.Itoa(size))
}

// 写入字符串内容
func (r *Report) Cell(x float64, y float64, content string) {
	r.addAtomicCell("C1|" + Ftoa(x) + "|" + Ftoa(y) + "|" + content)
}
func (r *Report) CellRight(x float64, y float64, w float64, content string) {
	r.addAtomicCell("CR|" + Ftoa(x) + "|" + Ftoa(y) + "|" +
		Ftoa(w) + "|" + content)
}

// 划线
func (r *Report) LineType(ltype string, width float64) {
	r.sumWork["__lw__"] = width
	r.addAtomicCell("LT|" + ltype + "|" + Ftoa(width))
}
func (r *Report) Line(x1 float64, y1 float64, x2 float64, y2 float64) {
	r.addAtomicCell("L|" + Ftoa(x1) + "|" + Ftoa(y1) + "|" + Ftoa(x2) +
		"|" + Ftoa(y2))
}
func (r *Report) LineH(x1 float64, y float64, x2 float64) {
	adj := r.sumWork["__lw__"] * 0.5
	r.addAtomicCell("LH|" + Ftoa(x1) + "|" + Ftoa(y+adj) + "|" + Ftoa(x2))
}
func (r *Report) LineV(x float64, y1 float64, y2 float64) {
	adj := r.sumWork["__lw__"] * 0.5
	r.addAtomicCell("LV|" + Ftoa(x+adj) + "|" + Ftoa(y1) + "|" + Ftoa(y2))
}

// 画特定的图形, 目前支持: 长方形, 椭圆两大类
func (r *Report) Rect(x1 float64, y1 float64, x2 float64, y2 float64) {
	r.addAtomicCell("R|" + Ftoa(x1) + "|" + Ftoa(y1) + "|" + Ftoa(x2) +
		"|" + Ftoa(y2))
}
func (r *Report) Oval(x1 float64, y1 float64, x2 float64, y2 float64) {
	r.addAtomicCell("O|" + Ftoa(x1) + "|" + Ftoa(y1) + "|" + Ftoa(x2) +
		"|" + Ftoa(y2))
}

// 设置当前的字体颜色, 线条颜色
func (r *Report) TextColor(red int, green int, blue int) {
	r.addAtomicCell("TC|" + strconv.Itoa(red) + "|" + strconv.Itoa(green) +
		"|" + strconv.Itoa(blue))
}
func (r *Report) LineColor(red int, green int, blue int) {
	r.addAtomicCell("LC|" + strconv.Itoa(red) + "|" + strconv.Itoa(green) +
		"|" + strconv.Itoa(blue))
}

func (r *Report) FillColor(red int, green int, blue int) {
	r.addAtomicCell("FC|" + strconv.Itoa(red) + "|" + strconv.Itoa(green) +
		"|" + strconv.Itoa(blue))
}

func (r *Report) GrayColor(x, y float64, w, h float64, gray float64) {
	if gray < 0 || gray > 1 {
		gray = 0.85
	}
	r.LineType("straight", h)
	r.GrayStroke(gray)
	r.LineH(x, y, x+w)
	r.LineType("straight", 0.01)
	r.GrayStroke(0)
}

func (r *Report) GrayFill(grayScale float64) {
	r.addAtomicCell("GF|" + Ftoa(grayScale))
}
func (r *Report) GrayStroke(grayScale float64) {
	r.addAtomicCell("GS|" + Ftoa(grayScale))
}

// 图片
func (r *Report) Image(path string, x1 float64, y1 float64, x2 float64, y2 float64) {
	r.addAtomicCell("I|" + path + "|" + Ftoa(x1) + "|" + Ftoa(y1) + "|" +
		Ftoa(x2) + "|" + Ftoa(y2))
}

// 添加变量
func (r *Report) Var(name string, val string) {
	r.addAtomicCell("V|" + name + "|" + val)
}
