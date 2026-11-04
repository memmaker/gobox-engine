package parity

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gobox/game"
	"por/rules"
)

// The State fields no golden renders, as por publishes them over the dev
// scenes. They were checked as the pieces of por's View classic builds from
// them; they are asserted on game.State itself now. Each stop of a scene is
// one line of wantState: the mode, the pace, and a hash of the groups below.

// groups are the State fields s holds, each hashed on its own.
func groups(s game.State) (out []string) {
	s = still(s)
	reach := 0
	if s.Combat != nil { // the squares in reach vary from run to run (por's own); how many do not
		c := *s.Combat
		reach, c.Reach = len(c.Reach), nil
		s.Combat = &c
	}
	s.Note.At = time.Time{}
	pace := -1
	if s.Pacing != nil { // Seq counts the states published: not a fact of the screen
		p := *s.Pacing
		pace, p.Seq = int(p.Pace), 0
		s.Pacing = &p
	}
	for _, g := range []struct {
		name string
		v    any
	}{
		{"pending", s.Pending}, {"holding", s.Holding}, {"powers", s.PowerPick}, {"door", s.Door}, // what the game waits on
		{"flags", s.Flags}, {"intro", s.Intro}, {"pacing", s.Pacing}, {"screen", []any{s.Cells, s.Ink}}, // status, title, pace, page size, colours
		{"page", s.Page}, {"read", s.Read}, {"note", s.Note.Text}, {"results", s.Results}, {"walls", s.Walls},
		{"place", []any{s.Where, s.Window}}, {"quests", s.Quests}, {"lore", s.Lore}, {"party", []any{s.Party, s.Selected, s.Purse}},
		{"sheet", []any{s.Sheet, s.SheetPage, s.Portrait, s.IconBox}}, {"items", []any{s.Items, s.PoolItems, s.PartyCap, s.Allowed}},
		{"shop", []any{s.Shop, s.Temple, s.TempleName, s.TempleOffers}}, {"loot", s.Loot}, {"map", s.Map},
		{"camp", []any{s.Bestiary, s.RestHeals, s.RestStudies}}, {"combat", []any{s.Combat, reach}}, {"training", s.Training},
		{"picker", s.Picker}, {"icon", []any{s.IconEditor, s.IconPalette}}, {"creation", s.Creation},
	} {
		out = append(out, g.name+"="+sum(g.v))
	}
	return append([]string{fmt.Sprintf("m%d p%d", s.Mode, pace)}, out...)
}

// still is s with the page's picture frames at 0: they follow the wall clock
// (the camp's animated picture), not the rules.
func still(s game.State) game.State {
	if s.Page != nil {
		p := *s.Page
		p.Layers = append([]game.Layer(nil), p.Layers...)
		for i := range p.Layers {
			p.Layers[i].Frame = 0
		}
		s.Page = &p
	}
	return s
}

// unsetLate are contract fields added after the hashes were taken, as JSON
// at their zero value: left out of the hash while unset, so a new field the
// rules do not fill changes nothing.
var unsetLate = []string{
	`"Reloads":false,"Pooled":null,`, // Shop
	`,"View":false`,                  // Ask
}

// sum is the first bytes of the hash of v's JSON.
func sum(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "unmarshalable: " + err.Error()
	}
	for _, f := range unsetLate {
		b = bytes.ReplaceAll(b, []byte(f), nil)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:3])
}

var stopsSeen = map[*testing.T]map[string]int{}

// at asserts the State at e's stop name equals the one por published there
// (wantState), the stops of one test counted in the order they come.
func at(t *testing.T, name string, e *rules.Engine) {
	t.Helper()
	if stopsSeen[t] == nil {
		stopsSeen[t] = map[string]int{}
		t.Cleanup(func() { delete(stopsSeen, t) })
	}
	stopsSeen[t][name]++
	atKey(t, fmt.Sprintf("%s/%s#%d", t.Name(), name, stopsSeen[t][name]), groups(e.Rules().State()))
}

// atKey is at for the groups g of the stop key.
func atKey(t *testing.T, key string, g []string) {
	t.Helper()
	got := g[0] + " " + sum(g[1:])
	if f := os.Getenv("I2_PROBE"); f != "" {
		l, _ := os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		fmt.Fprintf(l, "\t%q: %q,\n", key, got)
		l.Close()
		return
	}
	if want, ok := wantState[key]; !ok || got != want {
		t.Errorf("%s: State is %q, por published %q\n%s", key, got, want, strings.Join(g, "\n"))
	}
}

// The art the classic front draws, as por holds it: Data.Art's border pieces
// and the party menu's title and backdrop.
const (
	wantBorder = "ef7552"
	wantTitle  = "7ce7ab"
)
