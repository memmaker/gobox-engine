package rules1e

import (
	"fmt"
	"gobox/rules"
	"slices"
	"strings"
)

// The party and its roster: members, monsters, hired NPCs, reactions, creation picks (core/party.go).

// checkParty answers the Checks of the party topic.
func (r *Rules) checkParty(k rules.Check) (rules.Answer, bool) {
	s := r.S
	switch k := k.(type) {
	case rules.Modifiable:
		c := r.chars[k.Who]
		return rules.Answer{Yes: c != nil && s.Modifiable(c)}, true
	case rules.PaladinCan:
		c := r.chars[k.Who]
		used := AffLaidOnHands
		if k.Cure {
			used = AffCuredDisease
		}
		return rules.Answer{Yes: c != nil && s.PaladinCan(c, used)}, true
	case rules.PartySize:
		return rules.Answer{Value: int32(len(s.Party))}, true
	case rules.PartyAuto:
		return rules.Answer{Yes: s.PartyAuto()}, true
	case rules.CanJoin:
		if c := r.chars[k.Who]; c != nil {
			return rules.Answer{Yes: !s.JoinRefused(c)}, true
		}
		return rules.Answer{}, true
	case rules.InParty:
		return rules.Answer{Yes: r.inPartyAt(k.Who) >= 0}, true
	case rules.ControlledBefore:
		var a rules.Answer
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
	case rules.MoveExtremes:
		hi, lo := s.MoveExtremes()
		return rules.Answer{High: int32(hi), Low: int32(lo)}, true
	case rules.Diseased:
		c := r.chars[k.Who]
		return rules.Answer{Yes: c != nil && slices.ContainsFunc(Diseases[:], func(d AffectType) bool { return s.FindAffect(c, d) != nil })}, true
	case rules.PartyHasSkill:
		return rules.Answer{Yes: s.PartyHasSkill(k.N)}, true
	default:
		return rules.Answer{}, false
	}
}

// ImportedFlag is the bit of Sheet.ScriptFlag a character added from another game carries.
const ImportedFlag = 1

// resolveParty answers the Actions of the party topic.
func (r *Rules) resolveParty(a rules.Action) (rules.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case rules.MarkImported:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		if s.T().ImportFlag {
			c.ScriptFlag |= ImportedFlag
		}
	case rules.SetPoolPic:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		*c.PoolIDs.pic(a.Pic) = a.ID
	case rules.SetIconPics:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		p := a.Pics
		c.IconHead, c.IconBody, c.IconSize, c.Icon = byte(p.Head), byte(p.Body), byte(p.Size), p.PNG
		for i, k := range p.Colours {
			c.IconColour[i] = Colours(k.Low&0xF | k.High<<4)
		}
	case rules.SetPortrait:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		c.Portrait = a.PNG
	case rules.SetStatusAction:
		return rules.Outcome{}, true, r.setStatus(a)
	case rules.SetPolicy:
		return rules.Outcome{}, true, r.setPolicy(a)
	case rules.NewGroup:
		r.tmpl = nil
	case rules.EmptyParty:
		s.SetPartyCount(0)
	case rules.JoinParty:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		s.AddToParty(c)
	case rules.SetInactive:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		c.Okay = 0
	case rules.MoveMember:
		return rules.Outcome{}, true, s.MoveMember(r.inPartyAt(a.Who), int(a.To))
	case rules.SetOrder:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		c.Order = byte(a.Order)
	case rules.ForgetParty:
		s.Party, s.Pool = nil, nil
	case rules.GainExperience:
		for _, c := range r.party(a.Whom) {
			s.GainExperience(c, a.Points)
		}
	case rules.PickKind:
		k, err := kindNamed(a)
		if err != nil {
			return rules.Outcome{}, true, err
		}
		picks := map[rules.FieldID]int{FieldSex: int(k.Sex), FieldRace: slices.Index(PickRaces, k.Race),
			FieldAlign: int(k.Align), FieldClass: slices.Index(PickClasses(), k.Class)}
		return rules.Outcome{Effects: []rules.Effect{rules.Picked{Picks: picks}}}, true, nil
	case rules.MakeMonster:
		m, err := r.monster(a.Of)
		if err != nil {
			return rules.Outcome{}, true, fmt.Errorf("MakeMonster: %w", err)
		}
		c := &Character{}
		r.S.MakeMonster(c, m)
		return rules.Outcome{Effects: []rules.Effect{rules.MonsterMade{Who: r.ID(c)}}}, true, nil
	case rules.DressNPC:
		c := r.chars[a.Who]
		if c == nil {
			c = &Character{}
		}
		r.S.dressNPC(c)
		return rules.Outcome{Effects: []rules.Effect{rules.NPCDressed{Who: r.ID(c)}}}, true, nil
	case rules.SetRoster:
		r.chars[a.Who].PoolIDs.Roster = a.Roster
	case rules.EmptyMember:
		if r.empty == nil {
			r.empty = &Character{}
		}
		*r.empty = Character{}
		return rules.Outcome{Effects: []rules.Effect{rules.EmptyMemberIs{Who: r.ID(r.empty)}}}, true, nil
	case rules.PaladinHeal:
		c, t := r.chars[a.Who], r.chars[a.Target]
		if c == nil || t == nil {
			return rules.Outcome{}, true, errNoChar
		}
		s.Who = s.PartyIndex(c)
		var o rules.Outcome
		if hp := 2 * int(c.Levels[ClsPaladin]); s.Heal(t, hp, true) {
			o.Effects = append(o.Effects, rules.Healed{Unit: a.Target, HP: int32(hp)})
		}
		s.AddAffect(c, AffLaidOnHands, 1440, 0, 0)
		return o, true, nil
	case rules.PaladinCure:
		c, t := r.chars[a.Who], r.chars[a.Target]
		if c == nil || t == nil {
			return rules.Outcome{}, true, errNoChar
		}
		s.Who = s.PartyIndex(c)
		s.Rv.Cure = 1
		for _, d := range Diseases {
			s.RemoveAffect(t, d, nil)
		}
		s.Rv.Cure = 0
		s.AddAffect(c, AffCuredDisease, 0x2760, 0, 0)
		return rules.Outcome{}, true, nil
	case rules.HealParty:
		for _, c := range s.Party {
			c.HP, c.Status, c.Okay = c.MaxHP, StatusOkay, 1
		}
		return rules.Outcome{}, true, nil
	case rules.Toughen:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		s.Toughen(c)
		return rules.Outcome{}, true, nil
	case rules.TakePower:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		c.Powers = append(c.Powers, a.Power)
		return rules.Outcome{}, true, nil
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}

// factParty answers the Facts of the party topic.
func (r *Rules) factParty(id rules.CharID, f rules.Fact) (rules.Value, bool, error) {
	s, c := r.S, r.chars[id]
	switch f := f.(type) {
	case rules.PartyFact:
		p := make(rules.Party, len(s.Party))
		for i, m := range s.Party {
			p[i] = r.ID(m)
		}
		return p, true, nil
	case rules.ClassByNumber:
		if f.N < 0 || f.N >= len(PorClass) {
			return rules.ClassNone, true, nil
		}
		return PorClass[f.N], true, nil
	case rules.ModFact:
		if c == nil {
			return nil, true, errNoChar
		}
		a := slices.Index(rules.Abilities[:], f.Ability)
		if a < 0 {
			return nil, true, errNoFact
		}
		return rules.Number(s.Mod(c, Ability(a))), true, nil
	}
	switch f {
	case rules.AbilityFact{Ability: rules.Constitution}, rules.AbilityFact{Ability: rules.Charisma}, rules.FactExpValue, rules.FactExpPerHP,
		rules.FactMove, rules.FactControl, rules.FactRace, rules.FactAlignment, rules.FactActive, rules.FactLevelReady, rules.FactSide:
		if c == nil {
			return nil, true, errNoChar
		}
	}
	switch f {
	case rules.AbilityFact{Ability: rules.Constitution}:
		return rules.Number(c.Stats[StatCon]), true, nil
	case rules.AbilityFact{Ability: rules.Charisma}:
		return rules.Number(c.Stats[StatCha]), true, nil
	case rules.FactExpValue:
		return rules.Number(c.ExpValue), true, nil
	case rules.FactExpPerHP:
		return rules.Number(c.ExpPerHP), true, nil
	case rules.FactMove:
		return rules.Number(c.Move), true, nil
	case rules.FactControl:
		return rules.Number(c.Control), true, nil
	case rules.FactRace:
		if c.Race == RaceMonster && int(c.MonsterRace) >= len(RecordRace) {
			return rules.Number(c.MonsterRace), true, nil // outside the table: the record's own byte, as the original reads it
		}
		return c.Race, true, nil
	case rules.FactAlignment:
		if int(c.Alignment) >= len(RecordAlignment) {
			return rules.Number(c.Alignment), true, nil // SpecialReadZ1 $D8 reads the raw byte
		}
		return RecordAlignment[c.Alignment], true, nil
	case rules.FactActive:
		return rules.Yes(c.Okay != 0), true, nil
	case rules.FactLevelReady:
		return rules.Yes(levelReady(c)), true, nil
	case rules.FactSide:
		switch {
		case c.Side == SideMonsters:
			return rules.SideMonsters, true, nil
		case c.Side == SideParty && c.Auto != 0:
			return rules.SideAuto, true, nil
		}
		return rules.SideParty, true, nil
	}
	switch f := f.(type) {
	case rules.PoolPicsFact:
		if c == nil {
			return nil, true, errNoChar
		}
		p := c.PoolIDs
		return rules.PoolPics{PortraitHead: p.PortraitHeadID, PortraitBody: p.PortraitBodyID, IconHead: p.IconHeadID, IconBody: p.IconBodyID}, true, nil
	case rules.DraftIconFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return r.Creation().IconOf(c), true, nil
	case rules.IconPicsFact:
		if c == nil {
			return nil, true, errNoChar
		}
		ic := rules.IconPics{Head: int(c.IconHead), Body: int(c.IconBody), Size: int(c.IconSize), PNG: c.Icon}
		for i, b := range c.IconColour {
			ic.Colours[i] = iconColour(b)
		}
		return ic, true, nil
	case rules.FiguresFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return rules.Figures(s.Rules().Sheet(s, c)), true, nil
	case rules.PortraitFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return rules.Portrait{PNG: c.Portrait, Head: int(c.PortHead), Body: int(c.PortBody)}, true, nil
	case rules.SaveFact:
		if c == nil {
			return nil, true, errNoChar
		}
		for i, s := range porSave {
			if s == f.Save {
				return rules.Number(c.Saves[i]), true, nil
			}
		}
		return nil, true, errNoFact
	case rules.SkillFact:
		if c == nil {
			return nil, true, errNoChar
		}
		for i, s := range porSkill {
			if s == f.Skill {
				return rules.Number(c.Thief[i]), true, nil
			}
		}
		return nil, true, errNoFact
	case rules.WordFact:
		if c == nil {
			return nil, true, errNoChar
		}
		switch f.Of {
		case rules.FactStatus:
			return rules.Text(s.StatusName(c)), true, nil
		case rules.FactHarmed:
			var what []string
			for _, a := range c.Affects {
				if name := s.HarmfulAffect(a.Type); name != "" {
					what = append(what, name)
				}
			}
			return rules.Text(strings.Join(what, ", ")), true, nil
		}
		v, err := word(c, f)
		return v, true, err
	case rules.IconFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return rules.Number(c.Order), true, nil
	case rules.ReactionsFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return rules.Reactions(s.Reactions(c)), true, nil
	case rules.MemberCardFact:
		if c == nil {
			return nil, true, errNoChar
		}
		m := rules.MemberCard{Party: IsParty(c), Class: ClassName(c.Class), Portrait: c.Portrait}
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
	case rules.OwedFact:
		if c == nil {
			return nil, true, errNoChar
		}
		var o rules.PowerOwed
		if ps := r.S.Rules().Picks(r.S, c); len(ps) > 0 {
			o.What = ps[0].What
			for _, m := range ps[0].Options {
				o.Options = append(o.Options, rules.ManeuverCard{ID: rules.ManeuverID(m.ID), Name: m.Name, Tip: m.Tip, Aim: rules.ManeuverAim(m.Aim), Range: m.Range, Area: m.Area, Move: m.Move})
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
func kindNamed(k rules.PickKind) (Kind, error) {
	race, sex := slices.Index(RaceNames[:], k.Race), slices.Index(GenderNames[:], k.Gender)
	class, align := slices.Index(ClassNames[:], k.Class), slices.Index(AlignmentNames[:], k.Align)
	if race < 0 || sex < 0 || class < 0 || align < 0 {
		return Kind{}, fmt.Errorf("no kind %+v", k)
	}
	return Kind{Race: RecordRace[race], Sex: byte(sex), Class: ClassID(class), Align: byte(align)}, nil
}

// inPartyAt is the index of the first member of who's name and random id, or -1.
func (r *Rules) inPartyAt(who rules.CharID) int {
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
