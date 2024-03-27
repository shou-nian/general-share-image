package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
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
		log.Fatal(err)
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
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatalf("get share image failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	imgPath := fmt.Sprintf("./_img/_imag%d.png", rd.Int())
	img, err := os.Create(imgPath)
	if err != nil {
		log.Fatalf("create image failed: %v", err)
	}

	_, err = io.Copy(img, res.Body)
	if err != nil {
		log.Fatalf("copy image failed: %v", err)
	}
	err = img.Close()
	if err != nil {
		log.Fatalf("close image failed: %v", err)
	}

	return toImgPath(imgPath)
}

func GeneralImage(imgPath ImgPath, title string, other ...any) ImgPath {
	img, err := gg.LoadImage(string(imgPath))
	if err != nil {
		log.Fatalf("load image failed: %v", err)
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
		log.Fatalf("save image failed: %v", err)
	}

	return imgPath
}

func DeleteLocalStoryImage(path ImgPath) error {
	err := os.Remove(string(path))

	return err
}
