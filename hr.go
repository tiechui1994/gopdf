package gopdf

import (
	"github.com/tiechui1994/gopdf/core"
)

type HLine struct {
	pdf    *core.Report
	color  float64
	width  float64
	margin Scope
}

func NewHLine(pdf *core.Report) *HLine {
	unit := pdf.GetUnit()
	return &HLine{
		pdf:   pdf,
		color: 0,
		width: 0.1,
		margin: Scope{
			Left:   0,
			Right:  0,
			Top:    0.1 * unit,
			Bottom: 0.1 * unit,
		},
	}
}

func (h *HLine) SetColor(color float64) *HLine {
	if color < 0 || color > 1.0 {
		color = 0
	}

	h.color = color
	return h
}

func (h *HLine) SetMargin(margin Scope) *HLine {
	replaceMarign(&margin)
	h.margin = margin
	return h
}

func (h *HLine) SetWidth(width float64) *HLine {
	h.width = width
	return h
}

func (h *HLine) GenerateAtomicCell() {
	var (
		sx, sy = h.pdf.GetXY()
	)

	x := sx + h.margin.Left
	y := sy + h.margin.Top
	endY := h.pdf.GetPageEndY()
	if (sy >= endY || sy < endY) && sy+h.width > endY {
		h.pdf.AddNewPage(false)
		h.pdf.SetXY(h.pdf.GetPageStartXY())
		h.GenerateAtomicCell()
		return
	}

	cw, _ := h.pdf.GetContentWidthAndHeight()
	h.pdf.GrayColor(x, y, cw, h.width, h.color)

	x, _ = h.pdf.GetPageStartXY()
	h.pdf.SetXY(x, y+h.margin.Bottom+h.width)
}
