package core

import (
	"strconv"
	"strings"
	"io/ioutil"
	"os"
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

type CallBack func(report *Report)

type Report struct {
	IsMutiPage      bool
	FisrtPageNeedFH bool // 首页需要执行页眉和页脚
	Vars            map[string]string

	currX     float64
	currY     float64
	bands     map[string]*Band
	flags     map[string]bool
	sumWork   map[string]float64
	unit      float64 // 转换单位
	pageNo    int     // 记录当前的 Page 的页数
	converter *Converter

	pageWidth, pageHeight  float64
	pageStartX, pageStartY float64
	pageEndX, pageEndY     float64
	callbacks              []CallBack // 在PDF生成之后执行
}

// 写入PDF文件
func (r *Report) Execute(filename string) {
	r.execute(true)
	r.converter.GoPdf.WritePdf(filename)

	for i := range r.callbacks {
		r.callbacks[i](r)
	}
}

// 获取PDF内容
func (r *Report) GetBytesPdf() (ret []byte) {
	r.execute(true)
	ret = r.converter.GoPdf.GetBytesPdf()
	return
}

// 转换, 内容 -> PDF文件
func (r *Report) execute(exec bool) {
	if exec {
		if r.sumWork["__ft__"] == 0 {
			panic("footerY not set yet.")
		}
		r.pageNo = 1
		r.currY = 0

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

// 设置页脚的Y值
func (r *Report) SetFooterYbyFooterHeight(footerHeight float64) {
	if r.pageHeight == 0 {
		panic("Page size not yet specified")
	}
	r.sumWork["__ft__"] = r.pageHeight - footerHeight
}

func (r *Report) SetFooterY(footerY float64) {
	r.sumWork["__ft__"] = footerY
}

// 设置可用字体
func (r *Report) SetFonts(fmap []*FontMap) {
	r.converter.Fonts = fmap
}

// 构建新的PAGE
func (r *Report) SetAutoAddNewPage(resetpageNo bool) {
	r.flags[Flag_AutoAddNewPage] = true
	r.flags[Flag_ResetPageNo] = resetpageNo
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
	r.currY = 0
	r.ExecutePageHeader()
	r.addAtomicCell("v|PAGE|" + strconv.Itoa(r.pageNo))
}

func (r *Report) AddNewPageCheck(height float64) {
	if r.currY+height > r.sumWork["__ft__"] {
		r.AddNewPage(false)
	}
}

func (r *Report) ExecutePageFooter() {
	r.currY = r.sumWork["__ft__"]

	h := r.bands[Footer]
	if h != nil {
		height := (*h).GetHeight(r)
		(*h).Execute(r)
		r.currY += height
	}
}

func (r *Report) ExecutePageHeader() {
	h := r.bands[Header]
	if h != nil {
		height := (*h).GetHeight(r)
		(*h).Execute(r)
		r.currY += height
	}
}

func (r *Report) ExecuteDetail() {
	h := r.bands[Detail]
	if h != nil {
		//fmt.Printf("report.NewPage flag %v\n", r.flags["NewPageForce"])
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

		height := (*h).GetHeight(r)
		r.AddNewPageCheck(height)
		(*h).Execute(r)
		r.currY += height
	}
}

func (r *Report) RegisterBand(band Band, name string) {
	r.bands[name] = &band
}

// 换页坐标
func (r *Report) SetPageEndY(y float64) {
	r.pageEndY = y
}
func (r *Report) GetPageEndY() float64 {
	return r.pageEndY
}

// 页面开始坐标
func (r *Report) SetPageStartXY(x, y float64) {
	r.pageStartX = x
	r.pageStartY = y
}
func (r *Report) GetPageStartXY() (x, y float64) {
	return r.pageStartX, r.pageStartY
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

// 设置页面的尺寸, unit: mm pt in  size: A4 LTR, 目前支持常用的两种方式
func (r *Report) SetPage(size string, unit string, orientation string) {
	r.setUnit(unit)
	switch size {
	case "A4":
		switch orientation {
		case "P":
			r.addAtomicCell("P|" + unit + "|A4|P")
			r.pageWidth = 595.28 / r.unit
			r.pageHeight = 841.89 / r.unit
		case "L":
			r.addAtomicCell("P|" + unit + "|A4|L")
			r.pageWidth = 841.89 / r.unit
			r.pageHeight = 595.28 / r.unit
		}
	case "LTR":
		switch orientation {
		case "P":
			r.pageWidth = 612 / r.unit
			r.pageHeight = 792 / r.unit
			r.addAtomicCell("P1|" + unit + "|" + strconv.FormatFloat(r.pageWidth, 'f', 4, 64) +
				"|" + strconv.FormatFloat(r.pageHeight, 'f', 4, 64))
		case "L":
			r.pageWidth = 792 / r.unit
			r.pageHeight = 612 / r.unit
			r.addAtomicCell("P1|" + unit + "|" + strconv.FormatFloat(r.pageWidth, 'f', 4, 64) +
				"|" + strconv.FormatFloat(r.pageHeight, 'f', 4, 64))
		}
	}

	if r.pageEndY == 0 {
		r.pageEndY = r.pageHeight
	}

	r.execute(false)
}
func (r *Report) SetPageByDimension(unit string, width float64, height float64) {
	r.setUnit(unit)
	r.pageWidth = width
	r.pageHeight = height
	r.addAtomicCell("P1|" + unit + "|" + strconv.FormatFloat(width, 'f', 4, 64) +
		"|" + strconv.FormatFloat(height, 'f', 4, 64))
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
func (r *Report) StrokeColor(red int, green int, blue int) {
	r.addAtomicCell("SC|" + strconv.Itoa(red) + "|" + strconv.Itoa(green) +
		"|" + strconv.Itoa(blue))
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

type Band interface {
	GetHeight(report *Report) float64
	Execute(report *Report)
}

func CreateReport() *Report {
	report := new(Report)
	report.converter = new(Converter)

	report.Vars = make(map[string]string)
	report.bands = make(map[string]*Band)
	report.sumWork = make(map[string]float64)
	report.callbacks = make([]CallBack, 0)
	report.flags = make(map[string]bool)

	report.sumWork["__ft__"] = 0.0 // FooterY
	report.flags[Flag_AutoAddNewPage] = false
	report.flags[Flag_ResetPageNo] = false

	return report
}

type TemplateDetail struct {
}

func (h TemplateDetail) GetHeight() float64 {
	return 10
}
func (h TemplateDetail) Execute(report *Report) {
}
