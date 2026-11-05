package audio

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"

	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/oggvorbis"
)

// MusicPCM decodes a music file (name's extension: .mp3 or .ogg, any rate,
// mono or stereo) to s16 stereo PCM at 44100 Hz.
func MusicPCM(name string, b []byte) ([]byte, error) {
	var st []float32 // interleaved stereo, -1..1
	var rate int
	if strings.HasSuffix(strings.ToLower(name), ".mp3") {
		d, err := mp3.NewDecoder(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(d) // s16 stereo
		if err != nil && len(raw) == 0 {
			return nil, err
		}
		rate = d.SampleRate()
		st = make([]float32, len(raw)/2)
		for i := range st {
			st[i] = float32(int16(binary.LittleEndian.Uint16(raw[2*i:]))) / 32768
		}
	} else {
		f, format, err := oggvorbis.ReadAll(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		rate, st = format.SampleRate, make([]float32, 0, len(f)/format.Channels*2)
		for i := 0; i+format.Channels-1 < len(f); i += format.Channels {
			st = append(st, f[i], f[i+min(1, format.Channels-1)])
		}
	}
	return resample(st, rate), nil
}

// resample is interleaved stereo st at rate as s16 stereo at sndRate,
// linearly interpolated.
// ponytail: linear interpolation, a windowed-sinc filter if the aliasing is ever heard.
func resample(st []float32, rate int) []byte {
	frames := len(st) / 2
	if rate <= 0 || frames == 0 {
		return []byte{}
	}
	n := int(int64(frames) * sndRate / int64(rate))
	out := make([]byte, 0, n*4)
	for i := range n {
		x := float64(i) * float64(rate) / sndRate
		j := min(int(x), frames-1)
		k := min(j+1, frames-1)
		t := float32(x - float64(j))
		for c := range 2 {
			v := st[2*j+c]*(1-t) + st[2*k+c]*t
			out = binary.LittleEndian.AppendUint16(out, uint16(int16(max(-1, min(1, v))*32767)))
		}
	}
	return out
}
