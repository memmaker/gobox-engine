// Package script is the one decompiled form of ECL event scripts: Decode
// turns an ECL block into a Script (instruction-numbered commands), Write
// and Parse carry it as text, MarkExits and FoldVendorGates add what the
// frontend needs. The differences between the games' byte code are data: a
// typed Version, chosen at import; nothing here knows a game.
package ecl

// FormatRevision is the text form's revision; a change makes every import stale.
const FormatRevision = 2

// VersionID names one byte-code dialect.
type VersionID int

// Cmd is how a Version encodes one byte-code command.
type Cmd struct {
	Op    byte // the canonical command (form.go, ops.go)
	Fixed int  // operands always there
	Count int  // the operand (from 1) holding how many more follow; 0 none
	Known bool // the original's dispatcher has a case for it
}

// ExitPlace is a menu instruction of the decoded script of an area.
type ExitPlace struct{ Area, Instr int }

// NoExitRow in ExitOverrides: every row of the menu is a real choice.
const NoExitRow = -1

// Version is one dialect: where its code sits and what its commands are.
type Version struct {
	ID            VersionID
	Base          int               // the address of the code's first byte
	Cmds          [256]Cmd          // by byte-code command
	ExitOverrides map[ExitPlace]int // corrections of the exit heuristic
	// LoadMapSecondUnused: the dialect's engine ignores LoadMap's second
	// operand, so the text leaves it out.
	LoadMapSecondUnused bool
	// Skip[b] is how many operands the dialect's engine reads when it steps
	// over byte-code command b unrun (an If that fails: parse_command's own
	// table). It differs from Cmds[b].Fixed where the original's table did
	// (TestSkipCounts names each difference).
	Skip [256]byte
	// Calls names the engine routine each CALL operand (its raw value, an
	// address in the original) runs; an operand not listed runs nothing.
	Calls map[uint16]CallFn
	// Programs names what each PROGRAM operand does; one not listed does nothing.
	Programs map[byte]ProgramFn
	// Behaviour is how the dialect's engine runs the commands that differ
	// between the games (ruling S15).
	Behaviour Behaviour
}

// CombatScreen is what Combat opens when no monsters are loaded.
type CombatScreen int

const (
	CombatShop   CombatScreen = iota + 1 // the shop
	CombatTemple                         // the temple
	CombatVault                          // the vault (pod)
)

// Behaviour is the per-dialect behaviour of the commands whose encoding or
// effect differs between the games (one named behaviour each, decoded here).
type Behaviour struct {
	ReturnEndsScript  bool           // RETURN with no caller ends the script (BR's exitScript); otherwise the host's Exit hook runs and the script continues
	SetupMonsterSides bool           // SetupMonster has a 4th operand, the sides (bits 0..3) the monsters come from
	OutdoorsNear      bool           // SetupMonster outdoors asks for 2 squares whatever the script says (pod O055:03D7); BR keeps the asked distance
	InputTextOperand  bool           // InputString's first operand is the maximum length; otherwise the text window's width
	SelectMemberDrop  bool           // SelectMember's bit 7 refreshes and may drop a dead member; otherwise the bit is only masked off
	NoMonsterCombat   []CombatScreen // Combat with no monsters loaded: the first screen whose word is set opens (it clears the word)
	SymbolSetFrames   bool           // the symbol set picks only the frame set (pod's tileSet, pod-re eclmenu.pas:472 $39: Load8x8Set, no other files); otherwise it is the content set too (POR's, cotab's gameArea)
	// StoreByOperand: a variable operand's type says how many bytes a read or write takes (BR ecl.go
	// Width, GetN, PutN; buckrogers-go rules/ecl.go val, getVar, setVar); otherwise a cell is a word
	// whatever the operand (Pool's WriteGbvm)
	StoreByOperand bool
	// WhoEndsNoMember: a picked member ends the "no member" state a SelectMember that matched nobody left (BR
	// pO0D2_2F14 case 57, buckrogers-go opPickMember: noChar = false); the Pool's Who leaves it (cotab eclvm.pas EclWho does not touch characterNotFound)
	WhoEndsNoMember bool
	NPCSideBit      bool // AddNPC writes the side bit alone, the hostile one for monster 0x18 (por ovr04:2B7E); BR's EclAddNpc leaves the side as the record has it
	HitPadded       bool // a script's damage message pads the number to three columns (the Pool's Str(dmg:3)); BR's EclInflictDamage does not
	LostDuelResult  bool // a lost solo fight leaves the combat result at 0x80, the team beaten (BR postcbt.pas DetermineCombatResults); the Pool's leaves it 0
	NoCamp          bool // the dungeon menu has no camp: BR's explore.pas menu is Move, Area, View, Save, Change, Look; its only rest is the script's Program(9)
	FarSight        bool // the squares ahead counted before a wall run to 5 (BR eclmenu.pas MaxMonsterDist); otherwise to 2 (the Pool's ovr08:0482)
	ClearExitText   bool // ClearExit empties the text pane too (BR pO035_2DCC case 66, 2837); otherwise only the view comes back (pod O040:24AE)
	// ViewKept: a block's start keeps the view kind (3D view or overland) until the script's AreaKind command says otherwise (BR eclmenu.pas EclInit leaves the interface state alone); otherwise the start sets the 3D view (the Pool's ovr08:01B6)
	ViewKept bool
	// MessageDelayMs is the pause of one step of the game's message speed in ms (BR textrend.pas PauseOnly: Delay(messageSpeed * 100)); 0: the Pool's (the engine's messageDelayUnit)
	MessageDelayMs  int
	CreditsTreasure bool // TreasureOf's coins are one operand, the party's credits (BR eclvm.pas EclTreasure: poolCredits += var 1); otherwise platinum, gems and jewelry (pod O040:13E0)
	PictureAnimates bool // Picture marks the view drawn, so a wait under it cycles the picture's frames (cotab's and sotsb's b8B7D, pod's alike); otherwise only a step or Picture(255) does (POR's picDrawn, eclvm.pas:371, 1874)
}

// BRMessageDelayMs is BR's message speed unit: textrend.pas PauseOnly, Delay(messageSpeed * 100).
const BRMessageDelayMs = 100

// ProgramFn is a PROGRAM sub-command, by name (ruling S14: each game's
// codes map onto one set).
type ProgramFn int

const (
	ProgramNone      ProgramFn = iota
	ProgramPartyMenu           // the party menu
	ProgramEnding              // the adventure's ending
	ProgramCamp                // camp (pod: with its camp events); the script ends
	ProgramRest                // the rest menu (BR explore.pas RestMenu): the script goes on afterwards
	ProgramQuit                // leave the game
	ProgramSaveQuit            // save, then the ending
	ProgramUnported            // a sub-command the original has and the engine lacks: it stops the game, never silently nothing
)

// CallFn is an engine routine a script CALLs, by name.
type CallFn int

const (
	CallNone             CallFn = iota
	CallRedrawView              // redraw the view if the script changed it
	CallStepForward             // the party steps one square ahead
	CallWallAhead               // the wall the party faces into the wall variable
	CallDuelWin                 // the duel's winning setup
	CallDuelLose                // the duel's losing setup
	CallDuelSound               // the duel's sound, by the sound variable
	CallPause                   // a message pause (cotab, sotsb)
	CallNextFrame               // the picture's next frame (pod)
	CallSound                   // a sound (Buck Rogers: 45571)
	CallPictureAnimation        // the picture's animation, the front's (Buck Rogers: 26627)
)

// skips makes a Skip table of the groups: operand count -> commands.
func skips(groups map[byte][]byte) (t [256]byte) {
	for n, cs := range groups {
		for _, c := range cs {
			t[c] = n
		}
	}
	return t
}

// Current reports whether a text carrying id and revision is this version's.
func (v *Version) Current(id VersionID, revision int) bool {
	return id == v.ID && revision == FormatRevision
}

// commands makes the commands with n fixed operands of the groups' lists; a
// command's canonical number is its byte value.
func commands(groups map[int][]byte) (t [256]Cmd) {
	for n, cs := range groups {
		for _, c := range cs {
			t[c] = Cmd{Op: c, Fixed: n, Known: true}
		}
	}
	return t
}

// counted sets commands whose last fixed operand counts the rest.
func counted(t *[256]Cmd, fixedAndCount int, cs ...byte) {
	for _, c := range cs {
		t[c] = Cmd{Op: c, Fixed: fixedAndCount, Count: fixedAndCount, Known: true}
	}
}

// V1 is the dialect of the first engine (Pool of Radiance's ECL*.DAX).
var V1 = func() *Version {
	t := commands(map[int][]byte{
		0:  {OpExit, 0x0D, OpReturn, OpIfEq, OpIfNe, OpIfLt, OpIfGt, OpIfLe, OpIfGe, OpClearMonsters, OpCombat, 0x31, OpNewLine, 0x3A, 0x3D},
		1:  {OpGoto, OpGosub, 0x0A, OpPicture, OpPrint, OpPrintClear, OpPartyStr, OpNewECL, OpCall, OpFindItem, 0x34, OpProgram, 0x39, 0x3C},
		2:  {OpCompare, OpRandom, OpSave, OpInputNumber, OpInputString, 0x22, 0x36},
		3:  {OpAdd, OpSub, OpDiv, OpMul, 0x0B, 0x0C, OpLoadMap, 0x28, OpGetTable, OpAnd, OpOr, OpSaveTable, OpLoadWalls, 0x3B},
		4:  {OpCompareAnd, 0x23},
		5:  {0x2E},
		6:  {0x1E, OpParlay},
		8:  {OpTreasure},
		14: {0x29},
	})
	counted(&t, 3, OpVMenu)
	counted(&t, 2, OpOnGoto, OpOnGosub, OpHMenu)
	// the pager (Felix 2026-11-11, one for every game): unused by the first
	// engine, so at Pools of Darkness' own number
	set(&t, 31, OpJournal, 1, 0)
	// a quest handed over (the quest journal; Felix 2026-11-11: never by
	// matching its printed text): unused by every Gold Box dialect
	set(&t, 66, OpQuest, 1, 0)
	// POR's cabinets' CAST KNOCK (Felix d3: the Spell command is gone): unused
	// by every Gold Box dialect
	set(&t, 67, OpCastKnock, 1, 0)
	// POR's tavern game, over the party purse in 64-bit gold (a coin field is
	// 16 bits): unused by every Gold Box dialect
	set(&t, 68, OpTavernGamble, 0, 0)
	// SkipEclCmd's counts (por-re src/eclmenu.pas:1041; Felix 2026-11-11:
	// POR steps over a command as the original did)
	skip := skips(map[byte][]byte{
		1: {1, 2, 10, 14, 17, 18, 29, 32, 45, 50, 52, 54, 56, 57, 60},
		2: {3, 8, 9, 12, 15, 16, 31, 34},
		3: {4, 5, 6, 7, 11, 33, 40, 42, 47, 48, 53, 55, 59},
		4: {20, 35}, 5: {46}, 6: {30, 44}, 8: {39}, 14: {41},
	})
	return &Version{ID: 1, Base: 0x9900, Cmds: t, ExitOverrides: v1ExitOverrides, LoadMapSecondUnused: true, Skip: skip,
		// ovr04:2FE8's routine numbers, operand - 0x7FFF in the original
		Calls: map[uint16]CallFn{0x2C90: CallRedrawView, 0x8000: CallDuelWin, 0x8001: CallDuelLose, 0xBA03: CallDuelSound,
			0xC01E: CallStepForward, 0xC018: CallWallAhead},
		Programs:  map[byte]ProgramFn{0: ProgramPartyMenu, 8: ProgramEnding, 9: ProgramCamp},
		Behaviour: Behaviour{HitPadded: true, NPCSideBit: true, SelectMemberDrop: true, NoMonsterCombat: []CombatScreen{CombatShop, CombatTemple}}}
}()

// set makes byte-code command b the canonical command op with these operands.
func set(t *[256]Cmd, b, op byte, fixed, count int) {
	t[b] = Cmd{Op: op, Fixed: fixed, Count: count, Known: true}
}

// V2 is the dialect of the second engine (Curse of the Azure Bonds' and
// Secret of the Silver Blades' ECL*.DAX): the first's, with 65 added.
var V2 = func() *Version {
	t := V1.Cmds
	set(&t, 49, OpSpriteOff, 0, 0)
	set(&t, 62, OpDump, 0, 0)
	set(&t, 63, OpFindAffect, 1, 0)
	set(&t, 64, OpDestroy, 1, 0)
	set(&t, 65, OpGainExp, 2, 0)
	set(&t, 52, 0x34, 2, 0)
	// parse_command's counts (cotab O04C:1FD2, sotsb's); 65 is sotsb's only
	// (cotab's scripts have no GainExperience).
	skip := skips(map[byte][]byte{
		1: {1, 2, 10, 14, 17, 18, 29, 32, 45, 50, 52, 54, 56, 57, 60, 63, 64},
		2: {3, 8, 9, 15, 16, 31, 34, 65},
		3: {4, 5, 6, 7, 11, 12, 33, 40, 42, 47, 48, 53, 55, 59},
		4: {20, 35}, 5: {46}, 6: {30, 44}, 8: {39}, 14: {41},
	})
	// CALL's addresses: Curse of the Azure Bonds' (O03C:2F23) and Secret of
	// the Silver Blades' (O050:33C6) differ for RedrawView and DuelSound;
	// none collides, so the one dialect carries both.
	calls := map[uint16]CallFn{11792: CallRedrawView, 11723: CallRedrawView, 32768: CallDuelWin, 32769: CallDuelLose,
		45568: CallDuelSound, 45571: CallDuelSound, 49182: CallStepForward, 49176: CallWallAhead, 26627: CallPause}
	// PROGRAM: cotab's 0, 3 quit, 8 ending, 9 camp; sotsb uses 0 and 9
	programs := map[byte]ProgramFn{0: ProgramPartyMenu, 3: ProgramQuit, 8: ProgramEnding, 9: ProgramCamp}
	return &Version{ID: 2, Base: 32768, Cmds: t, Skip: skip, Calls: calls, Programs: programs, Behaviour: func() Behaviour { b := V1.Behaviour; b.PictureAnimates = true; return b }()}
}()

// V3 is Pools of Darkness' dialect.
var V3 = func() *Version {
	t := V2.Cmds
	set(&t, 12, 0x0C, 4, 0)
	set(&t, 29, OpPressKey, 0, 0)
	set(&t, 31, OpJournal, 1, 0)
	set(&t, 33, OpLoadArea, 2, 0)
	set(&t, 34, OpYesNo, 0, 0)
	set(&t, 35, OpClearExit, 0, 0)
	set(&t, 39, OpTreasureOf, 4, 4)
	set(&t, 41, OpNoOp, 0, 0)
	set(&t, 44, OpSound, 1, 0)
	set(&t, 49, OpMenuKeys, 2, 2)
	set(&t, 55, OpFindNPC, 2, 0)
	set(&t, 58, OpDelay, 0, 0) // the original pauses 600 ms (game delay)
	set(&t, 60, OpItemSet, 1, 0)
	// Pools of Darkness' parse_command counts (O055:1464)
	skip := skips(map[byte][]byte{
		1: {1, 2, 10, 14, 17, 18, 31, 32, 44, 45, 50, 52, 54, 56, 57, 60, 63, 64},
		2: {3, 8, 9, 15, 16, 65, 33, 55},
		3: {4, 5, 6, 7, 11, 40, 42, 47, 48, 53, 59},
		4: {12, 20}, 5: {46}, 6: {30},
	})
	// pod renumbered its routines (O040:1E4A)
	return &Version{ID: 3, Base: 32768, Cmds: t, Skip: skip, Calls: map[uint16]CallFn{1: CallRedrawView, 2: CallStepForward, 3: CallNextFrame},
		Programs:  map[byte]ProgramFn{0: ProgramPartyMenu, 1: ProgramSaveQuit, 2: ProgramQuit, 9: ProgramCamp},
		Behaviour: Behaviour{HitPadded: true, NPCSideBit: true, SetupMonsterSides: true, OutdoorsNear: true, InputTextOperand: true, NoMonsterCombat: []CombatScreen{CombatVault, CombatTemple}, PictureAnimates: true, SymbolSetFrames: true}}
}()

// V4 is the Buck Rogers engine's dialect (Countdown to Doomsday, Matrix Cubed).
var V4 = func() *Version {
	var t [256]Cmd
	for b, c := range V1.Cmds {
		if b < 0x40 {
			t[b] = c
		}
	}
	for _, c := range []struct {
		b, op        byte
		fixed, count int
	}{
		{12, 0x0C, 4, 0}, {29, OpShade, 2, 0}, {30, OpPartyStat, 6, 0}, {31, OpSpaceFight, 4, 0}, {34, OpSkill, 3, 0}, {35, OpSkillShow, 3, 0},
		{39, OpTreasureOf, 2, 2}, {41, OpPressKey, 0, 0}, {44, OpYesNo, 0, 0}, {45, OpCall, 1, 0}, {46, 0x2E, 5, 0}, {49, OpMenuKeys, 2, 2},
		{52, 0x34, 2, 0}, {55, OpLoadWalls, 3, 0}, {56, OpProgram, 1, 0}, {57, OpWho, 1, 0},
		{58, 0x3A, 0, 0}, {59, OpNothing, 0, 0}, {60, OpNothing, 0, 0}, {61, 0x31, 0, 0}, {62, OpDump, 0, 0}, {63, OpLogbook, 2, 0}, {64, OpTakeItem, 2, 0},
		{65, OpGainExp, 2, 0}, {66, OpClearExit, 0, 0}, {67, OpSound, 1, 0}, {68, OpNothing, 0, 0}, {69, OpWallDist, 2, 0}, {70, OpFor, 2, 0}, {71, OpNext, 0, 0},
		{72, OpStash, 1, 0}, {73, OpSkillDamage, 6, 0}, {74, OpDuel, 0, 0}, {75, OpShop, 1, 0}, {76, OpAreaKind, 2, 0},
	} {
		set(&t, c.b, c.op, c.fixed, c.count)
	}
	// BR's skip() (buckrogers-go/rules/ecl.go, pO0B8_1BEB): the counts its
	// table gives, not the commands' Fixed; every other command steps on by 0.
	// But 59 and 60: its rows for them (3 and 1) are never reached. Both
	// decode to OpNothing, as 68 does, and BR's executor runs every OpNothing
	// as 68, stepping over none (eclbr.VM.SetVersion takes the highest byte
	// of an op). The scripts hold Nothing() (ten in BR1), so they skip 0 here too.
	skip := skips(map[byte][]byte{
		1: {1, 2, 10, 14, 17, 18, 32, 45, 50, 52, 54, 56, 57, 67, 72, 75},
		2: {3, 8, 9, 15, 16, 31, 65, 64, 29, 63, 69, 70, 76},
		3: {4, 5, 6, 7, 11, 12, 33, 40, 42, 47, 48, 53, 55, 34, 35},
		4: {20}, 5: {46}, 6: {30, 73},
	})
	// opSpecial's addresses (buckrogers-go/rules/eclops.go): 11723 redraws the
	// view, 49182 steps ahead, 49176 sets the wall ahead (V2's routines); 45571
	// is the sound and 26627 the picture's animation, which V2 maps to
	// CallDuelSound and CallPause: BR's own meaning, so named separately.
	calls := map[uint16]CallFn{11723: CallRedrawView, 49182: CallStepForward, 49176: CallWallAhead, 45571: CallSound, 26627: CallPictureAnimation}
	return &Version{ID: 4, Base: 32768, Cmds: t, Skip: skip, Calls: calls,
		// BR1 pO035_2DCC case 56 (buckrogers-go opMenuCall): 0 the team menu, 9 the rest menu (eclvm.pas PROGRAM_REST_MENU: holds the script, which
		// goes on after it: not the Pool games' camp, which ends it)
		Programs:  map[byte]ProgramFn{0: ProgramPartyMenu, 9: ProgramRest},
		Behaviour: Behaviour{SetupMonsterSides: true, ReturnEndsScript: true, StoreByOperand: true, ClearExitText: true, WhoEndsNoMember: true, FarSight: true, NoCamp: true, LostDuelResult: true, ViewKept: true, CreditsTreasure: true, MessageDelayMs: BRMessageDelayMs}}
}()

// Versions are the dialects by ID.
var Versions = map[VersionID]*Version{V1.ID: V1, V2.ID: V2, V3.ID: V3, V4.ID: V4}
