package gopdf

// markdown_atomic.go：不参与复合子列表的叶子结点——垂直空隙、强制换行、图片。

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// Horizontal tolerance when deciding if the caret is on the same logical column as
// page edge, list hang, or blockquote body (floating PDF coordinates).
const mdSpaceLineXEpsilon = 1.5

// MdSpace 垂直留白（段间距、代码块后空隙等），可能绘制引用延续竖条。
type MdSpace struct {
	ElementBase
}

// GenerateAtomicCell 下移光标并可能绘制引用条；页底放不下时触发分页。
func (c *MdSpace) GenerateAtomicCell() (pagebreak, over bool, err error) {
	c.resetLayoutExtent()
	pageStartX, _ := c.pdf.GetPageStartXY()
	_, pageEndY := c.pdf.GetPageEndXY()
	x, y := c.pdf.GetXY()

	atStart := mdSpaceLineStart(x, pageStartX, &c.ElementBase)
	deltaY := mdSpaceVerticalDelta(&c.theme, c.lineHeight, atStart)
	spaceX := mdSpaceAnchorX(pageStartX, c.hangingIndentPt)
	spaceY := y + deltaY

	c.noteLayoutStart(spaceX, y)
	c.noteLayoutExtent(spaceX, spaceY)
	c.paintBlockquoteBars(spaceX, pageStartX, y, deltaY)

	bodyLH := c.theme.bodyLineHeight()
	if pageEndY-spaceY < bodyLH {
		return true, true, nil
	}

	c.pdf.SetXY(spaceX, spaceY)
	return false, true, nil
}

// mdSpaceLineStart mirrors legacy MdSpace “line start”: page content edge,
// list hang column, or blockquote body column (no list hang overlay).
//
// Uses a single epsilon throughout so float drift from MdText.Cell is handled
// consistently (the old mix of strict == and fuzzy checks was brittle).
func mdSpaceLineStart(x, pageStartX float64, e *ElementBase) bool {
	switch {
	case nearMdX(x, pageStartX):
		return true
	case e.hangingIndentPt > 0 && nearMdX(x, pageStartX+e.hangingIndentPt):
		return true
	case e.blockquote > 0 && e.flowColumnOffsetPt > 0 && e.hangingIndentPt == 0 &&
		nearMdX(x, pageStartX+e.flowColumnOffsetPt):
		return true
	default:
		return false
	}
}

func nearMdX(a, b float64) bool {
	return math.Abs(a-b) < mdSpaceLineXEpsilon
}

func mdSpaceAnchorX(pageStartX, hangingIndentPt float64) float64 {
	if hangingIndentPt > 0 {
		return pageStartX + hangingIndentPt
	}
	return pageStartX
}

// mdSpaceVerticalDelta computes how far down to advance for this gap node:
// explicit lineHeight (e.g. after fenced code); otherwise theme BreakGap-only
// on a fresh block line, or BreakGap + one body line mid-block.
func mdSpaceVerticalDelta(theme *MarkdownTheme, lineHeight float64, atStart bool) float64 {
	brk := theme.paraBreakGap()
	bodyLH := theme.bodyLineHeight()
	switch {
	case lineHeight > 0:
		if atStart {
			return lineHeight
		}
		return lineHeight + brk
	case atStart:
		return brk
	default:
		return brk + bodyLH
	}
}

func (c *MdSpace) paintBlockquoteBars(spaceX, pageStartX, y, deltaY float64) {
	if c.blockquote <= 0 {
		return
	}
	ext := mdLineHeight*0.72 + blockquoteBarVOverlap()*0.5
	barH := deltaY + ext
	barX := spaceX
	if c.quoteBarsLeftOffsetPt > 0 {
		barX = pageStartX + c.quoteBarsLeftOffsetPt
	}
	for i := 0; i < c.blockquote; i++ {
		c.pdf.BackgroundColor(barX+blockquoteBarOffset(i), y-ext, blockLen, barH, color_gray, "0000")
	}
}

// MdHardBreak 强制换行（Markdown 硬换行 / <br>）；indentX 用于嵌套列表首行对齐到标记列。
type MdHardBreak struct {
	ElementBase
	indentX float64
}

func (m *MdHardBreak) SetText(interface{}, ...string) {}

func (m *MdHardBreak) GetType() string {
	return TYPE_BR
}

// GenerateAtomicCell 将光标下移一行高（或触发分页）；引用路径下绘制连接竖条。
func (m *MdHardBreak) GenerateAtomicCell() (pagebreak, over bool, err error) {
	m.resetLayoutExtent()
	pageStartX, _ := m.pdf.GetPageStartXY()
	_, pageEndY := m.pdf.GetPageEndXY()
	x, y := m.pdf.GetXY()
	lh := m.lineHeight
	if lh == 0 {
		lh = m.theme.bodyLineHeight()
	}
	newY := y + lh
	m.noteLayoutStart(x, y)
	m.noteLayoutExtent(pageStartX+m.indentX, newY)

	if newY >= pageEndY || pageEndY-newY < mdScale(0.5/18.0) {
		return true, true, nil
	}
	if m.blockquote > 0 {
		ext := mdLineHeight * 0.1
		vO := blockquoteBarVOverlap()
		barH := lh + 2*ext + vO
		barTop := y - ext - vO*0.5
		barX := pageStartX
		if m.quoteBarsLeftOffsetPt > 0 {
			barX = pageStartX + m.quoteBarsLeftOffsetPt
		}
		for i := 0; i < m.blockquote; i++ {
			m.pdf.BackgroundColor(barX+blockquoteBarOffset(i), barTop, blockLen, barH, color_gray, "0000")
		}
	}
	newX := pageStartX + m.indentX
	m.pdf.SetXY(newX, newY)
	return false, true, nil
}

// MdImage 嵌入图片；本地路径或 http(s)。高度默认一行 bodyLineHeight。
type MdImage struct {
	ElementBase
	image  *Image
	height float64
}

// SetText 加载路径：支持 http(s) 下载至临时文件或本地路径；首个参数为 URL / 路径。
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
		i.height = i.theme.bodyLineHeight()
	}

	i.image = NewImageWithWidthAndHeight(filepath, 0, i.height, i.pdf)
}

// GenerateAtomicCell 委托底层 Image；加载失败时 image==nil 则跳过。
func (i *MdImage) GenerateAtomicCell() (pagebreak, over bool, err error) {
	i.resetLayoutExtent()
	if i.image == nil {
		return false, true, nil
	}
	return i.image.GenerateAtomicCell()
}

func (i *MdImage) GetType() string {
	return i.Type
}
