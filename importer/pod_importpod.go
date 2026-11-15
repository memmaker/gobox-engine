package importer

import (
	"encoding/json"
	"gobox/adventure"
	"gobox/rules1e"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"gobox/asset/orig"
	"gobox/audio"
	"gobox/rules"
)

// init registers Pools of Darkness's import with gobox/importer: the
// decompiled scripts with the game's fixes, the art sheets, and the maps and records into
// adventures/pod/{scripts,art,data}. The Amiga art is taken from a folder amiga
// beside the originals when there is one (data/dos, data/amiga).
func init() {
	Register(Game{
		Cover: true,
		ID:    "pod", Title: "Pools of Darkness", Release: 1991, Ruleset: "rules1e", Record: PodRecordLayout, Version: 20, // 20: rules.json namedItems (the named items' own ids, for their icons); 19: ITEMS row 71 (the halberd) the series' halberd, not the flail in the gauntlets' slot; 18: the frame sets (Border.Sets: BORDERS 0..3 by tileSet) and the Amiga window frame (Border.Amiga: FRAME.TLB sets); 17: the Amiga text font pooled as pod-amiga's (shown with the Amiga art); 16: GameSet 1 (set 2 read the missing GEO2, WALLDEF2, PIC2: the start area had no walls); 15: GameTables.ScriptFixes (the engine keys the texts as the import does); 14: GameTables.Walls (the 3D view's wall textures); 13: Window, UserSaves, Foreign in the manifest (playpool reads them); 12: 1: imported for pool (Sw-pod); 2: MON records as the rules' Sheet (Pf); 3: start money in gold (Rz); 4: GameTables.Art, items as char.Item (Pg); 5: spellsGone, no Restoration (Rz); 6: groundIds, itemKinds, moraleUncapped (Ph); 7: appraisal (Pi); 8: scripts gold only (Gc); 9: the Amiga music in sound/ (Pm); 10: AreaState, KeptAreaState, KeptAreaCells, Word02EF arrays (TestNoScalarTables); 11: debug party-variable menu and the Word00AB[159] write removed (Felix 2026-11-13); 12: SequenceStep.Credits on the credits picture (TITLE 3)
		Window: Size{W: 80 * 8, H: 50 * 8}, UserSaves: true, Foreign: []string{"sotsb"},
		Sources: []string{"*.DAX", "ITEMS", "GAME.EXE", "*.SMS", "*.ACF", "*.BMA"},
		Import:  PodImport, Amiga: &podAmiga,
		Start: Fixed(StartSequence()),
		End:   Fixed(EndSequence()),
	})
}

// Import writes the folder dst from the DOS originals in dos (ImportTo)
// and MUSIC.DAX's songs as MIDI with their OPL voices (glbr.WriteSongs).
func PodImport(dos, dst string) error {
	if err := PodImportTo(dos, dst); err != nil {
		return err
	}
	music, err := os.ReadFile(filepath.Join(dos, "MUSIC.DAX"))
	if err != nil {
		return err
	}
	if _, err := orig.WriteSongs(music, 0, filepath.Join(dst, orig.MusicDir)); err != nil {
		return err
	}
	if err := podImportRules(dos, filepath.Join(dst, "data")); err != nil {
		return err
	}
	sets, err := json.Marshal(TreasureSets())
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dst, adventure.DataDir, adventure.TreasureSetsFile), sets, 0o644); err != nil {
		return err
	}
	return writeSwitchData(dst)
}

// writeSwitchData writes what the shared engine reads of the game besides
// its tables: GameTables (gametables.json), the slot Layout, the journal book
// and the entries the scripts' Journal n name.
func writeSwitchData(dst string) error {
	data := filepath.Join(dst, adventure.DataDir)
	monsters, icons, count, status, control := AreaMonsterTable()
	area := adventure.AreaMonsters{Count: count, Status: status, Control: control}
	for i := range monsters {
		area.ByScreen = append(area.ByScreen, adventure.AreaMonster{Monster: int(monsters[i]), Icon: int(icons[i])})
	}
	sounds := map[byte]adventure.ScriptSound{}
	for n := 0; n <= math.MaxUint8; n++ {
		song, music, digitized := SoundKind(byte(n))
		switch {
		case music:
			sounds[byte(n)] = adventure.ScriptSound{Music: true, Song: song}
		case digitized:
			sounds[byte(n)] = adventure.ScriptSound{Cue: rules.SoundCue(strconv.Itoa(n))}
		}
	}
	tables := adventure.GameTables{
		Walls:       wallTextures("pod", podFloor, podSky),
		GameSet:     1,                                                    // pod has no game areas: every content file is set 1 (pod-re view3d LoadDax GEO1, WALLDEF1; eclmenu ECL1)
		FirstBlock:  16,                                                   // firstBlock
		BigPictures: adventure.BigPictures{First: 240, NotOverland: true}, // O040:07DB
		Dialect:     Dialect, AddNPCIcon: true, AreaMonsters: area, ScriptSounds: sounds,
		OriginalSave: PodOriginalSave, ScriptFixes: podScriptFixes,
		Symbols:    []adventure.SymbolLoad{{Set: 4, File: "8X8D0", Block: 202}}, // gameinit's Load8x8Set(202, 1): POD's symbols are 8X8D0's, not the content set's file
		MonsterSet: 1,                                                           // LoadCreatureData's MON1, LoadIconAnim's CPIC1
		Art: &adventure.ArtNames{ // LoadCombatTileSet's Load24x24Set (pod-re combatst.pas)
			DungeonTiles:  []adventure.TileLoad{{File: "DUNGCOM", From: adventure.TileBlockWalls, Last: 24}},
			WildTiles:     []adventure.TileLoad{{File: "WILDCOM", From: adventure.TileBlockAreaSet, Last: 38}},
			CombatSprites: "COMSPR",
			Title:         "TITLE",
		},
	}
	if err := adventure.WriteGameData(data, PodStaticData()); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(data, adventure.SoundFile), []byte(audio.SilentSoundData), 0o644); err != nil { // AdLib only: no speaker or Tandy tables
		return err
	}
	if err := WriteJSON(filepath.Join(data, adventure.GameTablesFile), tables); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(data, rules1e.LayoutFile), PodLayout1e()); err != nil {
		return err
	}
	return ImportJournal("pod", filepath.Join(dst, adventure.JournalDir)) // the entries are in the scripts' data (podScriptData)
}

// importRules writes GAME.EXE's tables (tables.json, data segment paragraph
// 2535) into dst/rules.json.
func podImportRules(dos, dst string) error {
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	im, err := orig.Load(l.Path("GAME.EXE"), 2535)
	if err != nil {
		return err
	}
	return orig.WriteRulesWith(tableSpec("pod"), dst, []orig.Extra{{Name: "namedItems", Value: rules1e.PodNamedItems}, {Name: "affectIds", Value: rules1e.PodAffectIDs}, {Name: "spellIds", Value: rules1e.PodSpellIDs}, {Name: "spellsGone", Value: rules1e.PodSpellsGone}, {Name: "itemSpellCode", Value: rules1e.SevenBitItemSpell}, {Name: "spellClasses", Value: rules1e.PodSpellClasses}, {Name: "affectVariants", Value: rules1e.PodAffectVariants}, {Name: "affectGroups", Value: rules1e.PodAffectGroups}, {Name: "giantRule", Value: rules1e.LaterGiants}, {Name: "magicResistance", Value: rules1e.PodMagicResist}, {Name: "spiritualHammer", Value: rules1e.PodHammer}, {Name: "oldAge", Value: rules1e.PodOldAge}, {Name: "casterLevel", Value: rules1e.PodCasterLevel}, {Name: "affectRule", Value: rules1e.PodAffectRule}, {Name: "itemPowers", Value: rules1e.PodItemPowers}, {Name: "creationStart", Value: PodCreationStart(filepath.Dir(dst))}, {Name: "town", Value: PodTown()}, {Name: "combatGround", Value: PodCombatGround()}, {Name: "groundIds", Value: rules1e.PodGroundIDs}, {Name: "itemKinds", Value: rules1e.SotsbItemKinds}, {Name: "moraleUncapped", Value: true}, {Name: "itemLibrary", Value: ItemLibrary(filepath.Dir(dst))}, {Name: "randomScroll", Value: RandomScroll()}, {Name: "appraisal", Value: rules1e.SotsbAppraisal}}, im) // dst is the adventure's data folder
}
