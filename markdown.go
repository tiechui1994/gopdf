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

	TYPE_PARAGRAPH = "paragraph"
	TYPE_HEADING   = "heading"
	TYPE_LIST      = "list"
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
	GenerateAtomicCell() (pagebreak, over bool, err error)
}

type abstractMarkDown struct{}

func (a *abstractMarkDown) SetText(interface{}, ...string) {
}
func (a *abstractMarkDown) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return false, true, nil
}

///////////////////////////////////////////////////////////////////

// Atomic component
type MdText struct {
	Type       string
	pdf        *core.Report
	font       core.Font
	lineHeight float64
	
	stoped    bool    // symbol stoped
	precision float64 // sigle text char length
	length    float64
	text      string // text content
	remain    string // renain texy
	link      string // special TYPE_LINK
	newlines  int

	needpadding bool    // need pading
	pading      float64 // padding left length
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
			c.pdf.BackgroundColor(x1, y, x2-x1, 18.0, "40,42,54",
				"1111", "40,42,54")
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
			c.pdf.Cell(x1, y-0.45, text)
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

	needpadding := c.needpadding
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
		width -= c.pading
	}
	defer func() {
		if needpadding {
			text = strings.Repeat(" ", int(c.pading/c.precision)) + text
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
	Type       string
	pdf        *core.Report
	lineHeight float64
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

type MdImage struct {
	abstractMarkDown
	pdf    *core.Report
	image  *Image
	Type   string
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
	err = i.image.GenerateAtomicCell()
	return
}

// Combination components

type MdHeader struct {
	abstractMarkDown
	pdf      *core.Report
	fonts    map[string]string
	children []mardown
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
	for i, comment := range h.children {
		pagebreak, over, err = comment.GenerateAtomicCell()
		if err != nil {
			return
		}

		if pagebreak {
			if over && i != len(h.children)-1 {
				h.children = h.children[i+1:]
				return pagebreak, len(h.children) == 0, nil
			}

			h.children = h.children[i:]
			return pagebreak, len(h.children) == 0, nil
		}
	}
	return false, true, nil
}

type MdParagraph struct {
	abstractMarkDown
	pdf      *core.Report
	fonts    map[string]string
	children []mardown
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
			child = &MdText{pdf: p.pdf, Type: TYPE_LINK}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text, tok.Href)
		case TYPE_TEXT:
			child = &MdText{pdf: p.pdf, Type: TYPE_TEXT}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_EM:
			child = &MdText{pdf: p.pdf, Type: TYPE_EM}
			child.SetText(p.fonts[FONT_IALIC], tok.Text)
		case TYPE_CODESPAN:
			child = &MdText{pdf: p.pdf, Type: TYPE_CODESPAN}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_CODE:
			child = &MdText{pdf: p.pdf, Type: TYPE_CODE}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_STRONG:
			child = &MdText{pdf: p.pdf, Type: TYPE_STRONG}
			text := tok.Text
			if len(tok.Tokens) > 0 && tok.Tokens[0].Type == TYPE_EM {
				text = tok.Tokens[0].Text
			}
			child.SetText(p.fonts[FONT_BOLD], text)
		case TYPE_IMAGE:
			child = &MdImage{pdf: p.pdf, Type: TYPE_IMAGE}
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
	for i, comment := range p.children {
		pagebreak, over, err = comment.GenerateAtomicCell()
		if err != nil {
			return
		}

		if pagebreak {
			if over && i != len(p.children)-1 {
				p.children = p.children[i+1:]
				return pagebreak, len(p.children) == 0, nil
			}

			p.children = p.children[i:]
			return pagebreak, len(p.children) == 0, nil
		}
	}
	return
}

type MdList struct {
	abstractMarkDown
	pdf      *core.Report
	fonts    map[string]string
	children []mardown
	Type     string

	headerWrited bool
	Ordered      bool
}

func (l *MdList) SetToken(token Token) error {
	if l.fonts == nil || len(l.fonts) == 0 {
		return fmt.Errorf("no fonts")
	}
	if token.Type != TYPE_LIST {
		return fmt.Errorf("invalid type")
	}
	l.Ordered = token.Ordered
	if !token.Ordered {
		for _, item := range token.Items {
			text := &MdText{pdf: l.pdf, Type: TYPE_TEXT}
			text.SetText(core.Font{Family: l.fonts[FONT_NORMAL], Size: 28}, " · ")
			l.children = append(l.children, text)
			for _, item_token := range item.Tokens {
				for _, tok := range item_token.Tokens {
					switch tok.Type {
					case TYPE_STRONG:
						strong := &MdText{pdf: l.pdf, Type: tok.Type}
						strong.SetText(l.fonts[FONT_BOLD], tok.Text)
						l.children = append(l.children, strong)
					case TYPE_LINK:
						link := &MdText{pdf: l.pdf, Type: tok.Type}
						link.SetText(l.fonts[FONT_NORMAL], tok.Text, tok.Href)
						l.children = append(l.children, link)
					case TYPE_TEXT:
						text := &MdText{pdf: l.pdf, Type: tok.Type}
						text.SetText(l.fonts[FONT_NORMAL], tok.Text)
						l.children = append(l.children, text)
					case TYPE_SPACE:
						space := &MdSpace{pdf: l.pdf, Type: tok.Type}
						l.children = append(l.children, space)
					}
				}
			}
			space := &MdSpace{pdf: l.pdf, Type: TYPE_SPACE}
			l.children = append(l.children, space)
		}
	}

	return nil
}

func (l *MdList) GenerateAtomicCell() (pagebreak, over bool, err error) {
	if l.Ordered {
		return false, true, nil
	}
	for i, comment := range l.children {
		pagebreak, over, err = comment.GenerateAtomicCell()
		if err != nil {
			return
		}

		if pagebreak {
			if over && i != len(l.children)-1 {
				l.children = l.children[i+1:]
				return pagebreak, len(l.children) == 0, nil
			}

			l.children = l.children[i:]
			return pagebreak, len(l.children) == 0, nil
		}
	}

	return
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

type MarkdownText struct {
	quote       bool
	pdf         *core.Report
	fonts       map[string]string
	contents    []mardown
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
			mt.contents = append(mt.contents, paragraph)
		case TYPE_LIST:
			list := &MdList{pdf: mt.pdf, fonts: mt.fonts, Type: token.Type}
			list.SetToken(token)
			mt.contents = append(mt.contents, list)
		case TYPE_HEADING:
			header := &MdHeader{pdf: mt.pdf, fonts: mt.fonts, Type: token.Type}
			header.SetToken(token)
			mt.contents = append(mt.contents, header)
		case TYPE_SPACE:
			space := &MdSpace{pdf: mt.pdf, Type: token.Type}
			mt.contents = append(mt.contents, space)
		case TYPE_LINK:
			link := &MdText{pdf: mt.pdf, Type: token.Type}
			link.SetText(mt.fonts[FONT_NORMAL], token.Text, token.Href)
			mt.contents = append(mt.contents, link)
		case TYPE_TEXT:
			text := &MdText{pdf: mt.pdf, Type: token.Type}
			text.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.contents = append(mt.contents, text)
		case TYPE_EM:
			em := &MdText{pdf: mt.pdf, Type: token.Type}
			em.SetText(mt.fonts[FONT_IALIC], token.Text)
			mt.contents = append(mt.contents, em)
		case TYPE_CODESPAN:
			codespan := &MdText{pdf: mt.pdf, Type: token.Type}
			codespan.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.contents = append(mt.contents, codespan)
		case TYPE_CODE:
			code := &MdText{pdf: mt.pdf, Type: token.Type, needpadding: true, pading: 15}
			code.SetText(mt.fonts[FONT_NORMAL], token.Text+"\n")
			mt.contents = append(mt.contents, code)
		case TYPE_STRONG:
			strong := &MdText{pdf: mt.pdf, Type: token.Type}
			strong.SetText(mt.fonts[FONT_BOLD], token.Text)
			mt.contents = append(mt.contents, strong)

		}
	}
}

func (mt *MarkdownText) GenerateAtomicCell() (err error) {
	if len(mt.contents) == 0 {
		return fmt.Errorf("not set text")
	}

	for i := 0; i < len(mt.contents); {
		content := mt.contents[i]
		pagebreak, over, err := content.GenerateAtomicCell()
		if err != nil {
			i++
			continue
		}

		//log.Println(pagebreak, over, content)

		if pagebreak {
			log.Println("break", over, content)
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
