package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"github.com/dwarvesf/glod"
	nct "github.com/dwarvesf/glod/nhaccuatui"
	"github.com/dwarvesf/glod/soundcloud"
	"github.com/dwarvesf/glod/youtube"
	"github.com/dwarvesf/glod/zing"
)

func main() {
	const (
		initNhacCuaTui string = "nhaccuatui"
		initZingMp3    string = "zing"
		initYoutube    string = "youtube"
		initSoundCloud string = "soundcloud"
	)

	var link string
	var directory string

	app := cli.NewApp()
	app.Name = "glod-cli"
	app.Usage = "Command line tool using glod library to download music/video from multiple source"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "link download",
			Value: "link",
			Usage: "Input zing/nhaccuatui/youtube/soundcloud link",
		},
		cli.StringFlag{
			Name:  "custom directory",
			Value: "dir",
			Usage: "The directory you want to save",
		},
	}
	app.Version = Version

	app.Action = func(c *cli.Context) {
		if len(c.Args()) <= 0 {
			cli.ShowAppHelp(c)
			return
		}
		link = c.Args()[0]
		if len(c.Args()) > 1 {
			directory = c.Args()[1]
		}

	}

	app.Run(os.Args)

	if link != "" {

		var glod glod.Glod

		if strings.Contains(link, initNhacCuaTui) {
			glod = &nct.NhacCuaTui{}
		} else if strings.Contains(link, initZingMp3) {
			glod = &zing.Zing{}
		} else if strings.Contains(link, initYoutube) {
			glod = &youtube.Youtube{}
		} else if strings.Contains(link, initSoundCloud) {
			glod = &soundcloud.SoundCloud{}
		}

		fmt.Println("Downloading...")

		listStream, err := glod.GetDirectLink(link)
		if err != nil {
			fmt.Println(err)
		}

		var wg sync.WaitGroup
		wg.Add(len(listStream))

		var barList []*pb.ProgressBar
		var name []string
		var respList []*http.Response

		for _, l := range listStream {
			temp := l

			_temp := temp
			//if youtube there is a step to split string
			if strings.Contains(link, initYoutube) || strings.Contains(link, initZingMp3) {
				splitUrl := strings.Split(_temp, "~")
				temp = splitUrl[0]
			}

			resp, _ := http.Get(temp)
			defer resp.Body.Close()

			respList = append(respList, resp)

			bar := pb.New(int(resp.ContentLength)).SetUnits(pb.U_BYTES)
			bar.ShowSpeed = true
			bar.ShowTimeLeft = true
			bar.ShowBar = true
			bar.ShowPercent = true

			if strings.Contains(link, initNhacCuaTui) {
				splitName := strings.Split(temp, "/")

				bar.Prefix(splitName[len(splitName)-1])
				name = append(name, splitName[len(splitName)-1])

			} else if strings.Contains(link, initZingMp3) {
				splitName := strings.Split(_temp, "~")
				bar.Prefix(splitName[1] + ".mp3")
				name = append(name, splitName[1]+".mp3")
			} else if strings.Contains(link, initYoutube) {
				splitName := strings.Split(_temp, "~")
				bar.Prefix(splitName[1])
				name = append(name, splitName[1])
			} else if strings.Contains(link, initSoundCloud) {
				splitName := strings.Split(temp, "/")
				bar.Prefix(splitName[4] + ".mp3")
				name = append(name, splitName[4]+".mp3")
			}

			barList = append(barList, bar)
		}

		pool, err := pb.StartPool(barList...)
		if err != nil {
			panic(err)
		}

		for i, bar := range barList {
			_bar := bar
			_i := i
			go func() {
				defer wg.Done()
				rd := _bar.NewProxyReader(respList[_i].Body)

				if _, err := os.Stat(directory); os.IsNotExist(err) {
					os.MkdirAll(directory, 0777)
				}
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				var fullNameFile string

				if directory == "" {
					fullNameFile = name[_i]
				} else {
					fullNameFile = directory + string(filepath.Separator) + name[_i]
				}

				out, err := os.Create(fullNameFile)
				defer out.Close()
				if err != nil {
					fmt.Println("Can not create file")
					return
				}

				io.Copy(out, rd)
				//sleep to perfect progress bar
				time.Sleep(500 * time.Millisecond)
			}()
		}
		wg.Wait()
		pool.Stop()
		fmt.Println("Finish.")

	}
}
