package games

import eclscript "gobox/ecl/script"

// journalFixes are the import's special cases (eclscript.Fix: the site
// exactly once or the import fails) at each site where the original named
// an entry of the printed Adventurer's Journal by its number ("AND YOU
// RECORD IT AS JOURNAL ENTRY", the number often a PRINT of its own after a
// routine's lead, or a variable's): the text keeps what the text window
// showed (a recording line and its number were left out) and
// Journal(entry) puts the entry up (journal.go). Made by following the
// PRINTs as the text window did; TestJournalSitesNamed checks no reference
// is left without its Journal. CHANGES: COTAB/SotSB script, journal
// references.
var journalFixes = []eclscript.Fix{
	// area 16: the original printed "... JOURNAL ENTRY " then "37.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    L_9B39()
    Print("37.")
`, New: `    L_9B39()
    Journal(JournalEntry37)
`},
	// area 16: the original printed "... JOURNAL ENTRY " then Local09, a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    L_9B39()
    Print(Local09)
`, New: `    L_9B39()
    Journal(Local09)
`},
	// area 16: the original printed "... JOURNAL ENTRY " then "47.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    L_9B39()
    Print("47.")
`, New: `    L_9B39()
    Journal(JournalEntry47)
`},
	// area 16: the original printed "... JOURNAL ENTRY " then "5.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    L_9B39()
    Print("5.")
`, New: `    L_9B39()
    Journal(JournalEntry5)
`},
	// area 16: the original printed "... JOURNAL ENTRY " then "21.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    L_9B39()
    Print("21.")
`, New: `    L_9B39()
    Journal(JournalEntry21)
`},
	// area 16: the original printed "... JOURNAL ENTRY " then Scratch2, a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    L_9B39()
    Print(Scratch2)
`, New: `    L_9B39()
    Journal(Scratch2)
`},
	// area 16: the original printed "... JOURNAL ENTRY " then "7.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    L_9B39()
    Print("7.")
`, New: `    L_9B39()
    Journal(JournalEntry7)
`},
	// area 16: the original printed "... JOURNAL ENTRY " then "39.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    L_9B39()
    Print("39.")
`, New: `    L_9B39()
    Journal(JournalEntry39)
`},
	// area 16: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: ``},
	// area 32: the original printed "... JOURNAL ENTRY " then " 57.", a reference to the printed Adventurer's Journal.
	{Area: 32, Old: `    L_9863()
    Print(" 57.")
`, New: `    L_9863()
    Journal(JournalEntry57)
`},
	// area 32: the original printed "... JOURNAL ENTRY " then " 31.", a reference to the printed Adventurer's Journal.
	{Area: 32, Old: `    L_9863()
    Print(" 31.")
`, New: `    L_9863()
    Journal(JournalEntry31)
`},
	// area 32: the original printed "AND YOU RECORD IT AS JOURNAL ENTRY", a reference to the printed Adventurer's Journal.
	{Area: 32, Old: `    Print("AND YOU RECORD IT AS JOURNAL ENTRY")
`, New: ``},
	// area 33: the original printed "... JOURNAL ENTRY " then " 22.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9AF0()
    Print(" 22.")
`, New: `    L_9AF0()
    Journal(JournalEntry22)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 60.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9AF0()
    Print(" 60.")
`, New: `    L_9AF0()
    Journal(JournalEntry60)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 25.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9AF0()
    Print(" 25.")
`, New: `    L_9AF0()
    Journal(JournalEntry25)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 68.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9D3B()
    Print(" 68.")
`, New: `    L_9D3B()
    Journal(JournalEntry68)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 41.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9D3B()
    Print(" 41.")
`, New: `    L_9D3B()
    Journal(JournalEntry41)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 32.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9D3B()
    Print(" 32.")
`, New: `    L_9D3B()
    Journal(JournalEntry32)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 19.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9AF0()
    Print(" 19.")
`, New: `    L_9AF0()
    Journal(JournalEntry19)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 1.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9D3B()
    Print(" 1.")
`, New: `    L_9D3B()
    Journal(JournalEntry1)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 71.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9AF0()
    Print(" 71.")
`, New: `    L_9AF0()
    Journal(JournalEntry71)
`},
	// area 33: the original printed "... JOURNAL ENTRY " then " 38.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    L_9D3B()
    Print(" 38.")
`, New: `    L_9D3B()
    Journal(JournalEntry38)
`},
	// area 33: the original printed "THE WELL REMINDS YOU OF THE MAP AT JOURNAL ENTRY 1.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    Text("THE WELL REMINDS YOU OF THE MAP AT JOURNAL ENTRY 1.")
`, New: `    Text("THE WELL REMINDS YOU OF THE MAP AT JOURNAL ENTRY 1.")
    Journal(JournalEntry1)
`},
	// area 33: the original printed "AND YOU RECORD IT AS JOURNAL ENTRY", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    Print("AND YOU RECORD IT AS JOURNAL ENTRY")
`, New: ``},
	// area 34: the original printed "... JOURNAL ENTRY " then " 43.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    L_9ABC()
    Print(" 43.")
`, New: `    L_9ABC()
    Journal(JournalEntry43)
`},
	// area 34: the original printed "... JOURNAL ENTRY " then " 16. YOU TAKE THE MAP ", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    L_9ABC()
    Print(" 16. YOU TAKE THE MAP ")
`, New: `    L_9ABC()
    Print(" 16. YOU TAKE THE MAP ")
    Journal(JournalEntry16)
`},
	// area 34: the original printed "... JOURNAL ENTRY " then " 45.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    L_9ABC()
    Print(" 45.")
`, New: `    L_9ABC()
    Journal(JournalEntry45)
`},
	// area 34: the original printed "... JOURNAL ENTRY " then " 69.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    L_9ABC()
    Print(" 69.")
`, New: `    L_9ABC()
    Journal(JournalEntry69)
`},
	// area 34: the original printed "... JOURNAL ENTRY " then " 55.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    L_9ABC()
    Print(" 55.")
`, New: `    L_9ABC()
    Journal(JournalEntry55)
`},
	// area 34: the original printed "... JOURNAL ENTRY " then " 49.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    L_9ABC()
    Print(" 49.")
`, New: `    L_9ABC()
    Journal(JournalEntry49)
`},
	// area 34: the original printed "... JOURNAL ENTRY " then " 40.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    L_9ABC()
    Print(" 40.")
`, New: `    L_9ABC()
    Journal(JournalEntry40)
`},
	// area 34: the original printed "... JOURNAL ENTRY " then " 53.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    L_9ABC()
    Print(" 53.")
`, New: `    L_9ABC()
    Journal(JournalEntry53)
`},
	// area 34: the original printed "AND YOU RECORD IT AS JOURNAL ENTRY", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    Print("AND YOU RECORD IT AS JOURNAL ENTRY")
`, New: ``},
	// area 50: the original printed "... JOURNAL ENTRY " then Local04, a reference to the printed Adventurer's Journal.
	{Area: 50, Old: `def L_988E() {
    Print(" AS JOURNAL ENTRY ")
    Print(Local04)
`, New: `def L_988E() {
    Print(" AS JOURNAL ENTRY ")
    Print(Local04)
    Journal(Local04)
`},
	// area 52: the original printed "... JOURNAL ENTRY " then Local04, a reference to the printed Adventurer's Journal.
	{Area: 52, Old: `def L_9CC5() {
    Print(" AS JOURNAL ENTRY ")
    Print(Local04)
`, New: `def L_9CC5() {
    Print(" AS JOURNAL ENTRY ")
    Print(Local04)
    Journal(Local04)
`},
	// area 64: the original printed "... JOURNAL ENTRY " then " 33.", a reference to the printed Adventurer's Journal.
	{Area: 64, Old: `    L_99D4()
    Print(" 33.")
`, New: `    L_99D4()
    Journal(JournalEntry33)
`},
	// area 64: the original printed "... JOURNAL ENTRY " then " 68.", a reference to the printed Adventurer's Journal.
	{Area: 64, Old: `    L_99D4()
    Print(" 68.")
`, New: `    L_99D4()
    Journal(JournalEntry68)
`},
	// area 64: the original printed "AND YOU RECORD IT AS JOURNAL ENTRY", a reference to the printed Adventurer's Journal.
	{Area: 64, Old: `    Print("AND YOU RECORD IT AS JOURNAL ENTRY")
`, New: ``},
	// area 65: the original printed "... JOURNAL ENTRY " then " 62.", a reference to the printed Adventurer's Journal.
	{Area: 65, Old: `    L_9B29()
    Print(" 62.")
`, New: `    L_9B29()
    Journal(JournalEntry62)
`},
	// area 65: the original printed "... JOURNAL ENTRY " then " 32.", a reference to the printed Adventurer's Journal.
	{Area: 65, Old: `    L_9B29()
    Print(" 32.")
`, New: `    L_9B29()
    Journal(JournalEntry32)
`},
	// area 65: the original printed "AND YOU RECORD IT AS JOURNAL ENTRY", a reference to the printed Adventurer's Journal.
	{Area: 65, Old: `    Print("AND YOU RECORD IT AS JOURNAL ENTRY")
`, New: ``},
	// area 66: the original printed "... JOURNAL ENTRY " then " 13.", a reference to the printed Adventurer's Journal.
	{Area: 66, Old: `    L_9B02()
    Print(" 13.")
`, New: `    L_9B02()
    Journal(JournalEntry13)
`},
	// area 66: the original printed "... JOURNAL ENTRY " then " 52.", a reference to the printed Adventurer's Journal.
	{Area: 66, Old: `    L_9B02()
    Print(" 52.")
`, New: `    L_9B02()
    Journal(JournalEntry52)
`},
	// area 66: the original printed "AND YOU RECORD IT AS JOURNAL ENTRY", a reference to the printed Adventurer's Journal.
	{Area: 66, Old: `    Print("AND YOU RECORD IT AS JOURNAL ENTRY")
`, New: ``},
	// area 68: the original printed "... JOURNAL ENTRY " then "3.", a reference to the printed Adventurer's Journal.
	{Area: 68, Old: `    L_9989()
    Print("3.")
`, New: `    L_9989()
    Journal(JournalEntry3)
`},
	// area 68: the original printed "... JOURNAL ENTRY " then "23.", a reference to the printed Adventurer's Journal.
	{Area: 68, Old: `    L_9989()
    Print("23.")
`, New: `    L_9989()
    Journal(JournalEntry23)
`},
	// area 68: the original printed "... JOURNAL ENTRY " then "58.", a reference to the printed Adventurer's Journal.
	{Area: 68, Old: `    L_9989()
    Print("58.")
`, New: `    L_9989()
    Journal(JournalEntry58)
`},
	// area 68: the original printed "... JOURNAL ENTRY " then "70.", a reference to the printed Adventurer's Journal.
	{Area: 68, Old: `    L_9989()
    Print("70.")
`, New: `    L_9989()
    Journal(JournalEntry70)
`},
	// area 68: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 68, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: ``},
	// area 80: the original printed "... JOURNAL ENTRY " then "2.", a reference to the printed Adventurer's Journal.
	{Area: 80, Old: `    L_9B09()
    Print("2.")
`, New: `    L_9B09()
    Print("2.")
    Journal(JournalEntry2)
`},
	// area 80: the original printed "YOU FIND A PACKET OF DOCUMENTS. ONE OF THEM IS THE LETTER FROM THE CLERK THAT YOU RECORDED AS JOURNAL ENTRY 69.", a reference to the printed Adventurer's Journal.
	{Area: 80, Old: `    Text("YOU FIND A PACKET OF DOCUMENTS. ONE OF THEM IS THE LETTER FROM THE CLERK THAT YOU RECORDED AS JOURNAL ENTRY 69.")
`, New: `    Text("YOU FIND A PACKET OF DOCUMENTS. ONE OF THEM IS THE LETTER FROM THE CLERK THAT YOU RECORDED AS JOURNAL ENTRY 69.")
    Journal(JournalEntry69)
`},
	// area 81: the original printed "... JOURNAL ENTRY " then Local04, a reference to the printed Adventurer's Journal.
	{Area: 81, Old: `def L_9CBE() {
    Print(" AS JOURNAL ENTRY ")
    Print(Local04)
`, New: `def L_9CBE() {
    Print(" AS JOURNAL ENTRY ")
    Print(Local04)
    Journal(Local04)
`},
	// area 82: the original printed "... JOURNAL ENTRY " then "56.", a reference to the printed Adventurer's Journal.
	{Area: 82, Old: `    L_9AC0()
    Print("56.")
`, New: `    L_9AC0()
    Print("56.")
    Journal(JournalEntry56)
`},
	// area 96: the original printed "... JOURNAL ENTRY " then "29.", a reference to the printed Adventurer's Journal.
	{Area: 96, Old: `    L_9D1D()
    Print("29.")
`, New: `    L_9D1D()
    Journal(JournalEntry29)
`},
	// area 96: the original printed "... JOURNAL ENTRY " then "64.", a reference to the printed Adventurer's Journal.
	{Area: 96, Old: `    L_9D1D()
    Print("64.")
`, New: `    L_9D1D()
    Journal(JournalEntry64)
`},
	// area 96: the original printed "... JOURNAL ENTRY " then "67.", a reference to the printed Adventurer's Journal.
	{Area: 96, Old: `    L_9D1D()
    Print("67.")
`, New: `    L_9D1D()
    Journal(JournalEntry67)
`},
	// area 96: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 96, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: ``},
	// area 97: the original printed "... JOURNAL ENTRY " then "36.", a reference to the printed Adventurer's Journal.
	{Area: 97, Old: `    Text("THE GIANT LEADER HANDS YOU A MAP")
    L_9659()
    Print("36.")
`, New: `    Text("THE GIANT LEADER HANDS YOU A MAP")
    L_9659()
    Journal(JournalEntry36)
`},
	// area 97: the original printed "... JOURNAL ENTRY " then "36.", a reference to the printed Adventurer's Journal.
	{Area: 97, Old: `    Text("YOU DISCOVER TWO MAPS IN THE GIANT'S BELONGINGS. YOU COPY ONE")
    L_9659()
    Print("36.")
`, New: `    Text("YOU DISCOVER TWO MAPS IN THE GIANT'S BELONGINGS. YOU COPY ONE")
    L_9659()
    Journal(JournalEntry36)
`},
	// area 97: the original printed "... JOURNAL ENTRY " then "66.", a reference to the printed Adventurer's Journal.
	{Area: 97, Old: `    L_9659()
    Print("66.")
`, New: `    L_9659()
    Journal(JournalEntry66)
`},
	// area 97: the original printed " AND YOU RECORD IT AS JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 97, Old: `    Print(" AND YOU RECORD IT AS JOURNAL ENTRY ")
`, New: ``},
	// area 16: the gem trade, 1 gem for every 100 platinum; the pooled coins are gold since Gc.
	// Why: DEFAULT (Felix may override): 100 platinum is 500 gold, so 1 gem for every 500 gold, the text to match.
	{Area: 16, Old: `    Text("'WE WILL GIVE YOU 1 GEM FOR EVERY 100 PLATINUM YOU GIVE US.'")
    L_9AF9()
    Text("YOU CAN TRADE ")
    if &Word7BAE < 100 {
        &Scratch1 = 0
        goto L_8E61
    }
    &Scratch1 = &Word7BAE
    &Scratch3 = &Word7BAE / 100
    &Scratch1 -= DivideRemainder
    goto L_8E61

    L_8E61:
    Print(&Scratch1)
    Print(" PLATINUM PIECES. HOW MANY WILL YOU TRADE?")
    &Scratch1 = InputNumber(5)
    if &Scratch1 != 0 { goto L_8E9D }
    goto L_8EE9

    L_8E9D:
    if &Scratch1 < 100 { &Scratch1 = 100 }
    if &Scratch1 > &Word7BAE { goto L_8F74 }
    &Scratch4 = &Scratch1 / 100
`, New: `    Text("'WE WILL GIVE YOU 1 GEM FOR EVERY 500 GOLD YOU GIVE US.'")
    L_9AF9()
    Text("YOU CAN TRADE ")
    if &Word7BAE < 500 {
        &Scratch1 = 0
        goto L_8E61
    }
    &Scratch1 = &Word7BAE
    &Scratch3 = &Word7BAE / 500
    &Scratch1 -= DivideRemainder
    goto L_8E61

    L_8E61:
    Print(&Scratch1)
    Print(" GOLD PIECES. HOW MANY WILL YOU TRADE?")
    &Scratch1 = InputNumber(5)
    if &Scratch1 != 0 { goto L_8E9D }
    goto L_8EE9

    L_8E9D:
    if &Scratch1 < 500 { &Scratch1 = 500 }
    if &Scratch1 > &Word7BAE { goto L_8F74 }
    &Scratch4 = &Scratch1 / 500
`},
}
