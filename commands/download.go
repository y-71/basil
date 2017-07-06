package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/dwarvesf/glod"
)

func runDownload() {
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

	_, err = downloadWithProgressBar(Link, listResponse, Output)
	if err != nil {
		logrus.WithError(err).Error("failed to download")
		return
	}

	logrus.Println("Finish.")
}

func downloadWithProgressBar(link string, listResponse []glod.Response, dir string) ([]ObjectResponse, error) {

	objs, err := getResponseListThenCreateFiles(link, listResponse, dir)
	if err != nil {
		logrus.WithError(err).Error("failed to create objects response")
		return nil, err
	}

	writeWithProgress(objs)

	return objs, nil
}

func downloadWithoutProgressBar(link string, listResponse []glod.Response, dir string) ([]ObjectResponse, error) {

	objs, err := getResponseListThenCreateFiles(link, listResponse, dir)
	if err != nil {
		logrus.WithError(err).Error("failed to create objects response")
		return nil, err
	}

	write(objs)

	return objs, nil
}
