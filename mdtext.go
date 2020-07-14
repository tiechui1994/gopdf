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
	FONT_BOLD   = "bold"
	FONT_IALIC  = "italic"
	FONT_NORMAL = "normal"
)

const (
	TEXT_NORMAL = "normal"
	TEXT_BOLD   = "bold"
	TEXT_IALIC  = "italic"
	TEXT_WARP   = "warp"
	TEXT_CODE   = "code"
)

const (
	defaultLineHeight   = 18.0
	defaultFontSize     = 15.0
	defaultWarpFontSize = 10.0
)

type mardown interface {
	SetText(fontFamily, text string)
	GenerateAtomicCell() error
}

type content struct {
	pdf        *core.Report
	Type       string
	font       core.Font
	lineHeight float64

	stoped    bool
	precision float64
	length    float64
	text      string
	remain    string
	newlines  int

	// when type is code can use
	needpadding bool
}

func (c *content) SetText(fontFamily, text string) {
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
	default:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	}
	c.font = font
	c.text = text
	c.remain = text
	c.lineHeight = defaultLineHeight
	c.pdf.Font(font.Family, font.Size, font.Style)
	c.pdf.SetFontWithStyle(font.Family, font.Style, font.Size)
	re := regexp.MustCompile(`[\n \t=#%@&"':<>,(){}_;/\?\.\+\-\=\^\$\[\]\!]`)

	subs := re.FindAllString(text, -1)
	if len(subs) > 0 {
		str := re.ReplaceAllString(text, "")
		c.length = c.pdf.MeasureTextWidth(str)
		c.precision = c.length / float64(len([]rune(str)))
	} else {
		c.length = c.pdf.MeasureTextWidth(text)
		c.precision = c.length / float64(len([]rune(text)))
	}
}

func (c *content) GenerateAtomicCell() error {
	x1, y := c.pdf.GetXY()
	x2, _ := c.pdf.GetPageEndXY()

	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)
	text, width, newline := c.GetSubText(x1, x2)
	for !c.stoped {
		if c.Type == TEXT_WARP {
			// other padding is testing data
			c.pdf.BackgroundColor(x1, y, width, 15.0, "248,248,255",
				"1111", "220,220,220")
			c.pdf.Cell(x1, y+3.15, text)
		} else if c.Type == TEXT_CODE {
			c.pdf.BackgroundColor(x1, y, x2-x1, 18.0, "248,248,255",
				"0000", "220,220,220")
			c.pdf.Cell(x1, y+3.15, text)
		} else {
			c.pdf.Cell(x1, y-0.45, text)
		}

		if newline {
			x1, _ = c.pdf.GetPageStartXY()
			y += c.lineHeight
		} else {
			x1 += width
		}

		c.pdf.SetXY(x1, y)
		text, width, newline = c.GetSubText(x1, x2)
	}

	return nil
}

type mdimage struct {
	pdf   *core.Report
	image *Image
}

func (mi *mdimage) SetText(fontFamily, filename string) {
	image := NewImage(filename, mi.pdf)
	mi.image = image
}

func (mi *mdimage) GenerateAtomicCell() error {
	mi.image.GenerateAtomicCell()
	return nil
}

type mdlink struct {
	pdf  *core.Report
	font core.Font
	url  string
	link string
}

func (ml *mdlink) SetText(link, url string) {
	ml.link = link
	ml.url = url
}

func (ml *mdlink) GenerateAtomicCell() error {
	x, y := ml.pdf.GetXY()
	ml.pdf.ExternalLink(x, y, 10, ml.link, ml.url)
	return nil
}

const (
	SORT_ORDER    = "order"
	SORT_DISORDER = "disorder"
)

type mdsort struct {
	pdf       *core.Report
	font      core.Font
	sortType  string
	sortIndex string
}

func (ms *mdsort) SetText(fontFamily, _ string) {
	ms.font = core.Font{Family: fontFamily, Size: 18.0}
}

func (ms *mdsort) GenerateAtomicCell() error {
	ms.pdf.Font(ms.font.Family, ms.font.Size, ms.font.Style)
	ms.pdf.SetFontWithStyle(ms.font.Family, ms.font.Style, ms.font.Size)

	var text string
	x, y := ms.pdf.GetXY()
	switch ms.sortType {
	case SORT_ORDER:
		text = fmt.Sprintf(" %v. ", ms.sortIndex)
		ms.pdf.Cell(x, y, text)

	case SORT_DISORDER:
		text = " · "
		ms.pdf.Cell(x, y, text)
	}

	length := ms.pdf.MeasureTextWidth(text)
	ms.pdf.SetXY(x+length, y)

	return nil
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

		log.Println(w-width > c.precision, width-w > c.precision)
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
	relink := regexp.MustCompile(`^\[(.*?)\]\((.*?)\)`)
	reimage := regexp.MustCompile(`^\!\[image\]\((.*?)\)`)
	rensort := regexp.MustCompile(`^\-( )+`)
	runes := []rune(text)
	n := len(runes)
	var (
		buf      bytes.Buffer
		contents []mardown
		md       mardown
		cuts     []string
	)
	if strings.HasPrefix(text, ">") {
		mt.quote = true
		mt.x, _ = mt.pdf.GetPageStartXY()
	}

	for i := 0; i < n; {
		switch runes[i] {
		case '*':
			if len(cuts) > 0 && (cuts[len(cuts)-1] == cut_wrap || cuts[len(cuts)-1] == cut_code) {
				buf.WriteRune(runes[i])
				i += 1
				continue
			}

			if buf.Len() > 0 {
				md.SetText(mt.GetFontFamily(md), buf.String())
				buf.Reset()
				contents = append(contents, md)
			}

			if i+1 < n && string(runes[i:i+2]) == cut_bold {
				if len(cuts) == 0 || cuts[len(cuts)-1] != cut_bold {
					md = &content{pdf: mt.pdf, Type: TEXT_BOLD}
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
				md = &content{pdf: mt.pdf, Type: TEXT_IALIC}
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
				md.SetText(mt.GetFontFamily(md), buf.String())
				buf.Reset()
				contents = append(contents, md)
			}

			// code text
			if i+2 < n && string(runes[i:i+3]) == cut_code && (i == 0 || runes[i-1] == '\n') {
				if len(cuts) == 0 || cuts[len(cuts)-1] != cut_code {
					md = &content{pdf: mt.pdf, Type: TEXT_CODE, needpadding: true}
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
					md = &content{pdf: mt.pdf, Type: TEXT_WARP}
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
				md = &content{pdf: mt.pdf, Type: TEXT_WARP}
				cuts = append(cuts, cut_wrap)
				buf.WriteRune(runes[i+1])
				i += 2
			} else {
				cuts = cuts[:len(cuts)-1]
				i += 1
			}

		case '!':
			temp := string(runes[i:])
			if reimage.MatchString(temp) {
				if buf.Len() > 0 {
					md.SetText(mt.GetFontFamily(md), buf.String())
					buf.Reset()
					contents = append(contents, md)
				}
				matchstr := reimage.FindString(temp)
				submatch := reimage.FindStringSubmatch(temp)
				c := &mdimage{pdf: mt.pdf}
				c.SetText("", submatch[1])
				contents = append(contents, c)
				i += len([]rune(matchstr))
				continue
			}
			buf.WriteRune(runes[i])
			i += 1

		case '[':
			temp := string(runes[i:])
			if relink.MatchString(temp) {
				if buf.Len() > 0 {
					md.SetText(mt.GetFontFamily(md), buf.String())
					buf.Reset()
					contents = append(contents, md)
				}

				matchstr := relink.FindString(temp)
				submatch := relink.FindStringSubmatch(temp)
				c := &mdlink{pdf: mt.pdf}
				c.SetText(submatch[1], submatch[2])
				contents = append(contents, c)
				i += len([]rune(matchstr))
				continue
			}
			buf.WriteRune(runes[i])
			i += 1

		case '-':
			temp := string(runes[i:])
			if rensort.MatchString(temp) {
				if buf.Len() > 0 {
					md.SetText(mt.GetFontFamily(md), buf.String())
					buf.Reset()
					contents = append(contents, md)
				}

				matchstr := rensort.FindString(temp)
				c := &mdsort{pdf: mt.pdf, sortType: SORT_DISORDER}
				c.SetText(mt.fonts[FONT_BOLD], "")
				contents = append(contents, c)
				i += len([]rune(matchstr))
				continue
			}
			buf.WriteRune(runes[i])
			i += 1

		case '>':
			if i == 0 || (i-1 > 0 && runes[i-1] == '\n' ) {
				i++
				continue
			}
			buf.WriteRune(runes[i])
			i++

		default:
			if buf.Len() == 0 {
				md = &content{pdf: mt.pdf, Type: TEXT_NORMAL}
			}
			buf.WriteRune(runes[i])
			i += 1
		}
	}

	for _, c := range contents {
		if cc, ok := c.(*content); ok {
			log.Println("type", cc.Type)
			blocks := strings.Split(cc.text, "\n")
			log.Println("text", len(blocks), blocks)
			log.Printf("\n\n+++++++++++++++++++++++++++++++++++\n\n")
		}
	}

	mt.contents = contents

	return mt
}

func (mt *MarkdownText) GetWritedLines() int {
	return mt.writedLines
}

func (mt *MarkdownText) GenerateAtomicCell() error {
	if len(mt.contents) == 0 {
		return fmt.Errorf("not set text")
	}

	for _, c := range mt.contents {
		c.GenerateAtomicCell()
	}

	return nil
}
