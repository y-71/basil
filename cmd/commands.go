package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"github.com/dwarvesf/glod"
	"github.com/dwarvesf/glod/chiasenhac"
	"github.com/dwarvesf/glod/facebook"
	nct "github.com/dwarvesf/glod/nhaccuatui"
	"github.com/dwarvesf/glod/soundcloud"
	"github.com/dwarvesf/glod/youtube"
	"github.com/dwarvesf/glod/zing"
)

const (
	initNhacCuaTui string = "nhaccuatui"
	initZingMp3    string = "zing"
	initYoutube    string = "youtube"
	initSoundCloud string = "soundcloud"
	initChiaSeNhac string = "chiasenhac"
	initFacebook   string = "facebook"
)

var link string
var directory string
var play bool

// List of
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "Media URL",
		Value: "link",
		Usage: "Input MP3/nhaccuatui/youtube/soundcloud link",
	},
	cli.StringFlag{
		Name:  "Custom output directory",
		Value: "dir",
		Usage: "The directory you want to save",
	},
}

// Action defines the main action for glod-cli
func Action(c *cli.Context) {

	if len(c.Args()) <= 0 {
		cli.ShowAppHelp(c)
		return
	}

	link = c.Args()[0]
	if len(c.Args()) > 1 {
		directory = c.Args()[1]
	}

	if len(c.Args()) > 2 {
		play, _ = strconv.ParseBool(c.Args()[2])
	}

	if link != "" {

		var glod glod.Source

		if strings.Contains(link, initNhacCuaTui) {
			glod = &nct.NhacCuaTui{}
		} else if strings.Contains(link, initZingMp3) {
			glod = &zing.Zing{}
		} else if strings.Contains(link, initYoutube) {
			glod = &youtube.Youtube{}
		} else if strings.Contains(link, initSoundCloud) {
			glod = &soundcloud.SoundCloud{}
		} else if strings.Contains(link, initChiaSeNhac) {
			glod = &chiasenhac.ChiaSeNhac{}
		} else if strings.Contains(link, initFacebook) {
			glod = &facebook.Facebook{}
		}

		fmt.Println("Retrieving metadata ...")

		listStream, err := glod.GetDirectLink(link)
		if err != nil {
			fmt.Println(err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(listStream))

		var barList []*pb.ProgressBar
		var name []string
		var respList []*http.Response

		// Retrieve list of URLs
		for _, l := range listStream {
			temp := l

			_temp := temp
			// if youtube there is a step to split string
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
			} else if strings.Contains(link, initChiaSeNhac) {
				splitName := strings.Split(temp, "~")
				splitNameSplash := strings.Split(splitName[0], "/")
				var nameBeforeSanitize = splitNameSplash[len(splitNameSplash)-1]
				var nameSanitized = strings.Replace(nameBeforeSanitize, "%20", " ", -1)
				bar.Prefix(nameSanitized)
				name = append(name, nameSanitized)
			} else if strings.Contains(link, initFacebook) {
				splitName := strings.Split(link, "/")
				bar.Prefix(splitName[len(splitName)-2] + ".mp4")
				name = append(name, splitName[len(splitName)-2]+".mp4")
			}

			barList = append(barList, bar)
		}

		pool, err := pb.StartPool(barList...)
		if err != nil {
			panic(err)
		}

		// Download list of media files
		fmt.Println("Downloading ...")
		var listFullName []string
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

				listFullName = append(listFullName, fullNameFile)

				out, err := os.Create(fullNameFile)
				defer out.Close()
				if err != nil {
					fmt.Println("Cannot create file")
					return
				}

				io.Copy(out, rd)

				// sleep to perfect progress bar
				time.Sleep(500 * time.Millisecond)
			}()
		}
		wg.Wait()
		pool.Stop()

		if runtime.GOOS == "darwin" {
			fmt.Println("Do you want to play it now?(y/n)")
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')

			if strings.TrimSpace(text) == "y" {
				fmt.Println("Playing...")
				for _, v := range listFullName {
					cmd := exec.Command("afplay", v)
					cmd.Start()
					cmd.Wait()
				}
			}
		}
		// cmd := exec.Command("sh", "-c", "afplay *.mp3")
		fmt.Println("Finish.")
	}
}
