package rules1e

import (
	"slices"
	"strconv"
)

// Encamp, ported from ovr16 (camp menu, magic: cast / memorize / scribe /
// display, alter: order / drop / speed / pics) and the rest of ovr21
// (resting, healing, study).

// campState holds the ovr16/ovr21 globals.
type campState struct {
	Rest     [7]int                   // DS:6DC2: rest time left, clock units (DS:6DC4 mins, 6DC6 tens, 6DC8 hours, 6DCA days)
	Heal     int                      // DS:6DD8: five-minute steps toward the daily healing
	Study    [MaxCombatants + 1]Count // DS:6DCF: five-minute steps until member i's next spell is learned
	Interval int                      // DS:49FB: steps since the last encounter roll
	RestMins int                      // extra: minutes of this rest (a long rest: LongRest)
	RestHour int                      // extra: five-minute steps of this rest toward the hour (hourTick)
	StopRest bool                     // extra: StopRest: the rest ends at its next step

	scriber map[*Item][3]*Character // extra: who chose to scribe each scroll spell (the party inventory is shared)

	Interrupted bool // extra: a rest was interrupted (camp closes)
	On          bool // extra: camp is up (CampMenu)
	SavedFlag   byte // extra: beginFlag and the picture's animation as camp found them
	SavedAnim   bool
}

// restStep is what five minutes of rest did (stepRest).
type restStep struct {
	Over        bool // the rest is over
	Healed      bool // the whole party healed (healTick)
	Interrupted bool // a monster interrupted it: over too
}

func (e *RuleState) SpellLevel(sp SpellID) SpellProp { return e.T().Spells[sp.Spell()][1] }
func (e *RuleState) SpellClass(sp SpellID) SpellClass {
	return SpellClass(e.T().Spells[sp.Spell()][0])
}

// StudyTime = ovr16:0000 (char): minutes of study for the spells waiting
// to be memorised (+0x80) or scribed; +0x2C gets the base hours (4, or 6
// for spells above 2nd level).
func (e *RuleState) StudyTime(c *Character) int {
	maxMem, maxScr, sumMem, sumScr := SpellProp(0), SpellProp(0), 0, 0
	for i := 0; i <= 0x14; i++ {
		if s := c.Memorized[i]; s.Pending() {
			l := e.SpellLevel(s)
			if l > maxMem {
				maxMem = l
			}
			sumMem += int(l)
		}
	}
	for _, it := range e.Carried(c) {
		if !e.IsScroll(it) {
			continue
		}
		for k := 1; k <= 3; k++ {
			if s := it.Spells[k-1]; s&ScrollScribing != 0 && e.ScribesIt(c, it, k-1) {
				l := e.SpellLevel(s.Spell())
				if l > maxScr {
					maxScr = l
				}
				sumScr += int(l)
			}
		}
	}
	base := 0
	if sumMem > 0 || sumScr > 0 {
		base = 4
	}
	if maxMem > 2 || maxScr > 2 {
		base = 6
	}
	c.Study = byte(base)
	return base*60 + sumMem*15 + sumScr*15
}

// clearMemorize = ovr16:0159 (char): forget the spells still to be memorised.
// TrimMemorized = EffHndlr.ForgetSpellsPastSlots: the cleric and
// magic-user spells held past c's slots (after a level drain) are
// forgotten, the last of each spell level first.
func (e *RuleState) TrimMemorized(c *Character) {
	used := map[slotKind]int{}
	for i, v := range c.Memorized {
		if v == 0 {
			continue
		}
		s := v.Spell()
		lvl, cls := e.SpellLevel(s), e.SpellClass(s)
		if cls != SpellClassCleric && cls != SpellClassMage || lvl < 1 || lvl > 3 {
			continue
		}
		k := slotKind{cls, lvl}
		if used[k]++; used[k] > int(c.SpellSlot(cls, int(lvl))) {
			c.Memorized[i] = 0
		}
	}
}

func clearMemorize(c *Character) {
	for i := 0; i <= 0x14; i++ {
		if c.Memorized[i].Pending() {
			c.Memorized[i] = 0
		}
	}
	c.Study = 0
}

// clearScribe = ovr16:019C (char): scroll spells are no longer being scribed.
func (e *RuleState) clearScribe(c *Character) {
	for _, it := range e.Carried(c) {
		if e.IsScroll(it) {
			for k := 1; k <= 3; k++ {
				it.Spells[k-1] &^= ScrollScribing
			}
		}
	}
}

// ClearStudy = ovr16:0214: nobody studies any more.
func (e *RuleState) ClearStudy() {
	e.Camp.scriber = nil
	for _, c := range e.Party {
		clearMemorize(c)
		e.clearScribe(c)
	}
}

// CanStudy = ovr16:036C (mode 1 cast, 2 memorize, 3 scribe): area +0x1CA
// forbids casting; the dead (status 1) and the fallen cannot study.
func (e *RuleState) CanStudy(mode int) bool {
	msg := ""
	c := e.Acting()
	switch {
	case mode == 1 && e.Area.NoMagic > 0:
		msg = "cannot cast spells here"
	case c.Status == StatusAnimated || c.Okay == 0: // extra: shorter than the original's "is in no condition to cast any spells", so it fits a note
		msg = [...]string{"cannot cast now", "cannot memorize now", "cannot scribe now"}[mode-1]
	}
	if msg == "" {
		return true
	}
	e.CombatSay(c, msg, true)
	return false
}

// partyStudy: the longest study time of the party in minutes (studyTime
// also sets each member's study hours, which Rest counts down).
// ScribesIt (extra): spell k of scroll it is being scribed by c, not by
// another member sharing the scroll.
func (e *RuleState) ScribesIt(c *Character, it *Item, k int) bool {
	by := e.Camp.scriber[it][k]
	return it.Spells[k].Scribing() && (by == nil || by == c)
}

// MarkScribe (extra): c scribes spell k of scroll it (c nil: nobody).
func (e *RuleState) MarkScribe(c *Character, it *Item, k int) {
	if e.Camp.scriber == nil {
		e.Camp.scriber = map[*Item][3]*Character{}
	}
	by := e.Camp.scriber[it]
	by[k] = c
	e.Camp.scriber[it] = by
}

// willStudy (extra): a rest would have someone memorize or scribe a
// spell, the planned spells taken up first as a rest does (on copies).
func (e *RuleState) willStudy() bool {
	for _, c := range e.Party {
		mem, study := c.Memorized, c.Study
		e.PlanSpells(c)
		t := e.StudyTime(c)
		c.Memorized, c.Study = mem, study
		if t > 0 {
			return true
		}
	}
	return false
}

func (e *RuleState) PartyStudy() int {
	most := 0
	for _, c := range e.Party {
		most = max(most, e.StudyTime(c))
	}
	return most
}

const SpellCureLight = 3

// HealClerics are the members who cast Cure Light Wounds.
func (e *RuleState) HealClerics() []*Character {
	var clerics []*Character
	for _, c := range e.Party {
		if c.SpellSlot(SpellClassCleric, 1) > 0 && c.SpellBook[SpellCureLight-1] != 0 {
			clerics = append(clerics, c)
		}
	}
	return clerics
}

// PlanSpells (extra): the spells of c.Plan that c does not hold are set to
// be memorized on the next rest, as far as the free slots go.
func (e *RuleState) PlanSpells(c *Character) {
	clearMemorize(c)
	e.TrimMemorized(c)
	held := map[SpellID]int{}
	for _, v := range c.Memorized {
		if v != 0 {
			held[v]++
		}
	}
	free := map[slotKind]int{} // per level and class, lazily from the slots
	for _, s := range c.Plan {
		if s == 0 {
			continue
		}
		if held[s] > 0 {
			held[s]--
			continue
		}
		lvl, cls := e.SpellLevel(s), e.SpellClass(s)
		k := slotKind{cls, lvl}
		if _, ok := free[k]; !ok {
			free[k] = int(c.SpellSlot(cls, int(lvl)))
			for _, v := range c.Memorized {
				if v != 0 && e.SpellLevel(v) == lvl && e.SpellClass(v) == cls {
					free[k]--
				}
			}
		}
		if free[k] <= 0 {
			continue
		}
		free[k]--
		if i := slices.Index(c.Memorized[:PORMemory], 0); i >= 0 {
			c.Memorized[i] = s | SpellPending
		}
	}
	e.sortSpells(c)
}

// ForgetCures clears c's memorized Cure Light Wounds (healPlanned: the
// cures left make way for the planned spells).
func (e *RuleState) ForgetCures(c *Character) {
	for i, v := range c.Memorized {
		if v == SpellCureLight {
			c.Memorized[i] = 0
		}
	}
}

// MemorizeCures replaces c's first-level cleric spells by Cure Light
// Wounds to be memorized, one per slot.
func (e *RuleState) MemorizeCures(c *Character) {
	var m [MemorySlots]SpellID
	k := 0
	for _, v := range c.Memorized {
		if v != 0 && !(e.SpellClass(v) == SpellClassCleric && e.SpellLevel(v) == 1) {
			m[k] = v
			k++
		}
	}
	for n := c.SpellSlot(SpellClassCleric, 1); n > 0 && k < PORMemory; n-- {
		m[k] = SpellCureLight | SpellPending
		k++
	}
	c.Memorized = m
}

// SpellsAllowed: the area allows spells.
func (e *RuleState) SpellsAllowed() bool { return e.Area.NoMagic == 0 }

// CastCures (extra): every cleric able to cast spends memorized Cure Light
// Wounds on the most wounded member while someone needs it.
func (e *RuleState) CastCures() {
	if !e.SpellsAllowed() {
		return
	}
	for _, c := range e.Party {
		if c.Status == StatusAnimated || c.Okay == 0 {
			continue
		}
		for i, v := range c.Memorized {
			if v != SpellCureLight {
				continue
			}
			var t *Character
			for _, p := range e.Party {
				if healable(p) && p.HP < p.MaxHP && e.FindAffect(p, affMummyDisease) == nil && (t == nil || int(p.MaxHP)-int(p.HP) > int(t.MaxHP)-int(t.HP)) {
					t = p
				}
			}
			if t == nil {
				return
			}
			c.Memorized[i] = 0
			e.SpellCaster, e.spellTarget, e.NSpellTargets = c, t, 1
			e.spellCureLightWounds()
		}
	}
	e.SpellCaster = nil
}

// healable: Heal works on c (its status).
func healable(c *Character) bool {
	switch c.Status {
	case 0, 1, 4, 5:
		return true
	}
	return false
}

// ClockMinutes (extra): the time of day in minutes.
func (e *RuleState) ClockMinutes() int {
	return 60*int(e.Area.Clock[3]) + 10*int(e.Area.Clock[2]) + int(e.Area.Clock[1])
}

// SkyDay is State's Flags.Day: whether the 3D view's sky, as the original
// colours it (Sky, or SkyInside on a cell > 0x7F), is a day sky: anything
// but dark grey (8, the night the scripts set after 14:00, and the
// graveyard's sky at any hour). Off the 3D view, or before a script set
// one, the clock decides (isDay).
func (e *RuleState) SkyDay(inside bool) bool {
	sky := e.Area.Sky
	if inside {
		sky = e.Area.SkyInside
	}
	if e.Area.In3D == 0 || sky == 0 {
		return e.IsDay()
	}
	return sky != 8
}

// IsDay (extra): day runs from 6:00 to 20:00.
func (e *RuleState) IsDay() bool {
	m := e.ClockMinutes()
	return m >= 6*60 && m < 20*60
}

// HealMinutes (extra): the rest until every member that resting can heal
// is at full hit points (1 HP a day, see healTick); 0 when none is.
func (e *RuleState) HealMinutes() int {
	days := 0
	for _, c := range e.Party {
		if healable(c) && e.FindAffect(c, affMummyDisease) == nil {
			days = max(days, int(c.MaxHP)-int(c.HP))
		}
	}
	if days == 0 {
		return 0
	}
	return days*24*60 - e.Camp.Heal*5
}

// sortSpells = ovr16:0824 (char): c's spell slots in spell order (empty
// slots first).
func (e *RuleState) sortSpells(c *Character) {
	s := c.Memorized[:]
	for i := 0; i <= 0x13; i++ {
		for j := i; j <= 0x14; j++ {
			if s[i].Spell() > s[j].Spell() {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// slotKind is a spell slot's class and level.
type slotKind struct {
	cls SpellClass
	lvl SpellProp
}

// Planned: how many spells of the level and class c's plan holds.
func (e *RuleState) Planned(c *Character, lvl SpellProp, cls SpellClass) int {
	n := 0
	for _, v := range c.Plan {
		if v != 0 && e.SpellLevel(v) == lvl && e.SpellClass(v) == cls {
			n++
		}
	}
	return n
}

// listedAffect = the affect types ovr16:0FBC names after the spell that
// gives them (cs:0x1050 checks).
func listedAffect(t AffectType) bool {
	switch {
	case t == AffBless, t == AffCurse, t == affDetectMagic, t >= AffProtEvil && t <= affFriends,
		t == AffReadMagic, t == affShield, t >= affFindTraps && t <= affInvisible, t == affMirrorImage,
		t == affEnfeebled, t >= affAnimated && t <= affCauseDisease, t >= AffBestowCurse && t <= AffHaste,
		t == affProtMissiles, t == AffSlow, t == affProtEvilRadius, t == affProtGoodRadius,
		t == affPrayer, t >= affSnakeCharm && t <= affAsleep:
		return true
	}
	return false
}

// affectNames = ovr16:0FBC's named affects that no spell list entry covers.
var affectNames = map[AffectType]string{
	affStudyingManual:   "Studying Manual of Bodily Health",
	affTrainingManual:   "Training with Manual of Bodily Health",
	AffFeatherFall:      "Feather Fall",
	affHelpless:         "Helpless",
	affPrayerBonus:      "Prayer",
	affSick:             "Cause Disease",
	affMummyDisease:     "Dreaded Mummy Disease",
	affRepulsed:         "Repulsed",
	AffPoisoned:         "Poisoned",
	AffRegenerating:     "Regenerating",
	AffFireResistance:   "Fire Resistance",
	AffMonsterInvisible: "Invisible",
	affCamouflaged:      "Camouflaged",
	affDisplaced:        "Displaced",
	AffLaidOnHands:      "Laid on Hands today",
	AffCuredDisease:     "Cured Disease this week",
}

// AffectList = the names ovr16:0FBC lists for c: its spell effects.
func (e *RuleState) AffectList(c *Character) []string {
	var names []string
	for _, a := range c.Affects {
		t := a.Type
		name, known := "", true
		if listedAffect(t) {
			for s := 1; s <= LastSpell && name == ""; s++ {
				if AffectType(e.T().Spells[s][10]) == t {
					name = e.SpellName(s)
				}
			}
			if name == "" {
				name = "Funky--" + strconv.Itoa(int(t))
			}
		} else if name, known = affectNames[t]; !known {
			continue
		}
		names = append(names, name)
	}
	return names
}

// SpellLists are c's spell book (how often each is memorized on a rest)
// and the spells on c's scrolls not in it (Count 1 while being scribed).
func (e *RuleState) SpellLists(c *Character) (book, scrolls []BookSpell) {
	spell := func(s SpellID, n int) BookSpell {
		lvl, cls := e.SpellLevel(s), e.SpellClass(s)
		return BookSpell{ID: int(s), Name: e.SpellName(int(s)), Level: SpellGroups[e.SpellRec(SpellID(s), 1)], Count: n,
			Planned: e.Planned(c, SpellProp(byte(lvl)), cls), Slots: int(c.SpellSlot(cls, int(lvl)))}
	}
	for i := 1; i <= 0x37 && i <= len(c.SpellBook); i++ {
		if c.SpellBook[i-1] != 0 {
			book = append(book, spell(SpellID(i), SpellCount(c.Plan[:], SpellID(i))))
		}
	}
	for _, it := range e.Carried(c) {
		if !e.IsScroll(it) {
			continue
		}
		for j, s := range it.Spells {
			if s.Scribing() && !e.ScribesIt(c, it, j) {
				continue // another member scribes it
			}
			if k := int(s.Spell()); k != 0 && k <= len(c.SpellBook) && c.SpellBook[k-1] == 0 && c.SpellSlot(e.SpellClass(SpellID(k)), int(e.SpellLevel(SpellID(k)))) > 0 { // only what c can scribe (scribe refuses the rest)
				n := 0
				if s&ScrollScribing != 0 {
					n = 1
				}
				scrolls = append(scrolls, spell(s.Spell(), n))
			}
		}
	}
	return book, scrolls
}

// learnable: s is a spell c's book can hold, no item's or monster's
// power (SpellClassPower; POR's book of 56 ended before its 57..67).
func (e *RuleState) learnable(c *Character, s SpellID) bool {
	return s != 0 && int(s) <= len(c.SpellBook) && int(s) < len(e.T().Spells) && e.SpellClass(s) != SpellClassPower
}

// inBook: s is learnable and of a level its class has slots for (the slot
// tables' columns: POR's 3, so its Restoration, level 7, is never listed).
func (e *RuleState) inBook(c *Character, s SpellID) bool {
	if !e.learnable(c, s) {
		return false
	}
	levels := MaxSpellLevel
	t := e.T()
	switch e.SpellClass(s) {
	case SpellClassCleric:
		levels = len(t.ClericSpells[0])
	case SpellClassMage:
		levels = len(t.MageSpells[0])
	}
	return int(e.SpellLevel(s)) <= levels
}

// PlanSpell (extra) memorizes spell s on every rest once more (delta > 0)
// or once less, as far as c's slots for its level go.
func (e *RuleState) PlanSpell(c *Character, spell, delta int) {
	s := SpellID(spell)
	if !e.learnable(c, s) || c.SpellBook[s-1] == 0 || !e.CanStudy(2) {
		return
	}
	if delta < 0 {
		for i := len(c.Plan) - 1; i >= 0; i-- {
			if c.Plan[i] == s {
				c.Plan[i] = 0
				break
			}
		}
		return
	}
	if e.Planned(c, SpellProp(byte(e.SpellLevel(s))), e.SpellClass(s)) < int(c.SpellSlot(e.SpellClass(s), int(e.SpellLevel(s)))) {
		if i := slices.Index(c.Plan[:PORMemory], 0); i >= 0 {
			c.Plan[i] = s
		}
	}
}

// Scribe = the pick of ovr16:0B5B: c scribes spell s from a scroll on
// the next rest, or no longer when it is being scribed.
func (e *RuleState) Scribe(c *Character, spell int) {
	s := SpellID(spell)
	if !e.learnable(c, s) || !e.CanStudy(3) {
		return
	}
	for _, it := range e.Carried(c) {
		if !e.IsScroll(it) {
			continue
		}
		for k := range it.Spells {
			if it.Spells[k] == ItemPower(s)|ScrollScribing && e.ScribesIt(c, it, k) {
				it.Spells[k] = ItemPower(s)
				e.MarkScribe(nil, it, k)
				return
			}
		}
	}
	switch {
	case c.SpellBook[s-1] != 0:
		e.narrate("You already know that spell")
		return
	case c.SpellSlot(e.SpellClass(s), int(e.SpellLevel(s))) == 0:
		e.narrate("You can not scribe that spell")
		return
	}
	for _, it := range e.Carried(c) { // (any item's spell bytes count here, not only scrolls)
		for k := range it.Spells {
			if it.Spells[k] == ItemPower(s) {
				it.Spells[k] |= ScrollScribing
				e.MarkScribe(c, it, k)
				return
			}
		}
	}
}

// ---- ovr21: resting ----

// subRest = ovr21:047A (count, unit): take count units off the rest time,
// borrowing from the larger units; all of it when that runs out.
func (e *RuleState) subRest(count, unit int) {
	r := &e.Camp.Rest
	if r[4] == 0 && r[3] == 0 && r[2] == 0 && r[1] == 0 {
		return
	}
	for count > r[unit] {
		k := unit + 1
		for r[k] == 0 && k < 5 {
			k++
		}
		if k == 5 {
			*r = [7]int{}
			count = 0
			continue
		}
		for j := k; j >= unit+1; j-- {
			r[j]--
			r[j-1] += clockLimits[j-1]
		}
	}
	r[unit] -= count
	e.NormalizeRest()
}

// NormalizeRest = ovr21:03A3: carries, months fold into days, at most 99.
func (e *RuleState) NormalizeRest() {
	r := &e.Camp.Rest
	for i := 0; i < 5; i++ {
		if r[i] >= clockLimits[i] {
			r[i+1] += r[i] / clockLimits[i]
			r[i] %= clockLimits[i]
		}
	}
	if r[5] > 0 {
		r[4] += clockLimits[4] * r[5]
		r[5] = 0
		if r[4] > 0x63 {
			r[4] = 0x63
		}
	}
}

// healTick = ovr21:0871: every 24 hours of rest everyone heals 1 HP;
// true when they did.
func (e *RuleState) healTick() bool {
	e.Camp.Heal++
	if e.Camp.Heal < 0x120 {
		return false
	}
	for _, c := range e.Party {
		e.Heal(c, 1, false)
	}
	e.narrate("The Whole Party Is Healed")
	e.Camp.Heal = 0
	return true
}

// memorizeOne = ovr21:0925 (char, var check): with check clear the first
// waiting spell is memorised (and check set); returns the level of the
// next waiting spell, 0 when none is left.
func (e *RuleState) memorizeOne(c *Character, check *bool) int {
	found := 0
	for i := 0; i <= 0x14 && found == 0; i++ {
		s := c.Memorized[i]
		if !s.Pending() {
			continue
		}
		if *check {
			found = int(e.SpellLevel(s))
			continue
		}
		c.Memorized[i] = s.Spell()
		e.announceCast(c, "has memorized", int(c.Memorized[i]))
		*check = true
	}
	return found
}

// ScribeOne = ovr21:09E7 (char, var check): memorizeOne for scrolls being
// scribed; the spell goes into the book and off the scroll.
func (e *RuleState) ScribeOne(c *Character, check *bool) int {
	found := 0
	for _, it := range e.Carried(c) {
		if found != 0 {
			break
		}
		if !e.IsScroll(it) {
			continue
		}
		for k := 1; k <= 3 && found == 0; k++ {
			s := it.Spells[k-1]
			if !s.Scribing() || !e.ScribesIt(c, it, k-1) {
				continue
			}
			if *check {
				found = int(e.SpellLevel(s.Spell()))
				continue
			}
			sp := s.Spell()
			c.SpellBook[int(sp)-1] = 1
			e.scrollUsed(c, it, sp)
			e.announceCast(c, "has scribed", int(sp))
			*check = true
		}
	}
	return found
}

// studyStep = the shared tail of ovr21:0B0A / 0BB2: learn what is due and
// time the next spell (3 steps per spell level).
func (e *RuleState) studyStep(i int, c *Character, check bool) {
	lvl := e.ScribeOne(c, &check)
	if lvl == 0 {
		lvl = e.memorizeOne(c, &check)
	}
	e.Camp.Study[i] = Count(byte(lvl * 3))
}

// memTick = ovr21:0B0A: every five minutes.
func (e *RuleState) memTick() {
	for i, c := range e.Party {
		if i+1 >= len(e.Camp.Study) {
			break
		}
		e.Camp.Study[i+1]--
		if e.Camp.Study[i+1] == 0 && c.Study == 0 {
			e.studyStep(i+1, c, false)
		}
	}
}

// hourTick = ovr21:0BB2 (var count): every hour the study hours run down.
func (e *RuleState) hourTick(count *int) {
	*count++
	if *count < 12 {
		return
	}
	*count = 0
	for i, c := range e.Party {
		if i+1 >= len(e.Camp.Study) {
			break
		}
		if int8(c.Study) > 0 {
			c.Study--
			if c.Study == 0 {
				e.studyStep(i+1, c, true)
			}
		}
	}
}

// BeginRest = Rest ovr21:0C9A's opening: the study steps and the effect
// ticks start over, the rest's minutes and hour count from 0.
func (e *RuleState) BeginRest() {
	for i := range e.Camp.Study {
		e.Camp.Study[i] = 0
	}
	for i := 1; i < len(e.EffTick); i++ {
		e.EffTick[i] = 1
	}
	e.Camp.RestMins, e.Camp.RestHour, e.Camp.StopRest = 0, 0, false
}

// RestDone: the rest is over before its next step: its time is up, a key
// stopped it (StopRest) or, untilHealed, nobody is wounded any more.
func (e *RuleState) RestDone(untilHealed bool) bool {
	r := &e.Camp.Rest
	return r[4] == 0 && r[3] == 0 && r[2] == 0 && r[1] == 0 || untilHealed && e.HealMinutes() == 0 || e.Camp.StopRest
}

// stepRest takes five minutes of the rest (Elapse).
func (e *RuleState) stepRest() (s restStep) {
	e.subRest(5, 1)
	e.AdvanceClock(5, 1)
	s.Healed = e.healTick()
	e.memTick()
	e.hourTick(&e.Camp.RestHour)
	if e.Camp.RestMins += 5; e.Camp.RestMins == LongRest { // the rule set's long rest (ruling 6: the original's does nothing)
		for _, c := range e.Party {
			e.Rules().Rested(e, c)
		}
	}
	if every := e.Game.RestEvery; every > 0 {
		e.Camp.Interval++
		if e.Camp.Interval >= int(every) {
			e.Camp.Interval = 0
			if e.Dice(1, 100) <= int(e.Game.RestChance) {
				e.narrate("The Party is rudely interrupted!")
				s.Over, s.Interrupted = true, true
				return
			}
		}
	}
	return
}

// MoveMember (extra) moves member from to place to in the marching order,
// while camp is up and only then; the caller selects it there.
func (e *RuleState) MoveMember(from, to int) error {
	n := len(e.Party)
	if !e.Camp.On || from < 0 || from >= n || to < 0 || to >= n || from == to {
		return ErrNotNow
	}
	c := e.Party[from]
	e.Party = slices.Insert(slices.Delete(e.Party, from, from+1), to, c)
	return nil
}

// SpellCount is how often s stands in list.
func SpellCount(list []SpellID, s SpellID) int {
	n := 0
	for _, v := range list {
		if v == s {
			n++
		}
	}
	return n
}
