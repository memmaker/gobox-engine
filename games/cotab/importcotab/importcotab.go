// Package importcotab registers Curse of the Azure Bonds's import with gobox/importer: the
// decompiled scripts with the game's fixes, the art sheets, the maps and records, the sound
// tables and the shipped saves into adventures/cotab/{scripts,art,data,sound,saves}. The Amiga art is taken from a folder amiga
// beside the originals when there is one (data/dos, data/amiga).
package importcotab

import (
	"encoding/json"
	"os"
	"path/filepath"
	"rules1e"

	"cotab"

	"gobox/audio/sounddata"
	"gobox/core"
	"gobox/dax"
	"gobox/exe"
	"gobox/exe/tablespec"
	"gobox/importer"
	"gobox/pool"
)

func init() {
	importer.Register(importer.Game{
		ID: "cotab", Title: "Curse of the Azure Bonds", Release: 1989, Ruleset: "rules1e", Record: RecordLayout, Version: 11, // 1: imported for pool (Sw-cotab); 2: MON records as the rules' Sheet (Pf); 3: start money in gold (Rz); 4: GameTables.Art (Pg); 5: spellsGone, no Restoration (Rz); 6: groundIds (Ph); 7: randomTreasure, cotab's own (Ct); 8: scripts gold only (Gc); 9: randomTreasure from rules1e's treasure/cotab.json, one roller for every game (Ct); 10: Locals, Scratch, CheckpointGuards arrays (Tb); 11: no Sequence.Remaster (the remaster credit line is gobox's)
		Sources: []string{"*.DAX", "ITEMS", "START.EXE", "GAME.OVR"},
		Import:  Import,
		Start:   importer.Fixed(start()),
		End:     importer.Fixed(end()),
	})
}

// whole is the 40x25 page.
var whole = importer.Cells{X2: 39, Y2: 24}

// start is COTAB's title (O039:0632 title_screen, rules/title.go): the SSI
// logo (TITLE 1), the credits page (O039:01DD, credits), the
// painting (TITLE 2), the painting with the Forgotten Realms banner (TITLE 4
// at cell 0, 11) and its tune (titleSound 13); the painting with the Curse
// logo (TITLE 3 at cell 6, 11) stays under the main menu. A short intro
// (CURSE.CFG's, the host's Full intro option off) is the SSI logo only.
func start() *importer.Sequence {
	painting := importer.Layer{Art: importer.Art{File: "TITLE", Block: 2}, At: whole}
	q := &importer.Sequence{
		Steps: []importer.SequenceStep{
			{Kind: importer.StepPicture, Layers: []importer.Layer{{Art: importer.Art{File: "TITLE", Block: 1}, At: whole}}},
			{Kind: importer.StepCredits},
			{Kind: importer.StepPicture, Layers: []importer.Layer{painting}},
			{Kind: importer.StepPicture, Layers: []importer.Layer{painting, {Art: importer.Art{File: "TITLE", Block: 4}, At: importer.Cells{X1: 0, Y1: 11, X2: 39, Y2: 24}}}, Song: "13"},
		},
		Backdrop: []importer.Layer{painting, {Art: importer.Art{File: "TITLE", Block: 3}, At: importer.Cells{X1: 6, Y1: 11, X2: 35, Y2: 24}}},
		Credits:  credits,
		Short:    []int{0},
		Full:     true, // CURSE.CFG's default; the host's Full intro option starts from it
	}
	return q
}

// credits are the title's credits page (O039:01DD): the lines in the order
// DisplayString prints them, at its places and in its colours.
var credits = []importer.Credit{
	{X: 2, Y: 1, Colour: 0xa, Text: `based on the tsr novel "azure bonds"`},
	{X: 6, Y: 2, Colour: 0xa, Text: "by:"},
	{X: 9, Y: 2, Colour: 0xb, Text: "kate novak"},
	{X: 20, Y: 2, Colour: 0xa, Text: "and"},
	{X: 24, Y: 2, Colour: 0xb, Text: "jeff grubb"},
	{X: 10, Y: 4, Colour: 0xa, Text: "scenario created by:"},
	{X: 11, Y: 5, Colour: 0xe, Text: "tsr, inc."},
	{X: 21, Y: 5, Colour: 0xa, Text: "and"},
	{X: 25, Y: 5, Colour: 0xe, Text: "ssi"},
	{X: 14, Y: 6, Colour: 0xb, Text: "jeff grubb"},
	{X: 11, Y: 7, Colour: 0xb, Text: "george mac donald"},
	{X: 1, Y: 9, Colour: 0xa, Text: "game created by:"},
	{X: 18, Y: 9, Colour: 0xe, Text: "ssi special projects"},
	{X: 2, Y: 11, Colour: 0xe, Text: "project leader:"},
	{X: 2, Y: 12, Colour: 0xb, Text: "george mac donald"},
	{X: 2, Y: 14, Colour: 0xe, Text: "programming:"},
	{X: 2, Y: 15, Colour: 0xb, Text: "scot bayless"},
	{X: 2, Y: 16, Colour: 0xb, Text: "russ brown"},
	{X: 2, Y: 17, Colour: 0xb, Text: "michael mancuso"},
	{X: 2, Y: 19, Colour: 0xe, Text: "development:"},
	{X: 2, Y: 20, Colour: 0xb, Text: "david shelley"},
	{X: 2, Y: 21, Colour: 0xb, Text: "michael mancuso"},
	{X: 2, Y: 22, Colour: 0xb, Text: "oran kangas"},
	{X: 22, Y: 11, Colour: 0xe, Text: "graphic arts:"},
	{X: 22, Y: 12, Colour: 0xb, Text: "tom wahl"},
	{X: 22, Y: 13, Colour: 0xb, Text: "fred butts"},
	{X: 22, Y: 14, Colour: 0xb, Text: "susan manley"},
	{X: 22, Y: 15, Colour: 0xb, Text: "mark johnson"},
	{X: 22, Y: 16, Colour: 0xb, Text: "cyrus lum"},
	{X: 22, Y: 18, Colour: 0xe, Text: "playtesting:"},
	{X: 22, Y: 19, Colour: 0xb, Text: "jim jennings"},
	{X: 22, Y: 20, Colour: 0xb, Text: "james kucera"},
	{X: 22, Y: 21, Colour: 0xb, Text: "rick white"},
	{X: 22, Y: 22, Colour: 0xb, Text: "robert daly"},
}

// Import writes the folder dst from the DOS originals in dos: the engine's
// part (ImportTo), then START.EXE's sound tables.
func Import(dos, dst string) error {
	if err := ImportTo(dos, dst); err != nil {
		return err
	}
	if err := importSound(dos, filepath.Join(dst, "data")); err != nil {
		return err
	}
	if err := importRules(dos, filepath.Join(dst, "data")); err != nil {
		return err
	}
	if err := importJournal(filepath.Join(dst, "data", pool.JournalDir)); err != nil {
		return err
	}
	data := filepath.Join(dst, "data")
	if err := pool.WriteGameData(data, StaticData()); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(data, pool.GameTablesFile), gameTables); err != nil {
		return err
	}
	return writeJSON(filepath.Join(data, rules1e.LayoutFile), Layout1e)
}

// importJournal writes COTAB's journal books and pictures (JournalFS, gobox's journal format) into dir.
// The entries file (names/entries.json) comes with the scripts.
func importJournal(dir string) error { return os.CopyFS(dir, JournalFS()) }

// gameTables is COTAB's pool.GameTables: the second engine's dialect (its NoMonsterCombat is the
// dialect's: shop, then temple), no NPC combat icon, the duel sounds' numbers (CALL 0xB200's
// DS:2612 table: 8 a step, 10 the fireball; any other plays a step). The portrait tables stay
// empty: COTAB's portraits come from the shared collection.
var gameTables = pool.GameTables{
	Dialect: core.DialectCOTAB,
	ScriptSounds: map[byte]pool.ScriptSound{
		8:  {Cue: core.CueStep},
		10: {Cue: core.CueFireball},
	},
	AddNPCIcon:   false,
	DefaultSound: &pool.ScriptSound{Cue: core.CueStep},
	ScriptFixes:  scriptFixes,
	OriginalSave: OriginalSave,
	Art: &pool.ArtNames{ // copyprot's and LoadCombatTileSet's Load24x24Set (cotab-re combatst.pas)
		Tiles:         []pool.TileLoad{{File: "TILES", Block: 1, Last: 25}, {File: "TILES", Block: 2, First: 26, Last: 21}},
		DungeonTiles:  []pool.TileLoad{{File: "DUNGCOM", Block: 1, Last: 24}, {File: "RANDCOM", Block: 1, First: 34, Last: 5}},
		WildTiles:     []pool.TileLoad{{File: "WILDCOM", Block: 1, Last: 32}, {File: "RANDCOM", Block: 1, First: 34, Last: 5}},
		CombatSprites: "COMSPR",
		Title:         "TITLE",
	},
}

// writeJSON writes v as indented JSON to path.
func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// importRules writes START.EXE's tables (tables.json, data segment paragraph
// 2750) into dst/json.
func importRules(dos, dst string) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	im, err := exe.Load(l.Path("START.EXE"), 2750)
	if err != nil {
		return err
	}
	return tablespec.WriteRulesWith(cotab.TablesJSON, dst, []tablespec.Extra{{Name: "affectIds", Value: rules1e.CotabAffectIDs}, {Name: "spellIds", Value: rules1e.CotabSpellIDs}, {Name: "spellsGone", Value: rules1e.CotabSpellsGone}, {Name: "spellClasses", Value: rules1e.CotabSpellClasses}, {Name: "affectVariants", Value: rules1e.CotabAffectVariants}, {Name: "affectGroups", Value: rules1e.CotabAffectGroups}, {Name: "giantRule", Value: rules1e.LaterGiants}, {Name: "casterLevel", Value: rules1e.CotabCasterLevel}, {Name: "affectRule", Value: rules1e.CotabAffectRule}, {Name: "itemPowers", Value: rules1e.CotabItemPowers}, {Name: "creationStart", Value: CreationStart()}, {Name: "importFlag", Value: true}, {Name: "town", Value: Town()}, {Name: "combatGround", Value: CombatGround()}, {Name: "groundIds", Value: rules1e.CotabGroundIDs}, {Name: "randomTreasure", Value: rules1e.CotabTreasure()}}, im)
}

// importSound decodes START.EXE's sound tables into dst/SoundFile.
func importSound(dos, dst string) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	exe, err := os.ReadFile(l.Path("START.EXE"))
	if err != nil {
		return err
	}
	snd, err := sounddata.Decode(exe, soundLayout)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, SoundFile), snd, 0o644)
}

// soundLayout is where START.EXE's driver keeps its tables (gobox audio API.md).
var soundLayout = sounddata.Layout{Segment: 0x699, JumpTable: 356, SpeakerSet: 368, TandySet: 778,
	SpeakerSongs: 1186, TandySongs: 1354, Songs: 21, Enabled: 336,
	Commands: map[int]string{5521: "reset", 5615: "return", 5622: "call", 5512: "switch", 5489: "loop", 5570: "set"}}
