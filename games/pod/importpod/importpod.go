// Package importpod registers Pools of Darkness's import with gobox/importer: the
// decompiled scripts with the game's fixes, the art sheets, and the maps and records into
// adventures/pod/{scripts,art,data}. The Amiga art is taken from a folder amiga
// beside the originals when there is one (data/dos, data/amiga).
package importpod

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"rules1e"
	"strconv"

	"gobox/audio/glbr"
	"gobox/audio/sounddata"
	"gobox/core"
	"gobox/dax"
	"gobox/ecl/script"
	"gobox/exe"
	"gobox/exe/tablespec"
	"gobox/importer"
	"gobox/pool"
	"pod"
)

func init() {
	importer.Register(importer.Game{
		ID: "pod", Title: "Pools of Darkness", Release: 1991, Ruleset: "rules1e", Record: RecordLayout, Version: 11, // 1: imported for pool (Sw-pod); 2: MON records as the rules' Sheet (Pf); 3: start money in gold (Rz); 4: GameTables.Art, items as char.Item (Pg); 5: spellsGone, no Restoration (Rz); 6: groundIds, itemKinds, moraleUncapped (Ph); 7: appraisal (Pi); 8: scripts gold only (Gc); 9: the Amiga music in sound/ (Pm); 10: AreaState, KeptAreaState, KeptAreaCells, Word02EF arrays (TestNoScalarTables); 11: debug party-variable menu and the Word00AB[159] write removed (Felix 2026-11-13)
		Sources: []string{"*.DAX", "ITEMS", "GAME.EXE", "*.SMS", "*.ACF", "*.BMA"},
		Import:  Import,
		Start:   importer.Fixed(StartSequence()),
		End:     importer.Fixed(EndSequence()),
	})
}

// Import writes the folder dst from the DOS originals in dos (ImportTo)
// and MUSIC.DAX's songs as MIDI with their OPL voices (glbr.WriteSongs).
func Import(dos, dst string) error {
	if err := ImportTo(dos, dst); err != nil {
		return err
	}
	music, err := os.ReadFile(filepath.Join(dos, "MUSIC.DAX"))
	if err != nil {
		return err
	}
	if _, err := glbr.WriteSongs(music, 0, filepath.Join(dst, glbr.MusicDir)); err != nil {
		return err
	}
	if err := importRules(dos, filepath.Join(dst, "data")); err != nil {
		return err
	}
	sets, err := json.Marshal(TreasureSets())
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dst, pool.DataDir, pool.TreasureSetsFile), sets, 0o644); err != nil {
		return err
	}
	return writeSwitchData(dst)
}

// writeSwitchData writes what the shared engine reads of the game besides
// its tables: GameTables (gametables.json), the slot Layout, the journal book
// and the entries the scripts' Journal n name.
func writeSwitchData(dst string) error {
	data := filepath.Join(dst, pool.DataDir)
	monsters, icons, count, status, control := AreaMonsterTable()
	area := pool.AreaMonsters{Count: count, Status: status, Control: control}
	for i := range monsters {
		area.ByScreen = append(area.ByScreen, pool.AreaMonster{Monster: int(monsters[i]), Icon: int(icons[i])})
	}
	sounds := map[byte]pool.ScriptSound{}
	for n := 0; n <= math.MaxUint8; n++ {
		song, music, digitized := SoundKind(byte(n))
		switch {
		case music:
			sounds[byte(n)] = pool.ScriptSound{Music: true, Song: song}
		case digitized:
			sounds[byte(n)] = pool.ScriptSound{Cue: core.SoundCue(strconv.Itoa(n))}
		}
	}
	tables := pool.GameTables{
		Dialect: Dialect, AddNPCIcon: true, AreaMonsters: area, ScriptSounds: sounds,
		OriginalSave: OriginalSave,
		Symbols:      []pool.SymbolLoad{{Set: 4, File: "8X8D0", Block: 202}}, // gameinit's Load8x8Set(202, 1): POD's symbols are 8X8D0's, not the content set's file
		MonsterSet:   1,                                                      // LoadCreatureData's MON1, LoadIconAnim's CPIC1
		Art: &pool.ArtNames{ // LoadCombatTileSet's Load24x24Set (pod-re combatst.pas)
			DungeonTiles:  []pool.TileLoad{{File: "DUNGCOM", From: pool.TileBlockWalls, Last: 24}},
			WildTiles:     []pool.TileLoad{{File: "WILDCOM", From: pool.TileBlockAreaSet, Last: 38}},
			CombatSprites: "COMSPR",
			Title:         "TITLE",
		},
	}
	if err := pool.WriteGameData(data, StaticData()); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(data, pool.SoundFile), []byte(sounddata.Silent), 0o644); err != nil { // AdLib only: no speaker or Tandy tables
		return err
	}
	if err := writeJSON(filepath.Join(data, pool.GameTablesFile), tables); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(data, rules1e.LayoutFile), Layout1e()); err != nil {
		return err
	}
	if err := WriteJournal(filepath.Join(dst, pool.JournalDir)); err != nil {
		return err
	}
	return writeJSON(filepath.Join(dst, "scripts", script.EntriesFile), JournalEntries())
}

// writeJSON writes v as indented JSON to path, making its folder.
func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// importRules writes GAME.EXE's tables (tables.json, data segment paragraph
// 2535) into dst/rules.json.
func importRules(dos, dst string) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	im, err := exe.Load(l.Path("GAME.EXE"), 2535)
	if err != nil {
		return err
	}
	return tablespec.WriteRulesWith(pod.TablesJSON, dst, []tablespec.Extra{{Name: "affectIds", Value: rules1e.PodAffectIDs}, {Name: "spellIds", Value: rules1e.PodSpellIDs}, {Name: "spellsGone", Value: rules1e.PodSpellsGone}, {Name: "itemSpellCode", Value: rules1e.SevenBitItemSpell}, {Name: "spellClasses", Value: rules1e.PodSpellClasses}, {Name: "affectVariants", Value: rules1e.PodAffectVariants}, {Name: "affectGroups", Value: rules1e.PodAffectGroups}, {Name: "giantRule", Value: rules1e.LaterGiants}, {Name: "magicResistance", Value: rules1e.PodMagicResist}, {Name: "spiritualHammer", Value: rules1e.PodHammer}, {Name: "oldAge", Value: rules1e.PodOldAge}, {Name: "casterLevel", Value: rules1e.PodCasterLevel}, {Name: "affectRule", Value: rules1e.PodAffectRule}, {Name: "itemPowers", Value: rules1e.PodItemPowers}, {Name: "creationStart", Value: CreationStart(filepath.Dir(dst))}, {Name: "town", Value: Town()}, {Name: "combatGround", Value: CombatGround()}, {Name: "groundIds", Value: rules1e.PodGroundIDs}, {Name: "itemKinds", Value: rules1e.SotsbItemKinds}, {Name: "moraleUncapped", Value: true}, {Name: "itemLibrary", Value: ItemLibrary(filepath.Dir(dst))}, {Name: "randomScroll", Value: RandomScroll()}, {Name: "appraisal", Value: rules1e.SotsbAppraisal}}, im) // dst is the adventure's data folder
}
