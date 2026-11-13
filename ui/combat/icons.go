package combat

import (
	"embed"
	"image"
	"image/png"
	"strings"
)

//go:embed all:icons
var iconFiles embed.FS

var icons = map[string]image.Image{} // by series and button word; nil: no icon, the letter shows

// iconPicture is series' art for the button named word (nil: none).
func iconPicture(series, word string) image.Image {
	key := series + "/" + word
	if im, ok := icons[key]; ok {
		return im
	}
	name := strings.ToLower(word)
	if word == "Cancel" {
		name = "end"
	}
	var im image.Image
	if f, err := iconFiles.Open("icons/" + series + "/" + name + ".png"); err == nil {
		im, _ = png.Decode(f)
		f.Close()
	}
	icons[key] = im
	return im
}
