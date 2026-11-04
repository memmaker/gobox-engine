package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/jfreymuth/oggvorbis"
)

// VorbisPCM decodes an Ogg Vorbis file (mono or stereo, 44100 Hz) to s16
// stereo PCM.
func VorbisPCM(b []byte) ([]byte, error) {
	f, format, err := oggvorbis.ReadAll(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if format.SampleRate != 44100 || format.Channels > 2 {
		return nil, fmt.Errorf("voice: %d Hz, %d channels, want 44100 Hz mono or stereo", format.SampleRate, format.Channels)
	}
	pcm := make([]byte, 0, len(f)/format.Channels*4)
	for i := 0; i < len(f); i += format.Channels {
		l, r := f[i], f[i+format.Channels-1]
		pcm = binary.LittleEndian.AppendUint16(pcm, uint16(int16(max(-1, min(1, l))*32767)))
		pcm = binary.LittleEndian.AppendUint16(pcm, uint16(int16(max(-1, min(1, r))*32767)))
	}
	return pcm, nil
}
