package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/dwarvesf/glod-cli/cmd"
)

func main() {
	app := cli.NewApp()
	app.Name = "glod-cli"
	app.Version = "1.0.3.2"
	app.Usage = "A small cli written in Go to help download music/video from multiple sources."
	app.Email = "dev@dwarvesf.com"
	app.Action = cmd.Action
	app.Flags = cmd.Flags
	app.Run(os.Args)
}
