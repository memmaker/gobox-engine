// Command podrock rebuilds the gobox pool's ceiling_pod_rock and
// floor_pod_rock from the original's own indoor backdrop, SKYGRND block 4
// (O12A:019D): its top 44 rows are the rock ceiling, the rest the floor.
// Each part is mirrored both ways into a seamless tile and scaled
// (nearest neighbour, so only the original's colours appear) to 256x256.
//
//	go run ./cmd/podrock [gamedir] [pooldir]
package main

import (
	"fmt"
	"gobox/asset"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"gobox/asset/orig"
)

const ceilRows, size = 44, 256

func main() {
	game, pool := "../../data/pod/extras/gog", "../../walls/series/pool"
	if len(os.Args) > 1 {
		game = os.Args[1]
	}
	if len(os.Args) > 2 {
		pool = os.Args[2]
	}
	l, err := orig.NewLoader(game)
	if err != nil {
		log.Fatal(err)
	}
	b := l.BlockShared("SKYGRND", 4)
	im := orig.ParseVGA(b, false, 0)
	if im == nil || im.Palette == nil {
		log.Fatal("SKYGRND 4: not a VGA picture")
	}
	w, h := im.Width*8, im.Height
	at := func(x, y int) color.RGBA { return im.Palette[im.Data[y*w+x]] }
	for _, p := range []struct {
		name   string
		y0, y1 int
	}{{"ceiling_pod_rock", 0, ceilRows}, {"floor_pod_rock", ceilRows, h}} {
		cw, ch := w, p.y1-p.y0
		out := image.NewRGBA(image.Rect(0, 0, size, size))
		var sum [3]int
		for y := range size {
			for x := range size {
				// mirror tile: 2cw x 2ch, reflected so every edge meets its twin
				mx, my := x*2*cw/size, y*2*ch/size
				if mx >= cw {
					mx = 2*cw - 1 - mx
				}
				if my >= ch {
					my = 2*ch - 1 - my
				}
				c := at(mx, p.y0+my)
				out.Set(x, y, c)
				sum[0], sum[1], sum[2] = sum[0]+int(c.R), sum[1]+int(c.G), sum[2]+int(c.B)
			}
		}
		if err := asset.WritePNG(filepath.Join(pool, p.name+".png"), out); err != nil {
			log.Fatal(err)
		}
		var src [3]int
		for y := p.y0; y < p.y1; y++ {
			for x := range w {
				c := at(x, y)
				src[0], src[1], src[2] = src[0]+int(c.R), src[1]+int(c.G), src[2]+int(c.B)
			}
		}
		n, sn := size*size, cw*ch
		fmt.Printf("%s: %dx%d from %dx%d, mean (%d,%d,%d), original (%d,%d,%d)\n", p.name, size, size, cw, ch,
			sum[0]/n, sum[1]/n, sum[2]/n, src[0]/sn, src[1]/sn, src[2]/sn)
	}
}
