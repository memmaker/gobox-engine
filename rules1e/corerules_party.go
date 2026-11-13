package rules1e

import (
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
	"slices"
	"strings"
)

// The party and its roster: members, monsters, hired NPCs, reactions, creation picks (core/party.go).

// checkParty answers the Checks of the party topic.
func (r *Rules) checkParty(k core.Check) (core.Answer, bool) {
	s := r.S
	switch k := k.(type) {
	case core.Modifiable:
		c := r.chars[k.Who]
		return core.Answer{Yes: c != nil && s.Modifiable(c)}, true
	case core.PaladinCan:
		c := r.chars[k.Who]
		used := AffLaidOnHands
		if k.Cure {
			used = AffCuredDisease
		}
		return core.Answer{Yes: c != nil && s.PaladinCan(c, used)}, true
	case core.PartySize:
		return core.Answer{Value: int32(len(s.Party))}, true
	case core.PartyAuto:
		return core.Answer{Yes: s.PartyAuto()}, true
	case core.CanJoin:
		if c := r.chars[k.Who]; c != nil {
			return core.Answer{Yes: !s.JoinRefused(c)}, true
		}
		return core.Answer{}, true
	case core.InParty:
		return core.Answer{Yes: r.inPartyAt(k.Who) >= 0}, true
	case core.ControlledBefore:
		var a core.Answer
		if r.chars[k.Who] == nil {
			return a, true
		}
		at := r.inPartyAt(k.Who)
		for i, m := range s.Party {
			if i == at {
				break
			}
			if m.Control < 0x80 {
				a.Value++
			}
		}
		return a, true
	case core.MoveExtremes:
		hi, lo := s.MoveExtremes()
		return core.Answer{High: int32(hi), Low: int32(lo)}, true
	case core.Diseased:
		c := r.chars[k.Who]
		return core.Answer{Yes: c != nil && slices.ContainsFunc(Diseases[:], func(d AffectType) bool { return s.FindAffect(c, d) != nil })}, true
	case core.PartyHasSkill:
		return core.Answer{Yes: s.PartyHasSkill(k.N)}, true
	default:
		return core.Answer{}, false
	}
}

// ImportedFlag is the bit of Sheet.ScriptFlag a character added from another game carries.
const ImportedFlag = 1

// resolveParty answers the Actions of the party topic.
func (r *Rules) resolveParty(a core.Action) (core.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case core.MarkImported:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		if s.T().ImportFlag {
			c.ScriptFlag |= ImportedFlag
		}
	case core.SetPoolPic:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		*c.PoolIDs.pic(a.Pic) = a.ID
	case core.SetIconPics:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		p := a.Pics
		c.IconHead, c.IconBody, c.IconSize, c.Icon = byte(p.Head), byte(p.Body), byte(p.Size), p.PNG
		for i, k := range p.Colours {
			c.IconColour[i] = Colours(k.Low&0xF | k.High<<4)
		}
	case core.SetPortrait:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		c.Portrait = a.PNG
	case core.SetStatusAction:
		return core.Outcome{}, true, r.setStatus(a)
	case core.SetPolicy:
		return core.Outcome{}, true, r.setPolicy(a)
	case core.NewGroup:
		r.tmpl = nil
	case core.EmptyParty:
		s.SetPartyCount(0)
	case core.JoinParty:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		s.AddToParty(c)
	case core.SetInactive:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		c.Okay = 0
	case core.MoveMember:
		return core.Outcome{}, true, s.MoveMember(r.inPartyAt(a.Who), int(a.To))
	case core.SetOrder:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		c.Order = byte(a.Order)
	case core.ForgetParty:
		s.Party, s.Pool = nil, nil
	case core.GainExperience:
		for _, c := range r.party(a.Whom) {
			s.GainExperience(c, a.Points)
		}
	case core.PickKind:
		k, err := kindNamed(a)
		if err != nil {
			return core.Outcome{}, true, err
		}
		picks := map[core.FieldID]int{FieldSex: int(k.Sex), FieldRace: slices.Index(PickRaces, k.Race),
			FieldAlign: int(k.Align), FieldClass: slices.Index(PickClasses(), k.Class)}
		return core.Outcome{Effects: []core.Effect{core.Picked{Picks: picks}}}, true, nil
	case core.MakeMonster:
		m, err := r.monster(a.Of)
		if err != nil {
			return core.Outcome{}, true, fmt.Errorf("MakeMonster: %w", err)
		}
		c := &Character{}
		r.S.MakeMonster(c, m)
		return core.Outcome{Effects: []core.Effect{core.MonsterMade{Who: r.ID(c)}}}, true, nil
	case core.DressNPC:
		c := r.chars[a.Who]
		if c == nil {
			c = &Character{}
		}
		r.S.dressNPC(c)
		return core.Outcome{Effects: []core.Effect{core.NPCDressed{Who: r.ID(c)}}}, true, nil
	case core.SetRoster:
		r.chars[a.Who].PoolIDs.Roster = a.Roster
	case core.EmptyMember:
		if r.empty == nil {
			r.empty = &Character{}
		}
		*r.empty = Character{}
		return core.Outcome{Effects: []core.Effect{core.EmptyMemberIs{Who: r.ID(r.empty)}}}, true, nil
	case core.PaladinHeal:
		c, t := r.chars[a.Who], r.chars[a.Target]
		if c == nil || t == nil {
			return core.Outcome{}, true, errNoChar
		}
		s.Who = s.PartyIndex(c)
		var o core.Outcome
		if hp := 2 * int(c.Levels[ClsPaladin]); s.Heal(t, hp, true) {
			o.Effects = append(o.Effects, core.Healed{Unit: a.Target, HP: int32(hp)})
		}
		s.AddAffect(c, AffLaidOnHands, 1440, 0, 0)
		return o, true, nil
	case core.PaladinCure:
		c, t := r.chars[a.Who], r.chars[a.Target]
		if c == nil || t == nil {
			return core.Outcome{}, true, errNoChar
		}
		s.Who = s.PartyIndex(c)
		s.Rv.Cure = 1
		for _, d := range Diseases {
			s.RemoveAffect(t, d, nil)
		}
		s.Rv.Cure = 0
		s.AddAffect(c, AffCuredDisease, 0x2760, 0, 0)
		return core.Outcome{}, true, nil
	case core.HealParty:
		for _, c := range s.Party {
			c.HP, c.Status, c.Okay = c.MaxHP, StatusOkay, 1
		}
		return core.Outcome{}, true, nil
	case core.Toughen:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		s.Toughen(c)
		return core.Outcome{}, true, nil
	case core.TakePower:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		c.Powers = append(c.Powers, a.Power)
		return core.Outcome{}, true, nil
	default:
		return core.Outcome{}, false, nil
	}
	return core.Outcome{}, true, nil
}

// factParty answers the Facts of the party topic.
func (r *Rules) factParty(id core.CharID, f core.Fact) (core.Value, bool, error) {
	s, c := r.S, r.chars[id]
	switch f := f.(type) {
	case core.PartyFact:
		p := make(core.Party, len(s.Party))
		for i, m := range s.Party {
			p[i] = r.ID(m)
		}
		return p, true, nil
	case core.ClassByNumber:
		if f.N < 0 || f.N >= len(PorClass) {
			return core.ClassNone, true, nil
		}
		return PorClass[f.N], true, nil
	case core.ModFact:
		if c == nil {
			return nil, true, errNoChar
		}
		a := slices.Index(core.Abilities[:], f.Ability)
		if a < 0 {
			return nil, true, errNoFact
		}
		return core.Number(s.Mod(c, Ability(a))), true, nil
	}
	switch f {
	case core.AbilityFact{Ability: core.Constitution}, core.AbilityFact{Ability: core.Charisma}, core.FactExpValue, core.FactExpPerHP,
		core.FactMove, core.FactControl, core.FactRace, core.FactAlignment, core.FactActive, core.FactLevelReady, core.FactSide:
		if c == nil {
			return nil, true, errNoChar
		}
	}
	switch f {
	case core.AbilityFact{Ability: core.Constitution}:
		return core.Number(c.Stats[StatCon]), true, nil
	case core.AbilityFact{Ability: core.Charisma}:
		return core.Number(c.Stats[StatCha]), true, nil
	case core.FactExpValue:
		return core.Number(c.ExpValue), true, nil
	case core.FactExpPerHP:
		return core.Number(c.ExpPerHP), true, nil
	case core.FactMove:
		return core.Number(c.Move), true, nil
	case core.FactControl:
		return core.Number(c.Control), true, nil
	case core.FactRace:
		if int(c.Race) >= len(goldbox.RecordRace) {
			return core.Number(c.Race), true, nil // outside the table: the record's own byte, as the original reads it
		}
		return goldbox.RecordRace[c.Race], true, nil
	case core.FactAlignment:
		if int(c.Alignment) >= len(goldbox.RecordAlignment) {
			return core.Number(c.Alignment), true, nil // SpecialReadZ1 $D8 reads the raw byte
		}
		return goldbox.RecordAlignment[c.Alignment], true, nil
	case core.FactActive:
		return core.Yes(c.Okay != 0), true, nil
	case core.FactLevelReady:
		return core.Yes(levelReady(c)), true, nil
	case core.FactSide:
		switch {
		case c.Side == SideMonsters:
			return core.SideMonsters, true, nil
		case c.Side == SideParty && c.Auto != 0:
			return core.SideAuto, true, nil
		}
		return core.SideParty, true, nil
	}
	switch f := f.(type) {
	case core.PoolPicsFact:
		if c == nil {
			return nil, true, errNoChar
		}
		p := c.PoolIDs
		return core.PoolPics{PortraitHead: p.PortraitHeadID, PortraitBody: p.PortraitBodyID, IconHead: p.IconHeadID, IconBody: p.IconBodyID}, true, nil
	case core.DraftIconFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return r.Creation().IconOf(c), true, nil
	case core.IconPicsFact:
		if c == nil {
			return nil, true, errNoChar
		}
		ic := core.IconPics{Head: int(c.IconHead), Body: int(c.IconBody), Size: int(c.IconSize), PNG: c.Icon}
		for i, b := range c.IconColour {
			ic.Colours[i] = iconColour(b)
		}
		return ic, true, nil
	case core.FiguresFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return core.Figures(s.Rules().Sheet(s, c)), true, nil
	case core.PortraitFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return core.Portrait{PNG: c.Portrait, Head: int(c.PortHead), Body: int(c.PortBody)}, true, nil
	case core.SaveFact:
		if c == nil {
			return nil, true, errNoChar
		}
		for i, s := range porSave {
			if s == f.Save {
				return core.Number(c.Saves[i]), true, nil
			}
		}
		return nil, true, errNoFact
	case core.SkillFact:
		if c == nil {
			return nil, true, errNoChar
		}
		for i, s := range porSkill {
			if s == f.Skill {
				return core.Number(c.Thief[i]), true, nil
			}
		}
		return nil, true, errNoFact
	case core.WordFact:
		if c == nil {
			return nil, true, errNoChar
		}
		switch f.Of {
		case core.FactStatus:
			return core.Text(s.StatusName(c)), true, nil
		case core.FactHarmed:
			var what []string
			for _, a := range c.Affects {
				if name := s.HarmfulAffect(a.Type); name != "" {
					what = append(what, name)
				}
			}
			return core.Text(strings.Join(what, ", ")), true, nil
		}
		v, err := word(c, f)
		return v, true, err
	case core.IconFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return core.Number(c.Order), true, nil
	case core.ReactionsFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return core.Reactions(s.Reactions(c)), true, nil
	case core.MemberCardFact:
		if c == nil {
			return nil, true, errNoChar
		}
		m := core.MemberCard{Party: IsParty(c), Class: ClassName(c.Class), Portrait: c.Portrait}
		for _, s := range c.Memorized {
			if s.Held() {
				m.Ready++
			}
		}
		for _, class := range c.SpellSlots {
			for _, n := range class {
				m.Slots += int(n)
			}
		}
		return m, true, nil
	case core.OwedFact:
		if c == nil {
			return nil, true, errNoChar
		}
		var o core.PowerOwed
		if ps := r.S.Rules().Picks(r.S, c); len(ps) > 0 {
			o.What = ps[0].What
			for _, m := range ps[0].Options {
				o.Options = append(o.Options, core.ManeuverCard{ID: core.ManeuverID(m.ID), Name: m.Name, Tip: m.Tip, Aim: core.ManeuverAim(m.Aim), Range: m.Range, Area: m.Area, Move: m.Move})
			}
		}
		return o, true, nil
	default:
		return nil, false, nil
	}
}

// Portrait tables DS:8C7.. (gender-major).
var (
	portHeads     = [2][8]int{{1, 2, 3, 4, 5, 8, 0xB, 0xC}, {6, 7, 9, 0xA, 0xD, 0xE, 6, 7}}
	portBodyCls0  = [2][5]int{{2, 4, 2, 4, 2}, {2, 4, 8, 2, 4}}
	portBodyOther = [2][5]int{{1, 3, 4, 5, 6}, {7, 9, 0xB, 7, 9}}
	portBodyCls5  = [2][5]int{{2, 2, 2, 2, 2}, {8, 0xA, 0xC, 8, 0xA}}
	portBodyCls6  = [2][5]int{{6, 6, 6, 6, 6}, {8, 0xA, 8, 0xA, 8}}
)

// dressNPC = ovr18:1262 after the load: a random id, a portrait by a
// random class of c's and its gender.
func (e *RuleState) dressNPC(c *Character) {
	c.RandomID = byte(e.Random(0x100))
	var classes []int
	for cls := 0; cls < 7; cls++ {
		if int8(c.Levels[cls]) > 0 { // jle on a signed byte (ovr18:12CA)
			classes = append(classes, cls)
		}
	}
	if len(classes) == 0 {
		classes = []int{2}
	}
	cls := classes[e.Dice(1, len(classes))-1]
	n := e.Dice(1, 5) - 1
	g := int(c.Sex) & 1
	switch cls {
	case 0:
		c.PortBody = byte(portBodyCls0[g][n])
	case 5:
		c.PortBody = byte(portBodyCls5[g][n])
	case 6:
		c.PortBody = byte(portBodyCls6[g][n])
	default:
		c.PortBody = byte(portBodyOther[g][n])
	}
	c.PortHead = byte(portHeads[g][e.Dice(1, 8)-1])
}

// kindNamed is the kind the names in k stand for.
func kindNamed(k core.PickKind) (Kind, error) {
	race, sex := slices.Index(RaceNames[:], k.Race), slices.Index(GenderNames[:], k.Gender)
	class, align := slices.Index(ClassNames[:], k.Class), slices.Index(AlignmentNames[:], k.Align)
	if race < 0 || sex < 0 || class < 0 || align < 0 {
		return Kind{}, fmt.Errorf("no kind %+v", k)
	}
	return Kind{Race: RaceID(race), Sex: byte(sex), Class: ClassID(class), Align: byte(align)}, nil
}

// inPartyAt is the index of the first member of who's name and random id, or -1.
func (r *Rules) inPartyAt(who core.CharID) int {
	c := r.chars[who]
	if c == nil {
		return -1
	}
	for i, m := range r.S.Party {
		if m.Name() == c.Name() && m.RandomID == c.RandomID {
			return i
		}
	}
	return -1
}
