package gopdf

import (
	"math"
	"log"

	"github.com/tiechui1994/gopdf/core"
	"fmt"
	"strings"
	"bytes"
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

type content struct {
	pdf        *core.Report
	Type       string
	font       core.Font
	lineHeight float64

	precision float64
	length    float64
	text      string
	remain    string
	newlines  int
}

func (c *content) SetText(fontFamily, text string) {
	var font core.Font
	switch c.Type {
	case TEXT_BOLD:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: "B"}
	case TEXT_IALIC:
		font = core.Font{Family: fontFamily, Size: defaultFontSize, Style: ""}
	case TEXT_WARP:
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
	c.length = c.pdf.MeasureTextWidth(text)
	c.precision = c.length / float64(len([]rune(text)))
}

// GetLength, remain text length
func (c *content) GetLength() float64 {
	return c.length
}

func (c *content) GenerateAtomicCell() error {
	x1, y := c.pdf.GetXY()
	x2, _ := c.pdf.GetPageEndXY()

	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)
	text, width := c.GetSubText(x1, x2)
	for text != "" {
		if c.Type == TEXT_WARP {
			// other padding is testing data
			c.pdf.BackgroundColor(x1, y, width, c.precision+4.4, "248,248,255",
				"1111", "220,220,220")
			c.pdf.Cell(x1, y+1.35, text)
		} else {
			c.pdf.Cell(x1, y-0.45, text)
		}

		if c.GetLength() > 0 {
			x1, _ = c.pdf.GetPageStartXY()
			y += c.lineHeight
		} else {
			x1 += width
		}

		c.pdf.SetXY(x1, y)
		text, width = c.GetSubText(x1, x2)
	}

	return nil
}

// GetSubText, Returns the content of a string of length x2-x1.
// This string is a substring of text.
// After return, the remain and length will change
func (c *content) GetSubText(x1, x2 float64) (string, float64) {
	width := math.Abs(x1 - x2)
	c.pdf.Font(c.font.Family, c.font.Size, c.font.Style)
	c.pdf.SetFontWithStyle(c.font.Family, c.font.Style, c.font.Size)
	length := c.pdf.MeasureTextWidth(c.remain)
	if length <= width {
		text := c.remain
		c.remain = ""
		c.length = 0
		return text, length
	}

	runes := []rune(c.remain)
	step := int(float64(len(runes)) * width / length)
	for i, j := 0, step; i < len(runes) && j < len(runes); {
		w := c.pdf.MeasureTextWidth(string(runes[i:j]))

		// less than precision
		if math.Abs(w-width) < c.precision {
			// try again, can more precise
			if j+1 < len(runes) {
				w1 := c.pdf.MeasureTextWidth(string(runes[i:j+1]))
				if math.Abs(w1-width) < c.precision {
					j = j + 1
					continue
				}
			}

			// reset
			c.remain = string(runes[j:])
			c.length = c.length - w
			c.newlines ++

			return string(runes[i:j]), w
		}

		if w > width {
			j--
		} else {
			j++
		}
	}

	return "", 0
}

type MarkdownText struct {
	quote       bool
	pdf         *core.Report
	fonts       map[string]string
	contents    []content
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

func (mt *MarkdownText) GetFontFamily(c content) string {
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

func (mt *MarkdownText) SetText(text string) *MarkdownText {

	runes := []rune(text)
	n := len(runes)
	var (
		buf      bytes.Buffer
		c        content
		contents []content
		lastStep string
	)
	if strings.HasPrefix(text, ">") {
		mt.quote = true
		mt.x, _ = mt.pdf.GetPageStartXY()
	}

	for i := 0; i < n; {
		switch runes[i] {
		case '*':
			if buf.Len() > 0 {
				c.SetText(mt.GetFontFamily(c), buf.String())
				buf.Reset()
				contents = append(contents, c)
				continue
			}

			if i+1 < n && runes[i+1] == '*' {
				c = content{pdf: mt.pdf, Type: TEXT_BOLD}
				i += 2
			} else {
				c = content{pdf: mt.pdf, Type: TEXT_IALIC}
				i += 1
			}

		case '`':
			if buf.Len() > 0 {
				c.SetText(mt.GetFontFamily(c), buf.String())
				buf.Reset()
				contents = append(contents, c)
				continue
			}

			log.Println("char", i, string(runes[i]))

			// buf.Len() == 0
			if i+1 < n && i+2 < n && runes[i+1] == '`' && runes[i+2] == '`' {
				i += 2
				c = content{pdf: mt.pdf, Type: TEXT_CODE}
			} else {
				c = content{pdf: mt.pdf, Type: TEXT_WARP}
				i += 1
			}

			log.Println("c", c)

		case '>':
			if mt.quote && (i-1 > 0 && runes[i+1] == '\n' || i == 0) {
				i++
			} else {
				buf.WriteRune(runes[i])
				i++
			}

		default:
			if buf.Len() == 0 {
				c = content{pdf: mt.pdf, Type: TEXT_NORMAL}
			}
			buf.WriteRune(runes[i])
			i += 1
		}
	}

	for _, c := range contents {
		log.Println("type", c.Type)
		log.Println("text", c.text)
		log.Printf("\n\n+++++++++++++++++++++++++++++++++++\n\n")
	}

	return mt
}

func (mt *MarkdownText) GetWritedLines() int {
	return mt.writedLines
}

func (mt *MarkdownText) GenerateAtomicCell() error {
	if len(mt.contents) == 0 {
		return fmt.Errorf("not set text")
	}

	return nil
}
