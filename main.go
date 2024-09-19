package main

import (
	"log/slog"
	"net/http"
	"regexp"
	"sync"
)

func main() {
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		slog.Info(r.Method + " " + r.RequestURI)

		url := r.FormValue("url")
		if ok, _ := regexp.MatchString(`(?m)(?P<origin>(?P<protocol>https?:)?//(?P<host>[a-z0-9A-Z-_.]+))(?P<port>:\d+)?(?P<path>[/a-zA-Z0-9-.]+)?(?P<search>\?[^#\n]+)?(?P<hash>#.*)?`, url); !ok {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("url is invalid, please check again."))
			return
		}
		if ok, _ := regexp.MatchString("aspecta.ai", url); !ok {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("please enter aspecta website url."))
			return
		}

		body, err := CrawlWebsiteStaticHTML(url)
		if err != nil {
			slog.Error(err.Error())
		}
		res, err := ProfilingTagToMap(body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("this url not found twitter share image or og protoc, please check again."))
			return
		}
		p := GetShareImage(res["image"])
		if p == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("this url not found share image, please check again."))
			return
		}
		p = GeneralImage(p, res["title"])
		defer func(path ImgPath) {
			err := DeleteLocalStoryImage(path)
			if err != nil {
				slog.Error(err.Error())
			}
		}(p)

		http.ServeFile(w, r, string(p))
		return
	})

	// index
	http.Handle("/", http.FileServer(http.Dir("static")))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := http.ListenAndServe("0.0.0.0:8080", nil)
		if err != nil {
			slog.Error(err.Error())
		}
	}()
	wg.Wait()
}
