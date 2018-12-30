package gopdf

import (
	"os"
	"image"
	"image/draw"
	"image/png"
	"image/jpeg"
	"image/color"

	"github.com/nfnt/resize"
	"time"
	"fmt"
	"path/filepath"
)

type imageCompress struct {
	quality int // 质量, 0-100
}

func (c *imageCompress) getWidthAndHeight(picturePath string) (w, h int, err error) {
	_, err = os.Stat(picturePath)
	if err != nil {
		return 0, 0, err
	}

	fd, err := os.Open(picturePath)
	if err != nil {
		return 0, 0, err
	}

	config, _, err := image.DecodeConfig(fd)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}

func (c *imageCompress) getPictureType(picturePath string) (pictureType string, err error) {
	_, err = os.Stat(picturePath)
	if err != nil {
		return "", err
	}

	fd, err := os.Open(picturePath)
	if err != nil {
		return "", err
	}

	_, pictureType, err = image.DecodeConfig(fd)
	if err != nil {
		return "", err
	}

	return pictureType, nil
}

func (c *imageCompress) compress(in, out string, width int) (err error) {
	if _, err = os.Stat(in); err != nil {
		return err
	}

	var (
		originImage image.Image
		pictureType string
		reader      *os.File
	)

	reader, err = os.Open(in)
	defer reader.Close()
	if err != nil {
		return err
	}

	if pictureType, err = c.getPictureType(in); err != nil {
		return err
	}

	if pictureType == "jpeg" {
		if originImage, err = jpeg.Decode(reader); err != nil {
			return err
		}
	} else if pictureType == "png" {
		if originImage, err = png.Decode(reader); err != nil {
			return err
		}
	}

	// 等比压缩
	iw, ih, err := c.getWidthAndHeight(in)
	if err != nil {
		return err
	}

	w := uint(width)
	h := uint(width * ih / iw)
	canvas := resize.Thumbnail(w, h, originImage, resize.Lanczos3)

	outFile, err := os.Create(out)
	defer outFile.Close()
	if err != nil {
		return err
	}

	if pictureType == "png" {
		if err = png.Encode(outFile, canvas); err != nil {
			return err
		}
	} else if pictureType == "jpeg" {
		if err = jpeg.Encode(outFile, canvas, &jpeg.Options{Quality: c.quality}); err != nil {
			return err
		}
	}

	return nil
}

var compress *imageCompress

func init() {
	compress = &imageCompress{
		quality: 75,
	}
}

func GetImageWidthAndHeight(inPath string) (w, h int) {
	w, h, _ = compress.getWidthAndHeight(inPath)
	return w, h
}
func GetImageType(inPath string) (imageType string, err error) {
	return compress.getPictureType(inPath)
}

func ImageCompress(inPath string, width, height int) (outPath string) {
	iw, ih, err := compress.getWidthAndHeight(inPath)
	if err != nil {
		panic(err)
	}
	if int(height*iw/ih) < width {
		width = int(height * iw / ih)
	}

	pType, err := compress.getPictureType(inPath)
	if err != nil {
		panic(err)
	}

	timstamp := fmt.Sprintf("%v", time.Now().UnixNano())
	out := filepath.Join(filepath.Dir(inPath), timstamp+"."+pType)
	err = compress.compress(inPath, out, width)
	if err != nil {
		os.Remove(out)
		panic(err)
	}

	return out
}

func ConvertPNG2JPEG(srcPath, dstPath string) (err error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	srcImage, err := png.Decode(srcFile)
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return
	}
	defer dstFile.Close()

	dstImage := image.NewRGBA(srcImage.Bounds())
	draw.Draw(dstImage, dstImage.Bounds(), srcImage, srcImage.Bounds().Min, draw.Src)

	return jpeg.Encode(dstFile, dstImage, nil)
}

func DrawPNG(srcPath string) {
	const (
		width  = 300
		height = 500
	)

	// 文件
	pngFile, _ := os.Create(srcPath)
	defer pngFile.Close()

	// Image, 进行绘图操作
	pngImage := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pngImage.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), uint8((x ^ y) % 256), uint8((x ^ y) % 256)})
		}
	}

	// 以png的格式写入文件
	png.Encode(pngFile, pngImage)
}
