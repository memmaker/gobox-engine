package importer

import "gobox/ecl"

// scriptFixes are pod's special cases in its decompiler (ecl.Fix: the
// site exactly once or the import fails).
var podScriptFixes = []ecl.Fix{
	// area 1: the developer debug menu "PARTY PLUS WHAT OFFSET? ( >= 750 EXITS ROUTINE)" poked MonsterHit[&Word00C1] (a raw write into bank 0 at offset 1+n, n 0..749).
	// Why: Felix 2026-11-13: remove the whole debug branch (the "SET ANY PARTY VARIABLES?" question and the poke loop); the routine just returns.
	{Area: 1, Old: `L_887C:
    Text("DO YOU WISH TO SET ANY PARTY VARIABLES?")
    if L_8820() { return }
    Text("PARTY PLUS WHAT OFFSET? ( >= 750 EXITS ROUTINE)")
    &Word00C1 = InputNumber(3)
    if 750 <= &Word00C1 { return }
    Text("WHAT WILL THE VARIABLE BE SET TO?")
    Word00C3 = InputNumber(3)
    MonsterHit[&Word00C1] = Word00C3
    goto L_887C
`, New: `L_887C:
    return
`},
	// area 2: the same debug menu (see area 1), Felix 2026-11-13.
	{Area: 2, Old: `    goto L_847D

    L_847D:
    Text("DO YOU WISH TO SET ANY PARTY VARIABLES?")
    if L_841E() { return }
    Text("PARTY PLUS WHAT OFFSET? ( >= 750 EXITS ROUTINE)")
    &Word00C1 = InputNumber(3)
    if 750 <= &Word00C1 { return }
    Text("WHAT WILL THE VARIABLE BE SET TO?")
    Word00C3 = InputNumber(3)
    MonsterHit[&Word00C1] = Word00C3
    goto L_847D
`, New: `    return
`},
	// area 3: the same debug menu (see area 1), Felix 2026-11-13.
	{Area: 3, Old: `    goto L_8475

    L_8475:
    Text("DO YOU WISH TO SET ANY PARTY VARIABLES?")
    if L_843A() { return }
    Text("PARTY PLUS WHAT OFFSET? ( >= 750 EXITS ROUTINE)")
    &Word00C1 = InputNumber(3)
    if 750 <= &Word00C1 { return }
    Text("WHAT WILL THE VARIABLE BE SET TO?")
    Word00C3 = InputNumber(3)
    MonsterHit[&Word00C1] = Word00C3
    goto L_8475
`, New: `    return
`},
	// area 22: the original wrote SelCharisma_1 into Word00AB[159] (bank 0 $14A), never read: an original bug with no effect.
	// Why: Felix 2026-11-13: drop the write.
	{Area: 22, Old: `    SelectMember(Word00AC)
    Word00AB[159] = SelCharisma_1
    goto L_8416
`, New: `    SelectMember(Word00AC)
    goto L_8416
`},
}
