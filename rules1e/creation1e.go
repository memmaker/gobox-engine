package rules1e

import (
	"errors"
	"fmt"
	"slices"

	"gobox/core"
	"gobox/core/goldbox"
)

// Creation is POR's core.CreationRules (ONE-BINARY.md Track C): the
// creation's legality, rolls and keeping over create.go, so that gobox's
// one flow (gobox/creation) holds the draft. A draft converts to a Character
// and back without dice (charOf, draftOf); only Enforce, Roll and Draft(NoChar)
// draw, in the order CreateCharacter and SetField draw.
type Creation struct {
	S    *RuleState
	Host CreationHost
	r    *Rules // the ids (Rules.Creation)
}

// CreationHost is goldbox.CreationHost (the engine side of the creation).
type CreationHost = goldbox.CreationHost

// as hands the selected member to the rules as the actor.
func (p Creation) as() *RuleState {
	p.S.Who = p.Host.Selected()
	return p.S
}

// firstPower is the FieldID of the first power pick (extra: rules4e); pick i is firstPower+i.
const firstPower core.FieldID = FieldClass + 1

// noPick is a power pick not answered yet (powerSlot.pick -1).
const noPick = -1

// index is the party position of the member c names (the ruleset's ids, as every core.Rules call).
func (p Creation) index(c core.CharID) (int, bool) {
	i := slices.Index(p.S.Party, p.r.Char(c))
	return i, c != core.NoChar && i >= 0
}

// with runs f on the RNG r (the Engine's Turbo Pascal seed is the same generator).
func (p Creation) with(r *core.RNG, f func()) {
	if r != nil {
		p.S.RandSeed = r.State
	}
	f()
	if r != nil {
		r.State = p.S.RandSeed
	}
}

func KindOf(d core.Draft) Kind {
	return Kind{
		Race:  PickRaces[d.Picks[FieldRace]],
		Sex:   byte(d.Picks[FieldSex]),
		Class: PickClasses()[d.Picks[FieldClass]],
		Align: byte(d.Picks[FieldAlign]),
	}
}

func (p Creation) Fields(d core.Draft) []core.FieldDef {
	races := make([]string, len(PickRaces))
	for i, r := range PickRaces {
		races[i] = RaceNames[r]
	}
	classes := PickClasses()
	cls := make([]string, len(classes))
	for i, cl := range classes {
		cls[i] = ClassNames[cl]
	}
	out := []core.FieldDef{
		{ID: FieldSex, Label: "Gender", Options: GenderNames[:]},
		{ID: FieldRace, Label: "Race", Options: races},
		{ID: FieldAlign, Label: "Alignment", Options: AlignmentNames[:]},
		{ID: FieldClass, Label: "Class", Options: cls},
	}
	if d.Sheet.Race == "" {
		return out
	}
	for i, s := range p.S.PowerSlots(p.CharOf(d)) {
		names := make([]string, len(s.Options))
		for x, m := range s.Options {
			names[x] = m.Name
		}
		out = append(out, core.FieldDef{ID: firstPower + core.FieldID(i), Label: s.What, Options: names})
	}
	return out
}

func (p Creation) Fits(d core.Draft, f core.FieldID, o int) bool {
	k := KindOf(d)
	switch f {
	case FieldSex:
		return true
	case FieldRace:
		return ClassFor(PickRaces[o], k.Class)
	case FieldAlign:
		return AlignFor(k.Class, byte(o))
	case FieldClass:
		cl := PickClasses()[o]
		return ClassFor(k.Race, cl) && AlignFor(cl, k.Align)
	}
	c := p.CharOf(d)
	ss := p.S.PowerSlots(c)
	i := int(f - firstPower)
	if i < 0 || i >= len(ss) || o < 0 || o >= len(ss[i].Options) {
		return false
	}
	id := ss[i].Options[o].ID
	return !slices.Contains(c.Powers, id) || (ss[i].Pick >= 0 && ss[i].Options[ss[i].Pick].ID == id)
}

func (p Creation) Enforce(d core.Draft, f core.FieldID, o int, r *core.RNG) core.Draft {
	var out core.Draft
	p.with(r, func() {
		if d.Sheet.Race == "" { // nothing built yet (Quick Start): the kind's fresh character, rolled
			d.Picks[f] = o
			pi := map[core.FieldID]int{FieldRace: FRace, FieldSex: FSex, FieldClass: FClass, FieldAlign: FAlign}[f]
			c := p.S.Fresh(KindOf(d).Enforce(pi), nil)
			p.S.RollCharacter(c)
			out = p.DraftOf(c, d.Of)
			out.Name, out.Head, out.Body, out.Portrait = d.Name, d.Head, d.Body, d.Portrait
			out.Icon.Head, out.Icon.Body, out.Icon.PNG = d.Icon.Head, d.Icon.Body, d.Icon.PNG
			return
		}
		cr := &SheetChar{C: p.CharOf(d), K: KindOf(d)}
		p.S.SetSheetField(cr, int(f), o)
		out = p.DraftOf(cr.C, d.Of)
	})
	return out
}

func (p Creation) Roll(d *core.Draft, k core.RollKind, r *core.RNG) (err error) {
	if k == core.RollPointBuy {
		for _, a := range Abilities {
			if _, ok := d.Buy[a]; !ok {
				return fmt.Errorf("point buy: no %v bought", a)
			}
		}
	}
	p.with(r, func() {
		c := p.CharOf(*d)
		switch {
		case k == core.RollPointBuy: // the bought bases, then the roll's rules and hit points (extra: gobox's Point Buy)
			p.S.SetBaseStats(c, func(i int) byte { return byte(d.Buy[Abilities[i]]) })
			p.S.RollHP(c)
		case k == core.RollMinimal:
			was := p.S.MinStats
			for i, a := range Abilities {
				if v, ok := d.Lows[a]; ok {
					p.S.MinStats[i] = int(v)
					if i == 0 {
						p.S.MinStats[0] = StrScore(Score(v), byte(d.Lows[core.StrengthPercent]))
					}
				}
			}
			if err = p.S.RollMinimal(c); err != nil { // a *core.MinimumError; the draft stays
				p.S.MinStats = was
				return
			}
		default:
			p.S.RollCharacter(c)
		}
		*d = p.DraftOf(c, d.Of)
	})
	return err
}

func (p Creation) Minimums(core.Draft) map[core.Ability]int32 {
	m := map[core.Ability]int32{}
	for i, v := range p.S.MinStats {
		if i == 0 {
			st, x := v/100, 0
			if v > 1800 && v <= 1900 {
				st, x = 18, v-1800
			}
			m[core.Strength], m[core.StrengthPercent] = int32(st), int32(x)
			continue
		}
		m[Abilities[i]] = int32(v)
	}
	return m
}

// QuickRerolls: quickPick.good keeps the roll after 2000 rerolls.
func (p Creation) QuickRerolls() int { return quickRerolls }

const quickRerolls = 2000

func (p Creation) QuickParty() []core.Template {
	return p.Host.QuickParty()
}

var _ core.CreationRules = Creation{}

// SetScore sets the typed score, kept within the race and class limits as
// the baseline's setStatTyped did (clamped, never refused).
func (p Creation) SetScore(d core.Draft, a core.Ability, v int32) (core.Draft, error) {
	c := p.CharOf(d)
	if a == core.StrengthPercent {
		p.as().SetStat(c, 0, Score(byte(c.Stats[0])), byte(v))
		return p.DraftOf(c, d.Of), nil
	}
	i := slices.Index(Abilities[:], a)
	if i < 0 {
		return d, errors.New("no such ability")
	}
	p.as().SetStat(c, i, Score(byte(v)), c.Str00)
	return p.DraftOf(c, d.Of), nil
}

func (p Creation) Make(d core.Draft) (core.CharID, core.RosterID, error) {
	c := p.CharOf(d)
	i, ok := p.index(d.Of)
	if !ok { // a new character goes to the roster (SaveCharacter), not the party
		if err := p.S.startKit(c); err != nil { // POD's kit, given on keeping (ruling 11)
			return core.NoChar, core.NoRoster, err
		}
		if err := p.Host.Keep(p.r.ID(c)); err != nil { // ovr18:069E
			return core.NoChar, core.NoRoster, err
		}
		return core.NoChar, c.PoolIDs.Roster, nil
	}
	c.Share = 1
	*p.S.Party[i] = *c
	p.S.Who = i
	p.S.ApplyItems(p.S.Party[i]) // 00BA:0BB8
	return d.Of, core.NoRoster, nil
}

func (p Creation) Join(id core.RosterID) (core.CharID, error) { return p.Host.Join(id) }

func (p Creation) Draft(id core.CharID) (core.Draft, error) {
	if id == core.NoChar { // CreateCharacter's start: a human male fighter (Enforce builds and rolls it)
		return core.Draft{Picks: map[core.FieldID]int{FieldRace: slices.Index(PickRaces, RaceHuman),
			FieldClass: slices.Index(PickClasses(), ClsFighter)}}, nil
	}
	i, ok := p.index(id)
	if !ok {
		return core.Draft{}, errors.New("no such member")
	}
	return p.DraftOf(p.S.Party[i], id), nil
}

// classCount is how many classes c has.
func classCount(c *Character) (n int32) {
	for k := 0; k <= 7; k++ {
		if int8(c.Levels[k]) > 0 {
			n++
		}
	}
	return n
}

// draftOf is c as a draft; of is the member it modifies (NoChar: new).
func (p Creation) DraftOf(c *Character, of core.CharID) core.Draft {
	d := core.Draft{
		Picks: map[core.FieldID]int{FieldSex: int(c.Sex), FieldRace: slices.Index(PickRaces, c.Race),
			FieldAlign: int(c.Alignment), FieldClass: slices.Index(PickClasses(), c.Class)},
		Name: c.Name(),
		Of:   of, Purse: core.Copper(uint64(c.Coins[CoinGold]) * copperPerGold), HitRoll: int32(c.HPRoll) * classCount(c), Ident: uint32(c.RandomID),
	}
	d.Head, d.Body = p.Host.Portrait(picsOf(c))
	d.Sheet.Race, d.Sheet.Sex, d.Sheet.Age = RecordRace[c.Race], []core.Sex{core.Male, core.Female}[c.Sex&1], int32(c.Age)
	d.Sheet.HP, d.Sheet.MaxHP = int32(c.HP), int32(c.MaxHP)
	d.Sheet.Scores = map[core.Ability]int32{core.StrengthPercent: int32(c.Str00)}
	for i, a := range Abilities {
		d.Sheet.Scores[a] = int32(c.Stats[i])
	}
	d.MinHP, d.MaxHP = int32(p.as().MinHitPoints(c)), int32(p.as().MaxHitPoints(c))
	d.Lows = p.Minimums(d)
	for i, s := range p.S.PowerSlots(c) {
		d.Picks[firstPower+core.FieldID(i)] = s.Pick
	}
	d.Icon = p.IconOf(c)
	d.Portrait = c.Portrait
	return d
}

// iconOf is c's combat icon as the draft holds it.
func (p Creation) IconOf(c *Character) core.Icon {
	ic := core.Icon{Size: core.IconSmall}
	ic.Head, ic.Body = p.Host.IconPics(picsOf(c))
	if c.IconSize == 2 {
		ic.Size = core.IconLarge
	}
	for i, b := range c.IconColour {
		ic.Colours[i] = iconColour(b)
	}
	ic.PNG = c.Icon
	return ic
}

// charOf is d as a Character, drawing no dice: the member it modifies with
// the sheet's changes while race and class stand, else the kind's record
// with the draft's rolled values (keeping the member's items, as fresh).
func (p Creation) CharOf(d core.Draft) *Character {
	k := KindOf(d)
	var m *Character
	if i, ok := p.index(d.Of); ok {
		m = p.S.Party[i]
	}
	var c *Character
	if m != nil && m.Race == k.Race && m.Class == k.Class {
		cp := *m
		c = &cp
		c.Sex, c.Alignment = k.Sex, k.Align
		for i, a := range Abilities {
			if st := Score(d.Sheet.Scores[a]); st != c.Stats[i] || (i == 0 && byte(d.Sheet.Scores[core.StrengthPercent]) != c.Str00) {
				p.as().SetStat(c, i, st, byte(d.Sheet.Scores[core.StrengthPercent]))
			}
		}
	} else {
		c = p.S.KindRecord(k, m)
		c.RandomID, c.Age = byte(d.Ident), uint16(d.Sheet.Age)
		for i, a := range Abilities {
			c.Stats[i] = Score(d.Sheet.Scores[a])
		}
		c.Str00 = byte(d.Sheet.Scores[core.StrengthPercent])
		count := p.S.FirstLevel(c)
		p.S.startExperience(c, count, false) // the levels, without their dice: the hit points are the draft's
		c.Coins[CoinGold] = uint16(uint64(d.Purse) / copperPerGold)
		c.HPRoll = byte(d.HitRoll / int32(count))
	}
	c.HP, c.MaxHP = byte(d.Sheet.HP), byte(d.Sheet.MaxHP)
	if d.Name != "" {
		c.SetName(d.Name)
	}
	c.IconSize = 1
	if d.Icon.Size == core.IconLarge {
		c.IconSize = 2
	}
	pics := picsOf(c)
	p.Host.SetPics(d, &pics)
	setPics(c, pics)
	for i, col := range d.Icon.Colours {
		c.IconColour[i] = Colours(col.High<<4 | col.Low)
	}
	c.Portrait, c.Icon = d.Portrait, d.Icon.PNG // checked when they came in (readPNG, iconShape)
	c.Powers = nil                              // the picks in their order (draftOf states every one)
	for i := 0; ; i++ {
		x, ok := d.Picks[firstPower+core.FieldID(i)]
		if !ok {
			break
		}
		if x != noPick {
			p.S.SetPower(c, i, x)
		}
	}
	return c
}

// Creation is the ruleset's creation, its host the Rules' Host.
func (r *Rules) Creation() Creation { return Creation{S: r.S, Host: r.Host, r: r} }

// picsOf is c's pictures for the host; setPics takes them back.
func picsOf(c *Character) goldbox.Pics {
	p := c.PoolIDs
	return goldbox.Pics{PortHead: c.PortHead, PortBody: c.PortBody, IconHead: c.IconHead, IconBody: c.IconBody, IconSize: c.IconSize,
		Pool: core.PoolPics{PortraitHead: p.PortraitHeadID, PortraitBody: p.PortraitBodyID, IconHead: p.IconHeadID, IconBody: p.IconBodyID}}
}

func setPics(c *Character, p goldbox.Pics) {
	c.PortHead, c.PortBody, c.IconHead, c.IconBody = p.PortHead, p.PortBody, p.IconHead, p.IconBody
	c.PoolIDs.PortraitHeadID, c.PoolIDs.PortraitBodyID, c.PoolIDs.IconHeadID, c.PoolIDs.IconBodyID = p.Pool.PortraitHead, p.Pool.PortraitBody, p.Pool.IconHead, p.Pool.IconBody
}

// iconColour is a record's icon part colours as core's (low nibble, high nibble).
func iconColour(b Colours) core.IconColour {
	return core.IconColour{Low: uint8(b & 0xF), High: uint8(b >> 4)}
}
