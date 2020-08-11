package gopdf

import (
	"math"
	"log"
	"fmt"
	"strings"
	"bytes"
	"regexp"
	"encoding/json"

	"github.com/tiechui1994/gopdf/core"
	"net/http"
	"os"
	"time"
	"io"
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
	defaultLineHeight       = 18.0
	defaultSpaceLineHeight  = 36.0
	defaultFontSize         = 15.0
	defaultCodeSpanFontSize = 10.0
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

type abstractMarkDown struct{}

func (a *abstractMarkDown) SetText(interface{}, ...string) {
}
func (a *abstractMarkDown) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return false, true, nil
}
func (a *abstractMarkDown) GetType() string {
	return ""
}

///////////////////////////////////////////////////////////////////

// Atomic component
type MdText struct {
	Type       string
	pdf        *core.Report
	font       core.Font
	padding    float64 // padding left length
	lineHeight float64 // line height

	stoped    bool    // symbol stoped
	precision float64 // sigle text char length
	length    float64
	text      string // text content
	remain    string // renain texy
	link      string // special TYPE_LINK
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
			c.font = core.Font{Family: family, Size: defaultCodeSpanFontSize, Style: ""}
		case TYPE_LINK, TYPE_TEXT:
			c.font = core.Font{Family: family, Size: defaultFontSize, Style: ""}
		}
	case core.Font:
		c.font = font.(core.Font)
	default:
		panic(fmt.Sprintf("invalid type: %v", c.Type))
	}

	text[0] = strings.Replace(text[0], "\t", "    ", -1)
	c.text = text[0]
	c.remain = text[0]
	if c.Type == TYPE_LINK {
		c.link = text[1]
	}
	c.lineHeight = defaultLineHeight
	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)
	re := regexp.MustCompile(`[\n \t=#%@&"':<>,(){}_;/\?\.\+\-\=\^\$\[\]\!]`)

	subs := re.FindAllString(c.text, -1)
	if len(subs) > 0 {
		str := re.ReplaceAllString(c.text, "")
		c.length = c.pdf.MeasureTextWidth(str)
		c.precision = c.length / float64(len([]rune(str)))
	} else {
		c.length = c.pdf.MeasureTextWidth(c.text)
		c.precision = c.length / float64(len([]rune(c.text)))
	}
}

func (c *MdText) GenerateAtomicCell() (pagebreak, over bool, err error) {
	pageEndX, pageEndY := c.pdf.GetPageEndXY()
	x1, y := c.pdf.GetXY()
	x2 := pageEndX

	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)

	text, width, newline := c.GetSubText(x1, x2)
	for !c.stoped {
		switch c.Type {
		case TYPE_CODESPAN:
			// bg: 248,248,255 text:1,1,1 line:	245,245,245,     Typora
			// bg: 249,242,244 text:199,37,78 line:	245,245,245  GitLab
			c.pdf.BackgroundColor(x1, y, width, 15.0, "249,242,244",
				"1111", "245,245,245")

			c.pdf.TextColor(199, 37, 78)
			c.pdf.Cell(x1, y+3.15, text)
			c.pdf.TextColor(1, 1, 1)
		case TYPE_CODE:
			// bg: 248,248,255, text: 1,1,1    Typora
			// bg: 40,42,54, text: 248,248,242  CSDN
			bg := "128,255,128"
			c.pdf.BackgroundColor(x1, y, x2-x1, 18.0, bg, "0000")
			c.pdf.TextColor(248, 248, 242)
			c.pdf.Cell(x1, y+3.15, text)
			c.pdf.TextColor(1, 1, 1)

		case TYPE_LINK:
			// text, blue
			c.pdf.TextColor(0, 0, 255)
			c.pdf.ExternalLink(x1, y+12.0, 15, text, c.link)
			c.pdf.TextColor(1, 1, 1)

			// line
			c.pdf.LineColor(0, 0, 255)
			c.pdf.LineType("solid", 0.4)
			c.pdf.LineH(x1, y+c.precision, x1+width)
			c.pdf.LineColor(1, 1, 1)
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

func (c *MdText) GetType() string {
	return c.Type
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
	abstractMarkDown
	pdf        *core.Report
	padding    float64
	lineHeight float64 // line height
	Type       string
}

func (c *MdSpace) GenerateAtomicCell() (pagebreak, over bool, err error) {
	_, pageEndY := c.pdf.GetPageEndXY()
	_, y := c.pdf.GetXY()

	if c.lineHeight == 0 {
		c.lineHeight = defaultSpaceLineHeight
	}

	x, _ := c.pdf.GetPageStartXY()
	y += c.lineHeight
	if pageEndY-y < c.lineHeight {
		return true, true, nil
	}
	c.pdf.SetXY(x, y)
	return false, true, nil
}

func (c *MdSpace) GetType() string {
	return c.Type
}

func (c *MdSpace) LineHeight() float64 {
	return c.lineHeight / 2
}

func (c *MdSpace) String() string {
	return fmt.Sprint("[type=space]")
}

type MdImage struct {
	abstractMarkDown
	pdf     *core.Report
	padding float64
	Type    string

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
		i.height = defaultLineHeight
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
	abstractMarkDown
	pdf      *core.Report
	fonts    map[string]string
	children []mardown
	padding  float64
	Type     string
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
		return defaultFontSize, defaultLineHeight
	case 5:
		return 12, 15
	case 6:
		return defaultCodeSpanFontSize, defaultLineHeight
	}

	return defaultFontSize, defaultLineHeight
}

func (h *MdHeader) SetToken(token Token) (err error) {
	if h.fonts == nil || len(h.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if token.Type != TYPE_HEADING {
		return fmt.Errorf("invalid type")
	}

	fontsize, lineheight := h.CalFontSizeAndLineHeight(token.Depth)
	font := core.Font{Family: h.fonts[FONT_BOLD], Size: fontsize}
	for _, v := range token.Tokens {
		switch v.Type {
		case TYPE_TEXT:
			text := &MdText{pdf: h.pdf, Type: v.Type, lineHeight: lineheight}
			text.SetText(font, v.Text)
			h.children = append(h.children, text)
		case TYPE_IMAGE:
			image := &MdImage{pdf: h.pdf, Type: v.Type, height: lineheight}
			h.children = append(h.children, image)
		}
	}

	space := &MdSpace{pdf: h.pdf, Type: TYPE_SPACE}
	h.children = append(h.children, space)

	return nil
}

func (h *MdHeader) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return CommonGenerateAtomicCell(&h.children)
}

type MdParagraph struct {
	abstractMarkDown
	pdf      *core.Report
	fonts    map[string]string
	children []mardown
	padding  float64
	Type     string
}

func (p *MdParagraph) SetToken(token Token) error {
	if p.fonts == nil || len(p.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if token.Type != TYPE_PARAGRAPH {
		return fmt.Errorf("invalid type")
	}

	for _, tok := range token.Tokens {
		var child mardown
		switch tok.Type {
		case TYPE_LINK:
			child = &MdText{pdf: p.pdf, Type: tok.Type, padding: p.padding}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text, tok.Href)
		case TYPE_TEXT:
			child = &MdText{pdf: p.pdf, Type: tok.Type, padding: p.padding}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_EM:
			child = &MdText{pdf: p.pdf, Type: tok.Type, padding: p.padding}
			child.SetText(p.fonts[FONT_IALIC], tok.Text)
		case TYPE_CODESPAN:
			child = &MdText{pdf: p.pdf, Type: tok.Type, padding: p.padding}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_CODE:
			child = &MdText{pdf: p.pdf, Type: tok.Type, padding: p.padding}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_STRONG:
			child = &MdText{pdf: p.pdf, Type: tok.Type, padding: p.padding}
			text := tok.Text
			if len(tok.Tokens) > 0 && tok.Tokens[0].Type == TYPE_EM {
				text = tok.Tokens[0].Text
			}
			child.SetText(p.fonts[FONT_BOLD], text)
		case TYPE_IMAGE:
			child = &MdImage{pdf: p.pdf, Type: tok.Type, padding: p.padding}
			child.SetText("", tok.Href)
		}

		if child == nil {
			log.Println(tok.Type)
			continue
		}

		p.children = append(p.children, child)
	}

	return nil
}

func (p *MdParagraph) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return CommonGenerateAtomicCell(&p.children)
}

type MdList struct {
	abstractMarkDown
	pdf      *core.Report
	fonts    map[string]string
	children []mardown
	padding  float64
	Type     string
}

func (l *MdList) SetToken(token Token) error {
	if l.fonts == nil || len(l.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if token.Type != TYPE_LIST {
		return fmt.Errorf("invalid type")
	}

	for index, item := range token.Items {
		for _, tok := range item.Tokens {
			// special handle "list", "space"
			switch tok.Type {
			case TYPE_LIST:
				space := &MdSpace{pdf: l.pdf, Type: TYPE_SPACE}
				l.children = append(l.children, space)

				list := &MdList{pdf: l.pdf, fonts: l.fonts, Type: TYPE_LIST, padding: l.padding + 17.7}
				list.SetToken(tok)
				l.children = append(l.children, list.children...)
				continue

			case TYPE_SPACE:
				space := &MdSpace{pdf: l.pdf, Type: tok.Type}
				l.children = append(l.children, space)
				continue

			case TYPE_BLOCKQUOTE:
				blockquote := &MdBlockQuote{pdf: l.pdf, fonts: l.fonts, Type: token.Type}
				blockquote.SetToken(token)
				l.children = append(l.children, blockquote.children...)
				continue

			case TYPE_CODE:
				code := &MdText{pdf: l.pdf, Type: tok.Type}
				code.SetText(l.fonts[FONT_NORMAL], tok.Text)
				l.children = append(l.children, code)
				continue
			}

			if token.Ordered {
				text := &MdText{pdf: l.pdf, Type: TYPE_TEXT, offsety: -0.45, padding: l.padding}
				text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: defaultFontSize}, fmt.Sprintf("%v. ", index+1))
				l.children = append(l.children, text)
			}

			if !token.Ordered {
				text := &MdText{pdf: l.pdf, Type: TYPE_TEXT, offsety: -6.3, padding: l.padding}
				text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: 28}, "· ")
				l.children = append(l.children, text)
			}

			switch tok.Type {
			case TYPE_STRONG:
				strong := &MdText{pdf: l.pdf, Type: tok.Type, padding: l.padding}
				strong.SetText(l.fonts[FONT_BOLD], tok.Text)
				l.children = append(l.children, strong)
			case TYPE_LINK:
				link := &MdText{pdf: l.pdf, Type: tok.Type, padding: l.padding}
				link.SetText(l.fonts[FONT_NORMAL], tok.Text, tok.Href)
				l.children = append(l.children, link)
			case TYPE_TEXT:
				text := &MdText{pdf: l.pdf, Type: tok.Type, padding: l.padding}
				text.SetText(l.fonts[FONT_NORMAL], tok.Text)
				l.children = append(l.children, text)
			}
		}

		space := &MdSpace{pdf: l.pdf, Type: TYPE_SPACE}
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
	abstractMarkDown
	pdf      *core.Report
	fonts    map[string]string
	children []mardown
	padding  float64
	Type     string
}

func (b *MdBlockQuote) SetToken(t Token) error {
	if b.fonts == nil || len(b.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if t.Type != TYPE_BLOCKQUOTE {
		return fmt.Errorf("invalid type")
	}

	b.padding = 17.10

	for _, token := range t.Tokens {
		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{pdf: b.pdf, fonts: b.fonts, Type: token.Type, padding: b.padding}
			paragraph.SetToken(token)
			b.children = append(b.children, paragraph.children...)
		case TYPE_LIST:
			list := &MdList{pdf: b.pdf, fonts: b.fonts, Type: token.Type, padding: b.padding}
			list.SetToken(token)
			b.children = append(b.children, list.children...)
		case TYPE_HEADING:
			header := &MdHeader{pdf: b.pdf, fonts: b.fonts, Type: token.Type, padding: b.padding}
			header.SetToken(token)
			b.children = append(b.children, header.children...)
		case TYPE_SPACE:
			space := &MdSpace{pdf: b.pdf, Type: token.Type, padding: b.padding}
			b.children = append(b.children, space)
		case TYPE_LINK:
			link := &MdText{pdf: b.pdf, Type: token.Type, padding: b.padding,}
			link.SetText(b.fonts[FONT_NORMAL], token.Text, token.Href, )
			b.children = append(b.children, link)
		case TYPE_TEXT:
			text := &MdText{pdf: b.pdf, Type: token.Type, padding: b.padding}
			text.SetText(b.fonts[FONT_NORMAL], token.Text)
			b.children = append(b.children, text)
		case TYPE_EM:
			log.Println(token.Text)
			em := &MdText{pdf: b.pdf, Type: token.Type, padding: b.padding}
			em.SetText(b.fonts[FONT_IALIC], token.Text)
			b.children = append(b.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{pdf: b.pdf, Type: token.Type, padding: b.padding}
			codespan.SetText(b.fonts[FONT_NORMAL], token.Text)
			b.children = append(b.children, codespan)
		case TYPE_CODE:
			code := &MdText{pdf: b.pdf, Type: token.Type, padding: b.padding}
			code.SetText(b.fonts[FONT_NORMAL], token.Text+"\n")
			b.children = append(b.children, code)
		case TYPE_STRONG:
			strong := &MdText{pdf: b.pdf, Type: token.Type, padding: b.padding}
			strong.SetText(b.fonts[FONT_BOLD], token.Text)
			b.children = append(b.children, strong)
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

func (mt *MarkdownText) SetTokens(tokens []Token) {
	log.Println("len", len(tokens))
	for _, token := range tokens {
		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{pdf: mt.pdf, fonts: mt.fonts, Type: token.Type}
			paragraph.SetToken(token)
			mt.children = append(mt.children, paragraph.children...)
		case TYPE_LIST:
			list := &MdList{pdf: mt.pdf, fonts: mt.fonts, Type: token.Type}
			list.SetToken(token)
			mt.children = append(mt.children, list.children...)
		case TYPE_HEADING:
			header := &MdHeader{pdf: mt.pdf, fonts: mt.fonts, Type: token.Type}
			header.SetToken(token)
			mt.children = append(mt.children, header.children...)
		case TYPE_BLOCKQUOTE:
			blockquote := &MdBlockQuote{pdf: mt.pdf, fonts: mt.fonts, Type: token.Type}
			blockquote.SetToken(token)
			mt.children = append(mt.children, blockquote.children...)
		case TYPE_SPACE:
			space := &MdSpace{pdf: mt.pdf, Type: token.Type}
			mt.children = append(mt.children, space)
		case TYPE_LINK:
			link := &MdText{pdf: mt.pdf, Type: token.Type}
			link.SetText(mt.fonts[FONT_NORMAL], token.Text, token.Href)
			mt.children = append(mt.children, link)
		case TYPE_TEXT:
			text := &MdText{pdf: mt.pdf, Type: token.Type}
			text.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.children = append(mt.children, text)
		case TYPE_EM:
			em := &MdText{pdf: mt.pdf, Type: token.Type}
			em.SetText(mt.fonts[FONT_IALIC], token.Text)
			mt.children = append(mt.children, em)
		case TYPE_CODESPAN:
			codespan := &MdText{pdf: mt.pdf, Type: token.Type}
			codespan.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.children = append(mt.children, codespan)
		case TYPE_CODE:
			code := &MdText{pdf: mt.pdf, Type: token.Type, padding: 15}
			code.SetText(mt.fonts[FONT_NORMAL], token.Text+"\n")
			mt.children = append(mt.children, code)
		case TYPE_STRONG:
			strong := &MdText{pdf: mt.pdf, Type: token.Type}
			strong.SetText(mt.fonts[FONT_BOLD], token.Text)
			mt.children = append(mt.children, strong)
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
