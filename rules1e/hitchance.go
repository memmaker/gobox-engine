package rules1e

// HitChance (extra, por's hover text) is a's chance in percent to hit b with one blow of
// the readied attack: attackSlots' number to reach and AttackHits' d20.
// ponytail: the affects that change the roll (bless, prayer, ...) are left
// out; a dry run of affect groups 0xB and the attacker/defender ones would add them.
func (e *RuleState) HitChance(a, b *Character, w Wield) int {
	if e.IsHelpless(b) {
		return 100
	}
	need := b.AC[0]
	if e.RangedFacing(a, b, w) {
		need = b.AC[1] - 2
	}
	bonus := int(need)
	e.RangePenalty(a, b, w, &bonus)
	need = ArmourClass(e.Rules().HitNeed(e, a, b, bonus))
	adj := int(int8(byte(e.Game.PartyHit)))
	if a.Side != SideParty {
		adj = int(int8(byte(e.Game.MonsterHit)))
	}
	n := 0
	for r := 2; r <= 20; r++ {
		roll := r
		if r == 20 {
			roll = 100
		}
		if roll+int(e.thac0(a, w))+adj >= int(need) {
			n++
		}
	}
	return n * 5
}
