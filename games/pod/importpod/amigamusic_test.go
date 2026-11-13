package importpod

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"gobox/amiga"
	eclscript "gobox/ecl/script"
)

// TestAmigaMusic: every Sound n the imported scripts play reaches a pod.mx
// song or a SOUNDS.GLB entry that is there (AmigaMusic, SoundKind).
func TestAmigaMusic(t *testing.T) {
	dos, am := filepath.Join("..", "..", "..", "data", "pod", "sources", "dos"), filepath.Join("..", "..", "..", "data", "pod", "sources", "amiga")
	mx, err1 := os.ReadFile(filepath.Join(am, AmigaMusic.Songs))
	glb, err2 := os.ReadFile(filepath.Join(am, AmigaMusic.Effects))
	if _, err := os.Stat(dos); err != nil || err1 != nil || err2 != nil {
		t.Skip("no originals")
	}
	m, err := amiga.NewMX(mx, glb)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := ImportScriptsTo(dos, dir); err != nil {
		t.Fatal(err)
	}
	lib := eclscript.Library{Version: eclscript.V3, Dir: dir, Data: scriptData()}
	seen := 0
	for area := range 256 {
		if _, err := os.Stat(eclscript.AreaPath(dir, area)); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		s, err := lib.Area(area)
		if err != nil {
			t.Fatal(err)
		}
		for _, in := range s.Code {
			if in.Op != eclscript.OpSound || in.Args[0].Ref != (eclscript.Ref{}) || in.Args[0].Type&0x80 != 0 {
				continue // only literal operands: a variable's sound is the script's at run time
			}
			n := in.Args[0].N
			song, music, digitized := SoundKind(byte(n))
			seen++
			switch {
			case music && (song < 1 || song > m.Len()): // SongOf nil: song n is the file's n-1
				t.Errorf("area %d: Sound(%d): song %d of %d", area, n, song, m.Len())
			case digitized && (n-2 >= len(AmigaMusic.EffectOf) || AmigaMusic.EffectOf[n-2] >= m.Effects()):
				t.Errorf("area %d: Sound(%d): no effect", area, n)
			}
		}
	}
	t.Logf("%d literal sounds", seen)
	if seen == 0 {
		t.Error("no Sound in the scripts")
	}
}
