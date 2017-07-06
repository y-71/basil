package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/goguard/goguard/version"
)

const (
	VERSION = "1.1.0"
)

func runVersion() {
	logrus.Printf("Glod-cli's version %s", version.VERSION)
}
