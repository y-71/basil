package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

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

	app := cli.NewApp()
	app.Name = "glod-cli"
	app.Usage = "Command line tool using glod library to download music/video from multiple source"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "link download",
			Value: "link",
			Usage: "Input zing/nhaccuatui/youtube/soundcloud link",
		},
	}
	app.Version = Version

	app.Action = func(c *cli.Context) {
		link = c.Args()[0]
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

		for _, l := range listStream {
			temp := l
			go func() {

				defer wg.Done()
				_temp := temp
				//if youtube there is a step to split string
				if strings.Contains(link, initYoutube) || strings.Contains(link, initZingMp3) {
					splitUrl := strings.Split(_temp, "~")
					temp = splitUrl[0]
				}

				resp, _ := http.Get(temp)
				defer resp.Body.Close()

				data, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					if strings.Contains(link, initNhacCuaTui) {
						splitName := strings.Split(temp, "/")
						ioutil.WriteFile(splitName[len(splitName)-1], data, 0644)
					} else if strings.Contains(link, initZingMp3) {
						splitName := strings.Split(_temp, "~")
						ioutil.WriteFile(splitName[1]+".mp3", data, 0644)
					} else if strings.Contains(link, initYoutube) {
						splitName := strings.Split(_temp, "~")
						ioutil.WriteFile(splitName[1], data, 0644)
					} else if strings.Contains(link, initSoundCloud) {
						splitName := strings.Split(temp, "/")
						ioutil.WriteFile(splitName[4]+".mp3", data, 0644)
					}
				}
			}()
		}
		wg.Wait()
		fmt.Println("Done.")

	}
}
