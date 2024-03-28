package main

import (
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

func main() {
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		slog.Info(r.Method + " " + r.RequestURI)

		url := r.FormValue("url")
		if !strings.HasPrefix(url, "") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("url is required"))
			return
		}

		if ok, _ := regexp.MatchString("aspecta.id", url); !ok {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("please enter aspecta website url."))
			return
		}

		body, err := CrawlWebsiteStaticHTML(url)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("url is invalid."))
			return
		}

		res := ProfilingTagToMap(body)
		if len(res) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("this url not found twitter tag, please try again."))
			return
		}
		p := GetShareImage(res["image"])
		if p == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("unknown error, please try again."))
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
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			slog.Error(err.Error())
		}
	}()
	wg.Wait()
}
