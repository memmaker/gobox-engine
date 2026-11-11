// Package creation is gobox's one character-creation flow over a ruleset's
// core.CreationRules, for every game: POR c70cbee's sheet (engine/create.go
// editSheet). The draft lives here; the ruleset only answers legality, rolls
// and keeps. Drawing is the front's (front/classic/creation.go).
package creation

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"gobox/core"
)

// Name limits as POR's inlineInput: printable 0x20..0x7A, at most NameMax.
const (
	NameMax   = 15
	nameFirst = 0x20
	nameLast  = 0x7A
	// quickTries: Quick Start rerolls a member until its Need is met, then
	// keeps the next roll whatever it is (POR quickPick.good).
	quickTries = 2000
)

// Draft is one character on the sheet.
type Draft struct {
	R      core.CreationRules
	RNG    *core.RNG
	D      core.Draft
	Fields []core.FieldDef
}

// New is a fresh character: the ruleset's start (Draft(NoChar): POR a
// human male fighter), each field enforced, rolled.
func New(r core.CreationRules, rng *core.RNG) (*Draft, error) {
	start, err := r.Draft(core.NoChar)
	if err != nil {
		return nil, err
	}
	return begin(r, rng, start.Picks)
}

// begin is a new draft from picks: each field enforced in the ruleset's
// order (a ruleset may build and roll on the first, POR's fresh), then rolled
// unless that rolled already. A field picks lacks starts on its first option.
func begin(r core.CreationRules, rng *core.RNG, picks map[core.FieldID]int) (*Draft, error) {
	d := &Draft{R: r, RNG: rng}
	d.D.Of = core.NoChar
	d.D.Picks = map[core.FieldID]int{}
	d.D.Points = map[core.FieldID][]int32{}
	d.Fields = r.Fields(d.D)
	for _, f := range d.Fields {
		if f.Kind == core.FieldChoice {
			d.D.Picks[f.ID] = picks[f.ID]
		}
	}
	for _, f := range d.Fields {
		if f.Kind == core.FieldChoice {
			d.SetField(f.ID, d.D.Picks[f.ID])
		}
	}
	d.sizeSkills()
	if d.D.Sheet.Scores == nil {
		if err := d.Roll(core.RollFull); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// Modify opens character c on the sheet.
func Modify(r core.CreationRules, rng *core.RNG, c core.CharID) (*Draft, error) {
	cd, err := r.Draft(c)
	if err != nil {
		return nil, err
	}
	cd.Of = c
	return &Draft{R: r, RNG: rng, D: cd, Fields: r.Fields(cd)}, nil
}

// Fits lists which options of field f are legal for the draft (grey: not).
func (d *Draft) Fits(f core.FieldDef) []bool {
	out := make([]bool, len(f.Options))
	for o := range out {
		out[o] = d.R.Fits(d.D, f.ID, o)
	}
	return out
}

// SetField picks option o of f; dependants that no longer fit move (Enforce).
func (d *Draft) SetField(f core.FieldID, o int) {
	d.D = d.R.Enforce(d.D, f, o, d.RNG)
	d.Fields = d.R.Fields(d.D)
	d.sizeSkills()
}

// sizeSkills starts a skills field the picks brought (BR career skills) unspent.
func (d *Draft) sizeSkills() {
	if d.D.Points == nil {
		d.D.Points = map[core.FieldID][]int32{}
	}
	for _, fd := range d.Fields {
		if fd.Kind == core.FieldSkills && len(d.D.Points[fd.ID]) != len(fd.Options) {
			d.D.Points[fd.ID] = make([]int32, len(fd.Options))
		}
	}
}

// Roll rolls the abilities and hit points; a minimum the ruleset cannot
// meet is a *core.MinimumError, the draft unchanged.
func (d *Draft) Roll(k core.RollKind) error { return d.R.Roll(&d.D, k, d.RNG) }

// ErrBudget: a skill spend past the field's budget or below zero.
var ErrBudget = errors.New("no points left")

// Spend moves n points (negative: refunds) onto option o of skills field f.
func (d *Draft) Spend(f core.FieldDef, o int, n int32) error {
	p := d.D.Points[f.ID]
	var spent int32
	for _, v := range p {
		spent += v
	}
	if p[o]+n < 0 || spent+n > f.Budget {
		return ErrBudget
	}
	p[o] += n
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
func (d *Draft) SetScore(a core.Ability, typed string) error {
	v, pct, err := ParseScore(typed)
	if err != nil {
		return err
	}
	nd, err := d.R.SetScore(d.D, a, v)
	if err != nil {
		return err
	}
	if a == core.Strength {
		if nd, err = d.R.SetScore(nd, core.StrengthPercent, pct); err != nil {
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

// FormatScore is ParseScore's inverse.
func FormatScore(v, pct int32) string {
	if pct == 0 {
		return strconv.Itoa(int(v))
	}
	return strconv.Itoa(int(v)) + "/" + strconv.Itoa(int(pct%100)/10) + strconv.Itoa(int(pct%10))
}

// Keep makes the character (POR SaveCharacter).
// A new character goes to the roster (its RosterID), a modified one stays
// the member it is (its CharID).
func (d *Draft) Keep() (core.CharID, core.RosterID, error) { return d.R.Make(d.D) }

// QuickParty makes the ruleset's ready-made party: each member from its
// template, rerolled until its Need is met, kept to the roster and joined.
func QuickParty(r core.CreationRules, rng *core.RNG) ([]core.CharID, error) {
	var out []core.CharID
	for _, t := range r.QuickParty() {
		d, err := begin(r, rng, t.Picks)
		if err != nil {
			return out, err
		}
		for _, f := range d.Fields { // a pick the template leaves open (POR 4e powers): its first fitting option (POR defaultPowers)
			if _, ok := t.Picks[f.ID]; ok || f.Kind != core.FieldChoice || d.D.Picks[f.ID] >= 0 {
				continue
			}
			if o := slices.Index(d.Fits(f), true); o >= 0 {
				d.SetField(f.ID, o)
			}
		}
		for i := 0; i < quickTries && !meets(d.D, t.Need); i++ {
			if err := d.Roll(core.RollFull); err != nil {
				return out, err
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

func meets(d core.Draft, need map[core.Ability]int32) bool {
	for a, v := range need {
		if d.Sheet.Scores[a] < v {
			return false
		}
	}
	return true
}
