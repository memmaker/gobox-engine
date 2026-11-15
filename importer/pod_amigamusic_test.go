package importer

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset/orig"
	"gobox/ecl"
)

// TestAmigaMusic: every Sound n the imported scripts play reaches a pod.mx
// song or a SOUNDS.GLB entry that is there (AmigaMusic, SoundKind).
func TestAmigaMusic(t *testing.T) {
	dos, am := filepath.Join("..", "data", "pod", "sources", "dos"), filepath.Join("..", "data", "pod", "sources", "amiga")
	mx, err1 := os.ReadFile(filepath.Join(am, PodAmigaMusic.Songs))
	glb, err2 := os.ReadFile(filepath.Join(am, PodAmigaMusic.Effects))
	if _, err := os.Stat(dos); err != nil || err1 != nil || err2 != nil {
		t.Skip("no originals")
	}
	m, err := orig.NewMX(mx, glb)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := PodImportScriptsTo(dos, dir); err != nil {
		t.Fatal(err)
	}
	lib := ecl.Library{Version: ecl.V3, Dir: dir, Data: podScriptData(), Fixes: podScriptFixes} // the import's own key: PodImportScriptsTo
	seen := 0
	for area := range 256 {
		if _, err := os.Stat(ecl.AreaPath(dir, area)); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		s, err := lib.Area(area)
		if err != nil {
			t.Fatal(err)
		}
		for _, in := range s.Code {
			if in.Op != ecl.OpSound || in.Args[0].Ref != (ecl.Ref{}) || in.Args[0].Type&0x80 != 0 {
				continue // only literal operands: a variable's sound is the script's at run time
			}
			n := in.Args[0].N
			song, music, digitized := SoundKind(byte(n))
			seen++
			switch {
			case music && (song < 1 || song > m.Len()): // SongOf nil: song n is the file's n-1
				t.Errorf("area %d: Sound(%d): song %d of %d", area, n, song, m.Len())
			case digitized && (n-2 >= len(PodAmigaMusic.EffectOf) || PodAmigaMusic.EffectOf[n-2] >= m.Effects()):
				t.Errorf("area %d: Sound(%d): no effect", area, n)
			}
		}
	}
	t.Logf("%d literal sounds", seen)
	if seen == 0 {
		t.Error("no Sound in the scripts")
	}
}
