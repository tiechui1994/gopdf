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

// MdSpace 垂直留白（段间距、代码块后空隙等），可能绘制引用延续竖条。
type MdSpace struct {
	ElementBase
}

// GenerateAtomicCell 下移光标并可能绘制引用条；页底放不下时触发分页。
func (c *MdSpace) GenerateAtomicCell() (pagebreak, over bool, err error) {
	c.resetLayoutExtent()
	var (
		spaceX, spaceY float64
		linehieght     = c.lineHeight
	)
	brk := c.theme.paraBreakGap()
	bodyLH := c.theme.bodyLineHeight()

	pageStartX, _ := c.pdf.GetPageStartXY()
	_, pageEndY := c.pdf.GetPageEndXY()
	x, y := c.pdf.GetXY()

	atBqBody := c.blockquote > 0 && c.listHangIndent == 0 && c.FlowInset > 0 &&
		math.Abs(x-(pageStartX+c.FlowInset)) < 1.5
	atLineStart := x == pageStartX || (c.listHangIndent > 0 && math.Abs(x-(pageStartX+c.listHangIndent)) < 1.5) || atBqBody

	if c.lineHeight > 0 {
		if atLineStart {
			linehieght = c.lineHeight
		} else {
			linehieght = c.lineHeight + brk
		}
	} else if atLineStart {
		linehieght = brk
	} else if linehieght == 0 {
		linehieght = brk + bodyLH
	} else {
		linehieght += brk
	}

	spaceX = pageStartX
	if c.listHangIndent > 0 {
		spaceX = pageStartX + c.listHangIndent
	}
	spaceY = y + linehieght

	c.noteLayoutStart(spaceX, y)
	c.noteLayoutExtent(spaceX, spaceY)

	if c.blockquote > 0 {
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

	if pageEndY-spaceY < bodyLH {
		return true, true, nil
	}

	c.pdf.SetXY(spaceX, spaceY)
	return false, true, nil
}

func (c *MdSpace) String() string {
	return fmt.Sprint("[type=space]")
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
