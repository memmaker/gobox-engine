package classic

import (
	"cmp"
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
	if a == nil || s.Mode != game.ModeAsk && !overMenu(s) {
		return nil, true
	}
	k, ok := askKinds[a.Kind]
	if !ok {
		return nil, false
	}
	q := &front.Ask{Kind: k, Menu: front.Menu{Prompt: a.Prompt}, Max: a.Max, Default: a.Default, Break: a.Break, Yes: a.Yes, Text: a.Typed, View: a.View}
	if k == front.AskConfirm && q.Prompt == "" { // the question is in the message: the yes/no still needs a line to be drawn (present.confirm)
		q.Prompt = cmp.Or(a.Text, "Yes or no?")
	}
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

// overMenu: the game keeps ModeExplore under a menu or yes/no it opens over the 3D
// view (camp's): the menu is up, not the exploring.
func overMenu(s game.State) bool {
	return (s.Mode == game.ModeExplore || s.Mode == game.ModeCombat) && s.Pending != nil && (s.Pending.Kind == game.Choice || s.Pending.Kind == game.YesNo) // a fight's own question (a member's View) over the map
}
