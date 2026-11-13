package playtest

import (
	"os"
	"testing"
	"time"

	"gobox/audio"
	"gobox/engine"
	"gobox/engine/playpool"
	"gobox/ui"
)

// Start-up as main does it, headless: each step and the first frames must be quick.
func TestStartupTiming(t *testing.T) {
	if err := os.Chdir(".."); err != nil {
		t.Skip(err)
	}
	start, t0 := time.Now(), time.Now()
	defer func() {
		if d := time.Since(start); d > time.Second {
			t.Errorf("start-up to the fifth frame took %v, want well under a second", d)
		}
	}()
	lap := func(what string) { t.Logf("%-16s %v", what, time.Since(t0)); t0 = time.Now() }
	eng, err := engine.NewPool("../../data/por/sources/assets", "../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	lap("pool.New")
	tables, err := eng.SoundData()
	if err != nil {
		t.Fatal(err)
	}
	drv, err := audio.NewDriver(tables, eng.Cfg.Sound)
	if err != nil {
		t.Fatal(err)
	}
	am, fx, pack, speaking, vol := eng.SoundParts()
	drv.Attach(am, fx, ui.NewSoundtrack(eng.SoundtrackDir(), engine.SoundtrackFiles), pack, speaking, vol)
	eng.SetSound(drv)
	lap("sound")
	data := playpool.Data(eng, "por")
	lap("Data")
	g := ui.NewFront(ui.NewGame(eng, data), data, drv)
	lap("NewFront")
	g.Start("")
	for i := range 5 {
		g.Frame(1280, 800)
		lap("frame " + string(rune('0'+i)))
	}
}
