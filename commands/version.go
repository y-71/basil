package commands

import (
	"github.com/Sirupsen/logrus"
)

const (
	VERSION = "1.1.0"
)

func runVersion() {
	logrus.Printf("Glod-cli's version %s", VERSION)
}
