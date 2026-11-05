package audio

import (
	"io"
	"os"
	"sync"
	"testing"

	"github.com/ebitengine/oto/v3"
)

var (
	otoOnce  sync.Once
	otoCtx   *oto.Context
	otoReady <-chan struct{}
	otoMu    sync.Mutex
	otoKeep  []*oto.Player // kept from the garbage collector
)

// Output plays r (signed 16-bit stereo, 44100 Hz) on the default device, once the
// device is ready (in the browser: after the first click or key). All readers share
// one context. Without a device it does nothing.
func Output(r io.Reader) {
	if os.Getenv("GOBOX_NOSOUND") != "" || testing.Testing() { // scripted runs and go test: no device opened, nothing heard
		return
	}
	otoOnce.Do(func() {
		c, ready, err := oto.NewContext(&oto.NewContextOptions{SampleRate: sndRate, ChannelCount: 2, Format: oto.FormatSignedInt16LE})
		if err == nil {
			otoCtx, otoReady = c, ready
		}
	})
	if otoCtx == nil {
		return
	}
	go func() {
		<-otoReady
		p := otoCtx.NewPlayer(r)
		// at least the 2048 frames (46 ms) a browser pulls at once, or the rest of
		// each pull plays as silence
		p.SetBufferSize(sndRate * 4 / 20) // 50 ms of 16-bit stereo
		p.Play()
		otoMu.Lock()
		otoKeep = append(otoKeep, p)
		otoMu.Unlock()
	}()
}
