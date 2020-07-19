package gopdf

import (
	"math"
	"log"
	"fmt"
	"strings"
	"bytes"
	"regexp"

	"github.com/tiechui1994/gopdf/core"
)

const (
	FONT_NORMAL = "normal"
	FONT_BOLD   = "bold"
	FONT_IALIC  = "italic"
)

const (
	TYPE_ORIGIN = "origin"

	TYPE_NORMAL = "normal"
	TYPE_BOLD   = "bold"
	TYPE_IALIC  = "italic"
	TYPE_WARP   = "warp"
	TYPE_CODE   = "code"
	TYPE_LINK   = "link"

	TYPE_HIGHLIGHT = "highlight"

	TYPE_HEADER_ONE = "1"
	TYPE_HEADER_TWO = "2"
	TYPE_HEADER_THR = "3"
	TYPE_HEADER_FOU = "4"
	TYPE_HEADER_FIV = "5"
	TYPE_HEADER_SIX = "6"

	TYPE_IMAGE = "image"

	TYPE_DELETE = "delete"

	TYPE_NSORT = "nsort"

	TYPE_REFER = "refer"
)

const (
	HEADER_ONE = iota + 1
	HEADER_TWO
	HEADER_THR
	HEADER_FOU
	HEADER_FIV
	HEADER_SIX
)

const (
	defaultLineHeight   = 18.0
	defaultFontSize     = 15.0
	defaultWarpFontSize = 10.0
)

type mardown interface {
	SetText(fontFamily string, text ...string)
	SetPadding(need bool, padding float64)
	GenerateAtomicCell() (pagebreak, over bool, err error)
}

type combined interface {
	AddChild(child mardown)
}

// common componet
type content struct {
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

func (c *content) SetText(fontFamily string, text ...string) {
	if len(text) == 0 {
		panic("text is invalid")
	}

	var font core.Font
	switch c.Type {
	case TYPE_BOLD:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	case TYPE_IALIC:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	case TYPE_WARP:
		font = core.Font{Family: fontFamily, Size: defaultWarpFontSize, Style: ""}
	case TYPE_CODE:
		font = core.Font{Family: fontFamily, Size: defaultWarpFontSize, Style: ""}
	case TYPE_LINK, TYPE_NORMAL:
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

func (c *content) SetPadding(need bool, padding float64) {
}

func (c *content) GenerateAtomicCell() (pagebreak, over bool, err error) {
	pageEndX, pageEndY := c.pdf.GetPageEndXY()
	x1, y := c.pdf.GetXY()
	x2 := pageEndX

	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)

	text, width, newline := c.GetSubText(x1, x2)
	for !c.stoped {
		if c.Type == TYPE_WARP {
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

func (c *content) String() string {
	text := strings.Replace(c.remain, "\n", "|", -1)
	return fmt.Sprintf("[type=%v,text=%v]", c.Type, text)
}

type mdimage struct {
	pdf   *core.Report
	image *Image
}

func (mi *mdimage) SetText(fontFamily string, filename ...string) {
	image := NewImage(filename[0], mi.pdf)
	mi.image = image
}

func (mi *mdimage) GenerateAtomicCell() (pagebreak, over bool, err error) {
	mi.image.GenerateAtomicCell()
	return
}

func (mi *mdimage) SetPadding(need bool, padding float64) {
}

// combined components

type header struct {
	Size       int
	pdf        *core.Report
	fonts      map[string]string
	children   []mardown
	lineHeight float64

	text          string
	remain        string
	needBreakLine bool
}

func (h *header) GetFontFamily(md mardown) string {
	switch md.(type) {
	case *content:
		c, _ := md.(*content)
		switch c.Type {
		case TYPE_NORMAL:
			return h.fonts[FONT_NORMAL]
		case TYPE_BOLD:
			return h.fonts[FONT_BOLD]
		case FONT_IALIC:
			return h.fonts[FONT_IALIC]
		default:
			return h.fonts[FONT_NORMAL]
		}
	}

	return ""
}

func (h *header) SetText(_ string, textargs ...string) {
	var (
		normal, code, itialic int
	)
	switch h.Size {
	case HEADER_ONE:
		normal = defaultFontSize + 12
		h.lineHeight = defaultLineHeight + 16
	case HEADER_TWO:
		normal = defaultFontSize + 8
		h.lineHeight = defaultLineHeight + 12
	case HEADER_THR:
		normal = defaultFontSize + 4
		h.lineHeight = defaultLineHeight + 8
	case HEADER_FOU:
		normal = defaultFontSize + 3
		h.lineHeight = defaultLineHeight + 6
	case HEADER_FIV:
		normal = defaultFontSize + 2
		h.lineHeight = defaultLineHeight + 4
	case HEADER_SIX:
		normal = defaultFontSize
		h.lineHeight = defaultLineHeight
	}

	_, _, _ = normal, code, itialic
	text := textargs[0]

	relink := regexp.MustCompile(`^\[(.*?)\]\((.*?)\)`)
	reimage := regexp.MustCompile(`^\!\[image\]\((.*?)\)`)
	runes := []rune(text)
	n := len(runes)
	log.Println("runes:", len(runes), "text:", strings.Replace(text, "\n", "|", -1))
	var (
		buf      bytes.Buffer
		contents []mardown
		sub      mardown  // basic composte
		cuts     []string // mark some cut, eg, **, *, `, ```
	)

	setsub := func(m mardown) {
		contents = append(contents, m)
	}

	defaultop := func(i *int) {
		if buf.Len() == 0 {
			sub = &content{pdf: h.pdf, Type: TYPE_NORMAL}
		}
		buf.WriteRune(runes[*i])
		*i += 1
	}

	curIsCode := func() bool {
		if len(cuts) == 0 {
			return false
		}

		last := cuts[len(cuts)-1]
		if last == "```" || last == "`" {
			return true
		}

		return false
	}

	for i := 0; i < n; {
		switch runes[i] {
		case '*':
			if len(cuts) > 0 && (cuts[len(cuts)-1] == cut_wrap || cuts[len(cuts)-1] == cut_code) {
				defaultop(&i)
				continue
			}

			if buf.Len() > 0 {
				sub.SetText(h.GetFontFamily(sub), buf.String())
				buf.Reset()
				setsub(sub)
			}

			if i+1 < n && string(runes[i:i+2]) == cut_bold {
				if len(cuts) == 0 || cuts[len(cuts)-1] != cut_bold {
					sub = &content{pdf: h.pdf, Type: TYPE_BOLD}
					cuts = append(cuts, cut_bold)
					if runes[i+2] == '`' {
						buf.WriteRune(runes[i+3])
						i += 4
					} else {
						buf.WriteRune(runes[i+2])
						i += 3
					}

				} else {
					cuts = cuts[:len(cuts)-1]
					i += 2
				}
				continue
			}

			if len(cuts) == 0 || cuts[len(cuts)-1] != cut_itaic {
				sub = &content{pdf: h.pdf, Type: TYPE_IALIC}
				cuts = append(cuts, cut_itaic)
				if runes[i+1] == '`' {
					buf.WriteRune(runes[i+2])
					i += 3
				} else {
					buf.WriteRune(runes[i+1])
					i += 2
				}
			} else {
				cuts = cuts[:len(cuts)-1]
				i += 1
			}

		case '`':
			if len(cuts) > 0 && (cuts[len(cuts)-1] == cut_itaic || cuts[len(cuts)-1] == cut_bold) {
				i += 1
				continue
			}

			if buf.Len() > 0 {
				sub.SetText(h.GetFontFamily(sub), buf.String())
				buf.Reset()
				setsub(sub)
			}

			// code text
			if i+2 < n && string(runes[i:i+3]) == cut_code && (i == 0 || runes[i-1] == '\n') {
				if len(cuts) == 0 || cuts[len(cuts)-1] != cut_code {
					sub = &content{pdf: h.pdf, Type: TYPE_CODE, needpadding: true}
					index := strings.Index(string(runes[i:]), "\n")
					cuts = append(cuts, cut_code)
					buf.WriteRune(runes[i+index+1])
					i = i + index + 2
				} else {
					cuts = cuts[:len(cuts)-1]
					i += 3
				}
				continue
			}

			// wrap
			if i+2 < n && string(runes[i:i+3]) == cut_code {
				if len(cuts) == 0 || cuts[len(cuts)-1] != cut_code {
					sub = &content{pdf: h.pdf, Type: TYPE_WARP}
					cuts = append(cuts, cut_code)
					buf.WriteRune(runes[i+3])
					i += 4
				} else {
					cuts = cuts[:len(cuts)-1]
					i += 3
				}
				continue
			}

			// wrap
			if len(cuts) == 0 || cuts[len(cuts)-1] != cut_wrap {
				sub = &content{pdf: h.pdf, Type: TYPE_WARP}
				cuts = append(cuts, cut_wrap)
				buf.WriteRune(runes[i+1])
				i += 2
			} else {
				cuts = cuts[:len(cuts)-1]
				i += 1
			}

		case '!':
			temp := string(runes[i:])
			if !curIsCode() && reimage.MatchString(temp) {
				if buf.Len() > 0 {
					sub.SetText(h.GetFontFamily(sub), buf.String())
					buf.Reset()
					setsub(sub)
				}
				matchstr := reimage.FindString(temp)
				submatch := reimage.FindStringSubmatch(temp)
				c := &mdimage{pdf: h.pdf}
				c.SetText("", submatch[1])
				setsub(c)
				i += len([]rune(matchstr))
			} else {
				defaultop(&i)
			}

		case '[':
			temp := string(runes[i:])
			if !curIsCode() && relink.MatchString(temp) {
				if buf.Len() > 0 {
					sub.SetText(h.GetFontFamily(sub), buf.String())
					buf.Reset()
					setsub(sub)
				}

				matchstr := relink.FindString(temp)
				submatch := relink.FindStringSubmatch(temp)
				c := &content{pdf: h.pdf, Type: TYPE_LINK}
				c.SetText(h.fonts[FONT_NORMAL], submatch[1], submatch[2])
				setsub(c)
				i += len([]rune(matchstr))
			} else {
				defaultop(&i)
			}

		default:
			defaultop(&i)
		}
	}

	if buf.Len() > 0 {
		sub.SetText(h.GetFontFamily(sub), buf.String())
		buf.Reset()
		setsub(sub)
	}

	h.children = contents
}

func (h *header) SetPadding(need bool, padding float64) {
}

func (h *header) GenerateAtomicCell() (pagebreak, over bool, err error) {
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

func (h *header) String() string {
	text := strings.Replace(h.remain, "\n", "|", -1)
	return fmt.Sprintf("[size=%v,text=%v]", h.Size, text)
}

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

func (bs *blocksort) SetPadding(need bool, padding float64) {
}

func (bs *blocksort) AddChild(child mardown) {
	bs.children = append(bs.children, child)
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

//
type blockwarp struct {
	pdf      *core.Report
	children []mardown

	padinglen float64
}

func (bw *blockwarp) SetText(fontFamily string, _ ...string) {
}

func (bw *blockwarp) SetPadding(need bool, padding float64) {
}

func (bw *blockwarp) AddChild(child mardown) {
	bw.children = append(bw.children, child)
}

func (bw *blockwarp) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return
}

// GetSubText, Returns the content of a string of length x2-x1.
// This string is a substring of text.
// After return, the remain and length will change
func (c *content) GetSubText(x1, x2 float64) (text string, width float64, newline bool) {
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

// pretreatment element
type pre struct {
	Type  string
	Value string
}

func (p *pre) String() string {
	value := strings.Replace(p.Value, "\n", "|", -1)
	return fmt.Sprintf("[type=%v, value=%v]", p.Type, value)
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

func (mt *MarkdownText) GetFontFamily(md mardown) string {
	switch md.(type) {
	case *content:
		c, _ := md.(*content)
		switch c.Type {
		case TYPE_NORMAL:
			return mt.fonts[FONT_NORMAL]
		case TYPE_BOLD:
			return mt.fonts[FONT_BOLD]
		case FONT_IALIC:
			return mt.fonts[FONT_IALIC]
		default:
			return mt.fonts[FONT_NORMAL]
		}
	}

	return ""
}

const (
	cut_code = "```"
	cut_wrap = "`"

	cut_bold  = "**"
	cut_itaic = "*"
)

var (
	// [\n \t=#%@&"':<>,(){}_;/\?\.\+\-\=\^\$\[\]\!]
	rerest = regexp.MustCompile(`^\n[\t ]*?\n(\n[\t ]*?\n)*`)

	// ^\*{2}\S
	// ((\S+\s|\s\S+|\S)+\n)*
	// \S+\*{2}
	// need handle "\n\s*\n" special condition
	reboldprefix = regexp.MustCompile(`^\*{2}\S`)
	reboldsuffix = regexp.MustCompile(`\S\*{2}$`)

	// \*\S
	// ((\S+\s|\s\S+|\S)+\n)*
	// \S+\*
	// need handle "\n\s*\n" special condition
	reitalic = regexp.MustCompile(`^\*\S((\S+\s|\s\S+|\S)+\n)*\S+\*`)

	reitalicprefix = regexp.MustCompile(`^\*\S`)
	reitalicsuffix = regexp.MustCompile(`\S\*$`)

	// need handle "\n\s*\n" special condition
	redel = regexp.MustCompile(`^~{2}\S((\S+\s|\s\S+|\S)+\n)*\S+~{2}`)

	recodeprefix     = regexp.MustCompile("^`{3}.*?\n")
	recodewarp       = regexp.MustCompile("^`{3}.*`{3}")
	retextwarpprefix = regexp.MustCompile("^`")

	// when current is '='
	// first global search
	rehighlight = regexp.MustCompile(`.*\n\s{0,3}=+$`)

	// remove \n\s*\n
	relink  = regexp.MustCompile(`^\[.*\n?(.*\n|.*)\]\(\n?(.*)\n?\)`)
	reimage = regexp.MustCompile(`^\!\[image\]\(\n?(.*)\n?\)`)

	// stoped by \n\s*\n
	rensort = regexp.MustCompile(`^-\s{1,4}((\S+\s|\s\S+|\S)+\n)*`)

	reheader = regexp.MustCompile(`^(#)+.*`)

	// need handle "\n\s*\n" special condition
	reref = regexp.MustCompile(`^>((\S+\s|\s\S+|\S)+\n)*`)
)

func (mt *MarkdownText) parseBoldText(i int) (ok bool, ) {
	return
}

func (mt *MarkdownText) SetText(text string) *MarkdownText {
	var (
		contents []mardown
	)
	pres := mt.PreProccesText(text)
	for _, pre := range pres {
		log.Printf("%v", pre.String())
		switch pre.Type {
		case TYPE_WARP, TYPE_CODE, TYPE_BOLD, TYPE_NORMAL:
			c := content{
				pdf:  mt.pdf,
				Type: pre.Type,
			}
			c.SetText(mt.GetFontFamily(&c), pre.Value)
			contents = append(contents, &c)
		}

	}

	mt.contents = contents

	return mt
}

func (mt *MarkdownText) GetWritedLines() int {
	return mt.writedLines
}

func (mt *MarkdownText) PreProccesText(text string) []pre {
	/*
	// header and line
	# xx
	## xx

	// bold and line
	xxx
	==

	// del line
	~~ xx ~~
	*/

	// when header (#), the '\n' can not replace
	// when code (```) include '\n', code text can not replace
	// when hence (`), if contains '\n[ ]*\n', not hence, not replace, else replace '\n' with ' '
	// when link ([]()), if [] contains '\n[ ]*\n', not link, if () contains more than two lines,
	// not link. else replace '\n' with ' '. and image is so on
	// when sort (-), if contains '\n[ ]+\n', after must new line, else replace '\n' with ' '
	// when wrap (*), if after|before * is ' ' or '\n[ ]*\n' not warp

	runes := []rune(text)
	n := len(runes)

	// break some 
	rebreak := regexp.MustCompile(`\n\s*\n`)

	type prehighlight struct {
		pre
		last int
	}

	var (
		i           int
		buf         bytes.Buffer
		pres        []pre
		ignoreindex map[int]prehighlight
	)
	ignoreindex = make(map[int]prehighlight)
	hgindexs := rehighlight.FindAllStringIndex(text, -1)
	if len(hgindexs) > 0 {
		for _, tuple := range hgindexs {
			index, last := tuple[0], tuple[1]
			brindex := strings.LastIndex(string(runes[index:last]), "\n")
			ignoreindex[index] = prehighlight{
				pre: pre{
					Type:  TYPE_HIGHLIGHT,
					Value: string(runes[index:brindex]),
				},
				last: last,
			}
		}
	}
	resetbuf := func() {
		if buf.Len() > 0 {
			pres = append(pres, pre{
				Type:  TYPE_NORMAL,
				Value: buf.String(),
			})
			buf.Reset()
		}
	}
	defaultop := func() {
		temp := string(runes[i:])
		rebreakseg := regexp.MustCompile(`^\n\s*\n`)

		if runes[i] == '\n' && rebreakseg.MatchString(temp) {
			buf.WriteRune('\n')
			resetbuf()
			matched := rebreakseg.FindString(temp)
			i += len([]rune(matched))
			return
		}

		buf.WriteRune(runes[i])
		i++
	}

	for i < n {
		if val, ok := ignoreindex[i]; ok {
			pres = append(pres, val.pre)
			i = val.last
			continue
		}

		temp := string(runes[i:])
		switch runes[i] {
		case '*':
			log.Println("--------->", string(temp[:8]))
			if reboldprefix.MatchString(temp) {
				resetbuf()
				k := i + 2
				ok, index := deepinSearch(runes[k:], "**", true)
				log.Println(ok, string(runes[i:k+index]), index)
				if ok && reboldsuffix.MatchString(string(runes[k:k+index])) {
					pres = append(pres, pre{
						Type:  TYPE_BOLD,
						Value: strings.Trim(string(runes[i:k+index]), "**"),
					})
					i = k + index
					continue
				}

				pres = append(pres, pre{
					Type:  TYPE_ORIGIN,
					Value: string(runes[i:k+index]),
				})
				i = k + index
				continue
			}

			if reitalicprefix.MatchString(temp) {
				resetbuf()
				matched := reitalic.FindString(temp)
				i += len([]rune(matched))
				if rebreak.MatchString(matched) {
					matched = rebreak.ReplaceAllString(matched, "\n")
					pres = append(pres, pre{
						Type:  TYPE_ORIGIN,
						Value: strings.Replace(matched, "\n", " ", -1),
					})
					continue
				}

				matched = strings.Replace(matched, "\n", " ", -1)
				pres = append(pres, pre{
					Type:  TYPE_IALIC,
					Value: strings.Trim(matched, "*"),
				})
				continue
			}

			defaultop()

		case '`':
			if (i == 0 || runes[i-1] == '\n') && recodeprefix.MatchString(temp) {
				resetbuf()
				prefix := recodeprefix.FindString(temp)
				k := i + len([]rune(prefix))
				ok, index := deepinSearch(runes[k:], "\n```", false)
				if ok {
					pres = append(pres, pre{
						Type:  TYPE_CODE,
						Value: strings.Trim(string(runes[k:k+index]), "```"),
					})
					i = k + index
					continue
				}

				pres = append(pres, pre{
					Type:  TYPE_ORIGIN,
					Value: string(runes[i:k+index]),
				})
				i = k + index
				continue
			} else if recodewarp.MatchString(temp) {
				resetbuf()
				matched := recodewarp.FindString(temp)
				i += len([]rune(matched))
				pres = append(pres, pre{
					Type:  TYPE_WARP,
					Value: strings.Trim(matched, "````"),
				})
				continue
			} else if retextwarpprefix.MatchString(temp) {
				resetbuf()
				prefix := retextwarpprefix.FindString(temp)
				k := i + len([]rune(prefix))
				ok, index := deepinSearch(runes[k:], "`", true)
				if ok {
					pres = append(pres, pre{
						Type:  TYPE_WARP,
						Value: strings.Trim(string(runes[i:k+index]), "`"),
					})
					i = k + index
					continue
				}

				pres = append(pres, pre{
					Type:  TYPE_ORIGIN,
					Value: string(runes[i:k+index]),
				})
				i = k + index
				continue
			}

			defaultop()

		case '#':
			if (i == 0 || runes[i-1] == '\n') && reheader.MatchString(temp) {
				resetbuf()
				matched := reheader.FindString(temp)
				i += len([]rune(matched))
				var size int
				matched = strings.TrimLeftFunc(matched, func(r rune) bool {
					if size == HEADER_SIX {
						return false
					}
					if r == '#' {
						size++
					}
					return r == '#'
				})
				pres = append(pres, pre{
					Type:  fmt.Sprintf("%v", size),
					Value: matched,
				})
				continue
			}
			defaultop()

		case '!':
			if reimage.MatchString(temp) {
				resetbuf()
				matched := reimage.FindString(temp)
				i += len([]rune(matched))
				if rebreak.MatchString(matched) {
					matched = rebreak.ReplaceAllString(matched, "\n")
					pres = append(pres, pre{
						Type:  TYPE_ORIGIN,
						Value: strings.Replace(matched, "\n", " ", -1),
					})
					continue
				}

				si := strings.Index(matched, "]")
				se := strings.LastIndex(matched, ")")
				pres = append(pres, pre{
					Type:  TYPE_IMAGE,
					Value: matched[si+2:se],
				})
				continue
			}
			defaultop()

		case '[':
			if relink.MatchString(temp) {
				resetbuf()
				matched := relink.FindString(temp)
				i += len([]rune(matched))
				if rebreak.MatchString(matched) {
					matched = rebreak.ReplaceAllString(matched, "\n")
					pres = append(pres, pre{
						Type:  TYPE_ORIGIN,
						Value: strings.Replace(matched, "\n", " ", -1),
					})
					continue
				}

				matched = strings.Replace(matched, "\n", " ", -1)
				pres = append(pres, pre{
					Type:  TYPE_LINK,
					Value: matched,
				})
				continue
			}
			defaultop()

		case '~':
			if redel.MatchString(temp) {
				resetbuf()
				matched := redel.FindString(temp)
				i += len([]rune(matched))
				if rebreak.MatchString(matched) {
					matched = rebreak.ReplaceAllString(matched, "\n")
					pres = append(pres, pre{
						Type:  TYPE_ORIGIN,
						Value: strings.Replace(matched, "\n", " ", -1),
					})
					continue
				}

				matched = strings.Replace(matched, "\n", " ", -1)
				pres = append(pres, pre{
					Type:  TYPE_DELETE,
					Value: strings.Trim(matched, "~~"),
				})
				continue
			}

		case '-':
			if (i == 0 || runes[i-1] == '\n') && rensort.MatchString(temp) {
				resetbuf()
				matched := rensort.FindString(temp)

				// break
				if rebreak.MatchString(matched) {
					index := rebreak.FindAllStringIndex(matched, 1)
					matched = matched[:index[0][0]+1]
				}

				i += len([]rune(matched))
				pres = append(pres, pre{
					Type:  TYPE_NSORT,
					Value: strings.Replace(matched, "\n", " ", -1),
				})
				continue
			}
			defaultop()

		case '>':
			if (i == 0 || runes[i-1] == '\n') && reref.MatchString(temp) {
				resetbuf()
				matched := reref.FindString(temp)
				if rebreak.MatchString(matched) {
					index := rebreak.FindAllStringIndex(matched, 1)
					matched = matched[:index[0][0]+1]
				}

				i += len([]rune(matched))

				rereplace := regexp.MustCompile(`\n\s*>`)
				matched = rereplace.ReplaceAllString(matched, " ")
				matched = strings.Replace(matched, "\n", " ", -1)
				pres = append(pres, pre{
					Type:  TYPE_REFER,
					Value: matched,
				})
				continue
			}
			defaultop()

		default:
			defaultop()
		}
	}

	resetbuf()

	return pres
}

func (mt *MarkdownText) GenerateAtomicCell() (err error) {
	if len(mt.contents) == 0 {
		return fmt.Errorf("not set text")
	}

	log.Println()
	log.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++")

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

func deepinSearch(runes []rune, cut string, igblankline bool) (ok bool, index int) {
	cs := len([]rune(cut))
	var (
		i        int
		cur, pre int
	)

	cur, pre = 0, -1
	for i < len(runes) {
		if igblankline {
			//  \f\n\r\t\v
			if runes[i] != ' ' && runes[i] != '\f' && runes[i] != '\r' && runes[i] != '\t' && runes[i] != '\v' {
				cur++
			}

			if runes[i] == '\n' {
				if cur == 0 && pre == 0 {
					return false, i
				}
				pre, cur = cur, 0
				i++
				continue
			}
		}

		if i+cs < len(runes) && string(runes[i:i+cs]) == cut {
			return true, i + cs
		}

		i++
	}

	return false, cs
}
