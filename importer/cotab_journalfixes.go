package importer

import "gobox/ecl"

// JournalFixes are the import's special cases (ecl.Fix: the site
// exactly once or the import fails) at each site where the original named
// a handout of the printed Adventurer's Journal by its number ("YOU RECORD
// IT IN JOURNAL ENTRY 31", the number often a PRINT of its own after a
// routine's lead): the reference leaves the text, as the text window
// showed it, and Journal(entry) puts the entry up (journal.go). Made by
// following the PRINTs as the text window did; TestJournalSitesNamed
// checks no reference is left. CHANGES: COTAB/SotSB script, journal
// references.
var JournalFixes = []ecl.Fix{
	// area 1: the original printed "... JOURNAL ENTRY " then "31. 'PERHAPS THE SAGE WILL HELP. YOU CAN GET WEAPONS FROM ", a reference to the printed Adventurer's Journal.
	{Area: 1, Old: `    L_9D99()
    Print("31. 'PERHAPS THE SAGE WILL HELP. YOU CAN GET WEAPONS FROM ")
`, New: `    L_9D99()
    Print(". 'PERHAPS THE SAGE WILL HELP. YOU CAN GET WEAPONS FROM ")
    Journal(JournalEntry31)
`},
	// area 1: the original printed "... JOURNAL ENTRY " then "17.", a reference to the printed Adventurer's Journal.
	{Area: 1, Old: `    L_9D99()
    Print("17.")
`, New: `    L_9D99()
    Print(".")
    Journal(JournalEntry17)
`},
	// area 1: the original printed "... JOURNAL ENTRY " then "38.", a reference to the printed Adventurer's Journal.
	{Area: 1, Old: `    L_9D99()
    Print("38.")
`, New: `    L_9D99()
    Print(".")
    Journal(JournalEntry38)
`},
	// area 1: the original printed "... JOURNAL ENTRY " then "19.", a reference to the printed Adventurer's Journal.
	{Area: 1, Old: `    L_9D99()
    Print("19.")
`, New: `    L_9D99()
    Print(".")
    Journal(JournalEntry19)
`},
	// area 1: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 1, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: `    Print(" AND YOU RECORD IT IN YOUR JOURNAL")
`},
	// area 2: the original printed "YOU FIND INFORMATION ON HIS BODY AND LOG IT AS JOURNAL ", a reference to the printed Adventurer's Journal.
	{Area: 2, Old: `    Print("YOU FIND INFORMATION ON HIS BODY AND LOG IT AS JOURNAL ")
`, New: `    Print("YOU FIND INFORMATION ON HIS BODY AND LOG IT IN YOUR JOURNAL")
`},
	// area 2: the original printed "ENTRY 4.", a reference to the printed Adventurer's Journal.
	{Area: 2, Old: `    Print("ENTRY 4.")
`, New: `    Print(".")
    Journal(JournalEntry4)
`},
	// area 3: the original printed "RECORD IT IN JOURNAL ENTRY 41.", a reference to the printed Adventurer's Journal.
	{Area: 3, Old: `    Print("RECORD IT IN JOURNAL ENTRY 41.")
`, New: `    Print("RECORD IT IN YOUR JOURNAL.")
    Journal(JournalEntry41)
`},
	// area 4: the original printed "... JOURNAL ENTRY " then "11.", a reference to the printed Adventurer's Journal.
	{Area: 4, Old: `    L_972A()
    Print("11.")
`, New: `    L_972A()
    Print(".")
    Journal(JournalEntry11)
`},
	// area 4: the original printed "... JOURNAL ENTRY " then "54.", a reference to the printed Adventurer's Journal.
	{Area: 4, Old: `    L_972A()
    Print("54.")
`, New: `    L_972A()
    Print(".")
    Journal(JournalEntry54)
`},
	// area 4: the original printed "... JOURNAL ENTRY " then "53.", a reference to the printed Adventurer's Journal.
	{Area: 4, Old: `    L_972A()
    Print("53.")
`, New: `    L_972A()
    Print(".")
    Journal(JournalEntry53)
`},
	// area 4: the original printed "... JOURNAL ENTRY " then "27.", a reference to the printed Adventurer's Journal.
	{Area: 4, Old: `    L_972A()
    Print("27.")
`, New: `    L_972A()
    Print(".")
    Journal(JournalEntry27)
`},
	// area 4: the original printed "... JOURNAL ENTRY " then "26.", a reference to the printed Adventurer's Journal.
	{Area: 4, Old: `    L_972A()
    Print("26.")
`, New: `    L_972A()
    Print(".")
    Journal(JournalEntry26)
`},
	// area 4: the original printed "... JOURNAL ENTRY " then "9. OTHER ITEMS ALSO COME TO YOUR ATTENTION.", a reference to the printed Adventurer's Journal.
	{Area: 4, Old: `    L_972A()
    Print("9. OTHER ITEMS ALSO COME TO YOUR ATTENTION.")
`, New: `    L_972A()
    Print(". OTHER ITEMS ALSO COME TO YOUR ATTENTION.")
    Journal(JournalEntry9)
`},
	// area 4: the original printed "... JOURNAL ENTRY " then "29.", a reference to the printed Adventurer's Journal.
	{Area: 4, Old: `    L_972A()
    Print("29.")
`, New: `    L_972A()
    Print(".")
    Journal(JournalEntry29)
`},
	// area 4: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 4, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: `    Print(" AND YOU RECORD IT IN YOUR JOURNAL")
`},
	// area 16: the original printed "REMARKS AS JOURNAL ENTRY 22.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    Print("REMARKS AS JOURNAL ENTRY 22.")
`, New: `    Print("REMARKS IN YOUR JOURNAL.")
    Journal(JournalEntry22)
`},
	// area 16: the original printed "YOU RECORD THE FIND AT JOURNAL ENTRY 52.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    Print("YOU RECORD THE FIND AT JOURNAL ENTRY 52.")
`, New: `    Print("YOU RECORD THE FIND AT JOURNAL ENTRY 52.")
    Journal(JournalEntry52)
`},
	// area 16: the original printed "YOU RECORD THIS AS JOURNAL ENTRY 34.", a reference to the printed Adventurer's Journal.
	{Area: 16, Old: `    Print("YOU RECORD THIS AS JOURNAL ENTRY 34.")
`, New: `    Print("YOU RECORD THIS IN YOUR JOURNAL.")
    Journal(JournalEntry34)
`},
	// area 17: the original printed "AS JOURNAL ENTRY 3.", a reference to the printed Adventurer's Journal.
	{Area: 17, Old: `    Print("AS JOURNAL ENTRY 3.")
`, New: `    Print("IN YOUR JOURNAL.")
    Journal(JournalEntry3)
`},
	// area 18: the original printed "IT'S CONTENTS AS JOURNAL ENTRY 46.", a reference to the printed Adventurer's Journal.
	{Area: 18, Old: `    Print("IT'S CONTENTS AS JOURNAL ENTRY 46.")
`, New: `    Print("IT'S CONTENTS IN YOUR JOURNAL.")
    Journal(JournalEntry46)
`},
	// area 18: the original printed "JOURNAL ENTRY 3.", a reference to the printed Adventurer's Journal.
	{Area: 18, Old: `    Print("JOURNAL ENTRY 3.")
`, New: `    Print("JOURNAL ENTRY 3.")
    Journal(JournalEntry3)
`},
	// area 18: the original printed "DOWN AS JOURNAL ENTRY 3.", a reference to the printed Adventurer's Journal.
	{Area: 18, Old: `    Print("DOWN AS JOURNAL ENTRY 3.")
`, New: `    Print("DOWN IN YOUR JOURNAL.")
    Journal(JournalEntry3)
`},
	// area 18: the original printed "RECORD IT AS JOURNAL ENTRY 20.", a reference to the printed Adventurer's Journal.
	{Area: 18, Old: `    Print("RECORD IT AS JOURNAL ENTRY 20.")
`, New: `    Print("RECORD IT IN YOUR JOURNAL.")
    Journal(JournalEntry20)
`},
	// area 32: the original printed "THIS AS JOURNAL ENTRY 32.", a reference to the printed Adventurer's Journal.
	{Area: 32, Old: `    Print("THIS AS JOURNAL ENTRY 32.")
`, New: `    Print("THIS IN YOUR JOURNAL.")
    Journal(JournalEntry32)
`},
	// area 32: the original printed "PURSE.  YOU ENTER THIS AS JOURNAL ENTRY 21.", a reference to the printed Adventurer's Journal.
	{Area: 32, Old: `    Print("PURSE.  YOU ENTER THIS AS JOURNAL ENTRY 21.")
`, New: `    Print("PURSE.  YOU ENTER THIS IN YOUR JOURNAL.")
    Journal(JournalEntry21)
`},
	// area 32: the original printed "YOU ENTER HER STATEMENT AS JOURNAL ENTRY 50.", a reference to the printed Adventurer's Journal.
	{Area: 32, Old: `    Print("YOU ENTER HER STATEMENT AS JOURNAL ENTRY 50.")
`, New: `    Print("YOU ENTER HER STATEMENT IN YOUR JOURNAL.")
    Journal(JournalEntry50)
`},
	// area 33: the original printed "AS JOURNAL ENTRY 51.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    Print("AS JOURNAL ENTRY 51.")
`, New: `    Print("IN YOUR JOURNAL.")
    Journal(JournalEntry51)
`},
	// area 33: the original printed "HIMSELF AND YOU RECORD HIS REMARKS AS JOURNAL ENTRY 12.", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    Print("HIMSELF AND YOU RECORD HIS REMARKS AS JOURNAL ENTRY 12.")
`, New: `    Print("HIMSELF AND YOU RECORD HIS REMARKS IN YOUR JOURNAL.")
    Journal(JournalEntry12)
`},
	// area 33: the original printed "AS JOURNAL ENTRY 51", a reference to the printed Adventurer's Journal.
	{Area: 33, Old: `    Print("AS JOURNAL ENTRY 51")
`, New: `    Print("IN YOUR JOURNAL")
    Journal(JournalEntry51)
`},
	// area 34: the original printed " RECORD HIS HURRIED WHISPERS AS JOURNAL ENTRY 12.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    Print(" RECORD HIS HURRIED WHISPERS AS JOURNAL ENTRY 12.")
`, New: `    Print(" RECORD HIS HURRIED WHISPERS IN YOUR JOURNAL.")
    Journal(JournalEntry12)
`},
	// area 34: the original printed "DEXAM SPEAKS.  YOU RECORD HIS SPEECH AS JOURNAL ", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    Text("DEXAM SPEAKS.  YOU RECORD HIS SPEECH AS JOURNAL ")
`, New: `    Text("DEXAM SPEAKS.  YOU RECORD HIS SPEECH IN YOUR JOURNAL")
`},
	// area 34: the original printed "ENTRY 30.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    Print("ENTRY 30.")
`, New: `    Print(".")
    Journal(JournalEntry30)
`},
	// area 34: the original printed "HIS STATEMENT IS IN JOURNAL ENTRY 7.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    Print("HIS STATEMENT IS IN JOURNAL ENTRY 7.")
`, New: `    Print("HIS STATEMENT IS IN YOUR JOURNAL.")
    Journal(JournalEntry7)
`},
	// area 34: the original printed "YOU PLACE IT IN YOUR JOURNAL AS ENTRY 59.", a reference to the printed Adventurer's Journal.
	{Area: 34, Old: `    Print("YOU PLACE IT IN YOUR JOURNAL AS ENTRY 59.")
`, New: `    Print("YOU PLACE IT IN YOUR JOURNAL AS ENTRY 59.")
    Journal(JournalEntry59)
`},
	// area 35: the original printed "DOWN AS JOURNAL ENTRY 23.", a reference to the printed Adventurer's Journal.
	{Area: 35, Old: `    Print("DOWN AS JOURNAL ENTRY 23.")
`, New: `    Print("DOWN IN YOUR JOURNAL.")
    Journal(JournalEntry23)
`},
	// area 50: the original printed "SHE CONTINUES, AND YOU RECORD IT IN JOURNAL", a reference to the printed Adventurer's Journal.
	{Area: 50, Old: `    Text("SHE CONTINUES, AND YOU RECORD IT IN JOURNAL")
`, New: `    Text("SHE CONTINUES, AND YOU RECORD IT IN YOUR JOURNAL")
`},
	// area 50: the original printed " ENTRY 44. WILL ", a reference to the printed Adventurer's Journal.
	{Area: 50, Old: `    Print(" ENTRY 44. WILL ")
`, New: `    Print(". WILL ")
    Journal(JournalEntry44)
`},
	// area 51: the original printed "RECORD AS JOURNAL ENTRY 15.", a reference to the printed Adventurer's Journal.
	{Area: 51, Old: `    Print("RECORD AS JOURNAL ENTRY 15.")
`, New: `    Print("RECORD IN YOUR JOURNAL.")
    Journal(JournalEntry15)
`},
	// area 64: the original printed "... JOURNAL ENTRY " then "25. THEN, THE SPIRIT FADES.", a reference to the printed Adventurer's Journal.
	{Area: 64, Old: `    L_9B84()
    Print("25. THEN, THE SPIRIT FADES.")
`, New: `    L_9B84()
    Print(". THEN, THE SPIRIT FADES.")
    Journal(JournalEntry25)
`},
	// area 64: the original printed "... JOURNAL ENTRY " then "33.", a reference to the printed Adventurer's Journal.
	{Area: 64, Old: `    L_9B84()
    Print("33.")
`, New: `    L_9B84()
    Print(".")
    Journal(JournalEntry33)
`},
	// area 64: the original printed "... JOURNAL ENTRY " then "56. 'HURRY ON, HURRY, HURRY!' HE LEAVES.", a reference to the printed Adventurer's Journal.
	{Area: 64, Old: `    L_9B84()
    Print("56. 'HURRY ON, HURRY, HURRY!' HE LEAVES.")
`, New: `    L_9B84()
    Print(". 'HURRY ON, HURRY, HURRY!' HE LEAVES.")
    Journal(JournalEntry56)
`},
	// area 64: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 64, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: `    Print(" AND YOU RECORD IT IN YOUR JOURNAL")
`},
	// area 66: the original printed "... JOURNAL ENTRY " then "5. WILL YOU GO WITH HIM?", a reference to the printed Adventurer's Journal.
	{Area: 66, Old: `    L_97C0()
    Print("5. WILL YOU GO WITH HIM?")
`, New: `    L_97C0()
    Print(". WILL YOU GO WITH HIM?")
    Journal(JournalEntry5)
`},
	// area 66: the original printed "... JOURNAL ENTRY " then "57.", a reference to the printed Adventurer's Journal.
	{Area: 66, Old: `    L_97C0()
    Print("57.")
`, New: `    L_97C0()
    Print(".")
    Journal(JournalEntry57)
`},
	// area 66: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 66, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: `    Print(" AND YOU RECORD IT IN YOUR JOURNAL")
`},
	// area 67: the original printed "JOURNAL ENTRY 48.", a reference to the printed Adventurer's Journal.
	{Area: 67, Old: `    Print("JOURNAL ENTRY 48.")
`, New: `    Print("JOURNAL ENTRY 48.")
    Journal(JournalEntry48)
`},
	// area 80: the original printed "... JOURNAL ENTRY " then "18. THEN, HE VANISHES.", a reference to the printed Adventurer's Journal.
	{Area: 80, Old: `    L_9BC5()
    Print("18. THEN, HE VANISHES.")
`, New: `    L_9BC5()
    Print(". THEN, HE VANISHES.")
    Journal(JournalEntry18)
`},
	// area 80: the original printed "... JOURNAL ENTRY " then "58.", a reference to the printed Adventurer's Journal.
	{Area: 80, Old: `    L_9BC5()
    Print("58.")
`, New: `    L_9BC5()
    Print(".")
    Journal(JournalEntry58)
`},
	// area 80: the original printed "... JOURNAL ENTRY " then "55.", a reference to the printed Adventurer's Journal.
	{Area: 80, Old: `    L_9BC5()
    Print("55.")
`, New: `    L_9BC5()
    Print(".")
    Journal(JournalEntry55)
`},
	// area 80: the original printed "... JOURNAL ENTRY " then "45. THEN, YOU PART COMPANY.", a reference to the printed Adventurer's Journal.
	{Area: 80, Old: `    L_9BC5()
    Print("45. THEN, YOU PART COMPANY.")
`, New: `    L_9BC5()
    Print(". THEN, YOU PART COMPANY.")
    Journal(JournalEntry45)
`},
	// area 80: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 80, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: `    Print(" AND YOU RECORD IT IN YOUR JOURNAL")
`},
	// area 81: the original printed "... JOURNAL ENTRY " then "36.", a reference to the printed Adventurer's Journal.
	{Area: 81, Old: `    L_9C0E()
    Print("36.")
`, New: `    L_9C0E()
    Print(".")
    Journal(JournalEntry36)
`},
	// area 81: the original printed " AND YOU RECORD IT IN JOURNAL ENTRY ", a reference to the printed Adventurer's Journal.
	{Area: 81, Old: `    Print(" AND YOU RECORD IT IN JOURNAL ENTRY ")
`, New: `    Print(" AND YOU RECORD IT IN YOUR JOURNAL")
`},
	// area 80: the original printed the note's entry, a number of a table
	// (T_9D10: 24, 35, 42), after the lead; the text window never showed it.
	{Area: 80, Old: `    L_9BC5()
    Scratch7 = T_9D10[Word4C9F]
    Print(Scratch7)
`, New: `    L_9BC5()
    Scratch7 = T_9D10[Word4C9F]
    Journal(Scratch7)
`},
}
