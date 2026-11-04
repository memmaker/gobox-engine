package classic

import (
	"gobox/front"
	"gobox/game"
)

// askKinds are the questions classic rebuilds from game.Ask alone.
var askKinds = map[game.AskKind]front.AskKind{
	game.Choice: front.AskMenu, game.Continue: front.AskMenu, game.Name: front.AskText,
	game.Number: front.AskNumber, game.PickMember: front.AskMember, game.PickItem: front.AskList,
	game.YesNo: front.AskConfirm, game.Square: front.AskSquare, game.Direction: front.AskMove,
}

// Ask is the question from s (nil: none up), and false where classic
// cannot rebuild it from game.Ask: the game's stays.
func Ask(s game.State) (*front.Ask, bool) {
	a := s.Pending
	if s.Mode != game.ModeAsk || a == nil {
		return nil, true
	}
	k, ok := askKinds[a.Kind]
	if !ok {
		return nil, false
	}
	q := &front.Ask{Kind: k, Menu: front.Menu{Prompt: a.Prompt}, Max: a.Max, Default: a.Default, Break: a.Break, Yes: a.Yes, Text: a.Typed, View: a.View}
	if a.Kind == game.Continue && a.Ms > 0 { // a timed pause; Ms 0: a message to acknowledge
		q.Kind, q.Ms = front.AskPause, a.Ms
	}
	for _, o := range a.Options {
		q.Items = append(q.Items, front.MenuItem{Text: o.Label, Key: byte(o.Key)})
	}
	for _, l := range a.Lines {
		ll := front.ListLine{Text: l.Text, Header: l.Header, Dim: l.Dim, Off: l.Off, Spell: l.Spell}
		for _, k := range l.Hides {
			if k >= 0 && k < len(a.Options) {
				ll.Without = append(ll.Without, a.Options[k].Label)
			}
		}
		q.List = append(q.List, ll)
	}
	return q, true
}
