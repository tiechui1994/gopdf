package gopdf

import (
	"os"
	"github.com/tiechui1994/gopdf/core"
	"path/filepath"
	"strings"
	"math"
)

type Image struct {
	pdf           *core.Report
	path          string
	width, height int
	margin        Scope
}

func NewImage(width, height float64, path string, pdf *core.Report) *Image {
	if _, err := os.Stat(path); err != nil {
		panic("the path error")
	}

	picturePath, _ := filepath.Abs(path)
	imageType, _ := GetImageType(picturePath)
	if imageType == "png" {
		index := strings.LastIndex(picturePath, ".")
		jpegPath := picturePath[0:index] + ".jpeg"
		err := ConvertPNG2JPEG(picturePath, jpegPath)
		if err != nil {
			panic(err)
		}
	}

	w := int(math.Trunc(width))
	h := int(math.Trunc(height))
	picturePath = ImageCompress(picturePath, w, h)
	image := &Image{
		pdf:    pdf,
		path:   picturePath,
		width:  w,
		height: h,
	}

	return image
}

func NewImageWithOutCompress(path string, pdf *core.Report) *Image {
	if _, err := os.Stat(path); err != nil {
		panic("the path error")
	}

	picturePath, _ := filepath.Abs(path)
	imageType, _ := GetImageType(picturePath)
	if imageType == "png" {
		index := strings.LastIndex(picturePath, ".")
		jpegPath := picturePath[0:index] + ".jpeg"
		err := ConvertPNG2JPEG(picturePath, jpegPath)
		if err != nil {
			panic(err)
		}
	}

	w, h := GetImageWidthAndHeight(picturePath)
	image := &Image{
		pdf:    pdf,
		path:   picturePath,
		width:  w,
		height: h,
	}

	return image
}

func (image *Image) GetHeight() float64 {
	return float64(image.height)
}

func (image *Image) SetMargin(margin Scope) *Image {
	image.margin = margin
	image.margin.Right = 0
	image.margin.Bottom = 0
	return image
}

func (image *Image) getImagePostion(sx, sy float64) (x, y float64) {
	x = sx + image.margin.Left
	y = sy + image.margin.Top
	return x, y
}

// 自动换行
func (image *Image) GenerateAtomicCell() error {
	var (
		sx, sy = image.pdf.GetXY()
	)

	x, y := image.getImagePostion(sx, sy)
	pageEndY := image.pdf.GetPageEndY()
	if y < pageEndY && y+float64(image.height) > pageEndY {
		image.pdf.AddNewPage(false)
	}

	image.pdf.Image(image.path, x, y, x+float64(image.width), y+float64(image.height))
	image.pdf.SetXY(sx, y+float64(image.height))
	return nil
}
