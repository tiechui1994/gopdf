package gopdf

// markdown_types.go：Markdown 渲染 AST 公共约定——组件接口、ElementBase（流式几何 + 可选 Margin/Padding）、
// 主题、LayoutContext，以及与 lexer 协作的正则/颜色。

import (
	"regexp"
	"strings"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/lex"
)

// Token 即 lex.Token，表示 lexer 输出的 Markdown 结点。
type Token = lex.Token

// markdownNode 为渲染侧组件接口（扁平 slice + GenerateAtomicCell 状态机）；实现包含 MdText、MdParagraph 等。
//
// GenerateAtomicCell 返回值约定：
//   - pagebreak：画完本趟前需要先换页再继续；
//   - over：本结点这一轮是否处于「可丢弃/已完结」语义（与 CommonGenerateAtomicCell 切片策略配合）。
type markdownNode interface {
	SetText(font interface{}, text ...string)
	GetType() string
	GenerateAtomicCell() (pagebreak, over bool, err error)
}

// MdBoxModel 按结点类型的外边距与内边距（pt）；块级由 GenerateAtomicCell 首尾统一施加，
// 行内主要在 MdText 上参与行宽与首行前位移。
type MdBoxModel struct {
	Margin  MdInsets
	Padding MdInsets
}

// MarkdownTheme 文档级默认字号、行距与段间空隙（pt），以及各 Markdown 结点类型的盒模型。
type MarkdownTheme struct {
	BaseFontSize float64 // pt，正文字号
	LineHeight   float64 // pt，单行步进
	BreakGap     float64 // pt，类似段间距的垂直空隙

	BoxParagraph  MdBoxModel
	BoxHeading    MdBoxModel
	BoxList       MdBoxModel
	BoxBlockQuote MdBoxModel
	BoxCodeBlock  MdBoxModel
	BoxTable      MdBoxModel
	BoxHR         MdBoxModel
	BoxInlineText MdBoxModel
	BoxLink       MdBoxModel
	BoxCodespan   MdBoxModel
	BoxStrong     MdBoxModel
	BoxEm         MdBoxModel
	BoxDel        MdBoxModel
}

// DefaultMarkdownTheme 与历史 markdown.go 中的 mdBase / mdLineHeight / mdBreakGap 一致；
// 各 Box* 默认为零，行为与未引入盒模型前一致。
func DefaultMarkdownTheme() MarkdownTheme {
	return MarkdownTheme{BaseFontSize: mdBase, LineHeight: mdLineHeight, BreakGap: mdBreakGap}
}

// BoxForInlineToken 返回行内 token 类型对应的主题盒（未知类型按正文 text）。
func (t MarkdownTheme) BoxForInlineToken(typ string) MdBoxModel {
	switch typ {
	case TYPE_LINK:
		return t.BoxLink
	case TYPE_CODESPAN:
		return t.BoxCodespan
	case TYPE_STRONG:
		return t.BoxStrong
	case TYPE_EM:
		return t.BoxEm
	case TYPE_DEL:
		return t.BoxDel
	case TYPE_TEXT, TYPE_IMAGE:
		return t.BoxInlineText
	default:
		return t.BoxInlineText
	}
}

// bodyLineHeight 返回正文行距（pt）；主题未设时退回 mdLineHeight。
func (t MarkdownTheme) bodyLineHeight() float64 {
	if t.LineHeight > 0 {
		return t.LineHeight
	}
	return mdLineHeight
}

// bodyFontSize 返回正文字号（pt）；主题未设时退回 mdBase。
func (t MarkdownTheme) bodyFontSize() float64 {
	if t.BaseFontSize > 0 {
		return t.BaseFontSize
	}
	return mdBase
}

// paraBreakGap 段间类垂直空隙（MdSpace 等与 breakHeight 叠加时使用）。
func (t MarkdownTheme) paraBreakGap() float64 {
	if t.BreakGap > 0 {
		return t.BreakGap
	}
	return mdBreakGap
}

// MdInsets 为 CSS 语义的四边间距（pt）；默认 0。应用在 MdText 时位于 flowColumnOffsetPt / 列表 hang 之后。
type MdInsets struct {
	Top, Right, Bottom, Left float64
}

// mdPoint 记录与 core.Report.GetXY 同一坐标系下的包围参考点。
type mdPoint struct {
	X, Y float64
}

// ElementBase 各类 Markdown 渲染结点的几何与主题基类：输出目标、流式栏位、引用/列表缩进、
// 盒模型与一次绘制趟的包围盒记录；具体绘制逻辑由 MdText、MdParagraph 等实现。
type ElementBase struct {
	pdf *core.Report // 目标版面：页坐标、边距、写入原子单元等，与 core.Report 同一套 pt 坐标系

	lineHeight float64 // MdSpace/MdHardBreak 等用于垂直留白或硬性断行的目标高度（pt）；MdText 正文行距由 theme.bodyLineHeight() 统一给出，标题见 MdText.headingStepPt
	blockquote int     // 引用嵌套深度（层数），>0 时在 MdSpace/MdHardBreak/MdText 等路径绘制左侧引用竖条

	Type string // 结点类型，与 lexer Token.Type 或常量 TYPE_* 对齐，用于主题盒模型与分支（如 "text"、"paragraph"）

	hangingIndentPt float64 // 列表挂起正文列相对页内容区左缘的水平偏移（pt），即 marker 后对齐正文栏的 X 基准

	quoteBarsLeftOffsetPt float64 // 引用左侧竖条纹样左缘相对页起点的偏移（pt）；>0 时竖条固定在 pageStartX+本值，用于列表嵌套引用等与正文列对齐

	flowColumnOffsetPt float64 // 流式正文列左缘相对页内容区起点的水平偏移（pt）：引用嵌套、根级围栏代码、嵌套列表正文列等累加；与 MdInsets 的 margin/padding 不同，参见 Padding

	Margin  MdInsets // 外边距（pt），常由 MarkdownTheme 按结点类型合并进 ElementBase，块级结点在顶部/底部排版时统一施加
	Padding MdInsets // 内边距（pt）；在 MdText 中参与行可用宽度、首行前顶距、分页时下沿预留，与 flowColumnOffsetPt 分工不同

	theme MarkdownTheme // 文档级主题：正文字号/行距/段间距及各类结点的默认盒模型

	// StartPt/EndPt：本次 GenerateAtomicCell 调用中，本结点（或叶子）绘制范围的近似轴对齐矩形（左上角 StartPt，右下角 EndPt 含义为最大触及的 X/Y）
	StartPt, EndPt      mdPoint
	extentStartRecorded bool // 本轮 GenerateAtomicCell 是否已写入包围盒起点 StartPt（noteLayoutStart / noteLayoutExtent 使用）
}

// TextHorizontalInsets 返回左右 Margin+Padding 之和，用于计算可用行宽。
func (e *ElementBase) TextHorizontalInsets() (left, right float64) {
	return e.Margin.Left + e.Padding.Left, e.Margin.Right + e.Padding.Right
}

// TextVerticalTopInset 首行前施加的 Margin.Top + Padding.Top（每个 MdText 芯片首次排版时一次）。
func (e *ElementBase) TextVerticalTopInset() float64 {
	return e.Margin.Top + e.Padding.Top
}

// TextVerticalBottomInset 从页底向上预留 Margin.Bottom + Padding.Bottom，用于分页判断。
func (e *ElementBase) TextVerticalBottomInset() float64 {
	return e.Margin.Bottom + e.Padding.Bottom
}

// resetLayoutExtent 在新的 GenerateAtomicCell 调用开始时清空布局包围盒记录。
func (e *ElementBase) resetLayoutExtent() {
	e.extentStartRecorded = false
	e.StartPt = mdPoint{}
	e.EndPt = mdPoint{}
}

// noteLayoutStart 首次绘制前记录 StartPt，并与 EndPt 对齐。
func (e *ElementBase) noteLayoutStart(x, y float64) {
	if !e.extentStartRecorded {
		e.StartPt = mdPoint{X: x, Y: y}
		e.EndPt = e.StartPt
		e.extentStartRecorded = true
	}
}

// noteLayoutExtent 扩展 EndPt 至更大 X/Y（假定纵向向下书写时 Y 增大）。
func (e *ElementBase) noteLayoutExtent(x, y float64) {
	if !e.extentStartRecorded {
		e.noteLayoutStart(x, y)
		return
	}
	if x > e.EndPt.X {
		e.EndPt.X = x
	}
	if y > e.EndPt.Y {
		e.EndPt.Y = y
	}
}

func (e *ElementBase) SetText(interface{}, ...string) {}

func (e *ElementBase) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return false, true, nil
}

func (e *ElementBase) GetType() string {
	return e.Type
}

// PDF 绘制用到的 RGB 分量字符串（与 util.RGB 等配合）。
const (
	color_black = "1,1,1"
	color_gray  = "128,128,128"
	color_white = "255,255,255"

	color_pink       = "199,37,78"
	color_lightgray  = "220,220,220"
	color_whitesmoke = "245,245,245"
	color_blue       = "0,0,255"
)

// re 缓存 MdText 宽度估算与代码块尾部换行判断用到的正则。
var re struct {
	breakline *regexp.Regexp
}

func init() {
	re.breakline = regexp.MustCompile(`\n{2,}$`)
}

// monoFamilyFrom 优先取 fonts["mono"]，否则退回正文字体族。
func monoFamilyFrom(fonts map[string]string) string {
	if fonts == nil {
		return ""
	}
	if m := fonts[FONT_MONO]; m != "" {
		return m
	}
	return fonts[FONT_NORMAL]
}

// hasBreakLine 判断 token 源码是否在末尾带有「硬换行」语义（代码块用正则，其余看后缀 \n）。
func hasBreakLine(token Token) bool {
	switch token.Type {
	case TYPE_CODE:
		return re.breakline.MatchString(token.Raw)
	default:
		return strings.HasSuffix(token.Raw, "\n")
	}
}

// repairText 在进入排版前按类型清理/保留字符（多数 inline 类型保留 \n 以便按源换行）。
func repairText(TYPE, text string) string {
	switch TYPE {
	case TYPE_CODE:
		return text
	case TYPE_TEXT, TYPE_STRONG, TYPE_EM, TYPE_CODESPAN, TYPE_LINK, TYPE_DEL:
		return text
	default:
		return text
	}
}

// LayoutContext 对 core.Report 的薄封装：分页判断与换页（顶层 MarkdownText 与 MdText 共用逻辑）。
type LayoutContext struct {
	pdf *core.Report
}

// NewLayoutContext 构造布局上下文。
func NewLayoutContext(r *core.Report) *LayoutContext {
	return &LayoutContext{pdf: r}
}

// NeedTextPageBreak 与 MdText 尾部一致：换新行前须满足「已换行」或「光标贴近右缘」且纵向空间不足。
func (lc *LayoutContext) NeedTextPageBreak(y, pageEndY, lh float64, newline bool, x1, pageEndX, precision float64) bool {
	_ = lc.pdf
	return (y >= pageEndY || pageEndY-y < lh) && (newline || absFloat(x1-pageEndX) < precision)
}

// BreakPage 追加一页并将光标重置到页起点（与 MarkdownText.GenerateAtomicCell 行为一致）。
func (lc *LayoutContext) BreakPage() {
	if lc.pdf == nil {
		return
	}
	nx, ny := lc.pdf.GetPageStartXY()
	lc.pdf.AddNewPage(false)
	lc.pdf.SetXY(nx, ny)
}

// absFloat 计算绝对值（避免额外依赖）。
func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
