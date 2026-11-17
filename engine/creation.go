// creation.go is gobox's one character-creation flow over a ruleset's
// core.CreationRules, for every game: POR c70cbee's sheet (engine/create.go
// editSheet). The draft lives here; the ruleset only answers legality, rolls
// and keeps. Drawing is the front's (front/classic/creation.go).

package engine

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"gobox/rules"
)

// Name limits as POR's inlineInput: printable 0x20..0x7A, at most NameMax.
const (
	NameMax   = 15
	nameFirst = 0x20
	nameLast  = 0x7A
)

// Draft is one character on the sheet.
type Draft struct {
	R      rules.CreationRules
	RNG    *rules.RNG
	D      rules.Draft
	Fields []rules.FieldDef
}

// NewDraft is a fresh character: the ruleset's start (Draft(NoChar): POR a
// human male fighter), each field enforced, rolled.
func NewDraft(r rules.CreationRules, rng *rules.RNG) (*Draft, error) {
	start, err := r.Draft(rules.NoChar)
	if err != nil {
		return nil, err
	}
	return begin(r, rng, start.Picks)
}

// begin is a new draft from picks: each field enforced in the ruleset's
// order (a ruleset may build and roll on the first, POR's fresh), then rolled
// unless that rolled already. A field picks lacks starts on its first option.
func begin(r rules.CreationRules, rng *rules.RNG, picks map[rules.FieldID]int) (*Draft, error) {
	d := &Draft{R: r, RNG: rng}
	d.D.Of = rules.NoChar
	d.D.Picks = map[rules.FieldID]int{}
	d.D.Points = map[rules.FieldID][]int32{}
	d.Fields = r.Fields(d.D)
	for _, f := range d.Fields {
		if f.Kind == rules.FieldChoice {
			d.D.Picks[f.ID] = picks[f.ID]
		}
	}
	for _, f := range d.Fields {
		if f.Kind == rules.FieldChoice {
			d.SetField(f.ID, d.D.Picks[f.ID])
		}
	}
	set := map[rules.FieldID]bool{}
	for again := true; again; { // a field a pick opens (XXVc: a warrior's weapon) takes its pick too, once
		again = false
		for _, f := range d.Fields {
			if p, ok := picks[f.ID]; ok && f.Kind == rules.FieldChoice && !set[f.ID] && d.D.Picks[f.ID] != p {
				set[f.ID] = true
				d.SetField(f.ID, p)
				again = true
				break
			}
		}
	}
	d.sizeSkills()
	if d.D.Sheet.Scores == nil {
		if err := d.Roll(rules.RollFull); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// Modify opens character c on the sheet.
func Modify(r rules.CreationRules, rng *rules.RNG, c rules.CharID) (*Draft, error) {
	cd, err := r.Draft(c)
	if err != nil {
		return nil, err
	}
	cd.Of = c
	return &Draft{R: r, RNG: rng, D: cd, Fields: r.Fields(cd)}, nil
}

// Fits lists which options of field f are legal for the draft (grey: not).
func (d *Draft) Fits(f rules.FieldDef) []bool {
	out := make([]bool, len(f.Options))
	for o := range out {
		out[o] = d.R.Fits(d.D, f.ID, o)
	}
	return out
}

// SetField picks option o of f; dependants that no longer fit move (Enforce).
func (d *Draft) SetField(f rules.FieldID, o int) {
	d.D = d.R.Enforce(d.D, f, o, d.RNG)
	d.Fields = d.R.Fields(d.D)
	d.sizeSkills()
}

// sizeSkills starts a skills field the picks brought (BR career skills) unspent.
func (d *Draft) sizeSkills() {
	if d.D.Points == nil {
		d.D.Points = map[rules.FieldID][]int32{}
	}
	for _, fd := range d.Fields {
		if fd.Kind == rules.FieldSkills && len(d.D.Points[fd.ID]) != len(fd.Options) {
			d.D.Points[fd.ID] = make([]int32, len(fd.Options))
		}
	}
}

// Roll rolls the abilities and hit points; a minimum the ruleset cannot
// meet is a *core.MinimumError, the draft unchanged.
func (d *Draft) Roll(k rules.RollKind) error { return d.R.Roll(&d.D, k, d.RNG) }

// ErrPointBuy: bought scores outside core's point-buy range or budget.
var ErrPointBuy = errors.New("Point buy over budget or out of range")

// PointBuy makes the abilities from the bases bought (core.PointBuySpent),
// then the ruleset's roll rules (Roll(RollPointBuy)); refused, the draft stays.
func (d *Draft) PointBuy(bases map[rules.Ability]int32) error {
	if _, ok := rules.PointBuySpent(bases); !ok || len(bases) == 0 {
		return ErrPointBuy
	}
	was := d.D.Buy
	d.D.Buy = bases
	err := d.Roll(rules.RollPointBuy)
	d.D.Buy = was
	return err
}

// Spend moves one point (by 1: onto, -1: back) on option o of the points field f: the ruleset's rules decide
// (rules.CreationRules.Spend); its refusal comes back, the draft unchanged.
func (d *Draft) Spend(f rules.FieldID, o int, by int32) error {
	nd, err := d.R.Spend(d.D, f, o, by)
	if err != nil {
		return err
	}
	d.D = nd
	d.Fields = d.R.Fields(d.D)
	return nil
}

// Rename sets the name typed, keeping only the characters the original took.
func (d *Draft) Rename(s string) {
	var b strings.Builder
	for _, c := range s {
		if c >= nameFirst && c <= nameLast && b.Len() < NameMax {
			b.WriteRune(c)
		}
	}
	d.D.Name = b.String()
}

// ErrScore: the typed text is no score.
var ErrScore = errors.New("not a score")

// SetScore sets ability a from its typed text: "15", or for strength "18/50"
// ("18/00" is 100).
func (d *Draft) SetScore(a rules.Ability, typed string) error {
	v, pct, err := ParseScore(typed)
	if err != nil {
		return err
	}
	nd, err := d.R.SetScore(d.D, a, v)
	if err != nil {
		return err
	}
	if a == rules.Strength {
		if nd, err = d.R.SetScore(nd, rules.StrengthPercent, pct); err != nil {
			return err
		}
	}
	d.D = nd
	return nil
}

// ParseScore reads "15" (pct 0) or "18/50" (pct 50; "18/5": 5; "18/00": 100);
// scores run 0..25 and only 18 takes a percent (POR parseMinStat).
func ParseScore(s string) (v, pct int32, err error) {
	s = strings.TrimSpace(s)
	base, ex, has := strings.Cut(s, "/")
	n, e := strconv.Atoi(base)
	if e != nil || n < 0 || n > 25 {
		return 0, 0, ErrScore
	}
	if !has {
		return int32(n), 0, nil
	}
	p, e := strconv.Atoi(ex)
	if e != nil || p < 0 || p > 99 || n != 18 {
		return 0, 0, ErrScore
	}
	if p == 0 {
		p = 100
	}
	return int32(n), int32(p), nil
}

// Keep makes the character (POR SaveCharacter).
// A new character goes to the roster (its RosterID), a modified one stays
// the member it is (its CharID).
func (d *Draft) Keep() (rules.CharID, rules.RosterID, error) { return d.R.Make(d.D) }

// QuickParty makes the ruleset's ready-made party: each member from its
// template, rerolled until its Need is met, kept to the roster and joined.
func QuickParty(r rules.CreationRules, rng *rules.RNG) ([]rules.CharID, error) {
	var out []rules.CharID
	for _, t := range r.QuickParty() {
		d, err := begin(r, rng, t.Picks)
		if err != nil {
			return out, err
		}
		for _, f := range d.Fields { // a pick the template leaves open (POR 4e powers): its first fitting option (POR defaultPowers)
			if _, ok := t.Picks[f.ID]; ok || f.Kind != rules.FieldChoice || d.D.Picks[f.ID] >= 0 {
				continue
			}
			if o := slices.Index(d.Fits(f), true); o >= 0 {
				d.SetField(f.ID, o)
			}
		}
		for i := 0; i < r.QuickRerolls() && !meets(d.D, t.Need); i++ {
			if err := d.Roll(rules.RollFull); err != nil {
				return out, err
			}
		}
		var points []rules.FieldID // the sheet's order: one points field may open another's options
		for _, f := range d.Fields {
			points = append(points, f.ID)
		}
		for _, f := range points { // a point on each in turn while any goes (buckrogers-go quickSkills)
			opts := t.Spend[f]
			for more := len(opts) > 0; more; {
				more = false
				for _, o := range opts {
					if d.Spend(f, o, 1) == nil {
						more = true
					}
				}
			}
		}
		d.D.Name, d.D.Head, d.D.Body, d.D.Icon.Head, d.D.Icon.Body = t.Name, t.Head, t.Body, t.IconHead, t.IconBody
		_, id, err := d.Keep()
		if err != nil {
			return out, err
		}
		c, err := r.Join(id)
		if err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, nil
}

func meets(d rules.Draft, need map[rules.Ability]int32) bool {
	for a, v := range need {
		if d.Sheet.Scores[a] < v {
			return false
		}
	}
	return true
}
