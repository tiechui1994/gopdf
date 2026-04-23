package gopdf

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/lex"
	"github.com/tiechui1994/gopdf/util"
)

const (
	FONT_NORMAL = "normal"
	FONT_BOLD   = "bold"
	FONT_IALIC  = "italic"
	FONT_MONO   = "mono"
)

const (
	TYPE_TEXT     = "text"
	TYPE_STRONG   = "strong"   // **strong**
	TYPE_EM       = "em"       // *em*
	TYPE_CODESPAN = "codespan" // `codespan`, ```codespan```
	TYPE_CODE     = "code"     //
	TYPE_LINK     = "link"     // [xx](http://www.link)
	TYPE_IMAGE    = "image"    // ![xxx](https://www.image)
	TYPE_DEL      = "del"      // ~~deleted text~~

	TYPE_SPACE = "space"

	TYPE_PARAGRAPH  = "paragraph"
	TYPE_HEADING    = "heading"
	TYPE_LIST       = "list"
	TYPE_BLOCKQUOTE = "blockquote"
	TYPE_TABLE      = "table"
	TYPE_HR         = "hr"
)

// Layout uses PDF points (pt) only — never device pixels — so output is stable across DPI/screen.
const (
	mdBase       = 12.0 // body font size (pt); typographic base unit
	mdLineHeight = 18.0 // one line step (pt), = mdBase * 1.5
	mdBreakGap   = mdLineHeight * (8.0 / 18.0)
)
const (
	lineHeight  = mdLineHeight
	breakHeight = mdBreakGap
	fontSize    = mdBase
)
const (
	spaceLen = mdLineHeight * (4.425 / 18.0)
	blockLen = spaceLen * 0.6
)

// listNestIndentWidth is the additional inset applied for each nested list level.
// Markdown examples in this repo describe nesting via 2-space indentation, so we keep this
// noticeably smaller than blockquoteIndentWidth (which includes a gutter + bar).
func listNestIndentWidth() float64 { return 2 * spaceLen }

// markdownBlockquoteIndentSteps is how many em-derived “space” units one blockquote tier indents
// (bar + text gutter). Used instead of scattered numeric offsets.
const markdownBlockquoteIndentSteps = 4

// blockquoteIndentWidth is the horizontal inset for one nesting level (padding and inter-bar step).
func blockquoteIndentWidth() float64 { return float64(markdownBlockquoteIndentSteps) * spaceLen }

// blockquoteBarOffset is the left edge of the level-th vertical bar (0 = bar for current depth).
func blockquoteBarOffset(level int) float64 { return float64(level) * blockquoteIndentWidth() }

// atMarkdownLineLeft is true at the start of a body line: page margin or list hang column.
// Blockquote padding must apply here; only checking the page margin misses list > blockquote.
func atMarkdownLineLeft(x1, pageStartX, listHangIndent float64) bool {
	if math.Abs(x1-pageStartX) < 0.5 {
		return true
	}
	if listHangIndent > 0 && math.Abs(x1-(pageStartX+listHangIndent)) < 0.5 {
		return true
	}
	return false
}

func stripListParagraphIndent(s string) string {
	// The lexer preserves some indentation inside list items (e.g. "  indent the next line...").
	// For PDF layout, we want those continuation paragraphs to align with the list body column,
	// not to render with an extra visual first-line indent.
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if len(ln) == 0 {
			continue
		}
		// remove up to 2 leading spaces
		j := 0
		for j < len(ln) && j < 2 && ln[j] == ' ' {
			j++
		}
		lines[i] = ln[j:]
	}
	return strings.Join(lines, "\n")
}

// mdScale returns a vertical/horizontal offset as a fraction of the line step (pt).
func mdScale(frac float64) float64 { return mdLineHeight * frac }

// codeBlockPad is inner padding (pt) between the grey code band edge and the monospace text.
func codeBlockPad() float64 { return mdLineHeight * (4.0 / 18.0) }

// listNestBreakBefore is vertical space before a nested list (tighter than a full line).
func listNestBreakBefore() float64 { return mdLineHeight * 0.30 }

// listCodeBreakBefore is vertical space before a list-item code block.
func listCodeBreakBefore() float64 { return lineHeight * 0.42 }

// codeBlockAfterGap is vertical space after a rendered code (pre) block so consecutive
// code blocks in the same column do not “stick” together.
func codeBlockAfterGap() float64 { return mdLineHeight * 0.5 }

// blockquoteBarVOverlap nudges the vertical bar slightly tall so segments between MdText
// and MdSpace meet without visible gaps in nested blockquote + code paths.
func blockquoteBarVOverlap() float64 { return mdLineHeight * 0.1 }

// unorderedListBulletPrefix alternates filled bullet vs hollow ring by nesting (• / ◦ / • / …).
func unorderedListBulletPrefix(nestLevel int) string {
	if nestLevel < 0 {
		nestLevel = 0
	}
	if nestLevel%2 == 0 {
		return "• "
	}
	return "◦ "
}

// listMarkerLeaderType is true for block tokens that should be preceded by this list item’s marker.
func listMarkerLeaderType(typ string) bool {
	switch typ {
	case TYPE_TEXT, TYPE_STRONG, TYPE_LINK, TYPE_EM, TYPE_CODESPAN, TYPE_DEL,
		TYPE_LIST, TYPE_BLOCKQUOTE, TYPE_CODE:
		return true
	default:
		return false
	}
}

// headingMarginTop/Bottom are block margins (pt) before/after heading lines — MdSpace at line
// start collapses to breakHeight, so headings use MdHardBreak for predictable spacing.
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

// codeFittedRuneIndex is the largest n with MeasureTextWidth(runes[:n]) <= avail.
// Used for TYPE_CODE only (pre must not use word-based wrapping, which over-breaks on spaces).
func (c *MdText) codeFittedRuneIndex(runes []rune, avail float64) int {
	if len(runes) == 0 {
		return 0
	}
	eps := c.precision
	if eps < 0.02 {
		eps = 0.02
	}
	if c.pdf.MeasureTextWidth(string(runes)) <= avail+eps {
		return len(runes)
	}
	if c.pdf.MeasureTextWidth(string(runes[0:1])) > avail+eps {
		return 1
	}
	lo, hi := 0, len(runes)
	for lo+1 < hi {
		mid := (lo + hi + 1) / 2
		wm := c.pdf.MeasureTextWidth(string(runes[0:mid]))
		if wm <= avail+eps {
			lo = mid
		} else {
			hi = mid
		}
	}
	if lo < 1 {
		return 1
	}
	return lo
}

// applyWordAwareSlice picks a break end in runes[i:end] so we prefer the last space (avoids "Ma"+"rkdown").
// Returns the visible line, its width, and cut = rune index in runes where the remainder starts (after consumed spaces).
func (c *MdText) applyWordAwareSlice(runes []rune, i, end int, avail float64) (line string, w float64, cut int) {
	cut = end
	if end <= i || end > len(runes) {
		return "", 0, i
	}
	if end < len(runes) && runes[end-1] != ' ' {
		lastSpace := -1
		for k := end - 1; k > i; k-- {
			if runes[k] == ' ' {
				lastSpace = k
				break
			}
		}
		if lastSpace > i {
			cand := string(runes[i:lastSpace])
			cw := c.pdf.MeasureTextWidth(cand)
			if cw <= avail+c.precision && len(cand) > 0 {
				k := lastSpace + 1
				for k < len(runes) && runes[k] == ' ' {
					k++
				}
				return cand, cw, k
			}
		}
	}
	seg := string(runes[i:end])
	return seg, c.pdf.MeasureTextWidth(seg), end
}

const (
	color_black = "1,1,1"
	color_gray  = "128,128,128"
	color_white = "255,255,255"

	color_pink       = "199,37,78"
	color_lightgray  = "220,220,220"
	color_whitesmoke = "245,245,245"
	color_blue       = "0,0,255"
)

var re struct {
	notwords  *regexp.Regexp
	breakline *regexp.Regexp
}

func init() {
	re.notwords = regexp.MustCompile(`[\n \t=#%@&"':<>,(){}_;/\?\.\+\-\=\^\$\[\]\!]`)
	re.breakline = regexp.MustCompile(`\n{2,}$`)
}

// monoFamilyFrom returns FONT_MONO when registered, otherwise the body font.
func monoFamilyFrom(fonts map[string]string) string {
	if fonts == nil {
		return ""
	}
	if m := fonts[FONT_MONO]; m != "" {
		return m
	}
	return fonts[FONT_NORMAL]
}

// Token is parse markdown result element
type Token = lex.Token

type mardown interface {
	SetText(font interface{}, text ...string)
	GetType() string
	GenerateAtomicCell() (pagebreak, over bool, err error)
}

type abstract struct {
	pdf            *core.Report // core reporter
	padding        float64      // padding left length
	lineHeight     float64      // line height
	blockquote     int          // the cuurent ele is blockquote
	Type           string
	listHangIndent float64 // pageStartX + this = list body text column (bullet wrap / blocks)
	// If >0, blockquote vertical bars are drawn at pageStartX+this (pt); list+bq need a stable
	// anchor so the bar does not follow x1 of wrapped/nested list lines. 0 = use legacy x1.
	blockquoteBarLeft float64
}

func (a *abstract) SetText(interface{}, ...string) {
}
func (a *abstract) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return false, true, nil
}
func (a *abstract) GetType() string {
	return a.Type
}

func hasBreakLine(token Token) bool {
	switch token.Type {
	case TYPE_CODE:
		return re.breakline.MatchString(token.Raw)
	default:
		return strings.HasSuffix(token.Raw, "\n")
	}
}

func repairText(TYPE, text string) string {
	switch TYPE {
	case TYPE_CODE:
		return text
	case TYPE_TEXT, TYPE_STRONG, TYPE_EM, TYPE_CODESPAN, TYPE_LINK, TYPE_DEL:
		// Keep '\n' so PDF can break lines like the Markdown source; use lexer "br" for hard breaks.
		return text
	default:
		return text
	}
}

///////////////////////////////////////////////////////////////////

// Atomic component
type MdText struct {
	abstract
	font core.Font

	stoped    bool    // symbol stoped
	precision float64 // sigle text char length
	text      string  // text content
	remain    string  // renain texy
	link      string  // special TYPE_LINK
	newlines  int

	offsetx float64
	offsety float64
}

func (c *MdText) SetText(font interface{}, texts ...string) {
	if len(texts) == 0 {
		panic("text is invalid")
	}

	switch font.(type) {
	case string:
		family := font.(string)
		switch c.Type {
		case TYPE_STRONG:
			c.font = core.Font{Family: family, Size: fontSize, Style: ""}
		case TYPE_EM:
			// Markdown emphasis is italic: use FONT_IALIC (a real italic TTF), not Style "I"
			// (gopdf only supports "I" when that style is embedded in the registered font).
			c.font = core.Font{Family: family, Size: fontSize, Style: ""}
		case TYPE_CODESPAN, TYPE_CODE:
			c.font = core.Font{Family: family, Size: fontSize, Style: ""}
		case TYPE_LINK, TYPE_TEXT:
			c.font = core.Font{Family: family, Size: fontSize, Style: ""}
		case TYPE_DEL:
			c.font = core.Font{Family: family, Size: fontSize, Style: ""}
		}
	case core.Font:
		c.font = font.(core.Font)
	default:
		panic(fmt.Sprintf("invalid type: %v", c.Type))
	}

	if c.lineHeight == 0 {
		switch c.Type {
		case TYPE_CODE, TYPE_CODESPAN:
			c.lineHeight = lineHeight
		case TYPE_TEXT, TYPE_LINK, TYPE_STRONG, TYPE_EM, TYPE_DEL:
			c.lineHeight = lineHeight
		}
	}

	text := strings.Replace(texts[0], "\t", "    ", -1)
	c.text = repairText(c.Type, text)
	c.remain = c.text
	if c.Type == TYPE_LINK {
		c.link = texts[1]
	}
	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)

	subs := re.notwords.FindAllString(c.text, -1)
	if len(subs) > 0 {
		str := re.notwords.ReplaceAllString(c.text, "")
		length := c.pdf.MeasureTextWidth(str)
		c.precision = length / float64(len([]rune(str)))
	} else {
		length := c.pdf.MeasureTextWidth(c.text)
		c.precision = length / float64(len([]rune(c.text)))
	}
}

func (c *MdText) GenerateAtomicCell() (pagebreak, over bool, err error) {
	lineheight := c.lineHeight
	pageStartX, _ := c.pdf.GetPageStartXY()
	pageEndX, pageEndY := c.pdf.GetPageEndXY()
	x1, y := c.pdf.GetXY()
	x2 := pageEndX

	// List body: (1) cursor still at page left after br/space → snap to text column;
	// (2) previous MdText fragment ended flush right → this token is a new source line → wrap to hang column.
	if c.listHangIndent > 0 {
		targetX := pageStartX + c.listHangIndent
		if x1 <= pageStartX+0.5 {
			x1 = targetX
			c.pdf.SetXY(x1, y)
		} else if x1 < targetX-0.5 {
			// After list marker, x is at pageStart+offset+marker; snap to the body column
			// (offset + padding + marker) when still short of the hang column.
			x1 = targetX
			c.pdf.SetXY(x1, y)
		} else if x1 >= pageEndX-5 {
			x1 = targetX
			y += lineheight
			c.pdf.SetXY(x1, y)
		}
	}

	// TYPE_CODE: apply c.padding as physical X, not via needpadding + width shrink in GetSubText
	// (that double-counts and can make the line width near zero → one character per line).
	if c.Type == TYPE_CODE && c.listHangIndent == 0 && c.padding > 0 {
		if math.Abs(x1-pageStartX) < 1.0 {
			x1 = pageStartX + c.padding
			c.pdf.SetXY(x1, y)
		}
	}

	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)

	tl, tr := x1, x2
	if c.Type == TYPE_CODE {
		p := codeBlockPad()
		if tr-tl > 2*p {
			tl, tr = tl+p, tr-p
		}
	}
	text, width, newline := c.GetSubText(tl, tr)
	for !c.stoped {
		// PDF Cell() uses baseline Y; BackgroundColor uses upper-left. Align with font ascender/descender.
		asc, desc := c.pdf.GetFontMetrics(c.font.Family, float64(c.font.Size))
		inlinePad := mdScale(0.35 / 18.0)
		emH := asc - desc
		if emH < 1 {
			emH = mdBase * 1.2
		}

		barTop := y - asc - inlinePad
		barH := lineheight + 2*inlinePad
		vO := blockquoteBarVOverlap()
		if c.Type == TYPE_CODE {
			cp := codeBlockPad()
			bh := emH + 2*cp
			if bh < lineheight+2*cp {
				bh = lineheight + 2*cp
			}
			barH = bh
			barTop = y - asc - cp
		}
		barH += vO
		barTop -= vO * 0.5
		atPage := math.Abs(x1-pageStartX) < 1.0
		atListText := c.listHangIndent > 0 && math.Abs(x1-(pageStartX+c.listHangIndent)) < 2.0
		// Fenced/typed code in blockquote uses x1 = pageStartX+c.padding (no list): bar must still draw.
		atBqCode := c.Type == TYPE_CODE && c.blockquote > 0 && c.listHangIndent == 0 && c.padding > 0 &&
			math.Abs(x1-(pageStartX+c.padding)) < 1.5
		if c.blockquote > 0 && (atPage || atListText || atBqCode) {
			barX := x1
			if c.blockquoteBarLeft > 0 {
				barX = pageStartX + c.blockquoteBarLeft
			} else if atBqCode {
				barX = pageStartX
			}
			for i := 0; i < c.blockquote; i++ {
				c.pdf.BackgroundColor(barX+blockquoteBarOffset(i), barTop, blockLen, barH, color_gray, "0000")
			}
		}

		switch c.Type {
		case TYPE_CODESPAN:
			bgTop := y - asc - inlinePad
			bgH := math.Max(emH+2*inlinePad, lineheight-mdScale(0.5/18.0))
			c.pdf.BackgroundColor(x1, bgTop, width, bgH, color_lightgray, "1111", color_whitesmoke)
			c.pdf.TextColor(util.RGB(color_pink))
			c.pdf.Cell(x1, y, text)
			c.pdf.TextColor(util.RGB(color_black))
		case TYPE_CODE:
			codePad := codeBlockPad()
			bgH := emH + 2*codePad
			if bgH < lineheight+2*codePad {
				bgH = lineheight + 2*codePad
			}
			bgTop := y - asc - codePad
			// Full-width band: left edge = body column (physical x1 for code, or list/bq insets).
			// c.padding is the blockquote (or root indent) body inset; do not use pageStartX when
			// x1 is already at pageStartX+c.padding.
			bgLeft := x1
			if c.listHangIndent > 0 {
				bgLeft = pageStartX + c.listHangIndent
			} else if c.blockquote > 0 {
				bgLeft = pageStartX + c.padding
			} else if c.padding > 0 {
				bgLeft = pageStartX + c.padding
			} else if math.Abs(x1-pageStartX) < 0.5 {
				bgLeft = pageStartX
			}
			fullW := x2 - bgLeft
			if fullW < 1 {
				fullW = x2 - x1
			}
			c.pdf.BackgroundColor(bgLeft, bgTop, fullW, bgH, color_whitesmoke, "0000")
			c.pdf.TextColor(util.RGB(color_black))
			c.pdf.Cell(x1+codePad, y, text)
			c.pdf.TextColor(util.RGB(color_black))

		case TYPE_LINK:
			// text
			c.pdf.TextColor(util.RGB(color_blue))
			c.pdf.ExternalLink(x1, y, lineheight, text, c.link)
			c.pdf.TextColor(util.RGB(color_black))
		case TYPE_DEL:
			dAsc, _ := c.pdf.GetFontMetrics(c.font.Family, float64(c.font.Size))
			if dAsc < 1 {
				dAsc = lineheight * 0.38
			}
			strikeY := y - dAsc*0.28
			c.pdf.TextColor(util.RGB(color_gray))
			c.pdf.Cell(x1, y, text)
			c.pdf.LineType("straight", 0.3)
			c.pdf.LineH(x1, strikeY, x1+width)
			c.pdf.TextColor(util.RGB(color_black))
		default:
			c.pdf.Cell(x1+c.offsetx, y+c.offsety, text)
		}

		if newline {
			if c.listHangIndent > 0 {
				x1 = pageStartX + c.listHangIndent
			} else if c.Type == TYPE_CODE && c.padding > 0 {
				x1 = pageStartX + c.padding
			} else {
				x1, _ = c.pdf.GetPageStartXY()
			}
			y += c.lineHeight
		} else {
			x1 += width
		}

		// need new page, x,y must statisfy condition
		if (y >= pageEndY || pageEndY-y < lineHeight) && (newline || math.Abs(x1-pageEndX) < c.precision) {
			return true, c.stoped, nil
		}

		c.pdf.SetXY(x1, y)
		tl, tr = x1, x2
		if c.Type == TYPE_CODE {
			p := codeBlockPad()
			if tr-tl > 2*p {
				tl, tr = tl+p, tr-p
			}
		}
		text, width, newline = c.GetSubText(tl, tr)
	}

	return false, c.stoped, nil
}

// GetSubText, Returns the content of a string of length x2-x1.
// This string is a substring of text.
// After return, the remain and length will change
func (c *MdText) GetSubText(x1, x2 float64) (text string, width float64, newline bool) {
	if len(c.remain) == 0 {
		c.stoped = true
		return "", 0, false
	}

	pageX, _ := c.pdf.GetPageStartXY()
	needpadding := c.padding > 0 && atMarkdownLineLeft(x1, pageX, c.listHangIndent)
	// Pre blocks: padding is applied as physical x1 in GenerateAtomicCell; do not also shrink
	// the inner line width (that caused one-character “lines” and wrong word breaks).
	if c.Type == TYPE_CODE {
		needpadding = false
	}
	remainText := c.remain
	index := strings.Index(c.remain, "\n")
	suffix := ""
	if index != -1 {
		newline = true
		remainText = c.remain[:index]
		suffix = c.remain[index:]
	}

	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)
	width = math.Abs(x1 - x2)
	length := c.pdf.MeasureTextWidth(remainText)

	if needpadding {
		width -= c.padding
	}
	defer func() {
		if needpadding {
			space := c.pdf.MeasureTextWidth(" ")
			text = strings.Repeat(" ", int(c.padding/space)) + text
			width += c.padding
		}
	}()

	if length <= width {
		if newline {
			c.remain = c.remain[index+1:]
		} else {
			c.remain = ""
		}
		return remainText, length, newline
	}

	runes := []rune(remainText)
	// Monospaced pre blocks: fit the longest prefix by measured width, no word-based breaking.
	if c.Type == TYPE_CODE {
		hi := c.codeFittedRuneIndex(runes, width)
		if hi < 1 {
			hi = 1
		}
		line := string(runes[0:hi])
		wline := c.pdf.MeasureTextWidth(line)
		c.remain = string(runes[hi:]) + suffix
		c.newlines++
		return line, wline, true
	}

	step := int(float64(len(runes)) * width / length)
	for i, j := 0, step; i < len(runes) && j < len(runes); {
		w := c.pdf.MeasureTextWidth(string(runes[i:j]))

		// less than precision
		if math.Abs(w-width) < c.precision {
			// real with more than page width
			if w-width > 0 {
				line, wline, cut := c.applyWordAwareSlice(runes, 0, j-1, width)
				c.remain = string(runes[cut:]) + suffix
				c.newlines++
				return line, wline, true
			}

			// try again, can more precise
			if j+1 < len(runes) {
				w1 := c.pdf.MeasureTextWidth(string(runes[i : j+1]))
				if math.Abs(w1-width) < c.precision {
					j = j + 1
					continue
				}
			}

			line, wline, cut := c.applyWordAwareSlice(runes, 0, j, width)
			c.remain = string(runes[cut:]) + suffix
			c.newlines++
			return line, wline, true
		}

		if w-width > 0 && w-width > c.precision {
			j--
		}
		if width-w > 0 && width-w > c.precision {
			j++
		}
	}

	return "", 0, false
}

func (c *MdText) String() string {
	text := strings.Replace(c.remain, "\n", "|", -1)
	return fmt.Sprintf("[type=%v,text=%v]", c.Type, text)
}

type MdSpace struct {
	abstract
}

func (c *MdSpace) GenerateAtomicCell() (pagebreak, over bool, err error) {
	var (
		spaceX, spaceY float64
		linehieght     = c.lineHeight
	)

	pageStartX, _ := c.pdf.GetPageStartXY()
	_, pageEndY := c.pdf.GetPageEndXY()
	x, y := c.pdf.GetXY()

	// After TYPE_CODE, x is at pageStartX+c.padding (not page margin) — still “line start”
	// in blockquote body, so we must not stack breakHeight on explicit codeBlockAfterGap().
	atBqBody := c.blockquote > 0 && c.listHangIndent == 0 && c.padding > 0 &&
		math.Abs(x-(pageStartX+c.padding)) < 1.5
	atLineStart := x == pageStartX || (c.listHangIndent > 0 && math.Abs(x-(pageStartX+c.listHangIndent)) < 1.5) || atBqBody

	if c.lineHeight > 0 {
		// E.g. gap after a code block: honor it; at body margin do not add another breakHeight.
		if atLineStart {
			linehieght = c.lineHeight
		} else {
			linehieght = c.lineHeight + breakHeight
		}
	} else if atLineStart {
		linehieght = breakHeight
	} else if linehieght == 0 {
		linehieght = breakHeight + lineHeight
	} else {
		linehieght += breakHeight
	}

	spaceX = pageStartX
	if c.listHangIndent > 0 {
		spaceX = pageStartX + c.listHangIndent
	}
	spaceY = y + linehieght

	if c.blockquote > 0 {
		// Upward ext only: bottom must be y+linehieght (= next SetXY) or inner-bar stumps show on the following line.
		ext := mdLineHeight*0.72 + blockquoteBarVOverlap()*0.5
		barH := linehieght + ext
		barX := spaceX
		if c.blockquoteBarLeft > 0 {
			barX = pageStartX + c.blockquoteBarLeft
		}
		for i := 0; i < c.blockquote; i++ {
			c.pdf.BackgroundColor(barX+blockquoteBarOffset(i), y-ext, blockLen, barH, color_gray, "0000")
		}
	}

	if pageEndY-spaceY < lineHeight {
		return true, true, nil
	}

	c.pdf.SetXY(spaceX, spaceY)
	return false, true, nil
}

func (c *MdSpace) String() string {
	return fmt.Sprint("[type=space]")
}

// MdHardBreak forces a new line (Markdown "  \\n" / <br>).
type MdHardBreak struct {
	abstract
	// indentX, when non-zero, sets the new line X to pageStartX+indentX (e.g. nested list marker column).
	indentX float64
}

func (m *MdHardBreak) SetText(interface{}, ...string) {}

func (m *MdHardBreak) GetType() string {
	return "br"
}

func (m *MdHardBreak) GenerateAtomicCell() (pagebreak, over bool, err error) {
	pageStartX, _ := m.pdf.GetPageStartXY()
	_, pageEndY := m.pdf.GetPageEndXY()
	_, y := m.pdf.GetXY()
	lh := m.lineHeight
	if lh == 0 {
		lh = lineHeight
	}
	newY := y + lh
	if newY >= pageEndY || pageEndY-newY < mdScale(0.5/18.0) {
		return true, true, nil
	}
	// Bridge blockquote vertical bars across paragraph/code boundaries (no MdSpace bar).
	if m.blockquote > 0 {
		ext := mdLineHeight * 0.3
		vO := blockquoteBarVOverlap()
		barH := lh + 2*ext + vO
		barTop := y - ext - vO*0.5
		barX := pageStartX
		if m.blockquoteBarLeft > 0 {
			barX = pageStartX + m.blockquoteBarLeft
		}
		for i := 0; i < m.blockquote; i++ {
			m.pdf.BackgroundColor(barX+blockquoteBarOffset(i), barTop, blockLen, barH, color_gray, "0000")
		}
	}
	// Reset to the page margin (or indentX for nested lists). The next MdText snaps using
	// listHangIndent; indentX must match the list marker column when the previous line ended
	// at pageStartX so we do not rely on a duplicate horizontal offset on the marker glyph.
	newX := pageStartX + m.indentX
	m.pdf.SetXY(newX, newY)
	return false, true, nil
}

type MdImage struct {
	abstract
	image  *Image
	height float64
}

func (i *MdImage) SetText(_ interface{}, filename ...string) {
	var filepath string
	if strings.HasPrefix(filename[0], "http") {
		response, err := http.DefaultClient.Get(filename[0])
		if err != nil {
			log.Println(err)
			return
		}

		imageType := response.Header.Get("Content-Type")
		switch imageType {
		case "image/png":
			filepath = fmt.Sprintf("/tmp/%v.png", time.Now().Unix())
			fd, _ := os.Create(filepath)
			io.Copy(fd, response.Body)
		case "image/jpeg":
			filepath = fmt.Sprintf("/tmp/%v.jpeg", time.Now().Unix())
			fd, _ := os.Create(filepath)
			io.Copy(fd, response.Body)
		}

	} else {
		filepath = filename[0]
	}

	if i.height == 0 {
		i.height = lineHeight
	}

	i.image = NewImageWithWidthAndHeight(filepath, 0, i.height, i.pdf)
}

func (i *MdImage) GenerateAtomicCell() (pagebreak, over bool, err error) {
	if i.image == nil {
		return false, true, nil
	}
	return i.image.GenerateAtomicCell()
}

func (i *MdImage) GetType() string {
	return i.Type
}

///////////////////////////////////////////////////////////////////

func CommonGenerateAtomicCell(children *[]mardown) (pagebreak, over bool, err error) {
	for i, comment := range *children {
		pagebreak, over, err = comment.GenerateAtomicCell()
		if err != nil {
			return
		}

		if pagebreak {
			if over && i != len(*children)-1 {
				*children = (*children)[i+1:]
				return pagebreak, len(*children) == 0, nil
			}

			*children = (*children)[i:]
			return pagebreak, len(*children) == 0, nil
		}
	}
	return false, true, nil
}

// Combination components

type MdMutiText struct {
	abstract
	fonts    map[string]string
	children []mardown
}

func (m *MdMutiText) getabstract(typ string) abstract {
	return abstract{
		pdf:               m.pdf,
		padding:           m.padding,
		blockquote:        m.blockquote,
		Type:              typ,
		listHangIndent:    m.listHangIndent,
		blockquoteBarLeft: m.blockquoteBarLeft,
	}
}

func (m *MdMutiText) SetToken(t Token) error {
	if m.fonts == nil || len(m.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_TEXT {
		return fmt.Errorf("invalid type")
	}

	n := len(t.Tokens)
	for i := 0; i < n; i++ {
		token := t.Tokens[i]
		abs := m.getabstract(token.Type)
		switch token.Type {
		case TYPE_TEXT:
			if len(token.Tokens) <= 1 {
				text := &MdText{abstract: abs}
				txt := token.Text
				if abs.listHangIndent > 0 {
					txt = stripListParagraphIndent(txt)
				}
				text.SetText(m.fonts[FONT_NORMAL], txt)
				m.children = append(m.children, text)
			} else {
				mutiltext := &MdMutiText{abstract: abs, fonts: m.fonts}
				mutiltext.SetToken(token)
				m.children = append(m.children, mutiltext.children...)
			}

		case TYPE_LINK:
			link := &MdText{abstract: abs}
			link.SetText(m.fonts[FONT_NORMAL], token.Text, token.Href)
			m.children = append(m.children, link)
		case TYPE_EM:
			em := &MdText{abstract: abs}
			em.SetText(m.fonts[FONT_IALIC], token.Text)
			m.children = append(m.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{abstract: abs}
			codespan.SetText(monoFamilyFrom(m.fonts), token.Text)
			m.children = append(m.children, codespan)
		case TYPE_STRONG:
			strong := &MdText{abstract: abs}
			text := token.Text
			if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
				text = token.Tokens[0].Text
			}
			strong.SetText(m.fonts[FONT_BOLD], text)
			m.children = append(m.children, strong)
		case TYPE_DEL:
			del := &MdText{abstract: abs}
			del.SetText(m.fonts[FONT_NORMAL], token.Text)
			m.children = append(m.children, del)
		}
	}

	return nil
}

type MdHeader struct {
	abstract
	fonts    map[string]string
	children []mardown
}

func (h *MdHeader) CalFontSizeAndLineHeight(size int) (fontsize int, lineheight float64) {
	var fs int
	switch size {
	case 1:
		fs = 22
	case 2:
		fs = 18
	case 3:
		fs = 16
	case 4:
		fs = 13
	case 5:
		fs = 12
	case 6:
		fs = 11
	default:
		fs = 14
	}
	lh := float64(fs) * 1.38
	minLH := mdLineHeight * 1.08
	if lh < minLH {
		lh = minLH
	}
	return fs, lh
}

func (h *MdHeader) getabstract(typ string) abstract {
	return abstract{
		pdf:               h.pdf,
		padding:           h.padding,
		blockquote:        h.blockquote,
		Type:              typ,
		listHangIndent:    h.listHangIndent,
		blockquoteBarLeft: h.blockquoteBarLeft,
	}
}

func (h *MdHeader) SetToken(t Token) (err error) {
	if h.fonts == nil || len(h.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_HEADING {
		return fmt.Errorf("invalid type")
	}

	fontsize, lineheight := h.CalFontSizeAndLineHeight(t.Depth)
	font := core.Font{Family: h.fonts[FONT_BOLD], Size: fontsize}

	absTop := h.getabstract("br")
	topBrk := &MdHardBreak{abstract: absTop}
	topBrk.lineHeight = headingMarginTop(t.Depth)
	h.children = append(h.children, topBrk)

	// Block headings have no inline tokens; use t.Text directly.
	if len(t.Tokens) == 0 && t.Text != "" {
		abs := h.getabstract(TYPE_TEXT)
		abs.lineHeight = lineheight
		text := &MdText{abstract: abs}
		text.SetText(font, t.Text)
		h.children = append(h.children, text)
	} else {
		for _, token := range t.Tokens {
			abs := h.getabstract(token.Type)
			switch token.Type {
			case TYPE_TEXT:
				abs.lineHeight = lineheight
				text := &MdText{abstract: abs}
				text.SetText(font, token.Text)
				h.children = append(h.children, text)
			case TYPE_IMAGE:
				abs.lineHeight = lineheight
				image := &MdImage{abstract: abs}
				h.children = append(h.children, image)
			}
		}
	}

	absBot := h.getabstract("br")
	botBrk := &MdHardBreak{abstract: absBot}
	botBrk.lineHeight = headingMarginBottom(t.Depth)
	h.children = append(h.children, botBrk)

	return nil
}

func (h *MdHeader) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return CommonGenerateAtomicCell(&h.children)
}

type MdParagraph struct {
	abstract
	fonts    map[string]string
	children []mardown
}

func (p *MdParagraph) getabstract(typ string) abstract {
	return abstract{
		pdf:               p.pdf,
		padding:           p.padding,
		blockquote:        p.blockquote,
		Type:              typ,
		listHangIndent:    p.listHangIndent,
		blockquoteBarLeft: p.blockquoteBarLeft,
	}
}

func (p *MdParagraph) SetToken(t Token) error {
	if p.fonts == nil || len(p.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_PARAGRAPH {
		return fmt.Errorf("invalid type")
	}

	for _, token := range t.Tokens {
		abs := p.getabstract(token.Type)
		switch token.Type {
		case "br":
			brk := &MdHardBreak{abstract: abs}
			brk.lineHeight = lineHeight
			p.children = append(p.children, brk)
		case TYPE_LINK:
			link := &MdText{abstract: abs}
			link.SetText(p.fonts[FONT_NORMAL], token.Text, token.Href)
			p.children = append(p.children, link)
		case TYPE_TEXT:
			text := &MdText{abstract: abs}
			text.SetText(p.fonts[FONT_NORMAL], token.Text)
			p.children = append(p.children, text)
		case TYPE_EM:
			em := &MdText{abstract: abs}
			em.SetText(p.fonts[FONT_IALIC], token.Text)
			p.children = append(p.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{abstract: abs}
			codespan.SetText(monoFamilyFrom(p.fonts), token.Text)
			p.children = append(p.children, codespan)
		case TYPE_CODE:
			code := &MdText{abstract: abs}
			code.SetText(monoFamilyFrom(p.fonts), token.Text)
			p.children = append(p.children, code)
		case TYPE_STRONG:
			strong := &MdText{abstract: abs}
			text := token.Text
			if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
				text = token.Tokens[0].Text
			}
			strong.SetText(p.fonts[FONT_BOLD], text)
			p.children = append(p.children, strong)
		case TYPE_IMAGE:
			image := &MdImage{abstract: abs}
			image.SetText("", token.Href)
			p.children = append(p.children, image)
		case TYPE_DEL:
			del := &MdText{abstract: abs}
			del.SetText(p.fonts[FONT_NORMAL], token.Text)
			p.children = append(p.children, del)
		default:
			// unknown inline token; skip
		}
	}

	return nil
}

func (p *MdParagraph) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return CommonGenerateAtomicCell(&p.children)
}

type MdList struct {
	abstract
	fonts     map[string]string
	children  []mardown
	nestLevel int
}

func (l *MdList) getabstract(typ string) abstract {
	return abstract{
		pdf:               l.pdf,
		padding:           0, // list block offset is in listHangIndent only; non-zero padding here adds fake leading spaces in GetSubText
		blockquote:        l.blockquote,
		Type:              typ,
		listHangIndent:    l.listHangIndent,
		blockquoteBarLeft: l.blockquoteBarLeft,
	}
}

func (l *MdList) SetToken(t Token) error {
	if l.fonts == nil || len(l.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_LIST {
		return fmt.Errorf("invalid type")
	}

	l.pdf.Font(l.fonts[FONT_NORMAL], int(fontSize), "")
	l.pdf.SetFontWithStyle(l.fonts[FONT_NORMAL], "", int(fontSize))

	var drewMarker bool
	for index, item := range t.Items {
		drewMarker = false
		n := len(item.Tokens)

		var marker string
		if t.Ordered {
			marker = fmt.Sprintf("%d. ", index+1)
		} else {
			marker = unorderedListBulletPrefix(l.nestLevel)
		}
		// listHangIndent is the x-offset of the list block (e.g. inside a blockquote in a list).
		itemHang := l.listHangIndent + l.padding + l.pdf.MeasureTextWidth(marker)

		for i, token := range item.Tokens {
			abs := l.getabstract(token.Type)
			abs.listHangIndent = itemHang

			if !drewMarker && listMarkerLeaderType(token.Type) {
				if t.Ordered {
					mabs := l.getabstract(TYPE_TEXT)
					// Marker column = list block origin + indent padding. Use listHangIndent only
					// (no extra offsetx): MdSpace already places the cursor on this column for item 2+,
					// and offsetx+padding was double-counting indent after inter-item space.
					mabs.listHangIndent = l.listHangIndent + l.padding
					mabs.padding = 0
					text := &MdText{abstract: mabs, offsetx: 0, offsety: mdScale(-0.45 / 18.0)}
					text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: int(fontSize)}, marker)
					l.children = append(l.children, text)
				}
				if !t.Ordered {
					mabs := l.getabstract(TYPE_TEXT)
					mabs.listHangIndent = l.listHangIndent + l.padding
					mabs.padding = 0
					text := &MdText{abstract: mabs, offsetx: 0, offsety: mdScale(-0.28 / 18.0)}
					text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: int(fontSize)}, marker)
					l.children = append(l.children, text)
				}
				drewMarker = true
			}

			switch token.Type {
			case TYPE_LIST:
				nestAbs := l.getabstract(token.Type)
				nestAbs.listHangIndent = l.listHangIndent
				// Align nested list with the parent item’s body column, then add one logical nest step.
				nestAbs.padding = itemHang - l.listHangIndent + listNestIndentWidth()

				nest := &MdHardBreak{abstract: l.getabstract("br")}
				nest.lineHeight = listNestBreakBefore()
				// First line of nested list: X must be the nested marker column (not raw page margin),
				// since list markers no longer apply padding via offsetx.
				nest.indentX = l.listHangIndent + nestAbs.padding
				l.children = append(l.children, nest)

				sub := &MdList{abstract: nestAbs, fonts: l.fonts, nestLevel: l.nestLevel + 1}
				sub.SetToken(token)
				l.children = append(l.children, sub.children...)
				continue

			case TYPE_SPACE:
				gap := &MdHardBreak{abstract: l.getabstract("br")}
				gap.lineHeight = mdBreakGap + mdLineHeight*0.48
				l.children = append(l.children, gap)
				continue

			case "br":
				brk := &MdHardBreak{abstract: abs}
				brk.lineHeight = lineHeight
				l.children = append(l.children, brk)
				continue

			case TYPE_BLOCKQUOTE:
				bq := &MdHardBreak{abstract: l.getabstract("br")}
				bq.lineHeight = lineHeight * 0.82
				l.children = append(l.children, bq)

				bqa := l.getabstract(TYPE_BLOCKQUOTE)
				bqa.listHangIndent = itemHang
				bqa.blockquote += 1
				bqa.padding += blockquoteIndentWidth()
				gut := mdLineHeight * (3.5 / 18.0)
				bqa.blockquoteBarLeft = itemHang - gut
				if bqa.blockquoteBarLeft < 0 {
					bqa.blockquoteBarLeft = 0
				}
				blockquote := &MdBlockQuote{abstract: bqa, fonts: l.fonts}
				blockquote.SetToken(token)
				l.children = append(l.children, blockquote.children...)

				if n > 0 && i == n-1 {
					for j := len(l.children) - 1; j >= 0; j-- {
						if sp, ok := l.children[j].(*MdSpace); ok {
							sp.blockquote -= 1
							break
						}
					}
				}
				continue

			case TYPE_CODE:
				cb := &MdHardBreak{abstract: l.getabstract("br")}
				cb.lineHeight = listCodeBreakBefore()
				l.children = append(l.children, cb)

				cabs := l.getabstract(TYPE_CODE)
				cabs.listHangIndent = itemHang
				code := &MdText{abstract: cabs}
				code.SetText(monoFamilyFrom(l.fonts), token.Text+"\n")
				l.children = append(l.children, code)
				ag := &MdSpace{abstract: l.getabstract(TYPE_SPACE)}
				ag.lineHeight = codeBlockAfterGap()
				l.children = append(l.children, ag)
				continue
			}

			abs.listHangIndent = itemHang
			switch token.Type {
			case TYPE_TEXT:
				mutiltext := &MdMutiText{abstract: abs, fonts: l.fonts}
				mutiltext.SetToken(token)
				l.children = append(l.children, mutiltext.children...)
				// Lexer drops single '\n' between list-item lines without a token; the next TYPE_TEXT
				// would otherwise continue on the same baseline (e.g. "...first line," + "but...").
				if i+1 < n && item.Tokens[i+1].Type == TYPE_TEXT {
					brLine := &MdHardBreak{abstract: l.getabstract("br")}
					brLine.lineHeight = lineHeight
					l.children = append(l.children, brLine)
				}
			case TYPE_STRONG:
				strong := &MdText{abstract: abs}
				text := token.Text
				if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
					text = token.Tokens[0].Text
				}
				strong.SetText(l.fonts[FONT_BOLD], text)
				l.children = append(l.children, strong)
			case TYPE_LINK:
				link := &MdText{abstract: abs}
				link.SetText(l.fonts[FONT_NORMAL], token.Text, token.Href)
				l.children = append(l.children, link)
			}
		}

		// After an item, place the cursor on the *marker* start column for this list, not the body
		// column (itemHang). If we use itemHang here, the next item’s marker is too far right and
		// the hang-indent snap logic never runs (listHangIndent for marker ≠ itemHang).
		abs := l.getabstract(TYPE_SPACE)
		abs.listHangIndent = l.listHangIndent + l.padding
		abs.blockquote -= 1
		space := &MdSpace{abstract: abs}
		l.children = append(l.children, space)
	}

	return nil
}

func (l *MdList) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return CommonGenerateAtomicCell(&l.children)
}

func (l *MdList) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "(list")
	for _, child := range l.children {
		fmt.Fprintf(&buf, "%v", child)
	}
	fmt.Fprint(&buf, ")")

	return buf.String()
}

type MdBlockQuote struct {
	abstract
	fonts    map[string]string
	children []mardown
}

func (b *MdBlockQuote) getabstract(typ string) abstract {
	return abstract{
		pdf:               b.pdf,
		padding:           b.padding,
		blockquote:        b.blockquote,
		Type:              typ,
		listHangIndent:    b.listHangIndent,
		blockquoteBarLeft: b.blockquoteBarLeft,
	}
}

func (b *MdBlockQuote) SetToken(t Token) error {
	if b.fonts == nil || len(b.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_BLOCKQUOTE {
		return fmt.Errorf("invalid type")
	}

	n := len(t.Tokens)
	for i := 0; i < n; i++ {
		token := t.Tokens[i]
		abs := b.getabstract(token.Type)
		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{abstract: abs, fonts: b.fonts}
			paragraph.SetToken(token)
			b.children = append(b.children, paragraph.children...)

			// last
			if i == n-1 {
				abs := b.getabstract(TYPE_SPACE)
				space := &MdSpace{abstract: abs}
				space.blockquote -= 1
				b.children = append(b.children, space)
			}

			if i < n-1 {
				pg := &MdHardBreak{abstract: b.getabstract("br")}
				pg.lineHeight = mdBreakGap + mdLineHeight*0.42
				b.children = append(b.children, pg)
				if t.Tokens[i+1].Type == TYPE_SPACE {
					i++
				}
			}

		case TYPE_LIST:
			if i > 0 {
				g := &MdHardBreak{abstract: b.getabstract("br")}
				g.lineHeight = lineHeight * 0.75
				b.children = append(b.children, g)
			}
			la := abs
			// Keep the same body column as blockquote text so nested list markers and lines align.
			la.listHangIndent = b.listHangIndent
			list := &MdList{abstract: la, fonts: b.fonts, nestLevel: 0}
			list.SetToken(token)
			b.children = append(b.children, list.children...)
		case TYPE_HEADING:
			if i > 0 {
				g := &MdHardBreak{abstract: b.getabstract("br")}
				g.lineHeight = mdLineHeight * 0.48
				b.children = append(b.children, g)
			}
			header := &MdHeader{abstract: abs, fonts: b.fonts}
			header.SetToken(token)
			b.children = append(b.children, header.children...)
		case TYPE_BLOCKQUOTE:
			if i > 0 {
				g := &MdHardBreak{abstract: b.getabstract("br")}
				g.lineHeight = mdLineHeight * 0.48
				b.children = append(b.children, g)
			}
			abs.blockquote += 1
			abs.padding += blockquoteIndentWidth()
			blockquote := &MdBlockQuote{abstract: abs, fonts: b.fonts}
			blockquote.SetToken(token)
			b.children = append(b.children, blockquote.children...)

			if n > 0 && i == n-1 {
				l := len(b.children)
				b.children[l-1].(*MdSpace).blockquote -= 1
			}
		case TYPE_TEXT:
			mutiltext := &MdMutiText{abstract: abs, fonts: b.fonts}
			mutiltext.SetToken(token)
			b.children = append(b.children, mutiltext.children...)
		case TYPE_SPACE:
			if i == len(t.Tokens)-1 {
				abs.blockquote -= 1
			}
			space := &MdSpace{abstract: abs}
			b.children = append(b.children, space)
		case TYPE_LINK:
			link := &MdText{abstract: abs}
			link.SetText(b.fonts[FONT_NORMAL], token.Text, token.Href)
			b.children = append(b.children, link)
		case TYPE_CODE:
			if i > 0 {
				g := &MdHardBreak{abstract: b.getabstract("br")}
				g.lineHeight = mdLineHeight * 0.52
				b.children = append(b.children, g)
			}
			code := &MdText{abstract: abs}
			code.SetText(monoFamilyFrom(b.fonts), token.Text+"\n")
			b.children = append(b.children, code)
			// One trailing space: gap after code (fenced with extra blank uses TYPE_TEXT as before).
			if hasBreakLine(token) {
				spa := b.getabstract(TYPE_TEXT)
				br := &MdSpace{abstract: spa}
				br.lineHeight = codeBlockAfterGap()
				b.children = append(b.children, br)
			} else {
				gap := &MdSpace{abstract: b.getabstract(TYPE_SPACE)}
				gap.lineHeight = codeBlockAfterGap()
				b.children = append(b.children, gap)
			}
		case TYPE_EM:
			em := &MdText{abstract: abs}
			em.SetText(b.fonts[FONT_IALIC], token.Text)
			b.children = append(b.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{abstract: abs}
			codespan.SetText(monoFamilyFrom(b.fonts), token.Text)
			b.children = append(b.children, codespan)
		case TYPE_STRONG:
			strong := &MdText{abstract: abs}
			text := token.Text
			if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
				text = token.Tokens[0].Text
			}
			strong.SetText(b.fonts[FONT_BOLD], text)
			b.children = append(b.children, strong)
		case TYPE_DEL:
			del := &MdText{abstract: abs}
			del.SetText(b.fonts[FONT_NORMAL], token.Text)
			b.children = append(b.children, del)
		}
	}

	l := len(b.children)
	if l > 0 {
		lastType := b.children[l-1].GetType()
		if lastType != TYPE_SPACE {
			abs := b.getabstract(TYPE_TEXT)
			abs.blockquote -= 1
			br := &MdSpace{abstract: abs}
			b.children = append(b.children, br)
		}
	}

	return nil
}

type MdTable struct {
	abstract
	fonts    map[string]string
	children []mardown
}

func (tb *MdTable) getabstract(typ string) abstract {
	return abstract{
		pdf:               tb.pdf,
		padding:           tb.padding,
		blockquote:        tb.blockquote,
		Type:              typ,
		listHangIndent:    tb.listHangIndent,
		blockquoteBarLeft: tb.blockquoteBarLeft,
	}
}

func (tb *MdTable) SetToken(t Token) error {
	if tb.fonts == nil || len(tb.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_TABLE {
		return fmt.Errorf("invalid type")
	}

	cols := len(t.Header)
	rows := len(t.Cells) + 1 // header + data rows

	if cols == 0 || rows == 0 {
		return nil
	}

	header := lex.PadAlignTo(t.Header, cols)
	align := lex.PadAlignTo(t.Align, cols)
	cells := make([][]string, len(t.Cells))
	for i := range t.Cells {
		cells[i] = lex.PadAlignTo(t.Cells[i], cols)
	}

	pageEndX, _ := tb.pdf.GetPageEndXY()
	pageStartX, _ := tb.pdf.GetPageStartXY()
	tableWidth := pageEndX - pageStartX

	table := NewTable(cols, rows, tableWidth, lineHeight, tb.pdf)
	table.SetMargin(core.Scope{})

	border := core.NewScope(4.0, 4.0, 4.0, 3.0)
	f := core.Font{Family: tb.fonts[FONT_BOLD], Size: int(fontSize), Style: ""}

	// header row
	for j, h := range header {
		cell := table.NewCell()
		tc := NewTextCell(table.GetColWidth(0, j), lineHeight, 1.0, tb.pdf)
		tc.SetFont(f).SetBorder(border).SetContent(strings.TrimSpace(h))
		tc.VerticalCentered()
		// apply alignment
		if j < len(align) {
			switch align[j] {
			case "center":
				tc.HorizontalCentered()
			case "right":
				tc.RightAlign()
			}
		}
		cell.SetElement(tc)
	}

	// data rows
	f = core.Font{Family: tb.fonts[FONT_NORMAL], Size: int(fontSize), Style: ""}
	for i, row := range cells {
		for j, val := range row {
			if j >= cols {
				break
			}
			cell := table.NewCell()
			tc := NewTextCell(table.GetColWidth(i+1, j), lineHeight, 1.0, tb.pdf)
			tc.SetFont(f).SetBorder(border).SetContent(strings.TrimSpace(val))
			tc.VerticalCentered()
			// apply alignment
			if j < len(align) {
				switch align[j] {
				case "center":
					tc.HorizontalCentered()
				case "right":
					tc.RightAlign()
				}
			}
			cell.SetElement(tc)
		}
	}

	tb.children = append(tb.children, &mdTableRenderer{table: table})

	return nil
}

func (tb *MdTable) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return CommonGenerateAtomicCell(&tb.children)
}

// mdTableRenderer wraps Table to implement the mardown interface
type mdTableRenderer struct {
	table *Table
}

func (m *mdTableRenderer) SetText(font interface{}, text ...string) {}

func (m *mdTableRenderer) GetType() string {
	return TYPE_TABLE
}

func (m *mdTableRenderer) GenerateAtomicCell() (pagebreak, over bool, err error) {
	err = m.table.GenerateAtomicCell()
	return false, true, err
}

type MdHr struct {
	abstract
}

func (hr *MdHr) GenerateAtomicCell() (pagebreak, over bool, err error) {
	pageStartX, _ := hr.pdf.GetPageStartXY()
	pageEndX, _ := hr.pdf.GetPageEndXY()
	_, y := hr.pdf.GetXY()

	// y is previous paragraph baseline (blank line before --- is skipped in SetTokens).
	above := mdLineHeight * (15.0 / 18.0)
	below := mdLineHeight * (12.0 / 18.0)
	ruleY := y + above
	hr.pdf.LineType("straight", 0.5)
	hr.pdf.LineH(pageStartX, ruleY, pageEndX)

	_, pageEndY := hr.pdf.GetPageEndXY()
	spaceY := ruleY + below
	if pageEndY-spaceY < lineHeight {
		return true, true, nil
	}

	hr.pdf.SetXY(pageStartX, spaceY)
	return false, true, nil
}

type MarkdownText struct {
	quote       bool
	pdf         *core.Report
	fonts       map[string]string
	children    []mardown
	x           float64
	writedLines int
}

func NewMarkdownText(pdf *core.Report, x float64, fonts map[string]string) (*MarkdownText, error) {
	px, _ := pdf.GetPageStartXY()
	if x < px {
		x = px
	}

	if fonts == nil || fonts[FONT_BOLD] == "" || fonts[FONT_IALIC] == "" || fonts[FONT_NORMAL] == "" {
		return nil, fmt.Errorf("invalid fonts")
	}

	mt := MarkdownText{
		pdf:   pdf,
		x:     x,
		fonts: fonts,
	}

	return &mt, nil
}

func (mt *MarkdownText) getabstract(typ string) abstract {
	return abstract{
		pdf:  mt.pdf,
		Type: typ,
	}
}

func (mt *MarkdownText) SetTokens(tokens []Token) {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		abs := mt.getabstract(token.Type)

		// A blank line before --- is tokenized as TYPE_SPACE + TYPE_HR. Drawing both the
		// paragraph-gap space and MdHr's offset stacked ~35pt and looked disconnected.
		if token.Type == TYPE_SPACE && i+1 < len(tokens) && tokens[i+1].Type == TYPE_HR {
			continue
		}

		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{abstract: abs, fonts: mt.fonts}
			paragraph.SetToken(token)
			mt.children = append(mt.children, paragraph.children...)
		case TYPE_LIST:
			la := abs
			la.listHangIndent = 0
			list := &MdList{abstract: la, fonts: mt.fonts, nestLevel: 0}
			list.SetToken(token)
			mt.children = append(mt.children, list.children...)
		case TYPE_HEADING:
			header := &MdHeader{abstract: abs, fonts: mt.fonts}
			header.SetToken(token)
			mt.children = append(mt.children, header.children...)
		case TYPE_BLOCKQUOTE:
			abs.blockquote = 1
			abs.padding += blockquoteIndentWidth()
			blockquote := &MdBlockQuote{abstract: abs, fonts: mt.fonts}
			blockquote.SetToken(token)
			mt.children = append(mt.children, blockquote.children...)
		case TYPE_TEXT:
			mutiltext := &MdMutiText{abstract: abs, fonts: mt.fonts}
			mutiltext.SetToken(token)
			mt.children = append(mt.children, mutiltext.children...)
		case TYPE_SPACE:
			space := &MdSpace{abstract: abs}
			mt.children = append(mt.children, space)
		case TYPE_LINK:
			link := &MdText{abstract: abs}
			link.SetText(mt.fonts[FONT_NORMAL], token.Text, token.Href)
			mt.children = append(mt.children, link)
		case TYPE_CODE:
			abs.padding = mdScale(15.0 / 18.0)
			code := &MdText{abstract: abs}
			code.SetText(monoFamilyFrom(mt.fonts), token.Text+"\n")
			mt.children = append(mt.children, code)

			abs.lineHeight = codeBlockAfterGap()
			space := &MdSpace{abstract: abs}
			mt.children = append(mt.children, space)
		case "br":
			brk := &MdHardBreak{abstract: abs}
			brk.lineHeight = lineHeight
			mt.children = append(mt.children, brk)
		case TYPE_EM:
			em := &MdText{abstract: abs}
			em.SetText(mt.fonts[FONT_IALIC], token.Text)
			mt.children = append(mt.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{abstract: abs}
			codespan.SetText(monoFamilyFrom(mt.fonts), token.Text)
			mt.children = append(mt.children, codespan)
		case TYPE_STRONG:
			strong := &MdText{abstract: abs}
			text := token.Text
			if len(token.Tokens) > 0 && token.Tokens[0].Type == TYPE_EM {
				text = token.Tokens[0].Text
			}
			strong.SetText(mt.fonts[FONT_BOLD], text)
			mt.children = append(mt.children, strong)
		case TYPE_DEL:
			del := &MdText{abstract: abs}
			del.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.children = append(mt.children, del)
		case TYPE_TABLE:
			tb := &MdTable{abstract: abs, fonts: mt.fonts}
			tb.SetToken(token)
			mt.children = append(mt.children, tb.children...)
		case TYPE_HR:
			hr := &MdHr{abstract: abs}
			mt.children = append(mt.children, hr)

			abs2 := mt.getabstract(TYPE_SPACE)
			space := &MdSpace{abstract: abs2}
			mt.children = append(mt.children, space)
		}
	}
}

func (mt *MarkdownText) GenerateAtomicCell() (err error) {
	if len(mt.children) == 0 {
		return fmt.Errorf("not set text")
	}

	for i := 0; i < len(mt.children); {
		child := mt.children[i]

		pagebreak, over, err := child.GenerateAtomicCell()
		if err != nil {
			i++
			continue
		}

		if pagebreak {
			if over {
				i++
			}
			newX, newY := mt.pdf.GetPageStartXY()
			mt.pdf.AddNewPage(false)
			mt.pdf.SetXY(newX, newY)
			continue
		}

		if over {
			i++
		}
	}

	return nil
}
