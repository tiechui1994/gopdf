package gopdf

// markdown_const.go：Markdown 渲染用到的常量（lexer type 名、字体槽位）以及与排版相关的间距工具函数。
// 坐标单位均为 PDF point（pt），与 core.Report 一致。

import (
	"math"
	"strings"

	"github.com/tiechui1994/gopdf/lex"
)

// 字体映射字典（NewMarkdownText 的 fonts 参数）使用的键名。
const (
	FONT_NORMAL = "normal"
	FONT_BOLD   = "bold"
	FONT_IALIC  = "italic"
	FONT_MONO   = "mono"
)

// lex.Token.Type 取值，与 lexer 输出一致（渲染侧据此分支）。
const (
	TYPE_TEXT     = "text"
	TYPE_STRONG   = "strong" // **strong**
	TYPE_EM       = "em"     // *em*
	TYPE_CODESPAN = "codespan"
	TYPE_CODE     = "code"
	TYPE_LINK     = "link"
	TYPE_IMAGE    = "image"
	TYPE_DEL      = "del"

	TYPE_SPACE = "space"

	TYPE_PARAGRAPH  = "paragraph"
	TYPE_HEADING    = "heading"
	TYPE_LIST       = "list"
	TYPE_BLOCKQUOTE = "blockquote"
	TYPE_TABLE      = "table"
	TYPE_HR         = "hr"
	
	TYPE_BR = "br"
)

// Layout uses PDF points (pt) only — never device pixels — so output is stable across DPI/screen.
const (
	mdBase       = 12.0 // body font size (pt); typographic base unit
	mdLineHeight = 18.0 // one line step (pt), = mdBase * 1.5
	mdBreakGap   = mdLineHeight * (8.0 / 18.0)
)

const (
	spaceLen = mdLineHeight * (4.425 / 18.0)
	blockLen = spaceLen * 0.6
)

// listNestIndentWidth 嵌套列表相对父级正文列额外缩进的宽度。
func listNestIndentWidth() float64 { return 2 * spaceLen }

const markdownBlockquoteIndentSteps = 4

// blockquoteIndentWidth 单层引用块相对页边或父内容区的水平缩进（栏条+正文槽）。
func blockquoteIndentWidth() float64 { return float64(markdownBlockquoteIndentSteps) * spaceLen }

// blockquoteBarOffset 绘制第 level 条竖向引用线时的左偏移。
func blockquoteBarOffset(level int) float64 { return float64(level) * blockquoteIndentWidth() }

// atMarkdownLineLeft 判断当前 x 是否处于「正文行首」：页左边距或列表 hang 列对齐位置。
func atMarkdownLineLeft(x1, pageStartX, listHangIndent float64) bool {
	if math.Abs(x1-pageStartX) < 0.5 {
		return true
	}
	if listHangIndent > 0 && math.Abs(x1-(pageStartX+listHangIndent)) < 0.5 {
		return true
	}
	return false
}

// stripListParagraphIndent 去掉列表项内延续行最多 2 个前导空格，避免与 hang 列叠加视觉缩进。
func stripListParagraphIndent(s string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if len(ln) == 0 {
			continue
		}
		j := 0
		for j < len(ln) && j < 2 && ln[j] == ' ' {
			j++
		}
		lines[i] = ln[j:]
	}
	return strings.Join(lines, "\n")
}

// mdScale 按行高 mdLineHeight 的比例返回偏移（pt）。
func mdScale(frac float64) float64 { return mdLineHeight * frac }

// codeBlockPad 代码灰底内侧与等宽字形之间的水平 padding（非 ElementBase.Padding）。
func codeBlockPad() float64 { return mdLineHeight * (4.0 / 18.0) }

func listNestBreakBefore() float64 { return mdLineHeight * 0.30 }

func listCodeBreakBefore() float64 { return mdLineHeight * 0.42 }

// codeBlockAfterGap 代码块绘制结束后追加的垂直空隙，避免多块紧贴。
func codeBlockAfterGap() float64 { return mdLineHeight * 0.5 }

// blockquoteBarVOverlap 竖条略加长，避免 MdText/MdSpace 衔接处露缝。
func blockquoteBarVOverlap() float64 { return mdLineHeight * 0.1 }

// unorderedListBulletPrefix 无序列表项目符号，按嵌套深度交替实心圆/空心圆。
func unorderedListBulletPrefix(nestLevel int) string {
	if nestLevel < 0 {
		nestLevel = 0
	}
	if nestLevel%2 == 0 {
		return "• "
	}
	return "◦ "
}

// listMarkerLeaderType 是否需要在本列表项首个内容 token 前绘制序号/项目符号。
func listMarkerLeaderType(typ string) bool {
	switch typ {
	case TYPE_TEXT, TYPE_STRONG, TYPE_LINK, TYPE_EM, TYPE_CODESPAN, TYPE_DEL,
		TYPE_LIST, TYPE_BLOCKQUOTE, TYPE_CODE:
		return true
	default:
		return false
	}
}

// headingMarginTop 标题顶部的强制换行间距（通过 MdHardBreak 的 lineHeight 实现）。
func headingMarginTop(depth int) float64 {
	switch depth {
	case 1:
		return mdLineHeight * 1.35
	case 2:
		return mdLineHeight * 1.15
	case 3:
		return mdLineHeight * 1.02
	case 4:
		return mdLineHeight * 0.95
	case 5:
		return mdLineHeight * 0.88
	case 6:
		return mdLineHeight * 0.82
	default:
		return mdLineHeight * 0.9
	}
}

// headingMarginBottom 标题底部的强制换行间距。
func headingMarginBottom(depth int) float64 {
	switch depth {
	case 1:
		return mdLineHeight * 1.28
	case 2:
		return mdLineHeight * 1.08
	case 3:
		return mdLineHeight * 0.96
	case 4:
		return mdLineHeight * 0.88
	case 5:
		return mdLineHeight * 0.8
	case 6:
		return mdLineHeight * 0.74
	default:
		return mdLineHeight * 0.85
	}
}
