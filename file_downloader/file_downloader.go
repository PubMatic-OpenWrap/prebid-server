package filedownloader

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
)

func DownloadMultipleFiles(urls []string) []error {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
		Transport: &http.Transport{
			ReadBufferSize:  10000,
			WriteBufferSize: 10000,
		},
	}

	t := time.Now()
	if err := os.MkdirAll("/tmp/assests", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, len(urls))
	for _, URL := range urls {
		go func(URL string) {

			// Create blank file
			fileName := strings.Split(URL, "/")
			file, err := os.Create("/tmp/assests/" + fileName[len(fileName)-1])
			if err != nil {
				errCh <- err
				return
			}

			// Put content on file
			resp, err := client.Get(URL)
			if err != nil {
				errCh <- err
				return
			}
			defer resp.Body.Close()

			_, err = io.Copy(file, resp.Body)
			if err != nil {
				errCh <- err
				return
			}
			defer file.Close()

			glog.Infof("Downloaded a file %s\n", fileName)

			errCh <- nil
		}(URL)
	}
	var errs []error
	for i := 0; i < len(urls); i++ {
		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		glog.Infof("Error downloading assets: %v", errs)
		return errs
	}
	glog.Infof("Time taken for Asset download: %v", time.Since(t))
	return nil
}

func RemoveAssests() error {
	return os.RemoveAll("/tmp/assests")
}
