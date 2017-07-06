package commands

import (
	"strings"

	"github.com/dwarvesf/glod"
	"github.com/dwarvesf/glod/chiasenhac"
	"github.com/dwarvesf/glod/facebook"
	"github.com/dwarvesf/glod/nhaccuatui"
	"github.com/dwarvesf/glod/soundcloud"
	"github.com/dwarvesf/glod/vimeo"
	"github.com/dwarvesf/glod/youtube"
	"github.com/dwarvesf/glod/zing"
)

func getGlod(link string) glod.Source {
	switch {
	case strings.Contains(link, initNhacCuaTui):
		return &nhaccuatui.NhacCuaTui{}
	case strings.Contains(link, initZingMp3):
		return &zing.Zing{}
	case strings.Contains(link, initYoutube):
		return &youtube.Youtube{}
	case strings.Contains(link, initSoundCloud):
		return &soundcloud.SoundCloud{}
	case strings.Contains(link, initChiaSeNhac):
		return &chiasenhac.ChiaSeNhac{}
	case strings.Contains(link, initFacebook):
		return &facebook.Facebook{}
	case strings.Contains(link, initVimeo):
		return &vimeo.Vimeo{}
	}

	return nil
}
