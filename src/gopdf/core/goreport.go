package core

import (
	//"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// 需要解决的问题: CurrY的控制权, 用户 -> 程序 -> 自动化操作
// 页面的三部分: Header Page Footer
const (
	Header = "Header"
	Footer = "Footer"
	Detail = "Detail"

	Summary      = "Summary"
	GroupHeader  = "GroupHeader"
	GroupSummary = "GroupSummary"

	// flags
	Flag_AutoAddNewPage = "AutoAddNewPage"
	Flag_ResetPageNo    = "ResetPageNo"
)

type pageMark struct {
	cellNo int
	pageNo int
}

type Report struct {
	Records    []interface{}
	DataPos    int
	Bands      map[string]*Band
	CurrX      float64
	CurrY      float64
	MaxGroup   int
	IsMutiPage bool
	Flags      map[string]bool
	SumWork    map[string]float64
	Vars       map[string]string

	convPt     float64 // 转换单位
	pageNo     int     // 记录当前的 Page 的页数
	converter  *Converter
	pageWidth  float64
	pageHeight float64
}

// 设置页脚的Y值
func (r *Report) SetFooterYbyFooterHeight(footerHeight float64) {
	if r.pageHeight == 0 {
		panic("Page size not yet specified")
	}
	r.SumWork["__ft__"] = r.pageHeight - footerHeight
}

func (r *Report) SetFooterY(footerY float64) {
	r.SumWork["__ft__"] = footerY
}

// 设置可用字体
func (r *Report) SetFonts(fmap []*FontMap) {
	r.converter.Fonts = fmap
}

// 构建新的PAGE
func (r *Report) AutoAddNewPage(resetpageNo bool) {
	r.Flags[Flag_AutoAddNewPage] = true
	r.Flags[Flag_ResetPageNo] = resetpageNo
}

// 转换, 内容 -> PDF文件
func (r *Report) Convert(exec bool) {
	if exec {
		if r.SumWork["__ft__"] == 0 {
			panic("footerY not set yet.")
		}
		r.pageNo = 1
		r.CurrY = 0
		r.ExecutePageHeader() // 页眉
		r.addAtomicCell("v|PAGE|" + strconv.Itoa(r.pageNo))
		//for r.DataPos = 0; r.DataPos < len(r.Records); r.DataPos++ {
		//	r.ExecuteDetail()
		//}
		r.ExecuteDetail()
		r.ExecuteSummary()    // ???
		r.ExecutePageFooter() // 页脚

		r.Pagination() // 分页
	}
	r.converter.Execute()
}

// 写入PDF文件
func (r *Report) Execute(filename string) {
	r.Convert(true)
	r.converter.GoPdf.WritePdf(filename)
}

// 获取PDF内容
func (r *Report) GetBytesPdf() (ret []byte) {
	r.Convert(true)
	ret = r.converter.GoPdf.GetBytesPdf()
	return
}

// 分页, 只有一个页面的PDF没有此操作
func (r *Report) Pagination() {
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
			h.cellNo = i
			h.pageNo = AtoiPanic(line[7:], line)
			list.Add(h)
			//fmt.Printf("hist %v \n", h)
		}
	}

	// 第二次遍历单元格, 检查 TotalPage
	for i, line := range lines {
		if strings.Index(line, "{#TotalPage#}") > -1 {
			total := r.getpageNoByCellNo(i, list)
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

// 获取 cellNo 对应的 pageNo
func (r *Report) getpageNoByCellNo(cellNo int, list *List) int {
	count := 0
	page := 0

	// 遍历到当前的cellNo, 当前的count记录的是list的索引
	for i, l := range list.GetAsArray() {
		if l.(*pageMark).cellNo >= cellNo {
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
	r.CurrY = 0
	r.ExecutePageHeader()
	r.addAtomicCell("v|PAGE|" + strconv.Itoa(r.pageNo))
}

func (r *Report) AddNewPageCheck(height float64) {
	if r.CurrY+height > r.SumWork["__ft__"] {
		r.AddNewPage(false)
	}
}

func (r *Report) ExecutePageFooter() {
	r.CurrY = r.SumWork["__ft__"]

	h := r.Bands[Footer]
	if h != nil {
		height := (*h).GetHeight(r)
		(*h).Execute(r)
		r.CurrY += height
	}
}

func (r *Report) ExecuteSummary() {
	h := r.Bands[Summary]
	if h != nil {
		height := (*h).GetHeight(r)
		r.AddNewPageCheck(height)
		(*h).Execute(r)
		r.CurrY += height
	}
}

func (r *Report) ExecutePageHeader() {
	h := r.Bands[Header]
	if h != nil {
		height := (*h).GetHeight(r)
		(*h).Execute(r)
		r.CurrY += height
	}
}

func (r *Report) ExecuteGroupHeader(level int) {
	for l := level; l > 0; l-- {
		h := r.Bands[GroupHeader+strconv.Itoa(l)]
		if h != nil {
			height := (*h).GetHeight(r)
			r.AddNewPageCheck(height)
			(*h).Execute(r)
			r.CurrY += height
		}
	}
}

func (r *Report) ExecuteGroupSummary(level int) {
	for l := 1; l <= level; l++ {
		h := r.Bands[GroupSummary+strconv.Itoa(l)]
		if h != nil {
			height := (*h).GetHeight(r)
			r.AddNewPageCheck(height)
			(*h).Execute(r)
			r.CurrY += height
		}
	}
}

func (r *Report) ExecuteDetail() {
	h := r.Bands[Detail]
	if h != nil {
		//fmt.Printf("report.NewPage flag %v\n", r.Flags["NewPageForce"])
		if r.Flags[Flag_AutoAddNewPage] {
			r.AddNewPage(r.Flags[Flag_ResetPageNo])
			r.Flags[Flag_AutoAddNewPage] = false
			r.Flags[Flag_ResetPageNo] = false
		}
		var deti interface{} = *h

		if r.MaxGroup > 0 {
			bfr := reflect.ValueOf(deti).MethodByName("BeforeAddNewPage")
			if bfr.IsValid() == false {
				panic("BeforeAddNewPage function not exist in Detail")
			}
			res := bfr.Call([]reflect.Value{reflect.ValueOf(*r)})
			level := res[0].Int()
			if level > 0 {
				r.ExecuteGroupHeader(int(level))
			}
		}

		height := (*h).GetHeight(r)
		r.AddNewPageCheck(height)
		(*h).Execute(r)
		r.CurrY += height

		if r.MaxGroup > 0 {
			aft := reflect.ValueOf(deti).MethodByName("AfterAddNewPage")
			if aft.IsValid() == false {
				panic("AfterAddNewPage function not exist in Detail")
			}
			res := aft.Call([]reflect.Value{reflect.ValueOf(*r)})
			level := res[0].Int()
			if level > 0 {
				r.ExecuteGroupSummary(int(level))
			}
		}
	}
}

func (r *Report) RegisterBand(band Band, name string) {
	r.Bands[name] = &band
}

func (r *Report) RegisterGroupBand(band Band, name string, level int) {
	r.Bands[name+strconv.Itoa(level)] = &band
	if r.MaxGroup < level {
		r.MaxGroup = level
	}
}

func (r *Report) addAtomicCell(s string) {
	r.converter.AddAtomicCell(s)
}

func (r *Report) Margin(top, left float64) {
	r.addAtomicCell("Marign|" + Ftoa(top) + "|" + Ftoa(left))
}

func (r *Report) Font(fontName string, size int, style string) {
	r.addAtomicCell("F|" + fontName + "|" + style + "|" + strconv.Itoa(size))
}
func (r *Report) Cell(x float64, y float64, content string) {
	r.addAtomicCell("C1|" + Ftoa(x) + "|" + Ftoa(y) + "|" + content)
}
func (r *Report) CellRight(x float64, y float64, w float64, content string) {
	r.addAtomicCell("CR|" + Ftoa(x) + "|" + Ftoa(y) + "|" +
		Ftoa(w) + "|" + content)
}

func (r *Report) LineType(ltype string, width float64) {
	r.SumWork["__lw__"] = width
	r.addAtomicCell("LT|" + ltype + "|" + Ftoa(width))
}
func (r *Report) Line(x1 float64, y1 float64, x2 float64, y2 float64) {
	r.addAtomicCell("L|" + Ftoa(x1) + "|" + Ftoa(r.CurrY+y1) + "|" + Ftoa(x2) +
		"|" + Ftoa(r.CurrY+y2))
}
func (r *Report) LineH(x1 float64, y float64, x2 float64) {
	adj := r.SumWork["__lw__"] * 0.5
	r.addAtomicCell("LH|" + Ftoa(x1) + "|" + Ftoa(r.CurrY+y+adj) + "|" + Ftoa(x2))
}
func (r *Report) LineV(x float64, y1 float64, y2 float64) {
	adj := r.SumWork["__lw__"] * 0.5
	r.addAtomicCell("LV|" + Ftoa(x+adj) + "|" + Ftoa(r.CurrY+y1) + "|" + Ftoa(r.CurrY+y2))
}

//SumWork["__lw__"] width adjust
func (r *Report) Rect(x1 float64, y1 float64, x2 float64, y2 float64) {
	r.addAtomicCell("R|" + Ftoa(x1) + "|" + Ftoa(r.CurrY+y1) + "|" + Ftoa(x2) +
		"|" + Ftoa(r.CurrY+y2))
}
func (r *Report) Oval(x1 float64, y1 float64, x2 float64, y2 float64) {
	r.addAtomicCell("O|" + Ftoa(x1) + "|" + Ftoa(r.CurrY+y1) + "|" + Ftoa(x2) +
		"|" + Ftoa(r.CurrY+y2))
}
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
func (r *Report) Image(path string, x1 float64, y1 float64, x2 float64, y2 float64) {
	r.addAtomicCell("I|" + path + "|" + Ftoa(x1) + "|" + Ftoa(r.CurrY+y1) + "|" +
		Ftoa(x2) + "|" + Ftoa(r.CurrY+y2))
}

func (r *Report) Var(name string, val string) {
	r.addAtomicCell("V|" + name + "|" + val)
}

func Ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// unit: mm pt in  size: A4 LTR, 目前支持常用的两种方式
func (r *Report) SetPage(size string, unit string, orientation string) {
	r.SetConv(unit)
	switch size {
	case "A4":
		switch orientation {
		case "P":
			r.addAtomicCell("P|" + unit + "|A4|P")
			r.pageWidth = 595.28 / r.convPt
			r.pageHeight = 841.89 / r.convPt
		case "L":
			r.addAtomicCell("P|" + unit + "|A4|L")
			r.pageWidth = 841.89 / r.convPt
			r.pageHeight = 595.28 / r.convPt
		}
	case "LTR":
		switch orientation {
		case "P":
			r.pageWidth = 612 / r.convPt
			r.pageHeight = 792 / r.convPt
			r.addAtomicCell("P1|" + unit + "|" + strconv.FormatFloat(r.pageWidth, 'f', 4, 64) +
				"|" + strconv.FormatFloat(r.pageHeight, 'f', 4, 64))
		case "L":
			r.pageWidth = 792 / r.convPt
			r.pageHeight = 612 / r.convPt
			r.addAtomicCell("P1|" + unit + "|" + strconv.FormatFloat(r.pageWidth, 'f', 4, 64) +
				"|" + strconv.FormatFloat(r.pageHeight, 'f', 4, 64))
		}
	}
	r.Convert(false)
}

//unit mm pt in
func (r *Report) SetConv(ut string) {
	switch ut {
	case "mm":
		r.convPt = 2.834645669
	case "pt":
		r.convPt = 1
	case "in":
		r.convPt = 72
	default:
		panic("This unit is not specified :" + ut)
	}
}

func (r *Report) GetConvPt() float64 {
	if r.convPt == 0.0 {
		panic("does not set convpt")
	}
	return r.convPt
}

// 设置尺寸
func (r *Report) SetPageByDimension(unit string, width float64, height float64) {
	r.SetConv(unit)
	r.pageWidth = width
	r.pageHeight = height
	r.addAtomicCell("P1|" + unit + "|" + strconv.FormatFloat(width, 'f', 4, 64) +
		"|" + strconv.FormatFloat(height, 'f', 4, 64))
	r.Convert(false)
}

func (r *Report) GetAtomicCells() *[]string {
	return &r.converter.AtomicCells
}

// 保存单元格
func (r *Report) SaveText(fileName string) {
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

// 设置文本字体
func (r *Report) SetFontWithStyle(family, style string, size int) {
	r.converter.GoPdf.SetFont(family, style, size)
}
func (r *Report) SetFont(family string, size int) {
	r.SetFontWithStyle(family, "", size)
}

// 基本操作
func (r *Report) WriteParagraph(content string, ) {

}

type Band interface {
	GetHeight(report *Report) float64
	Execute(report *Report)
}

func CreateReport() *Report {
	Report := new(Report)
	Report.Bands = make(map[string]*Band)
	Report.converter = new(Converter)
	Report.SumWork = make(map[string]float64)
	Report.Vars = make(map[string]string)
	Report.SumWork["__ft__"] = 0.0 // FooterY
	Report.Flags = make(map[string]bool)
	Report.Flags[Flag_AutoAddNewPage] = false
	Report.Flags[Flag_ResetPageNo] = false
	return Report
}

type TemplateDetail struct {
}

func (h TemplateDetail) GetHeight() float64 {
	return 10
}
func (h TemplateDetail) Execute(report *Report) {
}

func (h TemplateDetail) BeforeAddNewPage(report *Report) int {
	return 0
}

func (h TemplateDetail) AfterAddNewPage(report *Report) int {
	return 0
}
