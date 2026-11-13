package ui

import (
	"cmp"
	"gobox/game"
	"gobox/game/play"
)

// askKinds are the questions classic rebuilds from play.Ask alone.
var askKinds = map[game.AskKind]AskKind{
	game.Choice: AskMenu, game.Continue: AskMenu, game.Name: AskText,
	game.Number: AskNumber, game.PickMember: AskMember, game.PickItem: AskList,
	game.YesNo: AskConfirm, game.Square: AskSquare, game.Direction: AskMove,
}

// Ask is the question from s (nil: none up), and false where classic
// cannot rebuild it from play.Ask: the game's stays.
func Ask(mode game.Mode, a *play.Ask) (*FrontAsk, bool) {
	if a == nil || mode != game.ModeAsk && !OverMenu(mode, a) {
		return nil, true
	}
	k, ok := askKinds[a.Kind]
	if !ok {
		return nil, false
	}
	q := &FrontAsk{Kind: k, Menu: Menu{Prompt: a.Prompt}, Max: a.Max, Default: a.Default, Break: a.Break, Yes: a.Yes, Text: a.Typed, View: a.View, NoEsc: a.NoEsc, NoPick: a.NoPick}
	if a.Kind == game.Name && q.Max <= 0 { // no length given (0): the Gold Box name's 15, never a line that takes no letter
		q.Max = 15
	}
	if k == AskConfirm && q.Prompt == "" { // the question is in the message: the yes/no still needs a line to be drawn (present.confirm)
		q.Prompt = cmp.Or(a.Text, "Yes or no?")
	}
	if a.Kind == game.Continue && a.Ms > 0 { // a timed pause; Ms 0: a message to acknowledge
		q.Kind, q.Ms = AskPause, a.Ms
	}
	for _, o := range a.Options {
		q.Commands = append(q.Commands, Command{Text: o.Label, Key: byte(o.Key)})
	}
	for _, l := range a.Lines {
		ll := FrontEntry{Text: l.Text, Header: l.Header, Dim: l.Dim, Off: l.Off, Spell: l.Spell}
		for _, k := range l.Hides {
			if k >= 0 && k < len(a.Options) {
				ll.Without = append(ll.Without, a.Options[k].Label)
			}
		}
		q.Entries = append(q.Entries, ll)
	}
	return q, true
}

// overMenu: the game keeps ModeExplore under a menu or yes/no it opens over the 3D
// view (camp's): the menu is up, not the exploring.
func OverMenu(mode game.Mode, pending *play.Ask) bool {
	return pending != nil && (pending.Kind == game.YesNo || pending.Kind == game.Name || pending.Kind == game.Number) || // a yes/no or a typed line over any screen (the treasure's "claim it?", a shop's drop, the name after creation's Keep)
		(mode == game.ModeExplore || mode == game.ModeCombat) && pending != nil && (pending.Kind == game.Choice || pending.Kind == game.YesNo) // a fight's own question (a member's View) over the map
}
