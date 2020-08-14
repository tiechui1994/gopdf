package core

import (
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/tiechui1994/gopdf/util"
)

const (
	// A PDF page is divided into three parts, the header(Header), the footer(Footer) and
	// the main body(Detail).
	//
	// Each part has its own executor. The main actuator is necessary because it is the most
	// important content that constitutes the PDF page.
	Header = "Header"
	Footer = "Footer"
	Detail = "Detail"

	// flags
	Flag_AutoAddNewPage = "AutoAddNewPage"
	Flag_ResetPageNo    = "ResetPageNo"
)

var (
	rline *regexp.Regexp
)

func init() {
	rline, _ = regexp.Compile(`^[01]+$`)
}

type pageMark struct {
	lineNo int
	pageNo int
}

type Executor func(report *Report)
type CallBack func(report *Report)

type Report struct {
	FisrtPageNeedHeader bool // Wheather the first page need to execute the header
	FisrtPageNeedFooter bool // Wheather the first page need to execute the header
	Vars                map[string]string

	converter    *Converter           // converter engine (connect with third-party libraries)
	config       *Config              // PDF page config
	currX, currY float64              // position
	executors    map[string]*Executor // executors, include Header, Detail, and Footer
	flags        map[string]bool      // Mark (automatic paging and reset page number)
	pageNo       int                  // Record the number of pages of the current Page
	linew        float64              // line width

	// page info
	pageWidth, pageHeight       float64
	contentWidth, contentHeight float64
	pageStartX, pageStartY      float64
	pageEndX, pageEndY          float64

	// Callback function, executed after PDF generation
	callbacks []CallBack
}

func CreateReport() *Report {
	report := new(Report)
	report.converter = new(Converter)

	report.Vars = make(map[string]string)
	report.executors = make(map[string]*Executor)
	report.callbacks = make([]CallBack, 0)
	report.flags = make(map[string]bool)

	report.flags[Flag_AutoAddNewPage] = false
	report.flags[Flag_ResetPageNo] = false

	return report
}

func (report *Report) NoCompression() {
	report.converter.NoCompression()
}

/****************************************************************
Compression level:
	-2: Only Huffman compression is used,
	-1: Default value, compression level 6
	0:  Not compress,
	1:  The fastest compression, but the compression ratio is not the best
	9:  Maximum compression, but the execution efficiency is also the slowest
****************************************************************/
func (report *Report) CompressLevel(level int) {
	report.converter.CompressLevel(level)
}

// Execute, execute Executors and generate PDF file
func (report *Report) Execute(filepath string) {
	if report.config == nil {
		panic("please set page config")
	}

	report.execute(true)
	report.converter.WritePdf(filepath)

	for i := range report.callbacks {
		report.callbacks[i](report)
	}
}

// GetBytesPdf, get PDF file content
func (report *Report) GetBytesPdf() (ret []byte) {
	if report.config == nil {
		panic("please set page config")
	}

	report.execute(true)
	ret = report.converter.GetBytesPdf()
	return
}

// LoadCellsFromText, generate PDF file from cells file
func (report *Report) LoadCellsFromText(filepath string) error {
	return report.converter.ReadFile(filepath)
}

func (report *Report) execute(exec bool) {
	if exec {
		// execute the first page header
		report.executePageHeader()

		report.pageNo = 1
		report.currX, report.currY = report.GetPageStartXY()
		report.addAtomicCell("v|PAGE|" + strconv.Itoa(report.pageNo))
		report.executeDetail()
		// execute the last page footer
		report.executePageFooter()

		// pagination
		report.pagination()
	}

	report.converter.Execute()
}
func (report *Report) executePageFooter() {
	if !report.FisrtPageNeedFooter {
		report.FisrtPageNeedFooter = true
		return
	}

	curX, curY := report.GetXY()
	report.currY = report.config.endY
	report.currX = report.config.startX

	h := report.executors[Footer]
	if h != nil {
		(*h)(report)
	}
	report.SetXY(curX, curY)
}
func (report *Report) executePageHeader() {
	if !report.FisrtPageNeedHeader {
		report.FisrtPageNeedHeader = true
		return
	}

	curX, curY := report.GetXY()
	report.currY = 0
	report.currX = report.config.startX
	h := report.executors[Header]
	if h != nil {
		(*h)(report)
	}
	report.SetXY(curX, curY)
}
func (report *Report) executeDetail() {
	h := report.executors[Detail]
	if h != nil {
		if report.flags[Flag_AutoAddNewPage] {
			report.AddNewPage(report.flags[Flag_ResetPageNo])
			report.flags[Flag_AutoAddNewPage] = false
			report.flags[Flag_ResetPageNo] = false
		}

		(*h)(report)
	}
}

// 分页, 只有一个页面的PDF没有此操作
func (report *Report) pagination() {
	lines := report.converter.GetAutomicCells()
	list := new(List)

	// 第一次遍历单元格, 确定需要创建的PDF页
	for i, line := range lines {
		if len(line) < 8 {
			continue
		}
		if line[0:7] == "v|PAGE|" {
			h := new(pageMark)
			h.lineNo = i
			h.pageNo = parseIntPanic(line[7:], line)
			list.Add(h)
			//fmt.Printf("hist %v \n", h)
		}
	}

	// 第二次遍历单元格, 检查 TotalPage
	for i, line := range lines {
		if strings.Index(line, "{#TotalPage#}") > -1 {
			total := report.getpageNoBylineNo(i, list)
			//fmt.Printf("total :%v\n", total)
			lines[i] = strings.Replace(lines[i], "{#TotalPage#}", strconv.Itoa(total), -1)
		}
	}

	cells := make([]string, 0)
	for _, line := range lines {
		cells = append(cells, line)
	}

	report.converter.SetAutomicCells(cells)
}

// 获取 lineNo 对应的 pageNo
func (report *Report) getpageNoBylineNo(lineNo int, list *List) int {
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
func (report *Report) SetFonts(fmap []*FontMap) {
	report.converter.fonts = fmap
}

// 获取当前页面编号
func (report *Report) GetCurrentPageNo() int {
	return report.pageNo
}

// 添加新的页面
func (report *Report) AddNewPage(resetpageNo bool) {
	report.executePageFooter()

	report.addAtomicCell("NP") // 构建新的页面
	if resetpageNo {
		report.pageNo = 1
	} else {
		report.pageNo++
	}

	report.addAtomicCell("v|PAGE|" + strconv.Itoa(report.pageNo))
	report.SetXY(report.GetPageStartXY())

	report.executePageHeader()
}

func (report *Report) RegisterExecutor(execuror Executor, name string) {
	report.executors[name] = &execuror
}

func (report *Report) GetConfig() Config {
	config := *report.config
	return config
}

func (report *Report) GetPageEndXY() (x, y float64) {
	return report.pageEndX, report.pageEndY
}
func (report *Report) GetPageStartXY() (x, y float64) {
	return report.pageStartX, report.pageStartY
}
func (report *Report) GetContentWidthAndHeight() (width, height float64) {
	return report.contentWidth, report.contentHeight
}

// currX, currY, 坐标
func (report *Report) SetXY(currX, currY float64) {
	if currX > 0 {
		report.currX = currX
	}

	if currY > 0 {
		report.currY = currY
	}
}
func (report *Report) GetXY() (x, y float64) {
	return report.currX, report.currY
}

func (report *Report) SetMargin(dx, dy float64) {
	x, y := report.GetXY()
	report.SetXY(x+dx, y+dy)
}

// 设置页面的尺寸, unit: mm pt in  size: A4 LTR, 目前支持常用的两种方式
func (report *Report) SetPage(size string, orientation string) {
	unit := "pt"
	config, ok := defaultConfigs[size]
	if !ok {
		panic("the config not exists, please add config")
	}

	switch size {
	case "A4":
		switch orientation {
		case "P":
			report.addAtomicCell("P|" + unit + "|A4|P")
			report.pageWidth = config.width
			report.pageHeight = config.height
		case "L":
			report.addAtomicCell("P|" + unit + "|A4|L")
			report.pageWidth = config.height
			report.pageHeight = config.width
		}
	case "LTR":
		switch orientation {
		case "P":
			report.pageWidth = config.width
			report.pageHeight = config.height
			report.addAtomicCell("P|" + unit + "|" + strconv.FormatFloat(report.pageWidth, 'f', 4, 64) +
				"|" + strconv.FormatFloat(report.pageHeight, 'f', 4, 64))
		case "L":
			report.pageWidth = config.height
			report.pageHeight = config.width
			report.addAtomicCell("P  |" + unit + "|" + strconv.FormatFloat(report.pageWidth, 'f', 4, 64) +
				"|" + strconv.FormatFloat(report.pageHeight, 'f', 4, 64))
		}
	}

	report.contentWidth = config.contentWidth
	report.contentHeight = config.contentHeight

	report.pageStartX = config.startX
	report.pageStartY = config.startY
	report.pageEndX = config.endX
	report.pageEndY = config.endY
	report.config = config

	report.execute(false)
}

// 获取底层的所有的原子单元内容
func (report *Report) GetAtomicCells() *[]string {
	cells := report.converter.GetAutomicCells()
	return &cells
}

// 保存原子操作单元
func (report *Report) SaveAtomicCellText(filepath string) {
	cells := report.converter.GetAutomicCells()
	text := strings.Join(cells, "\n")
	ioutil.WriteFile(filepath, []byte(text), os.ModePerm)
}

// 计算文本宽度, 必须先调用 SetFontWithStyle() 或者 SetFont()
func (report *Report) MeasureTextWidth(text string) float64 {
	return report.converter.MeasureTextWidth(text)
}

// 设置当前文本字体, 先注册,后设置
func (report *Report) SetFontWithStyle(family, style string, size int) {
	report.converter.SetFont(family, style, size)
	report.addAtomicCell("F|" + family + "|" + style + "|" + strconv.Itoa(size))
}
func (report *Report) SetFont(family string, size int) {
	report.converter.SetFont(family, "", size)
	report.addAtomicCell("F|" + family + "|" + "" + "|" + strconv.Itoa(size))
}

func (report *Report) AddCallBack(callback CallBack) {
	report.callbacks = append(report.callbacks, callback)
}

/********************************************
 将特定的字符串转换成底层可以识别的原子操作符
*********************************************/
func (report *Report) addAtomicCell(s string) {
	report.converter.AddAtomicCell(s)
}

// 注册当前字体
func (report *Report) Font(fontName string, size int, style string) {
	report.addAtomicCell("F|" + fontName + "|" + style + "|" + strconv.Itoa(size))
}

// 写入字符串内容
func (report *Report) Cell(x float64, y float64, content string) {
	report.addAtomicCell("CL|" + util.Ftoa(x) + "|" + util.Ftoa(y) + "|" + content)
	report.SetXY(report.converter.GetXY())
}
func (report *Report) CellRight(x float64, y float64, w float64, content string) {
	report.addAtomicCell("CR|" + util.Ftoa(x) + "|" + util.Ftoa(y) + "|" +
		util.Ftoa(w) + "|" + content)
	report.SetXY(report.converter.GetXY())
}
func (report *Report) CellGray(x float64, y float64, content string, grayScale float64) {
	report.grayFill(grayScale)
	report.addAtomicCell("CL|" + util.Ftoa(x) + "|" + util.Ftoa(y) + "|" + content)
	report.grayFill(0)
	report.SetXY(report.converter.GetXY())
}

// 划线
func (report *Report) LineType(ltype string, width float64) {
	report.linew = width
	report.addAtomicCell("LT|" + ltype + "|" + util.Ftoa(width))
}
func (report *Report) Line(x1 float64, y1 float64, x2 float64, y2 float64) {
	report.addAtomicCell("L|" + util.Ftoa(x1) + "|" + util.Ftoa(y1) + "|" + util.Ftoa(x2) +
		"|" + util.Ftoa(y2))
}
func (report *Report) LineH(x1 float64, y float64, x2 float64) {
	adj := report.linew * 0.5
	report.addAtomicCell("LH|" + util.Ftoa(x1) + "|" + util.Ftoa(y+adj) + "|" + util.Ftoa(x2))
}
func (report *Report) LineV(x float64, y1 float64, y2 float64) {
	adj := report.linew * 0.5
	report.addAtomicCell("LV|" + util.Ftoa(x+adj) + "|" + util.Ftoa(y1) + "|" + util.Ftoa(y2))
}

// 画特定的图形, 目前支持: 长方形, 椭圆两大类
func (report *Report) Rect(x1 float64, y1 float64, x2 float64, y2 float64) {
	report.addAtomicCell("R|" + util.Ftoa(x1) + "|" + util.Ftoa(y1) + "|" + util.Ftoa(x2) +
		"|" + util.Ftoa(y2))
}
func (report *Report) Oval(x1 float64, y1 float64, x2 float64, y2 float64) {
	report.addAtomicCell("O|" + util.Ftoa(x1) + "|" + util.Ftoa(y1) + "|" + util.Ftoa(x2) +
		"|" + util.Ftoa(y2))
}

// 设置当前的字体颜色, 线条颜色
func (report *Report) TextDefaultColor() {
	report.addAtomicCell("TC|" + strconv.Itoa(1) + "|" + strconv.Itoa(1) +
		"|" + strconv.Itoa(1))
}

func (report *Report) LineDefaultColor() {
	report.addAtomicCell("LC|" + strconv.Itoa(1) + "|" + strconv.Itoa(1) +
		"|" + strconv.Itoa(1))
}

func (report *Report) TextColor(red int, green int, blue int) {
	report.addAtomicCell("TC|" + strconv.Itoa(red) + "|" + strconv.Itoa(green) +
		"|" + strconv.Itoa(blue))
}
func (report *Report) LineColor(red int, green int, blue int) {
	report.addAtomicCell("LC|" + strconv.Itoa(red) + "|" + strconv.Itoa(green) +
		"|" + strconv.Itoa(blue))
}

// color: 背景颜色
// line: 是否需要边框线条, "0000"不需要,  "1111"需要, "0110" 是需要 TOP,RIGHT 线条
func (report *Report) BackgroundColor(x, y, w, h float64, color string, line string) {
	if !rline.MatchString(line) {
		line = "0000"
	}
	if len(line) == 1 {
		line = strings.Repeat(line, 4)
	}
	if len(line) == 2 {
		line = strings.Repeat(line, 2)
	}
	if len(line) == 3 {
		line += "0"
	}

	red, green, blue := util.GetColorRGB(color)

	report.addAtomicCell("BC|" + util.Ftoa(x) + "|" + util.Ftoa(y) + "|" + util.Ftoa(w) + "|" +
		util.Ftoa(h) + "|" + strconv.Itoa(red) + "|" + strconv.Itoa(green) + "|" + strconv.Itoa(blue) + "|" + line)
}

// 线条灰度
func (report *Report) LineGrayColor(x, y float64, w, h float64, gray float64) {
	if gray < 0 || gray > 1 {
		gray = 0.85
	}
	report.LineType("straight", h)
	report.grayStroke(gray)
	report.LineH(x, y, x+w)
	report.LineType("straight", 0.01)
	report.grayStroke(0)
}

// 只用于划线
func (report *Report) grayStroke(grayScale float64) {
	if grayScale > 1.0 || grayScale < 0.0 {
		grayScale = 0
	}

	report.addAtomicCell("GS|" + util.Ftoa(grayScale))
}

// 只用于文本
func (report *Report) grayFill(grayScale float64) {
	if grayScale > 1.0 || grayScale < 0.0 {
		grayScale = 0
	}

	report.addAtomicCell("GF|" + util.Ftoa(grayScale))
}

// 图片
func (report *Report) Image(path string, x1 float64, y1 float64, x2 float64, y2 float64) {
	report.addAtomicCell("I|" + path + "|" + util.Ftoa(x1) + "|" + util.Ftoa(y1) + "|" +
		util.Ftoa(x2) + "|" + util.Ftoa(y2))
}

// 添加变量
func (report *Report) Var(name string, val string) {
	report.addAtomicCell("V|" + name + "|" + val)
}

// 外部链接
func (report *Report) ExternalLink(x, y, th float64, content, link string) {
	tw := report.MeasureTextWidth(content)
	if x+tw > report.config.endX {
		tw = report.config.endX - x
	}
	report.addAtomicCell("EL|" + util.Ftoa(x) + "|" + util.Ftoa(y) + "|" + util.Ftoa(tw) + "|" + util.Ftoa(th) + "|" +
		content + "|" + link)

	report.SetXY(x+tw, y)
}

func (report *Report) InternalLinkAnchor(x, y, th float64, content, anchor string) {
	tw := report.MeasureTextWidth(content)
	if x+tw > report.config.endX {
		tw = report.config.endX - x
	}

	report.addAtomicCell("ILA|" + util.Ftoa(x) + "|" + util.Ftoa(y) + "|" + util.Ftoa(tw) + "|" + util.Ftoa(th) + "|" +
		content + "|" + anchor)

	report.SetXY(x+tw, y)
}

func (report *Report) InternalLinkLink(x, y float64, content, anchor string) {
	tw := report.MeasureTextWidth(content)
	if x+tw > report.config.endX {
		tw = report.config.endX - x
	}

	report.addAtomicCell("ILL|" + util.Ftoa(x) + "|" + util.Ftoa(y) + "|" + util.Ftoa(tw) + "|" + content + "|" + anchor)

	report.SetXY(x+tw, y)
}
