package parity

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"por/rules"
)

// samePage says classic's text page equals por's: the pane, the prompt
// row, the view window's text and the pictures. It counts what it saw.
func samePage(t *testing.T, name string, e *rules.Engine, seen map[string]int) {
	t.Helper()
	v, s := e.View(), e.Rules().State()
	c := classic.View(v, s)
	c = classic.Page(c, s)
	if !reflect.DeepEqual(c.Pane, v.Pane) || !reflect.DeepEqual(c.Prompt, v.Prompt) || !reflect.DeepEqual(c.ViewText, v.ViewText) {
		t.Errorf("%s: text\n got %+v %+v %+v\nwant %+v %+v %+v", name, c.Pane, c.Prompt, c.ViewText, v.Pane, v.Prompt, v.ViewText)
	}
	if !reflect.DeepEqual(c.Pictures, v.Pictures) {
		t.Errorf("%s: pictures\n got %+v\nwant %+v", name, c.Pictures, v.Pictures)
	}
	if v.Pane != nil && v.Pane.Shown() {
		seen["pane"]++
	}
	if len(v.Prompt) > 0 {
		seen["prompt"]++
	}
	if len(v.ViewText) > 0 {
		seen["view"]++
	}
	if len(v.Pictures) > 0 {
		seen["pictures"]++
	}
}

// TestPageParity: the text page at each stop of the dev scenes, Enter on.
func TestPageParity(t *testing.T) {
	seen := map[string]int{}
	for _, scene := range []string{"", "password", "parlay", "who", "journal", "talljournal", "runes", "lockeddoor", "shop", "temple", "appraise", "loot", "effects", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 12; i++ {
			samePage(t, scene, e, seen)
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	for _, k := range []string{"pane", "prompt", "view", "pictures"} {
		if seen[k] == 0 {
			t.Errorf("no %s seen: %v", k, seen)
		}
	}
	t.Logf("page parts compared: %v", seen)
}
