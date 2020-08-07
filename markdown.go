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
	TYPE_LINK     = "link"     // [xx](http://ww)

	TYPE_SPACE = "space"

	TYPE_PARAGRAPH = "paragraph"
	TYPE_HEADING   = "heading"
	TYPE_LIST      = "list"
)

const (
	defaultLineHeight       = 18.0
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
	SetText(fontFamily string, text ...string)
	GenerateAtomicCell() (pagebreak, over bool, err error)
}

type abstractMarkDown struct{}

func (a *abstractMarkDown) SetText(string, ...string) {
}
func (a *abstractMarkDown) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return
}

///////////////////////////////////////////////////////////////////

// Atomic component
type MdText struct {
	Type       string
	pdf        *core.Report
	font       core.Font
	lineHeight float64

	stoped    bool
	precision float64
	length    float64
	text      string
	remain    string
	link      string // special TYPE_LINK
	newlines  int

	// when type is code can use
	needpadding bool
	padinglen   float64
}

func (c *MdText) SetText(fontFamily string, text ...string) {
	if len(text) == 0 {
		panic("text is invalid")
	}

	var font core.Font
	switch c.Type {
	case TYPE_STRONG:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	case TYPE_EM:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	case TYPE_CODESPAN, TYPE_CODE:
		font = core.Font{Family: fontFamily, Size: defaultCodeSpanFontSize, Style: ""}
	case TYPE_LINK, TYPE_TEXT:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	default:
		panic(fmt.Sprintf("invalid type: %v", c.Type))
	}

	c.font = font
	text[0] = strings.Replace(text[0], "\t", "    ", -1)
	c.text = text[0]
	c.remain = text[0]
	if c.Type == TYPE_LINK {
		c.link = text[1]
	}
	c.lineHeight = defaultLineHeight
	c.pdf.Font(font.Family, font.Size, font.Style)
	c.pdf.SetFontWithStyle(font.Family, font.Style, font.Size)
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
		if c.Type == TYPE_CODESPAN {
			// other padding is testing data
			// bg: 248,248,255 text:1,1,1 line:	245,245,245,     Typora
			// bg: 249,242,244 text:199,37,78 line:	245,245,245  GitLab
			c.pdf.BackgroundColor(x1, y, width, 15.0, "249,242,244",
				"1111", "245,245,245")

			c.pdf.TextColor(199, 37, 78)
			c.pdf.Cell(x1, y+3.15, text)
			c.pdf.TextColor(1, 1, 1)
		} else if c.Type == TYPE_CODE {
			// bg: 248,248,255, text: 1,1,1    Typora
			// bg: 40,42,54, text: 248,248,242  CSDN
			c.pdf.BackgroundColor(x1, y, x2-x1, 18.0, "40,42,54",
				"1111", "40,42,54")
			c.pdf.TextColor(248, 248, 242)
			c.pdf.Cell(x1, y+3.15, text)
			c.pdf.TextColor(1, 1, 1)
		} else if c.Type == TYPE_LINK {
			// text, blue
			c.pdf.TextColor(0, 0, 255)
			c.pdf.ExternalLink(x1, y+12.0, 15, text, c.link)
			c.pdf.TextColor(1, 1, 1)

			// line
			c.pdf.LineColor(0, 0, 255)
			c.pdf.LineType("solid", 0.4)
			c.pdf.LineH(x1, y+c.precision, x1+width)
			c.pdf.LineColor(1, 1, 1)
		} else {
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

	needpadding := c.Type == TYPE_CODE && c.needpadding
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
		width -= c.pdf.MeasureTextWidth("  ")
	}
	defer func() {
		if needpadding {
			text = "  " + text
		}
	}()

	if length <= width {
		if newline {
			c.remain = c.remain[index+1:]
			c.needpadding = c.Type == TYPE_CODE
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
	pdf   *core.Report
	image *Image
}

func (mi *MdImage) SetText(fontFamily string, filename ...string) {
	image := NewImage(filename[0], mi.pdf)
	mi.image = image
}

func (mi *MdImage) GenerateAtomicCell() (pagebreak, over bool, err error) {
	mi.image.GenerateAtomicCell()
	return
}

// Combination components

type MdHeader struct {
	Size       int
	pdf        *core.Report
	fonts      map[string]string
	children   []mardown
	lineHeight float64

	text          string
	remain        string
	needBreakLine bool
}

func (h *MdHeader) GetFontFamily(md mardown) string {
	switch md.(type) {
	case *MdText:
		c, _ := md.(*MdText)
		switch c.Type {
		case TYPE_TEXT:
			return h.fonts[FONT_NORMAL]
		case TYPE_STRONG:
			return h.fonts[FONT_BOLD]
		case TYPE_EM:
			return h.fonts[FONT_IALIC]
		default:
			return h.fonts[FONT_NORMAL]
		}
	}

	return ""
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
	return
}

func (h *MdHeader) String() string {
	text := strings.Replace(h.remain, "\n", "|", -1)
	return fmt.Sprintf("[size=%v,text=%v]", h.Size, text)
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
			child = &MdText{
				pdf:  p.pdf,
				Type: TYPE_LINK,
			}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text, tok.Href)
		case TYPE_TEXT:
			child = &MdText{
				pdf:  p.pdf,
				Type: TYPE_TEXT,
			}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_EM:
			child = &MdText{
				pdf:  p.pdf,
				Type: TYPE_EM,
			}
			child.SetText(p.fonts[FONT_IALIC], tok.Text)
		case TYPE_CODESPAN:
			child = &MdText{
				pdf:  p.pdf,
				Type: TYPE_CODESPAN,
			}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_CODE:
			child = &MdText{
				pdf:  p.pdf,
				Type: TYPE_CODE,
			}
			child.SetText(p.fonts[FONT_NORMAL], tok.Text)
		case TYPE_STRONG:
			child = &MdText{
				pdf:  p.pdf,
				Type: TYPE_STRONG,
			}
			child.SetText(p.fonts[FONT_BOLD], tok.Text)
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

///////////////////////////////////////////////////////

const (
	SORT_ORDER    = "order"
	SORT_DISORDER = "disorder"
)

type blocksort struct {
	pdf      *core.Report
	font     core.Font
	children []mardown

	headerWrited bool
	sortType     string
	sortIndex    string
}

func (bs *blocksort) SetText(fontFamily string, _ ...string) {
	if bs.sortType == SORT_DISORDER {
		bs.font = core.Font{Family: fontFamily, Size: 40.0}
	} else {
		bs.font = core.Font{Family: fontFamily, Size: 18.0}
	}

}

func (bs *blocksort) GenerateAtomicCell() (pagebreak, over bool, err error) {
	bs.pdf.Font(bs.font.Family, bs.font.Size, bs.font.Style)
	bs.pdf.SetFontWithStyle(bs.font.Family, bs.font.Style, bs.font.Size)

	if !bs.headerWrited {
		var text string
		x, y := bs.pdf.GetXY()
		switch bs.sortType {
		case SORT_ORDER:
			text = fmt.Sprintf(" %v. ", bs.sortIndex)
			bs.pdf.Cell(x, y, text)

		case SORT_DISORDER:
			text = " · "
			bs.pdf.Cell(x, y-13, text)
		}

		length := bs.pdf.MeasureTextWidth(text)
		bs.pdf.SetXY(x+length, y)
		bs.headerWrited = true
	}

	for i, comment := range bs.children {
		pagebreak, over, err = comment.GenerateAtomicCell()
		if err != nil {
			return
		}

		if pagebreak {
			if over && i != len(bs.children)-1 {
				bs.children = bs.children[i+1:]
				return pagebreak, len(bs.children) == 0, nil
			}

			bs.children = bs.children[i:]
			return pagebreak, len(bs.children) == 0, nil
		}
	}

	return
}

func (bs *blocksort) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "(blocksort")
	for _, child := range bs.children {
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
	for _, token := range tokens {
		switch token.Type {
		case TYPE_PARAGRAPH:
			paragraph := &MdParagraph{
				pdf:   mt.pdf,
				fonts: mt.fonts,
				Type:  token.Type,
			}
			paragraph.SetToken(token)
			mt.contents = append(mt.contents, paragraph)
		case TYPE_SPACE:
			space := &MdSpace{pdf: mt.pdf, Type: token.Type, lineHeight: defaultLineHeight}
			mt.contents = append(mt.contents, space)
		case TYPE_LINK:
			link := &MdText{
				pdf:  mt.pdf,
				Type: TYPE_LINK,
			}
			link.SetText(mt.fonts[FONT_NORMAL], token.Text, token.Href)
			mt.contents = append(mt.contents, link)
		case TYPE_TEXT:
			text := &MdText{
				pdf:  mt.pdf,
				Type: TYPE_TEXT,
			}
			text.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.contents = append(mt.contents, text)
		case TYPE_EM:
			em := &MdText{
				pdf:  mt.pdf,
				Type: TYPE_EM,
			}
			em.SetText(mt.fonts[FONT_IALIC], token.Text)
			mt.contents = append(mt.contents, em)
		case TYPE_CODESPAN:
			codespan := &MdText{
				pdf:  mt.pdf,
				Type: TYPE_CODESPAN,
			}
			codespan.SetText(mt.fonts[FONT_NORMAL], token.Text)
			mt.contents = append(mt.contents, codespan)
		case TYPE_CODE:
			code := &MdText{
				pdf:  mt.pdf,
				Type: TYPE_CODE,
			}
			code.SetText(mt.fonts[FONT_NORMAL], token.Text+"\n")
			mt.contents = append(mt.contents, code)
		case TYPE_STRONG:
			strong := &MdText{
				pdf:  mt.pdf,
				Type: TYPE_STRONG,
			}
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
