package library

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

const (
	maxAttempts    = 4
	maxConcurrents = 10
)

var sem = make(chan struct{}, maxConcurrents)

func bulkDownload(list []string, output string, progressFunc func(float32)) error {
	var errFlag bool
	var wg sync.WaitGroup

	total := len(list)
	downloaded := int32(0)
	for _, v := range list {
		wg.Add(1)
		go func(link string) {
			defer func() {
				wg.Done()
				d := float32(atomic.AddInt32(&downloaded, 1))
				progressFunc(d / float32(total))
			}()

			var err error
			for i := 0; i < maxAttempts; i++ {
				sem <- struct{}{}
				err = download(link, output)
				<-sem
				if err == nil {
					break
				}
			}
			if err != nil {
				log.Printf("Failed to download: %s", err)
				errFlag = true
			}
		}(v)
	}
	wg.Wait()
	progressFunc(1)

	if errFlag {
		return errors.New("Lack of aac files")
	}
	return nil
}

func download(link, output string) error {
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, fileName := filepath.Split(link)
	file, err := os.Create(filepath.Join(output, fileName))
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}
