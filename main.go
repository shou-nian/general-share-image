package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"image/color"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	var url = "https://aspecta.id/u/xihe"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	resultMap := func(doc *goquery.Document) map[string]string {
		result := make(map[string]string)
		doc.Find("head meta").Each(func(i int, selection *goquery.Selection) {
			nameTag, exists := selection.Attr("name")
			if !exists || !strings.HasPrefix(nameTag, "twitter:") {
				return
			}
			contentTag, exists := selection.Attr("content")
			if !exists {
				return
			}
			result[nameTag[8:]] = contentTag
			return
		})
		return result
	}(doc)
	fmt.Println(resultMap)

	p := getShareImage(resultMap["image"])
	fmt.Println(p)

	p = generalImage(p, resultMap["title"])
	err = deleteLocalStoryImage(p)
	if err != nil {
		slog.Info(err.Error())
	}
	fmt.Println("done!")
}

type ImgPath string

func toImgPath(path string) ImgPath {
	return ImgPath(path)
}

func getShareImage(url string) ImgPath {
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

func deleteLocalStoryImage(path ImgPath) error {
	err := os.Remove(string(path))

	return err
}

func generalImage(imgPath ImgPath, title string, other ...any) ImgPath {
	img, err := gg.LoadImage(string(imgPath))
	if err != nil {
		log.Fatalf("load image failed: %v", err)
	}

	ctx := gg.NewContextForImage(img)
	// 加载系统默认字体
	if err := ctx.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 48); err != nil {
		panic(err)
	}
	ctx.SetColor(color.White)

	ctx.DrawString(title, 100, 100)
	err = ctx.SavePNG(string(imgPath))
	if err != nil {
		log.Fatalf("save image failed: %v", err)
	}

	return imgPath
}
