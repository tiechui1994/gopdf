package gopdf

import (
	"math"
	"log"
	"fmt"
	"strings"
	"bytes"
	"regexp"
	"encoding/json"
	"net/http"
	"os"
	"time"
	"io"

	"github.com/tiechui1994/gopdf/core"
	"github.com/tiechui1994/gopdf/util"
)

const (
	FONT_NORMAL = "normal"
	FONT_BOLD   = "bold"
	FONT_IALIC  = "italic"
)

const (
	TYPE_TEXT     = "text"
	TYPE_STRONG   = "strong"   // **strong**
	TYPE_EM       = "em"       // *em*
	TYPE_CODESPAN = "codespan" // `codespan`, ```codespan```
	TYPE_CODE     = "code"     //
	TYPE_LINK     = "link"     // [xx](http://www.link)
	TYPE_IMAGE    = "image"    // ![xxx](https://www.image)

	TYPE_SPACE = "space"

	TYPE_PARAGRAPH  = "paragraph"
	TYPE_HEADING    = "heading"
	TYPE_LIST       = "list"
	TYPE_BLOCKQUOTE = "blockquote"
)

const (
	mimLineHeight   = 18.0
	mimFontSize     = 10.0
	defaultFontSize = 15.0
)

const (
	spaceLen = 4.425
	blockLen = 0.6 * spaceLen
)

const (
	color_black = "1,1,1"
	color_gray  = "128,128,128"
	color_white = "255,255,255"

	color_pink       = "199,37,78"
	color_lightgray  = "220,220,220"
	color_whitesmoke = "245,245,245"
	color_blue       = "0,0,255"
)

// Token is parse markdown result element
type Token struct {
	Type string `json:"type"`
	Raw  string `json:"raw"`
	Text string `json:"text"`

	// list
	Ordered bool            `json:"ordered"`
	Start   json.RawMessage `json:"start"`
	Loose   bool            `json:"loose"`
	Task    bool            `json:"task"`
	Items   []Token         `json:"items"`

	// heading
	Depth int `json:"depth"`

	// link
	Href  string `json:"href"`
	Title string `json:"title"`

	Tokens []Token `json:"tokens"`
}

type mardown interface {
	SetText(font interface{}, text ...string)
	GetType() string
	GenerateAtomicCell() (pagebreak, over bool, err error)
}

type abstract struct {
	pdf        *core.Report // core reporter
	padding    float64      // padding left length
	lineHeight float64      // line height
	blockquote int          // the cuurent ele is blockquote
	Type       string
}

func (a *abstract) SetText(interface{}, ...string) {
}
func (a *abstract) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return false, true, nil
}
func (a *abstract) GetType() string {
	return a.Type
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

func (c *MdText) SetText(font interface{}, text ...string) {
	if len(text) == 0 {
		panic("text is invalid")
	}

	switch font.(type) {
	case string:
		family := font.(string)
		switch c.Type {
		case TYPE_STRONG:
			c.font = core.Font{Family: family, Size: defaultFontSize, Style: ""}
		case TYPE_EM:
			c.font = core.Font{Family: family, Size: defaultFontSize, Style: "U"}
		case TYPE_CODESPAN, TYPE_CODE:
			c.font = core.Font{Family: family, Size: mimFontSize, Style: ""}
		case TYPE_LINK, TYPE_TEXT:
			c.font = core.Font{Family: family, Size: defaultFontSize, Style: ""}
		}
	case core.Font:
		c.font = font.(core.Font)
	default:
		panic(fmt.Sprintf("invalid type: %v", c.Type))
	}

	if c.lineHeight == 0 {
		switch c.Type {
		case TYPE_CODE, TYPE_CODESPAN:
			c.lineHeight = 15.0
		case TYPE_TEXT, TYPE_LINK, TYPE_STRONG, TYPE_EM:
			c.lineHeight = 18.0
		}
	}

	text[0] = strings.Replace(text[0], "\t", "    ", -1)
	c.text = text[0]
	c.remain = text[0]
	if c.Type == TYPE_LINK {
		c.link = text[1]
	}
	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)
	re := regexp.MustCompile(`[\n \t=#%@&"':<>,(){}_;/\?\.\+\-\=\^\$\[\]\!]`)

	subs := re.FindAllString(c.text, -1)
	if len(subs) > 0 {
		str := re.ReplaceAllString(c.text, "")
		length := c.pdf.MeasureTextWidth(str)
		c.precision = length / float64(len([]rune(str)))
	} else {
		length := c.pdf.MeasureTextWidth(c.text)
		c.precision = length / float64(len([]rune(c.text)))
	}
}

func (c *MdText) GenerateAtomicCell() (pagebreak, over bool, err error) {
	pageStartX, _ := c.pdf.GetPageStartXY()
	pageEndX, pageEndY := c.pdf.GetPageEndXY()
	x1, y := c.pdf.GetXY()
	x2 := pageEndX

	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)

	text, width, newline := c.GetSubText(x1, x2)
	for !c.stoped {
		if c.padding > 0 && x1 == pageStartX && c.blockquote > 0 {
			for i := 0; i < c.blockquote; i++ {
				c.pdf.BackgroundColor(x1+float64(i*4)*spaceLen, y, blockLen, c.lineHeight, color_gray, "0000")
			}
		}

		switch c.Type {
		case TYPE_CODESPAN:
			c.pdf.BackgroundColor(x1, y, width, c.lineHeight, color_lightgray, "1111", color_whitesmoke)
			c.pdf.TextColor(util.RGB(color_pink))
			c.pdf.Cell(x1, y+3.15, text)
			c.pdf.TextColor(util.RGB(color_black))
		case TYPE_CODE:
			if c.blockquote > 0 {
				offsetx := float64(c.blockquote-1)*4*spaceLen + (4*spaceLen - blockLen)
				c.pdf.BackgroundColor(x1+offsetx, y, x2-x1-offsetx, c.lineHeight, color_whitesmoke, "0000")
			} else {
				c.pdf.BackgroundColor(x1, y, x2-x1, c.lineHeight, color_whitesmoke, "0000")
			}
			c.pdf.TextColor(util.RGB(color_black))
			c.pdf.Cell(x1, y+3.15, text)
			c.pdf.TextColor(util.RGB(color_black))

		case TYPE_LINK:
			// text
			c.pdf.TextColor(util.RGB(color_blue))
			c.pdf.ExternalLink(x1, y+12.0, 15, text, c.link)
			c.pdf.TextColor(util.RGB(color_black))
		default:
			c.pdf.Cell(x1+c.offsetx, y+c.offsety, text)
		}

		if newline {
			x1, _ = c.pdf.GetPageStartXY()
			y += c.lineHeight
		} else {
			x1 += width
		}

		// need new page, x,y must statisfy condition
		if (y >= pageEndY || pageEndY-y < c.lineHeight) && (newline || math.Abs(x1-pageEndX) < c.precision) {
			return true, c.stoped, nil
		}

		c.pdf.SetXY(x1, y)
		text, width, newline = c.GetSubText(x1, x2)
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
	needpadding := c.padding > 0 && pageX == x1
	remainText := c.remain
	index := strings.Index(c.remain, "\n")
	if index != -1 {
		newline = true
		remainText = c.remain[:index]
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
	step := int(float64(len(runes)) * width / length)
	for i, j := 0, step; i < len(runes) && j < len(runes); {
		w := c.pdf.MeasureTextWidth(string(runes[i:j]))

		// less than precision
		if math.Abs(w-width) < c.precision {
			// real with more than page width
			if w-width > 0 {
				w = c.pdf.MeasureTextWidth(string(runes[i:j-1]))
				c.remain = strings.TrimPrefix(c.remain, string(runes[i:j-1]))
				// reset
				c.newlines ++
				return string(runes[i:j-1]), w, true
			}

			// try again, can more precise
			if j+1 < len(runes) {
				w1 := c.pdf.MeasureTextWidth(string(runes[i:j+1]))
				if math.Abs(w1-width) < c.precision {
					j = j + 1
					continue
				}
			}

			c.remain = strings.TrimPrefix(c.remain, string(runes[i:j]))
			// reset
			c.newlines ++
			return string(runes[i:j]), w, true
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
	var blockquotey float64

	offsety := c.lineHeight
	pageEndX, pageEndY := c.pdf.GetPageEndXY()
	x, y := c.pdf.GetXY()
	if x <= pageEndX {
		if c.lineHeight == 0 {
			c.lineHeight = mimLineHeight
			offsety = 2 * c.lineHeight
			blockquotey = y + mimLineHeight
		} else {
			offsety = 2 * c.lineHeight
			blockquotey = y + c.lineHeight
		}
	}

	if c.lineHeight == 0 {
		c.lineHeight = mimLineHeight
		offsety = c.lineHeight
		blockquotey = y
	}

	x, _ = c.pdf.GetPageStartXY()
	y += offsety

	if c.blockquote > 0 {
		for i := 0; i < c.blockquote; i++ {
			c.pdf.BackgroundColor(x+float64(i*4)*spaceLen, blockquotey, blockLen, c.lineHeight, color_gray, "0000")
		}
	}

	if pageEndY-y < mimLineHeight {
		return true, true, nil
	}

	c.pdf.SetXY(x, y)
	return false, true, nil
}

func (c *MdSpace) LineHeight() float64 {
	return c.lineHeight / 2
}

func (c *MdSpace) String() string {
	return fmt.Sprint("[type=space]")
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
		i.height = mimLineHeight
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

type MdHeader struct {
	abstract
	fonts    map[string]string
	children []mardown
}

func (h *MdHeader) CalFontSizeAndLineHeight(size int) (fontsize int, lineheight float64) {
	switch size {
	case 1:
		return 20, 24
	case 2:
		return 18, 21
	case 3:
		return 16, 19
	case 4:
		return defaultFontSize, mimLineHeight
	case 5:
		return 12, 15
	case 6:
		return mimFontSize, mimLineHeight
	}

	return defaultFontSize, mimLineHeight
}

func (h *MdHeader) getabstract(typ string) abstract {
	return abstract{
		pdf:        h.pdf,
		padding:    h.padding,
		blockquote: h.blockquote,
		Type:       typ,
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

	abs := h.getabstract(TYPE_SPACE)
	abs.lineHeight = lineheight
	space := &MdSpace{abstract: abs}
	h.children = append(h.children, space)

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
		pdf:        p.pdf,
		padding:    p.padding,
		blockquote: p.blockquote,
		Type:       typ,
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
			codespan.SetText(p.fonts[FONT_NORMAL], token.Text)
			p.children = append(p.children, codespan)
		case TYPE_CODE:
			code := &MdText{abstract: abs}
			code.SetText(p.fonts[FONT_NORMAL], token.Text)
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
		}
	}

	return nil
}

func (p *MdParagraph) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return CommonGenerateAtomicCell(&p.children)
}

type MdList struct {
	abstract
	fonts    map[string]string
	children []mardown
}

func (l *MdList) getabstract(typ string) abstract {
	return abstract{
		pdf:        l.pdf,
		padding:    l.padding,
		blockquote: l.blockquote,
		Type:       typ,
	}
}

func (l *MdList) SetToken(t Token) error {
	if l.fonts == nil || len(l.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_LIST {
		return fmt.Errorf("invalid type")
	}

	for index, item := range t.Items {
		for _, token := range item.Tokens {
			abs := l.getabstract(token.Type)
			// special handle "list", "space"
			switch token.Type {
			case TYPE_LIST:
				space := &MdSpace{abstract: l.getabstract(TYPE_SPACE)}
				l.children = append(l.children, space)

				abs.padding += 4 * spaceLen
				list := &MdList{abstract: abs, fonts: l.fonts}
				list.SetToken(token)
				l.children = append(l.children, list.children...)
				continue

			case TYPE_SPACE:
				space := &MdSpace{abstract: abs}
				l.children = append(l.children, space)
				continue

			case TYPE_BLOCKQUOTE:
				abs.blockquote += 1
				abs.padding += 4 * spaceLen
				blockquote := &MdBlockQuote{abstract: abs, fonts: l.fonts}
				blockquote.SetToken(token)
				l.children = append(l.children, blockquote.children...)
				continue

			case TYPE_CODE:
				code := &MdText{abstract: abs}
				code.SetText(l.fonts[FONT_NORMAL], token.Text)
				l.children = append(l.children, code)

				space := &MdSpace{abstract: l.getabstract(TYPE_SPACE)}
				l.children = append(l.children, space)
				continue
			}

			if token.Ordered {
				text := &MdText{abstract: abs, offsety: -0.45}
				text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: defaultFontSize}, fmt.Sprintf("%v. ", index+1))
				l.children = append(l.children, text)
			}

			if !token.Ordered {
				text := &MdText{abstract: abs, offsety: -6.3}
				text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: 28}, "· ")
				l.children = append(l.children, text)
			}

			switch token.Type {
			case TYPE_STRONG:
				strong := &MdText{abstract: abs}
				strong.SetText(l.fonts[FONT_BOLD], token.Text)
				l.children = append(l.children, strong)
			case TYPE_LINK:
				link := &MdText{abstract: abs}
				link.SetText(l.fonts[FONT_NORMAL], token.Text, token.Href)
				l.children = append(l.children, link)
			case TYPE_TEXT:
				text := &MdText{abstract: abs}
				text.SetText(l.fonts[FONT_NORMAL], token.Text)
				l.children = append(l.children, text)
			}
		}

		space := &MdSpace{abstract: l.getabstract(TYPE_SPACE)}
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
		pdf:        b.pdf,
		padding:    b.padding,
		blockquote: b.blockquote,
		Type:       typ,
	}
}

func (b *MdBlockQuote) SetToken(t Token) error {
	if b.fonts == nil || len(b.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_BLOCKQUOTE {
		return fmt.Errorf("invalid type")
	}

	for i, token := range t.Tokens {
		abs := b.getabstract(token.Type)
		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{abstract: abs, fonts: b.fonts,}
			paragraph.SetToken(token)
			b.children = append(b.children, paragraph.children...)
		case TYPE_LIST:
			list := &MdList{abstract: abs, fonts: b.fonts}
			list.SetToken(token)
			b.children = append(b.children, list.children...)
		case TYPE_HEADING:
			header := &MdHeader{abstract: abs, fonts: b.fonts,}
			header.SetToken(token)
			b.children = append(b.children, header.children...)
		case TYPE_BLOCKQUOTE:
			abs.blockquote += 1
			abs.padding += 4 * spaceLen
			blockquote := &MdBlockQuote{abstract: abs, fonts: b.fonts}
			blockquote.SetToken(token)
			b.children = append(b.children, blockquote.children...)
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
		case TYPE_TEXT:
			text := &MdText{abstract: abs}
			text.SetText(b.fonts[FONT_NORMAL], token.Text)
			b.children = append(b.children, text)
		case TYPE_EM:
			em := &MdText{abstract: abs}
			em.SetText(b.fonts[FONT_IALIC], token.Text)
			b.children = append(b.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{abstract: abs}
			codespan.SetText(b.fonts[FONT_NORMAL], token.Text)
			b.children = append(b.children, codespan)
		case TYPE_CODE:
			code := &MdText{abstract: abs}
			code.SetText(b.fonts[FONT_NORMAL], token.Text)
			b.children = append(b.children, code)

			// code new line
			abs := b.getabstract(TYPE_SPACE)
			if i == len(t.Tokens)-1 {
				abs.blockquote -= 1
			}
			space := &MdSpace{abstract: abs}
			b.children = append(b.children, space)
		case TYPE_STRONG:
			strong := &MdText{abstract: abs}
			strong.SetText(b.fonts[FONT_BOLD], token.Text)
			b.children = append(b.children, strong)
		}
	}

	if len(b.children) > 0 {
		if b.children[len(b.children)-1].GetType() != TYPE_SPACE {
			abs := b.getabstract(TYPE_SPACE)
			abs.blockquote = 0
			space := &MdSpace{abstract: abs}
			b.children = append(b.children, space)
		}
	}

	return nil
}

func (b *MdBlockQuote) GenerateAtomicCell() (pagebreak, over bool, err error) {
	var ignore bool
	if len(b.children) > 0 && b.children[len(b.children)-1].GetType() == TYPE_SPACE {
		ignore = true
	}

	var (
		x      float64
		y1, y2 float64
	)

	x, _ = b.pdf.GetPageStartXY()
	_, y1 = b.pdf.GetXY()
	color := "175,238,238"
	for i, comment := range b.children {
		pagebreak, over, err = comment.GenerateAtomicCell()
		if err != nil {
			return
		}

		if over && i == len(b.children)-1 && ignore {
			_, y2 = b.pdf.GetXY()
			ignoreHeight := comment.(*MdSpace).LineHeight()
			y2 -= ignoreHeight + 5
		}

		if pagebreak {
			_, y2 = b.pdf.GetXY()
			b.pdf.BackgroundColor(x, y1, 5.0, y2-y1, color,
				"0000", color)
			if over && i != len(b.children)-1 {
				b.children = b.children[i+1:]
				return pagebreak, len(b.children) == 0, nil
			}

			b.children = b.children[i:]
			return pagebreak, len(b.children) == 0, nil
		}
	}

	if y2 == 0 {
		_, y2 = b.pdf.GetXY()
	}

	b.pdf.BackgroundColor(x, y1, 5.0, y2-y1, color,
		"0000", color)

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
	for _, token := range tokens {
		abs := mt.getabstract(token.Type)
		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{abstract: abs, fonts: mt.fonts}
			paragraph.SetToken(token)
			mt.children = append(mt.children, paragraph.children...)
		case TYPE_LIST:
			list := &MdList{abstract: abs, fonts: mt.fonts}
			list.SetToken(token)
			mt.children = append(mt.children, list.children...)
		case TYPE_HEADING:
			header := &MdHeader{abstract: abs, fonts: mt.fonts}
			header.SetToken(token)
			mt.children = append(mt.children, header.children...)
		case TYPE_BLOCKQUOTE:
			abs.blockquote = 1
			abs.padding += 4 * spaceLen
			blockquote := &MdBlockQuote{abstract: abs, fonts: mt.fonts}
			blockquote.SetToken(token)
			mt.children = append(mt.children, blockquote.children...)
		case TYPE_SPACE:
			space := &MdSpace{abstract: abs}
			mt.children = append(mt.children, space)
		case TYPE_LINK:
			link := &MdText{abstract: abs}
			link.SetText(mt.fonts[FONT_NORMAL], token.Text, token.Href)
			mt.children = append(mt.children, link)
		case TYPE_TEXT:
			text := &MdText{abstract: abs}
			text.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.children = append(mt.children, text)
		case TYPE_EM:
			em := &MdText{abstract: abs}
			em.SetText(mt.fonts[FONT_IALIC], token.Text)
			mt.children = append(mt.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{abstract: abs}
			codespan.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.children = append(mt.children, codespan)
		case TYPE_CODE:
			abs.padding = 15.0
			code := &MdText{abstract: abs}
			code.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.children = append(mt.children, code)

			abs := mt.getabstract(TYPE_SPACE)
			space := &MdSpace{abstract: abs}
			mt.children = append(mt.children, space)
		case TYPE_STRONG:
			strong := &MdText{abstract: abs}
			strong.SetText(mt.fonts[FONT_BOLD], token.Text)
			mt.children = append(mt.children, strong)
		}
	}
}

func (mt *MarkdownText) GenerateAtomicCell() (err error) {
	log.Println("children", len(mt.children))
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
