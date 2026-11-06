package app

import "testing"

// optsFront has its Esc menu open: the script's keys must reach it, not Key.
type optsFront struct {
	Front
	pad, keys []Key
	updates   int
}

func (f *optsFront) Tick()            {}
func (f *optsFront) Act()             {}
func (f *optsFront) Options() bool    { return true }
func (f *optsFront) OptionsPad(k Key) { f.pad = append(f.pad, k) }
func (f *optsFront) UpdateOptions()   { f.updates++ }
func (f *optsFront) Key(k Key)        { f.keys = append(f.keys, k) }

func TestScriptReachesOptions(t *testing.T) {
	f := &optsFront{}
	g := &frontGame{f: f}
	for _, k := range ParseKeys(`\D\r`, nil) {
		g.Step(k)
	}
	if len(f.pad) != 2 || f.updates != 2 || len(f.keys) != 0 {
		t.Fatalf("pad %v updates %d keys %v", f.pad, f.updates, f.keys)
	}
}
