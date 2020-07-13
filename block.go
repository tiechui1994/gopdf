package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
	"strings"
	"github.com/tiechui1994/gopdf/util"
	"math"
	"log"
)

type Block interface {
}

type TextBlock struct {
	pdf  *core.Report
	font core.Font

	contents     []string
	width        float64
	x            float64
	startY, endY float64
	lineHeight   float64
	lineWidth    float64
}

func NewTextBlock(pdf *core.Report, font core.Font) (*TextBlock) {
	x1, _ := pdf.GetPageStartXY()
	x2, _ := pdf.GetPageEndXY()

	x1 += 10.0

	tb := &TextBlock{
		pdf:        pdf,
		font:       font,
		x:          x1,
		width:      x2 - x1,
		lineHeight: 18.0,
		lineWidth:  3.0,
	}
	return tb
}

func (tb *TextBlock) SetContent(content string) *TextBlock {
	convertStr := strings.Replace(content, "\t", "    ", -1)

	var (
		blocks       = strings.Split(convertStr, "\n") // 分行
		contentWidth = tb.width
	)

	// 必须检查字体
	if util.IsEmpty(tb.font) {
		panic("there no avliable font")
	}

	tb.pdf.Font(tb.font.Family, tb.font.Size, tb.font.Style)
	tb.pdf.SetFontWithStyle(tb.font.Family, tb.font.Style, tb.font.Size)
	for _, block := range blocks {
		if tb.pdf.MeasureTextWidth(block) <= contentWidth {
			tb.contents = append(tb.contents, block)
			continue
		}

		tokens := splitN(tb.pdf, tb.font, block, contentWidth)
		tb.contents = append(tb.contents, tokens...)
	}

	return nil
}

func (tb *TextBlock) GenerateAtomicCell() error {
	_, y1 := tb.pdf.GetXY()
	y2 := y1 + float64(len(tb.contents))*tb.lineHeight
	x, _ := tb.pdf.GetPageStartXY()

	tb.pdf.LineGrayColor(x, y1, 3, y2-y1, 0.6)

	tb.pdf.Font(tb.font.Family, tb.font.Size, tb.font.Style)
	tb.pdf.SetFontWithStyle(tb.font.Family, tb.font.Style, tb.font.Size)
	y := y1
	for _, block := range tb.contents {
		tb.pdf.Cell(tb.x, y, block)
		y += tb.lineHeight
	}

	return nil
}

func splitN(pdf *core.Report, font core.Font, content string, width float64) []string {
	pdf.Font(font.Family, font.Size, font.Style)
	pdf.SetFontWithStyle(font.Family, font.Style, font.Size)
	length := pdf.MeasureTextWidth(content)
	if length <= width {
		return []string{content}
	}

	var blocks []string
	runes := []rune(content)
	step := int(float64(len(runes)) * width / length)
	precision := pdf.MeasureTextWidth(content) / float64(len(runes))
	for i, j := 0, step; i < len(runes) && j < len(runes); {
		w := pdf.MeasureTextWidth(string(runes[i:j]))
		// last split
		if w < width && j == len(runes)-1 {
			blocks = append(blocks, string(runes[i:j]))
			break
		}

		log.Println(math.Abs(w-width), precision, i, j)

		// less than precision
		if math.Abs(w-width) < precision {
			// try again, can more precise
			if j+1 < len(runes) {
				w1 := pdf.MeasureTextWidth(string(runes[i:j+1]))
				if math.Abs(w1-width) < precision {
					j = j + 1
					continue
				}
			}

			blocks = append(blocks, string(runes[i:j]))
			i = j
			j += step
			if j >= len(runes) {
				j = len(runes) - 1
			}
		}

		if w > width {
			j--
		} else {
			j++
		}
	}

	return blocks
}
