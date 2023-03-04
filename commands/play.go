package commands

import (
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

func runPlay() {
	if Link == "" {
		logrus.Error("link is empty")
		return
	}

	glod := getGlod(Link)
	if glod == nil {
		logrus.Error("source not yet supported")
		return
	}

	logrus.Println("Retrieving metadata ...")

	listResponse, err := glod.GetDirectLink(Link)
	if err != nil {
		logrus.WithError(err).Error("failed to get direct link")
		return
	}

	objs, err := downloadWithoutProgressBar(Link, listResponse, PlayDir)
	if err != nil {
		logrus.WithError(err).Error("failed to download")
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cleanUp()
		play(objs)
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-c:
		cleanUp()
	}

	wg.Wait()

	logrus.Println("Finish.")
}

func play(objs []ObjectResponse) {
	if runtime.GOOS != "darwin" {
		logrus.Info("sorry, this functionality only supports MacOS currently")
		return
	}

	logrus.Println("Playing...")
	for _, v := range objs {
		cmd := exec.Command("afplay", v.Name)
		cmd.Start()
		cmd.Wait()
	}
}
func cleanUp() {
	if err := os.RemoveAll(PlayDir); err != nil {
		logrus.WithError(err).Error("failed to remove dir")
	}
}
