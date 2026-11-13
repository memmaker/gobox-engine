package blobber

import (
	"image"
	"image/color"
	"math"

	"gobox/gpu"
)

// The sun and the far mountains of the open sky (the original draws
// a sun picture by the hour and a mountain range when facing north, in
// daylight). Bearings are compass radians (0 north, pi/2 east).
var (
	texSun     *gpu.Image
	texMount   *gpu.Image
	sunRadius  = 5 * math.Pi / 180
	mountWidth = 100 * math.Pi / 180 // the range's span across the north
)

// drawCelestial draws the sun and the mountains over the sky backdrop.
func drawCelestial(dst *gpu.Image, c cam, hy float64) {
	v := cur
	if v == nil || v.Night {
		return
	}
	if v.Mountains {
		if texMount == nil {
			texMount = gpu.NewImageFromImage(paintMountains(1024, 128))
		}
		// the range spans bearings -mountWidth/2..mountWidth/2, drawn in vertical strips
		const strips = 32
		b := texMount.Bounds()
		hgt := focal * math.Tan(7*math.Pi/180)
		for i := 0; i < strips; i++ {
			a0 := -mountWidth/2 + mountWidth*float64(i)/strips
			a1 := a0 + mountWidth/strips
			d0, d1 := wrap(a0-c.ang), wrap(a1-c.ang)
			if math.Abs(d0) > 1.2 || math.Abs(d1) > 1.2 {
				continue
			}
			x0, x1 := float64(W)/2+focal*math.Tan(d0), float64(W)/2+focal*math.Tan(d1)
			sub := texMount.SubImage(image.Rect(b.Dx()*i/strips, 0, b.Dx()*(i+1)/strips, b.Dy())).(*gpu.Image)
			op := &gpu.DrawImageOptions{}
			op.GeoM.Scale((x1-x0+1)/float64(b.Dx()/strips), hgt/float64(b.Dy()))
			op.GeoM.Translate(x0, hy-hgt+1)
			dst.DrawImage(sub, op)
		}
	}
	if v.SunUp {
		d := wrap(v.SunAz - c.ang)
		if math.Abs(d) > 1.3 {
			return
		}
		if texSun == nil {
			texSun = gpu.NewImageFromImage(paintSun(128))
		}
		x := float64(W)/2 + focal*math.Tan(d)
		y := hy - focal*math.Tan(v.SunEl)/math.Cos(d)
		r := 3 * focal * math.Tan(sunRadius) // the texture's glow reaches 3 radii
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(2*r/128, 2*r/128)
		op.GeoM.Translate(x-r, y-r)
		dst.DrawImage(texSun, op)
	}
}

// wrap brings an angle into -pi..pi.
func wrap(a float64) float64 { return math.Remainder(a, 2*math.Pi) }

// paintSun: a white-gold disc of a third of the size in a soft glow.
func paintSun(n int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			r := math.Hypot(float64(x)-float64(n)/2+0.5, float64(y)-float64(n)/2+0.5) / (float64(n) / 6) // 1 at the disc's edge
			var a float64
			switch {
			case r < 0.95:
				a = 1
			case r < 1.05:
				a = 1 - (r-0.95)/0.1*0.6
			default:
				a = 0.4 * math.Pow(1/r, 3)
			}
			g := 255 - 40*math.Min(r, 1)
			img.SetNRGBA(x, y, color.NRGBA{255, uint8(g), uint8(g - 60*math.Min(r, 1)), uint8(255 * math.Min(a, 1))})
		}
	}
	return img
}

// paintMountains: two hazy ridges of fractal peaks with snow on the tops.
func paintMountains(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	ridge := func(seed float64, x int) float64 { // 0..1 height of the ridge at x
		v, amp, f := 0.0, 0.5, 1.0
		for o := 0; o < 6; o++ {
			t := float64(x) / float64(w) * f * 6
			v += amp * (math.Sin(t*1.7+seed*3.1+math.Sin(t*0.6+seed)*2)*0.5 + 0.5)
			amp, f = amp*0.5, f*2.1
		}
		edge := math.Sin(math.Pi * float64(x) / float64(w)) // the range fades at its ends
		return v * (0.35 + 0.65*math.Sqrt(edge)) * edge
	}
	for x := 0; x < w; x++ {
		back, front := ridge(1.3, x)*0.95, ridge(4.7, x)*0.7
		for y := 0; y < h; y++ {
			hgt := 1 - float64(y)/float64(h) // 1 at the top
			var c color.NRGBA
			switch {
			case hgt < front:
				c = color.NRGBA{112, 124, 140, 255}
				if front-hgt < 0.12*front && front > 0.45 {
					c = color.NRGBA{226, 232, 240, 255}
				}
			case hgt < back:
				c = color.NRGBA{150, 166, 186, 255}
				if back-hgt < 0.15*back && back > 0.5 {
					c = color.NRGBA{236, 240, 246, 255}
				}
			default:
				continue
			}
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}
