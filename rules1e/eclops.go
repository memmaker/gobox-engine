package rules1e

// ECL opcode handlers, ported from ovr04. Each is named after its opcode;
// the dispatch table (ovr04:2F2D) maps the byte at the program counter to its handler.

// EffectTable = EclVm.EffectTableEntry: the Textile House's thieves-only
// well looks for Feather Fall in a table at 0x4900 (entries 0..63 an affect
// type, entry +64 the party index of its owner) that the original engine
// never kept (the Pascal's fix): the party's affects, in party order.
func (e *RuleState) EffectTable(i int) (uint16, bool) {
	const n = 64
	if i >= 2*n {
		return 0, false
	}
	k := 0
	for p, c := range e.Party {
		for _, a := range c.Affects {
			switch k {
			case i:
				return uint16(a.Type), true
			case i - n:
				return uint16(p), true
			}
			k++
		}
	}
	return 0, true
}
