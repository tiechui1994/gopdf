package gopdf

import (
	"os"
	"github.com/tiechui1994/gopdf/core"
)

type Image struct {
	pdf           *core.Report
	path          string
	width, height float64
	margin        Scope
}

func NewImage(width, height float64, path string, pdf *core.Report) *Image {
	if _, err := os.Stat(path); err != nil {
		panic("the path error")
	}

	image := &Image{
		pdf:    pdf,
		path:   path,
		width:  width,
		height: height,
	}

	return image
}

func (image *Image) GetHeight() float64 {
	return image.height
}

func (image *Image) SetHeight(height float64) {
	image.height = height
}

func (image *Image) SetMargin(margin Scope) *Image {
	image.margin = margin
	if margin.Top != 0 {
		margin.Bottom = 0
	}
	return image
}

func (image *Image) SetBorder(border Scope) {
}

// 自动换行
func (image *Image) GenerateAtomicCell() error {
	return nil
}
