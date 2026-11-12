package pool

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gobox/ecl/script"
)

// The area event scripts are text (data/assets/scripts/area<N>.txt, README.md
// there): the import decompiles the ECL byte code (gobox/ecl/script) and the
// engine runs that text.

func (e *Engine) scriptDir() string { return filepath.Join(e.adv, "scripts") }

// loadScript parses area's script into e.vm.Script and lays its tables out
// in vm.Code. A script that is missing or wrong is reported and runs as an
// empty one (every event exits).
func (e *Engine) loadScript(area int) {
	clear(e.vm.Code)
	s, err := e.parseScriptFile(area)
	if errors.Is(err, script.ErrNotImported) {
		panic(err) // the import must run (ImportScripts): no area plays without its script
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "script:", err)
		s = &script.Script{Area: area, Code: []script.Instr{{Op: 0}}, Labels: map[string]int{}, At: map[int]string{}}
	}
	e.vm.Script = s
	copy(e.vm.Code, s.Data)
	e.ext.stopAt, e.ext.interact = map[int]bool{}, nil
	for _, l := range interactions(e.scriptDir())[area] {
		if i, ok := s.Labels[l]; ok {
			e.ext.stopAt[i] = true
		}
	}
	e.purseWho = map[int]bool{}
	for _, w := range purseWhos {
		if i, ok := s.Labels[w.label]; ok && w.area == area {
			e.purseWho[i+w.by] = true
		}
	}
	e.native = map[int]func(*Engine){}
	for at, f := range nativeScripts {
		if at.area == area {
			if i, ok := s.Labels[at.label]; ok {
				e.native[i] = f
			}
		}
	}
}

func (e *Engine) parseScriptFile(area int) (*script.Script, error) {
	l := script.Library{Version: script.V1, Dir: e.scriptDir(), Data: os.DirFS(e.scriptDir()), Fixes: ScriptFixes}
	return l.Area(area)
}
