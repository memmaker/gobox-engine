package importer

import (
	"gobox/adventure"
	"gobox/game"
	"gobox/rules1e"
	"os"
	"path/filepath"

	"gobox/asset"
	"gobox/asset/orig"
	"gobox/audio"
	"gobox/rules"
)

// init registers Curse of the Azure Bonds's import with gobox/importer: the
// decompiled scripts with the game's fixes, the art sheets, the maps and records, the sound
// tables and the shipped saves into adventures/cotab/{scripts,art,data,sound,saves}. The Amiga art is taken from a folder amiga
// beside the originals when there is one (data/dos, data/amiga).
func init() {
	Register(Game{
		Cover: true,
		ID:    "cotab", Title: "Curse of the Azure Bonds", Release: 1989, Ruleset: "rules1e", Record: CotabRecordLayout, Version: 19, // 19: the Amiga text font pooled as cotab-amiga's (shown with the Amiga art); 18: Amiga delta frames of another size laid over the first (the opening hand steady); 17: the journal in the adventure's journal/ (it went to data/journal: the engine found none) and GameTables.BigPictures; 16: GameTables.GameSet, FirstBlock (the start); 15: GameTables.Walls (the 3D view's wall textures); 14: credits' Colour a game.Ink (C1); 13: data/look.json (ReadyMark "Yes "); 12: Window, UserSaves, Foreign in the manifest (playpool reads them); 11: 1: imported for pool (Sw-cotab); 2: MON records as the rules' Sheet (Pf); 3: start money in gold (Rz); 4: GameTables.Art (Pg); 5: spellsGone, no Restoration (Rz); 6: groundIds (Ph); 7: randomTreasure, cotab's own (Ct); 8: scripts gold only (Gc); 9: randomTreasure from rules1e's treasure/cotab.json, one roller for every game (Ct); 10: Locals, Scratch, CheckpointGuards arrays (Tb); 11: no Sequence.Remaster (the remaster credit line is gobox's)
		Window: Size{W: asset.ScreenW, H: asset.ScreenH}, UserSaves: true, Foreign: []string{"por"},
		Sources: []string{"*.DAX", "ITEMS", "START.EXE", "GAME.OVR"},
		Import:  CotabImport, Amiga: &cotabAmiga,
		Start: Fixed(cotabStart()),
		End:   Fixed(cotabEnd()),
	})
}

// whole is the 40x25 page.
var cotabWhole = adventure.Cells{X2: 39, Y2: 24}

// start is COTAB's title (O039:0632 title_screen, rules/title.go): the SSI
// logo (TITLE 1), the credits page (O039:01DD, credits), the
// painting (TITLE 2), the painting with the Forgotten Realms banner (TITLE 4
// at cell 0, 11) and its tune (titleSound 13); the painting with the Curse
// logo (TITLE 3 at cell 6, 11) stays under the main menu. A short intro
// (CURSE.CFG's, the host's Full intro option off) is the SSI logo only.
func cotabStart() *adventure.Sequence {
	painting := adventure.Layer{Art: adventure.Art{File: "TITLE", Block: 2}, At: cotabWhole}
	q := &adventure.Sequence{
		Steps: []adventure.SequenceStep{
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{{Art: adventure.Art{File: "TITLE", Block: 1}, At: cotabWhole}}},
			{Kind: adventure.StepCredits},
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{painting}},
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{painting, {Art: adventure.Art{File: "TITLE", Block: 4}, At: adventure.Cells{X1: 0, Y1: 11, X2: 39, Y2: 24}}}, Song: "13"},
		},
		Backdrop: []adventure.Layer{painting, {Art: adventure.Art{File: "TITLE", Block: 3}, At: adventure.Cells{X1: 6, Y1: 11, X2: 35, Y2: 24}}},
		Credits:  cotabCredits,
		Short:    []int{0},
		Full:     true, // CURSE.CFG's default; the host's Full intro option starts from it
	}
	return q
}

// credits are the title's credits page (O039:01DD): the lines in the order
// DisplayString prints them, at its places and in its colours.
var cotabCredits = []adventure.Credit{
	{X: 2, Y: 1, Colour: game.InkLightGreen, Text: `based on the tsr novel "azure bonds"`},
	{X: 6, Y: 2, Colour: game.InkLightGreen, Text: "by:"},
	{X: 9, Y: 2, Colour: game.InkLightCyan, Text: "kate novak"},
	{X: 20, Y: 2, Colour: game.InkLightGreen, Text: "and"},
	{X: 24, Y: 2, Colour: game.InkLightCyan, Text: "jeff grubb"},
	{X: 10, Y: 4, Colour: game.InkLightGreen, Text: "scenario created by:"},
	{X: 11, Y: 5, Colour: game.InkYellow, Text: "tsr, inc."},
	{X: 21, Y: 5, Colour: game.InkLightGreen, Text: "and"},
	{X: 25, Y: 5, Colour: game.InkYellow, Text: "ssi"},
	{X: 14, Y: 6, Colour: game.InkLightCyan, Text: "jeff grubb"},
	{X: 11, Y: 7, Colour: game.InkLightCyan, Text: "george mac donald"},
	{X: 1, Y: 9, Colour: game.InkLightGreen, Text: "game created by:"},
	{X: 18, Y: 9, Colour: game.InkYellow, Text: "ssi special projects"},
	{X: 2, Y: 11, Colour: game.InkYellow, Text: "project leader:"},
	{X: 2, Y: 12, Colour: game.InkLightCyan, Text: "george mac donald"},
	{X: 2, Y: 14, Colour: game.InkYellow, Text: "programming:"},
	{X: 2, Y: 15, Colour: game.InkLightCyan, Text: "scot bayless"},
	{X: 2, Y: 16, Colour: game.InkLightCyan, Text: "russ brown"},
	{X: 2, Y: 17, Colour: game.InkLightCyan, Text: "michael mancuso"},
	{X: 2, Y: 19, Colour: game.InkYellow, Text: "development:"},
	{X: 2, Y: 20, Colour: game.InkLightCyan, Text: "david shelley"},
	{X: 2, Y: 21, Colour: game.InkLightCyan, Text: "michael mancuso"},
	{X: 2, Y: 22, Colour: game.InkLightCyan, Text: "oran kangas"},
	{X: 22, Y: 11, Colour: game.InkYellow, Text: "graphic arts:"},
	{X: 22, Y: 12, Colour: game.InkLightCyan, Text: "tom wahl"},
	{X: 22, Y: 13, Colour: game.InkLightCyan, Text: "fred butts"},
	{X: 22, Y: 14, Colour: game.InkLightCyan, Text: "susan manley"},
	{X: 22, Y: 15, Colour: game.InkLightCyan, Text: "mark johnson"},
	{X: 22, Y: 16, Colour: game.InkLightCyan, Text: "cyrus lum"},
	{X: 22, Y: 18, Colour: game.InkYellow, Text: "playtesting:"},
	{X: 22, Y: 19, Colour: game.InkLightCyan, Text: "jim jennings"},
	{X: 22, Y: 20, Colour: game.InkLightCyan, Text: "james kucera"},
	{X: 22, Y: 21, Colour: game.InkLightCyan, Text: "rick white"},
	{X: 22, Y: 22, Colour: game.InkLightCyan, Text: "robert daly"},
}

// Import writes the folder dst from the DOS originals in dos: the engine's
// part (ImportTo), then START.EXE's sound tables.
func CotabImport(dos, dst string) error {
	if err := CotabImportTo(dos, dst); err != nil {
		return err
	}
	if err := cotabImportSound(dos, filepath.Join(dst, "data")); err != nil {
		return err
	}
	if err := cotabImportRules(dos, filepath.Join(dst, "data")); err != nil {
		return err
	}
	if err := ImportJournal("cotab", filepath.Join(dst, adventure.JournalDir)); err != nil {
		return err
	}
	data := filepath.Join(dst, "data")
	if err := adventure.WriteGameData(data, CotabStaticData()); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(data, adventure.GameTablesFile), cotabGameTables); err != nil {
		return err
	}
	if err := writeLook(data, cotabLook); err != nil {
		return err
	}
	return WriteJSON(filepath.Join(data, rules1e.LayoutFile), CotabLayout1e)
}

// gameTables is COTAB's pool.GameTables: the second engine's dialect (its NoMonsterCombat is the
// dialect's: shop, then temple), no NPC combat icon, the duel sounds' numbers (CALL 0xB200's
// DS:2612 table: 8 a step, 10 the fireball; any other plays a step). The portrait tables stay
// empty: COTAB's portraits come from the shared collection.
var cotabGameTables = adventure.GameTables{
	Walls:       wallTextures("cotab", cotabFloor, cotabSky),
	GameSet:     2,                                 // START.PAS: gameArea := 2
	FirstBlock:  1,                                 // the adventure's first script, no last ECL saved
	BigPictures: adventure.BigPictures{First: 120}, // O03C:0873
	Dialect:     rules.DialectCOTAB,
	ScriptSounds: map[byte]adventure.ScriptSound{
		8:  {Cue: rules.CueStep},
		10: {Cue: rules.CueFireball},
	},
	AddNPCIcon:   false,
	DefaultSound: &adventure.ScriptSound{Cue: rules.CueStep},
	ScriptFixes:  cotabScriptFixes,
	OriginalSave: CotabOriginalSave,
	Art: &adventure.ArtNames{ // copyprot's and LoadCombatTileSet's Load24x24Set (cotab-re combatst.pas)
		Tiles:         []adventure.TileLoad{{File: "TILES", Block: 1, Last: 25}, {File: "TILES", Block: 2, First: 26, Last: 21}},
		DungeonTiles:  []adventure.TileLoad{{File: "DUNGCOM", Block: 1, Last: 24}, {File: "RANDCOM", Block: 1, First: 34, Last: 5}},
		WildTiles:     []adventure.TileLoad{{File: "WILDCOM", Block: 1, Last: 32}, {File: "RANDCOM", Block: 1, First: 34, Last: 5}},
		CombatSprites: "COMSPR",
		Title:         "TITLE",
	},
}

// importRules writes START.EXE's tables (tables.json, data segment paragraph
// 2750) into dst/json.
func cotabImportRules(dos, dst string) error {
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	im, err := orig.Load(l.Path("START.EXE"), 2750)
	if err != nil {
		return err
	}
	return orig.WriteRulesWith(tableSpec("cotab"), dst, []orig.Extra{{Name: "affectIds", Value: rules1e.CotabAffectIDs}, {Name: "spellIds", Value: rules1e.CotabSpellIDs}, {Name: "spellsGone", Value: rules1e.CotabSpellsGone}, {Name: "spellClasses", Value: rules1e.CotabSpellClasses}, {Name: "affectVariants", Value: rules1e.CotabAffectVariants}, {Name: "affectGroups", Value: rules1e.CotabAffectGroups}, {Name: "giantRule", Value: rules1e.LaterGiants}, {Name: "casterLevel", Value: rules1e.CotabCasterLevel}, {Name: "affectRule", Value: rules1e.CotabAffectRule}, {Name: "itemPowers", Value: rules1e.CotabItemPowers}, {Name: "creationStart", Value: CotabCreationStart()}, {Name: "importFlag", Value: true}, {Name: "town", Value: CotabTown()}, {Name: "combatGround", Value: CotabCombatGround()}, {Name: "groundIds", Value: rules1e.CotabGroundIDs}, {Name: "randomTreasure", Value: rules1e.CotabTreasure()}}, im)
}

// importSound decodes START.EXE's sound tables into dst/SoundFile.
func cotabImportSound(dos, dst string) error {
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	exe, err := os.ReadFile(l.Path("START.EXE"))
	if err != nil {
		return err
	}
	snd, err := audio.DecodeSoundData(exe, cotabSoundLayout)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, CotabSoundFile), snd, 0o644)
}

// soundLayout is where START.EXE's driver keeps its tables (gobox audio API.md).
var cotabSoundLayout = audio.SoundDataLayout{Segment: 0x699, JumpTable: 356, SpeakerSet: 368, TandySet: 778,
	SpeakerSongs: 1186, TandySongs: 1354, Songs: 21, Enabled: 336,
	Commands: map[int]string{5521: "reset", 5615: "return", 5622: "call", 5512: "switch", 5489: "loop", 5570: "set"}}
