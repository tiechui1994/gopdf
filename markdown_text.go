package gopdf

// markdown_text.go：MdText 原子文本（及代码块行）测量、换行与绘制。

import (
	"fmt"
	"math"
	"strings"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/util"
)

// MdText 表示一段不可分割排版单元的文字（段落碎片、代码行、行内代码、链接等），内部用 remain 跨页续画。
// 正文片段行距一律用主题 bodyLineHeight()；标题片段由 MdHeader 设置 headingStepPt 按级别指定行距。
type MdText struct {
	ElementBase
	font core.Font

	// headingStepPt 仅用于标题行：>0 时作为该标题片段的行距（pt）；正文片段保持为 0。
	headingStepPt float64

	stoped    bool
	// precision：SetText 中写入的版面测量容差（pt），用于换行二分、词界比较与 NeedTextPageBreak 的右缘贴近判断；由字体探测推算，非「单字平均宽度」。
	precision float64
	text      string
	remain    string
	link      string
	newlines  int

	offsetx float64
	offsety float64
}

// SetText 设置字型与测量基准：font 可为字体族 string 或 core.Font；LINK 需 texts[1]=href。
func (c *MdText) SetText(font interface{}, texts ...string) {
	if len(texts) == 0 {
		panic("text is invalid")
	}

	fs := int(c.theme.bodyFontSize())

	switch font.(type) {
	case string:
		family := font.(string)
		switch c.Type {
		case TYPE_STRONG:
			c.font = core.Font{Family: family, Size: fs, Style: ""}
		case TYPE_EM:
			c.font = core.Font{Family: family, Size: fs, Style: ""}
		case TYPE_CODESPAN, TYPE_CODE:
			c.font = core.Font{Family: family, Size: fs, Style: ""}
		case TYPE_LINK, TYPE_TEXT:
			c.font = core.Font{Family: family, Size: fs, Style: ""}
		case TYPE_DEL:
			c.font = core.Font{Family: family, Size: fs, Style: ""}
		}
	case core.Font:
		c.font = font.(core.Font)
	default:
		panic(fmt.Sprintf("invalid type: %v", c.Type))
	}

	text := strings.Replace(texts[0], "\t", "    ", -1)
	c.text = repairText(c.Type, text)
	c.remain = c.text
	if c.Type == TYPE_LINK {
		c.link = texts[1]
	}
	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)

	c.precision = c.layoutMeasurementTolerancePt()
}

// layoutMeasurementTolerancePt 用于宽度比较（换行二分、前缀拟合）；由空格、拉丁与典型全角字形测量推导，不用「去掉标点后的平均一字宽」（对 CJK 会严重偏小）。
func (c *MdText) layoutMeasurementTolerancePt() float64 {
	const floorPt = 0.035
	sw := c.pdf.MeasureTextWidth(" ")
	mi := math.Max(c.pdf.MeasureTextWidth("m"), c.pdf.MeasureTextWidth("W"))
	id := math.Max(c.pdf.MeasureTextWidth("国"), c.pdf.MeasureTextWidth("ひ"))
	em := mi
	if id > em {
		em = id
	}
	rel := math.Max(sw*0.08, em*0.012)
	out := math.Max(floorPt, rel)
	maxTol := math.Max(sw*0.45, float64(c.font.Size)*0.06)
	if out > maxTol {
		out = maxTol
	}
	return out
}

// lineAdvancePt 行距（pt）：正文统一为主题正文行距；标题由 MdHeader 写入 headingStepPt（按级别字阶推算）。
func (c *MdText) lineAdvancePt() float64 {
	if c.headingStepPt > 0 {
		return c.headingStepPt
	}
	return c.theme.bodyLineHeight()
}

// GenerateAtomicCell 绘制本片段剩余内容直至分页或写完。
//
// 坐标约定：flowX 为「流式列」左缘（含 flowColumnOffsetPt、列表 hang）；笔画左缘为 flowX + 水平 Margin/Padding；
// 行宽右缘为 pageEndX 减去右侧 inset。引用竖条对齐使用 colLeft（flowX），避免把 CSS padding 当成栏位偏移。
func (c *MdText) GenerateAtomicCell() (pagebreak, over bool, err error) {
	c.resetLayoutExtent()
	lineheight := c.lineAdvancePt()
	lc := NewLayoutContext(c.pdf)
	pageStartX, _ := c.pdf.GetPageStartXY()
	pageEndX, pageEndY := c.pdf.GetPageEndXY()

	hPadL, hPadR := c.TextHorizontalInsets()
	pageEndXEff := pageEndX - hPadR
	pageEndYEff := pageEndY - c.TextVerticalBottomInset()

	flowX, y := c.pdf.GetXY()

	if c.hangingIndentPt > 0 {
		targetX := pageStartX + c.hangingIndentPt
		if flowX <= pageStartX+0.5 {
			flowX = targetX
			c.pdf.SetXY(flowX, y)
		} else if flowX < targetX-0.5 {
			flowX = targetX
			c.pdf.SetXY(flowX, y)
		} else if flowX >= pageEndX-5 {
			flowX = targetX
			y += lineheight
			c.pdf.SetXY(flowX, y)
		}
	}

	if c.Type == TYPE_CODE && c.hangingIndentPt == 0 && c.flowColumnOffsetPt > 0 {
		if math.Abs(flowX-pageStartX) < 1.0 {
			flowX = pageStartX + c.flowColumnOffsetPt
			c.pdf.SetXY(flowX, y)
		}
	}

	vTop := c.TextVerticalTopInset()
	if vTop > 0 && c.remain == c.text && len(c.remain) > 0 {
		y += vTop
		c.pdf.SetXY(flowX, y)
	}

	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)

	tl, tr := flowX+hPadL, pageEndXEff
	if c.Type == TYPE_CODE {
		p := codeBlockPad()
		if tr-tl > 2*p {
			tl, tr = tl+p, tr-p
		}
	}
	text, width, newline := c.GetSubText(tl, tr)

	for !c.stoped {
		x1 := flowX + hPadL
		colLeft := flowX

		c.noteLayoutStart(x1, y)

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
		atPage := math.Abs(colLeft-pageStartX) < 1.0
		atListText := c.hangingIndentPt > 0 && math.Abs(colLeft-(pageStartX+c.hangingIndentPt)) < 2.0
		atBqCode := c.Type == TYPE_CODE && c.blockquote > 0 && c.hangingIndentPt == 0 && c.flowColumnOffsetPt > 0 &&
			math.Abs(colLeft-(pageStartX+c.flowColumnOffsetPt)) < 1.5
		if c.blockquote > 0 && (atPage || atListText || atBqCode) {
			barX := colLeft
			if c.quoteBarsLeftOffsetPt > 0 {
				barX = pageStartX + c.quoteBarsLeftOffsetPt
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
			bgLeft := x1
			if c.hangingIndentPt > 0 {
				bgLeft = pageStartX + c.hangingIndentPt
			} else if c.blockquote > 0 {
				bgLeft = pageStartX + c.flowColumnOffsetPt
			} else if c.flowColumnOffsetPt > 0 {
				bgLeft = pageStartX + c.flowColumnOffsetPt
			} else if math.Abs(colLeft-pageStartX) < 0.5 {
				bgLeft = pageStartX
			}
			fullW := pageEndXEff - bgLeft
			if fullW < 1 {
				fullW = pageEndXEff - x1
			}
			c.pdf.BackgroundColor(bgLeft, bgTop, fullW, bgH, color_whitesmoke, "0000")
			c.pdf.TextColor(util.RGB(color_black))
			c.pdf.Cell(x1+codePad, y, text)
			c.pdf.TextColor(util.RGB(color_black))

		case TYPE_LINK:
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

		c.noteLayoutExtent(x1+width, y)

		if newline {
			if c.hangingIndentPt > 0 {
				flowX = pageStartX + c.hangingIndentPt
			} else if c.Type == TYPE_CODE && c.flowColumnOffsetPt > 0 {
				flowX = pageStartX + c.flowColumnOffsetPt
			} else {
				flowX, _ = c.pdf.GetPageStartXY()
			}
			y += lineheight
			c.noteLayoutExtent(flowX+hPadL, y)
		} else {
			flowX += width
		}

		lhCheck := lineheight
		cursorX := flowX + hPadL
		if lc.NeedTextPageBreak(y, pageEndYEff, lhCheck, newline, cursorX, pageEndXEff, c.precision) {
			return true, c.stoped, nil
		}

		c.pdf.SetXY(flowX+hPadL, y)
		tl, tr = flowX+hPadL, pageEndXEff
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

// GetSubText 在 [x1,x2] 可用宽度内取下一段可见文本；必要时按词或按字折断；更新 remain。
// x1/x2 已为扣除 Margin/Padding 后的内缘坐标；needpadding 分支仍按 flowColumnOffsetPt 在「逻辑行首」补空格。
func (c *MdText) GetSubText(x1, x2 float64) (text string, width float64, newline bool) {
	if len(c.remain) == 0 {
		c.stoped = true
		return "", 0, false
	}

	pageX, _ := c.pdf.GetPageStartXY()
	hL, _ := c.TextHorizontalInsets()
	needpadding := c.flowColumnOffsetPt > 0 && atMarkdownLineLeft(x1-hL, pageX, c.hangingIndentPt)
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
		width -= c.flowColumnOffsetPt
	}
	defer func() {
		if needpadding {
			space := c.pdf.MeasureTextWidth(" ")
			text = strings.Repeat(" ", int(c.flowColumnOffsetPt/space)) + text
			width += c.flowColumnOffsetPt
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

	maxFit := c.fitPrefixRuneCount(runes, width, c.precision)
	if maxFit < 1 {
		maxFit = 1
	}
	line, wline, cut := c.applyWordAwareSlice(runes, 0, maxFit, width)
	if line == "" && len(runes) > 0 {
		line = string(runes[:1])
		wline = c.pdf.MeasureTextWidth(line)
		cut = 1
	}
	if cut <= 0 && len(runes) > 0 {
		line = string(runes[:1])
		wline = c.pdf.MeasureTextWidth(line)
		cut = 1
	}
	c.remain = string(runes[cut:]) + suffix
	c.newlines++
	return line, wline, true
}

// codeFittedRuneIndex 对 TYPE_CODE：按测量宽度二分，取最长前缀 rune 数（不按词断开）。
func (c *MdText) codeFittedRuneIndex(runes []rune, avail float64) int {
	if len(runes) == 0 {
		return 0
	}
	return c.fitPrefixRuneCount(runes, avail, c.precision)
}

// fitPrefixRuneCount 返回最大 n，使 MeasureTextWidth(runes[:n]) <= avail+eps；整段可放入时返回 len(runes)。
func (c *MdText) fitPrefixRuneCount(runes []rune, avail, eps float64) int {
	if len(runes) == 0 {
		return 0
	}
	if c.pdf.MeasureTextWidth(string(runes)) <= avail+eps {
		return len(runes)
	}
	if c.pdf.MeasureTextWidth(string(runes[0:1])) > avail+eps {
		return 1
	}
	lo, hi := 1, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if c.pdf.MeasureTextWidth(string(runes[:mid])) <= avail+eps {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	if lo < 1 {
		return 1
	}
	return lo
}

// applyWordAwareSlice 在 runes[i:end] 内优先在空格处断开，避免英文单词从中劈开。
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
