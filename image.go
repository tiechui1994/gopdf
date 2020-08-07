package gopdf

import (
	"os"
	"path/filepath"
	"github.com/tiechui1994/gopdf/core"
	"fmt"
	"log"
	"time"
)

type Image struct {
	pdf           *core.Report
	path          string
	width, height float64
	margin        core.Scope
	tempFilePath  string
}

func NewImage(path string, pdf *core.Report) *Image {
	if _, err := os.Stat(path); err != nil {
		panic(fmt.Sprintf("the path error, %v", path))
	}

	dstPath := fmt.Sprintf("/tmp/%v.jpeg", time.Now().UnixNano())
	srcPath, _ := filepath.Abs(path)
	err := Convert2JPEG(srcPath, dstPath)
	if err != nil {
		log.Println(err)
		return nil
	}

	w, h := GetImageWidthAndHeight(dstPath)
	image := &Image{
		pdf:          pdf,
		path:         dstPath,
		width:        float64(w),
		height:       float64(h),
		tempFilePath: dstPath,
	}
	if dstPath != "" {
		pdf.AddCallBack(image.delTempImage)
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

	dstPath := fmt.Sprintf("/tmp/%v.jpeg", time.Now().UnixNano())
	srcPath, _ := filepath.Abs(path)
	err := Convert2JPEG(srcPath, dstPath)
	if err != nil {
		return nil
	}

	w, h := GetImageWidthAndHeight(dstPath)
	if width > 0 && height > 0 {
		if float64(h)*width/float64(w) > height {
			width = float64(w) * height / float64(h)
		} else {
			height = float64(h) * width / float64(w)
		}
	} else if width > 0 {
		height = float64(h) * width / float64(w)
	} else if height > 0 {
		width = float64(w) * height / float64(h)
	}

	image := &Image{
		pdf:          pdf,
		path:         dstPath,
		width:        width,
		height:       height,
		tempFilePath: dstPath,
	}

	if dstPath != "" {
		pdf.AddCallBack(image.delTempImage)
	}

	return image
}

func (image *Image) SetMargin(margin core.Scope) *Image {
	margin.ReplaceMarign()
	image.margin = margin
	return image
}

func (image *Image) GetHeight() float64 {
	return image.height
}
func (image *Image) GetWidth() float64 {
	return image.width
}


func (image *Image) GenerateAtomicCell() error {
	var (
		sx, sy = image.pdf.GetXY()
	)

	x, y := sx+image.margin.Left, sy+image.margin.Top
	pageEndX, pageEndY := image.pdf.GetPageEndXY()
	if y < pageEndY && y+float64(image.height) > pageEndY {
		image.pdf.AddNewPage(false)
	}

	image.pdf.Image(image.path, x, y, x+float64(image.width), y+float64(image.height))
	if x+float64(image.width) >= pageEndX {
		sx, _ = image.pdf.GetPageStartXY()
		image.pdf.SetXY(sx, y+float64(image.height)+image.margin.Bottom)
	} else {
		image.pdf.SetXY(x+float64(image.width), y)
	}

	return nil
}

func (image *Image) delTempImage(report *core.Report) {
	if image.tempFilePath == "" {
		return
	}

	if _, err := os.Stat(image.tempFilePath); err != nil {
		return
	}

	os.Remove(image.tempFilePath)
}
