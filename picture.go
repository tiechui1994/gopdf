package gopdf

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"fmt"
	"io"

	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
	"golang.org/x/image/tiff"
	"math"
)

const (
	PNG  = "png"
	BMP  = "bmp"
	WEBP = "webp"
	TIFF = "tiff"
	JPEG = "jpeg"
)

func Convert2JPEG(srcPath string, dstPath string) error {
	_, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	fd, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	_, pictureType, err := image.DecodeConfig(fd)
	if err != nil {
		return err
	}

	switch pictureType {
	case JPEG:
		writer, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, fd)
		return err
	case PNG:
		return ConvertPNG2JPEG(srcPath, dstPath)
	case WEBP:
		return ConvertWEBP2JPEG(srcPath, dstPath)
	case BMP:
		return ConvertWEBP2JPEG(srcPath, dstPath)
	case TIFF:
		return ConvertTIFF2JPEG(srcPath, dstPath)
	default:
		return fmt.Errorf("invalid picture type")
	}
}

func GetImageWidthAndHeight(picturePath string) (w, h int) {
	var err error
	_, err = os.Stat(picturePath)
	if err != nil {
		panic("the image path: " + picturePath + " not exist")
	}

	fd, err := os.Open(picturePath)
	if err != nil {
		panic("open image error")
	}

	config, _, err := image.DecodeConfig(fd)
	if err != nil {
		panic("decode image error")
	}

	return config.Width, config.Height
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

func ConvertBMP2JPEG(srcPath, dstPath string) (err error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	srcImage, err := bmp.Decode(srcFile)
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

func ConvertWEBP2JPEG(srcPath, dstPath string) (err error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	srcImage, err := webp.Decode(srcFile)
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

func ConvertTIFF2JPEG(srcPath, dstPath string) (err error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	srcImage, err := tiff.Decode(srcFile)
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
		width  = 200
		height = 200
	)

	// 文件
	pngFile, _ := os.Create(srcPath)
	defer pngFile.Close()

	// Image, 进行绘图操作
	pngImage := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(pngImage, pngImage.Bounds(), image.White, image.ZP, draw.Src)

	a, b, c := 5.70, 7.0, 2.2
	for p := 0.0; p <= 720.0; p += 0.03125 {
		x := int(20*((a-b)*math.Cos(p)+c*math.Cos((a/b-1)*p))) + 100
		y := int(20*((a-b)*math.Sin(p)-c*math.Sin((a/b-1)*p))) + 100

		pngImage.Set(x, y, color.RGBA{
			R: uint8(250),
			G: uint8(4),
			B: uint8(4),
			A: uint8(255),
		})
	}

	// 以 png 的格式写入文件
	png.Encode(pngFile, pngImage)
}
