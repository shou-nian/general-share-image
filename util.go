package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

type ImgPath string

func toImgPath(path string) ImgPath {
	return ImgPath(path)
}

func CrawlWebsiteStaticHTML(url string) (io.ReadCloser, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func ProfilingTagToMap(body io.ReadCloser) map[string]string {
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			slog.Error("close body failed: ", err.Error())
		}
	}(body)
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	res := func(doc *goquery.Document) map[string]string {
		resMap := make(map[string]string)
		doc.Find("head meta").Each(func(i int, selection *goquery.Selection) {
			nameTag, exists := selection.Attr("name")
			if !exists || !strings.HasPrefix(nameTag, "twitter") {
				return
			}
			contentTag, exists := selection.Attr("content")
			if !exists {
				return
			}
			resMap[nameTag[8:]] = contentTag
		})
		return resMap
	}(doc)

	return res
}

func GetShareImage(url string) ImgPath {
	res, err := http.Get(url)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	if res.StatusCode != http.StatusOK {
		slog.Error("get share image failed: " + err.Error())
		return ""
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(res.Body)

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	imgPath := fmt.Sprintf("./_img/_imag%d.png", rd.Int())
	img, err := os.Create(imgPath)
	if err != nil {
		slog.Error("create image failed: " + err.Error())
		return ""
	}

	_, err = io.Copy(img, res.Body)
	if err != nil {
		slog.Error("copy image failed: " + err.Error())
		return ""
	}
	err = img.Close()
	if err != nil {
		slog.Error("close image failed: " + err.Error())
		return ""
	}

	return toImgPath(imgPath)
}

func GeneralImage(imgPath ImgPath, title string, other ...any) ImgPath {
	img, err := loadImage(imgPath)
	if err != nil {
		slog.Error("load image failed: %v", err)
	}
	ctx := gg.NewContextForImage(img)

	// 加载默认字体
	if err := ctx.LoadFontFace("./_fonts/arial.ttf", 20); err != nil {
		panic(err)
	}
	ctx.SetColor(color.White)

	ctx.DrawString(title, 50, 500)
	err = ctx.SavePNG(string(imgPath))
	if err != nil {
		slog.Error("save image failed: %v" + err.Error())
	}

	return imgPath
}

func DeleteLocalStoryImage(path ImgPath) error {
	err := os.Remove(string(path))

	return err
}

func loadImage(oldImgPath ImgPath) (image.Image, error) {
	img, err := gg.LoadImage(string(oldImgPath))
	if err != nil {
		return nil, err
	}
	// 获取并调整图片大小
	bounds := img.Bounds()
	width, height := bounds.Size().X, bounds.Size().Y
	if width > 1024 || height > 536 {
		// 使用Lanczos3插值算法进行图像缩放
		newImage := resize.Resize(1024, 536, img, resize.Lanczos3)

		// 删除原来的图片
		err := DeleteLocalStoryImage(oldImgPath)
		if err != nil {
			return nil, err
		}
		// 在原路径上创建新图片
		outputFile, err := os.Create(string(oldImgPath))
		if err != nil {
			return nil, err
		}
		defer func(outputFile *os.File) {
			err := outputFile.Close()
			if err != nil {
				slog.Error(err.Error())
			}
		}(outputFile)
		err = png.Encode(outputFile, newImage)
		if err != nil {
			return nil, err
		}
		return newImage, nil
	}

	return img, nil
}
