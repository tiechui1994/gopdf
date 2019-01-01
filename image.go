package gopdf

import (
	"os"
	"github.com/tiechui1994/gopdf/core"
	"path/filepath"
	"strings"
)

type Image struct {
	pdf           *core.Report
	path          string
	width, height float64
	margin        Scope
	tempFilePath  string
}

func NewImage(path string, pdf *core.Report) *Image {
	if _, err := os.Stat(path); err != nil {
		panic("the path error")
	}

	var tempFilePath string
	picturePath, _ := filepath.Abs(path)
	imageType, _ := GetImageType(picturePath)
	if imageType == "png" {
		index := strings.LastIndex(picturePath, ".")
		tempFilePath = picturePath[0:index] + ".jpeg"
		err := ConvertPNG2JPEG(picturePath, tempFilePath)
		if err != nil {
			panic(err)
		}
		picturePath = tempFilePath
	}

	w, h := GetImageWidthAndHeight(picturePath)
	image := &Image{
		pdf:          pdf,
		path:         picturePath,
		width:        float64(w / 10),
		height:       float64(h / 10),
		tempFilePath: tempFilePath,
	}
	if tempFilePath != "" {
		pdf.AddCallBack(image.deleteTempImage)
	}

	return image
}

func NewImageWithWidthAndHeight(path string, width, height float64, pdf *core.Report) *Image {
	contentWidth, contentHeight := pdf.GetContentWidthAndHeight()
	if width > contentWidth {
		width = contentWidth
	}
	if height > contentHeight {
		height = contentHeight
	}

	if _, err := os.Stat(path); err != nil {
		panic("the path error")
	}

	var tempFilePath string
	picturePath, _ := filepath.Abs(path)
	imageType, _ := GetImageType(picturePath)

	if imageType == "png" {
		index := strings.LastIndex(picturePath, ".")
		tempFilePath = picturePath[0:index] + ".jpeg"
		err := ConvertPNG2JPEG(picturePath, tempFilePath)
		if err != nil {
			panic(err.Error())
		}
		picturePath = tempFilePath
	}

	w, h := GetImageWidthAndHeight(picturePath)
	if float64(h)*width/float64(w) > height {
		width = float64(w) * height / float64(h)
	} else {
		height = float64(h) * width / float64(w)
	}
	image := &Image{
		pdf:          pdf,
		path:         picturePath,
		width:        width,
		height:       height,
		tempFilePath: tempFilePath,
	}

	if tempFilePath != "" {
		pdf.AddCallBack(image.deleteTempImage)
	}

	return image
}

func (image *Image) GetHeight() float64 {
	return image.height
}

func (image *Image) GetWidth() float64 {
	return image.width
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
	sx, _ = image.pdf.GetPageStartXY()
	image.pdf.SetXY(sx, y+float64(image.height))
	return nil
}

func (image *Image) deleteTempImage(report *core.Report) {
	if image.tempFilePath == "" {
		return
	}

	if _, err := os.Stat(image.tempFilePath); err != nil {
		return
	}

	os.Remove(image.tempFilePath)
}
