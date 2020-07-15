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
	TEXT_NORMAL = "normal"
	TEXT_BOLD   = "bold"
	TEXT_IALIC  = "italic"
	TEXT_WARP   = "warp"
	TEXT_CODE   = "code"
	TEXT_LINK   = "link"
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

type combinedmark interface {
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
	link      string // special TEXT_LINK
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
	case TEXT_BOLD:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	case TEXT_IALIC:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	case TEXT_WARP:
		font = core.Font{Family: fontFamily, Size: defaultWarpFontSize, Style: ""}
	case TEXT_CODE:
		font = core.Font{Family: fontFamily, Size: defaultWarpFontSize, Style: ""}
	case TEXT_LINK, TEXT_NORMAL:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	default:
		panic(fmt.Sprintf("invalid type: %v", c.Type))
	}

	c.font = font
	text[0] = strings.Replace(text[0], "\t", "    ", -1)
	c.text = text[0]
	c.remain = text[0]
	if c.Type == TEXT_LINK {
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
		if c.Type == TEXT_WARP {
			// other padding is testing data
			// bg: 248,248,255 text:1,1,1 line:	245,245,245,     Typora
			// bg: 249,242,244 text:199,37,78 line:	245,245,245  GitLab
			c.pdf.BackgroundColor(x1, y, width, 15.0, "249,242,244",
				"1111", "245,245,245")

			c.pdf.TextColor(199, 37, 78)
			c.pdf.Cell(x1, y+3.15, text)
			c.pdf.TextColor(1, 1, 1)
		} else if c.Type == TEXT_CODE {
			// bg: 248,248,255, text: 1,1,1    Typora
			// bg: 40,42,54, text: 248,248,242  CSDN
			c.pdf.BackgroundColor(x1, y, x2-x1, 18.0, "40,42,54",
				"0000", "0,0,0")
			c.pdf.TextColor(248, 248, 242)
			c.pdf.Cell(x1, y+3.15, text)
			c.pdf.TextColor(1, 1, 1)
		} else if c.Type == TEXT_LINK {
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

type header struct {
	Size       int
	pdf        *core.Report
	font       core.Font
	lineHeight float64

	text          string
	remain        string
	needBreakLine bool
}

func (h *header) SetText(fontFamily string, text ...string) {
}

func (h *header) SetPadding(need bool, padding float64) {
}

func (h *header) GenerateAtomicCell() (pagebreak, over bool, err error) {
	return
}

func (h *header) String() string {
	text := strings.Replace(h.remain, "\n", "|", -1)
	return fmt.Sprintf("[size=%v,text=%v]", h.Size, text)
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

	needpadding := c.Type == TEXT_CODE && c.needpadding
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
			c.needpadding = c.Type == TEXT_CODE
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
	c, ok := md.(*content)
	if !ok {
		return ""
	}

	switch c.Type {
	case TEXT_NORMAL:
		return mt.fonts[FONT_NORMAL]
	case TEXT_BOLD:
		return mt.fonts[FONT_BOLD]
	case FONT_IALIC:
		return mt.fonts[FONT_IALIC]
	default:
		return mt.fonts[FONT_NORMAL]
	}
}

const (
	cut_code = "```"
	cut_wrap = "`"

	cut_bold  = "**"
	cut_itaic = "*"
)

func (mt *MarkdownText) SetText(text string) *MarkdownText {
	// [\n \t=#%@&"':<>,(){}_;/\?\.\+\-\=\^\$\[\]\!]
	relink := regexp.MustCompile(`^\[(.*?)\]\((.*?)\)`)
	reimage := regexp.MustCompile(`^\!\[image\]\((.*?)\)`)
	rensort := regexp.MustCompile(`^\-( )+`)
	rerest := regexp.MustCompile(`^\n[\t ]*?\n(\n[\t ]*?\n)*`)
	runes := []rune(text)
	n := len(runes)
	var (
		buf      bytes.Buffer
		contents []mardown
		main     combinedmark // combined composte
		sub      mardown      // basic composte
		cuts     []string     // mark some cut, eg, **, *, `, ```
	)
	if strings.HasPrefix(text, ">") {
		mt.quote = true
		mt.x, _ = mt.pdf.GetPageStartXY()
	}

	restmain := func() {
		main = nil
	}

	setsub := func(m mardown) {
		mainval, mainok := m.(combinedmark)

		// parent exsit, add to parent
		if main != nil {
			main.AddChild(m)
			log.Printf("main %+v", main)
			return
		}

		// parsent not exsit, set to parent
		if main == nil && mainok {
			main = mainval
			contents = append(contents, m)
			return
		}

		log.Printf("sub %+v", m)
		contents = append(contents, m)
	}

	defaultop := func(i *int) {
		if buf.Len() == 0 {
			sub = &content{pdf: mt.pdf, Type: TEXT_NORMAL}
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
				sub.SetText(mt.GetFontFamily(sub), buf.String())
				buf.Reset()
				setsub(sub)
			}

			if i+1 < n && string(runes[i:i+2]) == cut_bold {
				if len(cuts) == 0 || cuts[len(cuts)-1] != cut_bold {
					sub = &content{pdf: mt.pdf, Type: TEXT_BOLD}
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
				sub = &content{pdf: mt.pdf, Type: TEXT_IALIC}
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
				sub.SetText(mt.GetFontFamily(sub), buf.String())
				buf.Reset()
				setsub(sub)
			}

			// code text
			if i+2 < n && string(runes[i:i+3]) == cut_code && (i == 0 || runes[i-1] == '\n') {
				if len(cuts) == 0 || cuts[len(cuts)-1] != cut_code {
					sub = &content{pdf: mt.pdf, Type: TEXT_CODE, needpadding: true}
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
					sub = &content{pdf: mt.pdf, Type: TEXT_WARP}
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
				sub = &content{pdf: mt.pdf, Type: TEXT_WARP}
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
					sub.SetText(mt.GetFontFamily(sub), buf.String())
					buf.Reset()
					setsub(sub)
				}
				matchstr := reimage.FindString(temp)
				submatch := reimage.FindStringSubmatch(temp)
				c := &mdimage{pdf: mt.pdf}
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
					sub.SetText(mt.GetFontFamily(sub), buf.String())
					buf.Reset()
					setsub(sub)
				}

				matchstr := relink.FindString(temp)
				submatch := relink.FindStringSubmatch(temp)
				c := &content{pdf: mt.pdf, Type: TEXT_LINK}
				c.SetText(mt.fonts[FONT_NORMAL], submatch[1], submatch[2])
				setsub(c)
				i += len([]rune(matchstr))
			} else {
				defaultop(&i)
			}

		case '-':
			temp := string(runes[i:])
			if !curIsCode() && rensort.MatchString(temp) {
				if buf.Len() > 0 {
					sub.SetText(mt.GetFontFamily(sub), buf.String())
					buf.Reset()
					setsub(sub)
				}

				matchstr := rensort.FindString(temp)
				c := &blocksort{pdf: mt.pdf, sortType: SORT_DISORDER}
				c.SetText(mt.fonts[FONT_BOLD], "")
				setsub(c)
				i += len([]rune(matchstr))
			} else {
				defaultop(&i)
			}

		case '>':
			if !curIsCode() && (i == 0 || (i-1 > 0 && runes[i-1] == '\n' )) {
				i++
			} else {
				defaultop(&i)
			}

		case '\n':
			if i == 0 || main == nil {
				defaultop(&i)
				continue
			}

			temp := string(runes[i:])
			if !curIsCode() && rerest.MatchString(temp) {
				if buf.Len() > 0 {
					buf.WriteRune(runes[i])
					sub.SetText(mt.GetFontFamily(sub), buf.String())
					buf.Reset()
					setsub(sub)
				}

				restmain()

				matchstr := rerest.FindString(temp)
				i += len([]rune(matchstr))
			} else {
				defaultop(&i)
			}

		default:
			defaultop(&i)
		}
	}

	if buf.Len() > 0 {
		sub.SetText(mt.GetFontFamily(sub), buf.String())
		buf.Reset()
		setsub(sub)
	}

	//for _, c := range contents {
	//	if cc, ok := c.(*content); ok {
	//		log.Println("type", cc.Type)
	//		blocks := strings.Split(cc.text, "\n")
	//		log.Println("text", len(blocks), blocks)
	//		log.Printf("\n\n+++++++++++++++++++++++++++++++++++\n\n")
	//	}
	//}

	mt.contents = contents

	return mt
}

func (mt *MarkdownText) GetWritedLines() int {
	return mt.writedLines
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
