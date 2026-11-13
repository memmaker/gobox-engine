// vace.go decodes SSI's VACE animations (Pools of Darkness INTRO and
// FINAL): the .ACF control script, the .BMA frames and the .SMS songs it
// cues. FORMAT.md has the layouts. No drawing here: frames come out as
// image.Paletted with their palette and how long each stays up.

package ui

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Tick is VACE's timer tick without music (the BIOS clock, 18.2065 Hz).
const Tick = time.Second * 65536 / 1193182

// MusicTick is VACE's tick with music on: SOUND.COM sets the PIT to
// divisor 0x4000 (72.8227 Hz) and VACE counts every INT 8.
const MusicTick = time.Second * 16384 / 1193182

// VaceFrame is one picture shown for Delay.
type VaceFrame struct {
	Image *image.Paletted
	Delay time.Duration
}

// Cue is a music command at At from the start: Op "load" (Song names the
// .SMS file, in Slot), "play" (Slot) or "stop".
type Cue struct {
	At   time.Duration
	Op   string
	Song string
	Slot int
}

// Anim is a decoded animation.
type Anim struct {
	Frames []VaceFrame
	Cues   []Cue
}

// Len is the animation's running time.
func (a *Anim) Len() (d time.Duration) {
	for _, f := range a.Frames {
		d += f.Delay
	}
	return d
}

// Load decodes base.ACF and base.BMA in dir (e.g. "INTRO").
func Load(dir, base string) (*Anim, error) {
	acf, err := os.ReadFile(filepath.Join(dir, base+".ACF"))
	if err != nil {
		return nil, err
	}
	bma, err := os.ReadFile(filepath.Join(dir, base+".BMA"))
	if err != nil {
		return nil, err
	}
	return Decode(acf, bma)
}

type chunk struct{ off, frames int }

// Decode runs the control script acf over the frames bma.
func Decode(acf, bma []byte) (*Anim, error) {
	type rec struct {
		op int
		w  []uint16
	}
	var recs []rec
	for i := 0; i+2 <= len(acf); {
		n := int(acf[i])
		if n < 2 || n%2 != 0 || i+n > len(acf) {
			return nil, fmt.Errorf("vace: bad record at %d", i)
		}
		w := make([]uint16, n/2-1)
		for k := range w {
			w[k] = binary.LittleEndian.Uint16(acf[i+2+2*k:])
		}
		recs = append(recs, rec{int(acf[i+1]), w})
		i += n
	}
	// The chunk table (op 5): chunk i's frame count and byte size, in BMA order.
	var chunks []chunk
	off := 0
	for _, r := range recs {
		if r.op == 5 && len(r.w) >= 3 {
			chunks = append(chunks, chunk{off, int(r.w[1])})
			off += int(r.w[2]) // ponytail: low word only, as VACE; chunks stay under 64K
		}
	}
	if off > len(bma) {
		return nil, errors.New("vace: BMA shorter than its chunk table")
	}
	var (
		a       Anim
		pal     = make(color.Palette, 256)
		canvas  = image.NewPaletted(image.Rect(0, 0, 320, 200), pal)
		delay   time.Duration
		now     time.Duration
		pending *image.Paletted
		next    int // global frame number
	)
	// With music on (VACE 226A) a frame waits 261/arg of the music's ticks,
	// without 62/arg BIOS ticks; a script that loads a song plays with it.
	num, tick := 62, Tick
	for _, r := range recs {
		if r.op == 8 {
			num, tick = 261, MusicTick
		}
	}
	for i := range pal {
		pal[i] = color.RGBA{A: 255}
	}
	flush := func() {
		if pending != nil {
			pending.Palette = append(color.Palette(nil), pal...)
			a.Frames = append(a.Frames, VaceFrame{pending, delay})
			now += delay
			pending = nil
		}
	}
	for _, r := range recs {
		switch r.op {
		case 0: // speed: a frame every num/arg ticks
			if len(r.w) > 0 && r.w[0] != 0 {
				delay = time.Duration(num/int(r.w[0])) * tick
			}
		case 1: // end
			flush()
			return &a, nil
		case 2: // palette entry: hi(w0) index, lo(w0) R, hi(w1) G, lo(w1) B, 6-bit
			flush() // a shown frame keeps the palette it was shown with; these entries are the next frame's
			if len(r.w) >= 2 {
				c6 := func(v uint16) uint8 { v &= 0x3f; return uint8(v<<2 | v>>4) }
				pal[r.w[0]>>8] = color.RGBA{c6(r.w[0]), c6(r.w[1] >> 8), c6(r.w[1]), 255}
			}
		case 3: // show the next frame
			flush()
			c, k := 0, next
			for c < len(chunks) && k >= chunks[c].frames {
				k -= chunks[c].frames
				c++
			}
			if c == len(chunks) {
				continue // past the last frame: VACE skips it
			}
			if err := drawFrame(canvas, bma, chunks[c].off, k); err != nil {
				return nil, fmt.Errorf("vace: frame %d: %w", next, err)
			}
			next++
			pending = &image.Paletted{Pix: append([]uint8(nil), canvas.Pix...), Stride: canvas.Stride, Rect: canvas.Rect}
		case 8: // load song: lo(w0) slot, w1..w4 the name (hi, lo), space padded
			if len(r.w) >= 5 {
				var b []byte
				for _, w := range r.w[1:5] {
					b = append(b, byte(w>>8), byte(w)) // VACE reads hi then lo
				}
				a.Cues = append(a.Cues, Cue{At: now, Op: "load", Song: strings.TrimRight(string(b), " \x00"), Slot: int(r.w[0] & 0xff)})
			}
		case 9: // play the loaded song
			slot := 0
			for _, c := range a.Cues {
				if c.Op == "load" {
					slot = c.Slot
				}
			}
			a.Cues = append(a.Cues, Cue{At: now, Op: "play", Slot: slot})
		case 10:
			a.Cues = append(a.Cues, Cue{At: now, Op: "stop"})
		case 11: // clear the screen to colour 0 (shown with the next frame)
			clear(canvas.Pix)
		}
	}
	flush()
	return &a, nil
}

// drawFrame draws frame k of the chunk at off in bma over dst (a delta:
// only its spans change).
func drawFrame(dst *image.Paletted, bma []byte, off, k int) error {
	if off+6+4*(k+1) > len(bma) {
		return errors.New("chunk header past the end")
	}
	p := off + int(binary.LittleEndian.Uint32(bma[off+6+4*k:]))
	b := bma
	if p+4 > len(b) {
		return errors.New("frame past the end")
	}
	p += 4 // width, height (320x200)
	get := func() (int, error) {
		if p >= len(b) {
			return 0, errors.New("truncated")
		}
		p++
		return int(b[p-1]), nil
	}
	for {
		y, err := get()
		if err != nil {
			return err
		}
		if y == 0xff {
			return nil
		}
		for last := false; !last; {
			if p+4 > len(b) {
				return errors.New("truncated span")
			}
			x := int(binary.LittleEndian.Uint16(b[p:]))
			last = x&0x8000 != 0
			x &= 0x7fff
			n := int(b[p+3]) // data bytes; b[p+2] is the span's pixels
			p += 4
			end := p + n
			if end > len(b) {
				return errors.New("truncated span data")
			}
			row := y * dst.Stride
			put := func(c uint8) {
				if x < dst.Stride && y < dst.Rect.Dy() {
					dst.Pix[row+x] = c
				}
				x++
			}
			for p < end {
				t := int(b[p])
				p++
				cnt := t>>1 + 1
				if t&1 != 0 { // a run of one colour
					if p >= end {
						return errors.New("truncated run")
					}
					for range cnt {
						put(b[p])
					}
					p++
				} else { // literal pixels
					if p+cnt > end {
						return errors.New("truncated literal")
					}
					for _, c := range b[p : p+cnt] {
						put(c)
					}
					p += cnt
				}
			}
		}
	}
}

// Chunk is one block of an .SMS song: Tag "GLBR" (the instruments) or
// "SLBR" (the score), as stored (big-endian 68000 words inside).
type Chunk struct {
	Tag  string
	Data []byte
}
