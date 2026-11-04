package classic

import (
	"gobox/front"
	"gobox/game"
)

// askKinds are the questions classic rebuilds from game.Ask alone.
var askKinds = map[game.AskKind]front.AskKind{
	game.Choice: front.AskMenu, game.Continue: front.AskMenu, game.Name: front.AskText,
	game.Number: front.AskNumber, game.PickMember: front.AskMember,
}

// Ask is the question from s (nil: none up), and false where classic
// cannot rebuild it from game.Ask (a list's lines, a pause's length): the
// game's stays.
func Ask(s game.State) (*front.Ask, bool) {
	a := s.Pending
	if s.Mode != game.ModeAsk || a == nil {
		return nil, true
	}
	k, ok := askKinds[a.Kind]
	if !ok || a.Kind == game.Continue { // ponytail: Continue is a pause or an empty menu, game.Ask has no Ms to tell them apart
		return nil, false
	}
	q := &front.Ask{Kind: k, Menu: front.Menu{Prompt: a.Prompt}, Max: a.Max, Default: a.Default}
	for _, o := range a.Options {
		q.Items = append(q.Items, front.MenuItem{Text: o.Label, Key: byte(o.Key)})
	}
	return q, true
}
