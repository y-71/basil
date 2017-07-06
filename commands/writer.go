package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Sirupsen/logrus"
	"github.com/cheggaaa/pb"
	"github.com/dwarvesf/glod"
)

func write(objs []ObjectResponse) {
	var wg sync.WaitGroup
	wg.Add(len(objs))
	for _, v := range objs {
		go func(o ObjectResponse) {
			defer wg.Done()
			defer o.Resp.Body.Close()

			out, err := os.OpenFile(o.Name, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			defer out.Close()
			if err != nil {
				logrus.WithError(err).Errorf("cannot open file, file name = %s", o.Name)
				return
			}

			_, err = io.Copy(out, v.Resp.Body)
			if err != nil {
				logrus.WithError(err).Errorf("cannot copy file, file name = %s", o.Name)
				return
			}
		}(v)
	}
	wg.Wait()
}

func writeWithProgress(objs []ObjectResponse) {
	var wg sync.WaitGroup
	wg.Add(len(objs))

	var barList []*pb.ProgressBar
	for _, o := range objs {
		bar := pb.New(int(o.Resp.ContentLength)).SetUnits(pb.U_BYTES)
		bar.ShowSpeed = true
		bar.ShowTimeLeft = true
		bar.ShowBar = true
		bar.ShowPercent = true
		bar.Prefix(o.Name)

		barList = append(barList, bar)
	}

	pool, err := pb.StartPool(barList...)
	if err != nil {
		logrus.WithError(err).Error("cannot start pool")
		return
	}

	for i, v := range objs {
		go func(o ObjectResponse) {
			defer wg.Done()
			defer o.Resp.Body.Close()

			rd := barList[i].NewProxyReader(o.Resp.Body)

			out, err := os.OpenFile(o.Name, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			defer out.Close()
			if err != nil {
				logrus.WithError(err).Errorf("cannot open file, file name = %s", o.Name)
				return
			}

			_, err = io.Copy(out, rd)
			if err != nil {
				logrus.WithError(err).Errorf("cannot copy file, file name = %s", o.Name)
				return
			}

			time.Sleep(500 * time.Millisecond)
		}(v)
	}
	pool.Stop()
	wg.Wait()
}

func getResponseListThenCreateFiles(link string, listResponse []glod.Response, dir string) ([]ObjectResponse, error) {
	var objs []ObjectResponse
	var wg sync.WaitGroup
	wg.Add(len(listResponse))

	if dir != "" {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0777)
		}
	}

	for _, v := range listResponse {
		go func(r glod.Response) {
			defer wg.Done()
			temp := r.StreamURL
			// if youtube there is a step to split string
			if strings.Contains(link, initYoutube) || strings.Contains(link, initZingMp3) || strings.Contains(link, initVimeo) {
				splitUrl := strings.Split(temp, "~")
				temp = splitUrl[0]
			}

			resp, err := http.Get(temp)
			if err != nil {
				logrus.WithError(err).Error("failed to get response from  stream")
				return
			}

			fullName := fmt.Sprintf("%s%s", r.Title, ".mp3")
			if dir != "" {
				fullName = fmt.Sprintf("%s%s%s%s", dir, string(filepath.Separator), r.Title, ".mp3")
			}

			fullName = strings.Map(func(r rune) rune {
				if unicode.IsSpace(r) {
					return -1
				}
				return r
			}, fullName)

			out, err := os.Create(fullName)
			defer out.Close()
			if err != nil {
				logrus.WithError(err).Error("cannot create file")
				return
			}
			objs = append(objs, ObjectResponse{
				resp,
				fullName,
			})
		}(v)
	}
	wg.Wait()

	return objs, nil
}
