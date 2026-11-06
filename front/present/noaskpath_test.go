package present

import (
	"testing"

	"gobox/front"
	"gobox/input"
)

// TestNoPickEnterAnyPath: Enter and Esc through the real Key handler never
// answer a NoPick/NoEsc choice, over a scene's text too; the hotkey does.
func TestNoPickEnterAnyPath(t *testing.T) {
	ask := &front.Ask{Kind: front.AskMenu, NoPick: true, NoEsc: true,
		Menu: front.Menu{Commands: []front.Command{{Text: "Yes", Key: 'Y'}, {Text: "No", Key: 'N'}}}}
	pane := &front.Pane{Lines: [][]front.PaneRun{{{Text: "A CHEST."}}}}
	for i, v := range []front.View{{Ask: ask}, {Ask: ask, Pane: pane}, {Ask: ask, Pane: pane, Exploring: true}} {
		p := &Presenter{}
		p.Key(keyOf(input.Enter), v)
		p.Key(keyOf(input.Esc), v)
		if len(p.inv.acts) != 0 {
			t.Fatalf("view %d: answered %v", i, p.inv.acts)
		}
		p.Key('n', v)
		if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActChoose || p.inv.acts[0].N != 1 {
			t.Fatalf("N: %v", p.inv.acts)
		}
	}
}
