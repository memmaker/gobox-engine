package importer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"gobox/adventure"
	"gobox/asset"
	"gobox/asset/orig"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"gobox/rules/xxvc"
	"gobox/rules1e"
)

// BRSources are the originals both games' imports read.
var BRSources = []string{"*.DAX", "ITEMS", "START.EXE", "GAME.OVR", "*" + saveExt, "*" + itemsExt, "*" + effectsExt, "SAVGAM*.DAT"}

// BREXE is where a game's START.EXE keeps what the import reads from it.
type BREXE struct {
	DS        int // the data segment's paragraph in the unpacked image (gobox/exe)
	ItemWords int // GDConsts.itemNameParts[1]: String[20] rows
	NumWords  int // its rows
	// AmmoPrices is GDConsts.ammoPrices[5]: array[5..18] of Word.
	AmmoPrices int
	// ShopLists is GDConsts.itemListStarts[1] (16 Bytes), itemListCounts (16 Bytes) right after it,
	// itemListEntries (129 Bytes) after those: the lists ECL's Shop stocks a shop from.
	ShopLists int
	// ShipWeapons is GDConsts.shipWeapons[1] (TShipWeaponDef rows),
	// WeaponShort weaponShortNames[1] (String[8]), ShipTypes shipTypeNames[1]
	// (String[17]).
	ShipWeapons, WeaponShort, ShipTypes int
	// Explosion is ShipCbt DamageCrew's hurt when the party's hull goes.
	Explosion xxvc.Explosion
	// Space is GDConsts.orbitPositions[0] (58 rows of 3 Bytes), the orbit tables right after it, and MapSteps GDConsts.mapStepX (8
	// ShortInts) with mapStepY after: the rocket's system map; 0: none yet (BR2).
	Space, MapSteps int
	// Creation is the game's character-making numbers that are code in the Pascal (chargen.pas, skills.pas); its tables
	// are read at CareerTables.
	Creation xxvc.Creation
	// CareerTables is where GDConsts keeps experienceNeeded and careerBaseCombat (a row of ExperienceLevels longwords,
	// of BaseCombatLevels bytes, per career) and careerInitialSkills (12 (skill, value) byte pairs per career; 0: none).
	CareerTables BRCareerTables
}

// BRCareerTables are the offsets and row lengths of the career tables in the data segment.
type BRCareerTables struct {
	Experience, ExperienceLevels int // experienceNeeded[career, 2..]: LongInts, levels 2..2+ExperienceLevels-1
	BaseCombat, BaseCombatLevels int // careerBaseCombat[career, 0..]: Bytes
	InitialSkills                int // careerInitialSkills[career, 1..12]: skill, value; 0 when the game has none
}

// Ammunition item types (gdconsts.pas ammoPrices' bounds).
const (
	firstAmmoType = 5
	lastAmmoType  = 18
)

// itemWordSize is an itemNameParts row (String[20]).
const itemWordSize = 21

// registerBR registers adventure id (br1, br2) with title, its START.EXE and
// its title sequence start (Game.Start).
func registerBR(id, title string, e BREXE, start func(dos string) (*adventure.Sequence, error)) {
	Register(Game{ID: id, Title: title, Ruleset: "rulesxxvc", Sources: BRSources, Version: 25 /* 25: the Quick Start team's Gizmo needs "tch" (the ability key; was "tech", dropped); 24: the layouts name the ship fight's cells PostCombatMode enc=BrPostCombat (was FightState) and ShipOutcome enc=BrShipEnd (was SpaceOwnShip); 23: one import: the old engines' own files are no longer written (data/MON0, ITEM0, ITEMS, START, OVR, 8X8D1, SHIPS blocks; the partial script layout); 22: the layouts name the display options ViewFlags (was SaveFlags) and the GEO map Geo (was GeoBlock), the engine's world words; the shipped SAVGAMA.DAT becomes the slot "GAME A"; 21: party menu P/R as the Pascal (gamedata); 20: training: PartyExp layout name, Creation level-up numbers and MaxLevel; party menu P/R as the Pascal; 19: the layout names Party[725] LeftMap and Party[786] SymbolSet (the engine words, were OffMap and Disk, unbound); 18: BR1's places carry the map's area title (adventure.Place.Name); 17: BR2 layout names the rocket and ship cells (BodySquares, RocketX, RocketY, Ship*, Foe*) as BR1; 16: xxvc.Tables.Creation (start level, skill points, experience, base combat, BR2's initial skills); 15: the layout names the sky cells Sky, SkyInside; 14: BR2 gets walls, backs, places, game data, overland size and system map (BREXE.Space, MapSteps); 13: BR2 gets its Log Book (importer/br2/journal, from buckrogers2-go data/journal); 12: the layout names the step's previous square PrevX, PrevY; 11: the layout names the rocket's cells (BodySquares, RocketX, RocketY) and the import writes the system map (adventure.SpaceMap); 10: the layout names the party's ship cells (Ship*) and the last foe's (Foe*); 9: the characters' combat icon (the record's CHARS body); 8: the characters' age (the record's age byte); 7: the layout names the overland square MapX, MapY; 6: the layout comment names StockShop.Terms; 5: the shop's item lists (xxvc.Tables.ShopLists); 4: the script commands FindItem (was HasItem), AddNPC (was NPCJoin), SkillShow, TakeItem, Stash, SkillDamage; 3: a monster's isCompanion, npc byte and READIED items in the tables (xxvc.Monster); 2: the script commands Program (was MenuCall) and Who (was PickMember); 1: the credits' Colour a game.Ink (gobox C1) */, Start: start, Import: func(dos, dst string) error { return ImportBR(id, e, dos, dst) }})
}

// ImportBR writes adventure id's folder dst from the DOS originals in dos: the script blocks under the game's full layout, the
// GEO maps and rule tables, the roster, the pictures table, the art, the Log Book, the front's data (gamedata.json), the silent
// sound table, gametables.json and the originals' saved games as slots. A game without a full layout or a start entry in brStarts is an error.
//
// Not written yet: BR's sound tables (sound.json is the silent one, brSilentSound), portraits.json, the cover and docs (no
// launcher entry reads them from a BR folder).
func ImportBR(id string, e BREXE, dos, dst string) error {
	full, err := fs.ReadFile(embeddedScripts(id), ecl.GlobalLayout)
	if err != nil {
		return fmt.Errorf("%s has no full script layout yet: %w", id, err)
	}
	start, ok := brStarts[id]
	if !ok {
		return fmt.Errorf("%s has no start-up values (brStarts)", id)
	}
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	scripts := filepath.Join(dst, "scripts")
	g := filepath.Join(scripts, ecl.GlobalLayout)
	if err := os.MkdirAll(filepath.Dir(g), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(g, full, 0o644); err != nil {
		return err
	}
	if _, err := ecl.Import(ecl.V4, orig.DAXSource(l), os.DirFS(scripts), scripts, ecl.FoldGates, nil); err != nil {
		return err
	}
	data := filepath.Join(dst, adventure.DataDir)
	if err := ImportData(l, e, data); err != nil {
		return err
	}
	if err := writeRoster(dst); err != nil {
		return err
	}
	if err := asset.WritePictures(map[string]string{"br1": "buckrogers", "br2": "buckrogers2"}[id], dst); err != nil {
		return err
	}
	if err := orig.ImportArt(dos, filepath.Join(filepath.Dir(dos), "amiga"), filepath.Join(dst, adventure.ArtDir)); err != nil {
		return err
	}
	if start.Logbook != nil { // BR2 has no journal yet
		if err := ImportJournal(id, filepath.Join(dst, adventure.JournalDir)); err != nil {
			return err
		}
	}
	front := game.Data{Series: brSeries, ID: id}
	if start.Data != nil {
		front = start.Data(id)
	}
	if err := adventure.WriteGameData(data, front); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(data, adventure.SoundFile), []byte(brSilentSound), 0o644); err != nil {
		return err
	}
	im, err := orig.Load(l.Path("START.EXE"), e.DS)
	if err != nil {
		return err
	}
	tables := adventure.GameTables{
		Dialect:       rules.DialectBR,
		AddNPCIcon:    true, // saveload.pas AddCreatureToParty loads the NPC's CPIC icon (both games)
		FirstBlock:    start.FirstBlock,
		GameSet:       start.GameSet,
		Logbook:       start.Logbook,
		PictureSets:   start.PictureSets,
		BigPictures:   adventure.BigPictures{First: start.FirstBigPicture},
		OriginalSave:  rules1e.SaveLayout{CodeBase: start.CodeBase, CodeSize: start.CodeSize}, // eclLayout reads only these two
		Symbols:       start.Symbols,
		Art:           &start.Art,
		Places:        start.Places,
		ScriptSounds:  brScriptSounds(),
		OverlandSteps: start.Overland,
		Space:         SpaceTables(im.Mem[im.DS:], e),
	}
	if start.Walls {
		// the 3D view's textures: buckrogers-go rules/wallindex.txt as walls/<id>.txt (chunk (value-1)/5+1, wall (value-1)%5+1); the
		// flats and the BACK picture are not published yet (a view's floor and sky: backFlats by BACK block, not by area)
		tables.Walls = wallTextures(id, "", "")
	}
	if err := WriteJSON(filepath.Join(data, adventure.GameTablesFile), tables); err != nil {
		return err
	}
	return importBRGames(dos, dst, im.Mem[im.DS:], e) // the originals' saved games become slots
}

// BRMonster is a BR monster's records: MON0CHA, SPC (9-byte effects), ITM (items).
type BRMonster = adventure.Monster[BRRecord, [effectSize]byte, BRItem]

const (
	effectSize   = 9  // a MON0SPC effect
	templateSize = 16 // an ITEM0 block 1 template (TItemTemplate)
)

// ImportData writes the GEO maps, the rule tables (ITEMS, ITEM0's templates, START.EXE's tables, MON0's monsters typed) and the
// originals' saved characters into dst.
func ImportData(l *orig.Loader, e BREXE, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, f := range []string{"GEO1", "GEO2", "GEO3", "GEO4", "GEO5", "GEO6"} {
		if l.Exists(f + ".DAX") {
			if err := adventure.WriteBlocks(dst, f, f+".bin: GEO blocks, one after another (importbr)", orig.Blocks(l, f)); err != nil {
				return err
			}
		}
	}
	b, err := os.ReadFile(l.Path("ITEMS"))
	if err != nil {
		return err
	}
	t := Tables(b, l.BlockShared("ITEM0", templateBlock))
	im, err := orig.Load(l.Path("START.EXE"), e.DS)
	if err != nil {
		return err
	}
	t.ItemWords = ItemWords(im.Mem[im.DS:], e)
	t.AmmoPrices = AmmoPrices(im.Mem[im.DS:], e)
	t.ShopLists = ShopLists(im.Mem[im.DS:], e)
	t.ShipWeapons = ShipWeapons(im.Mem[im.DS:], e)
	t.Explosion = e.Explosion
	t.Creation = CareerCreation(im.Mem[im.DS:], e)
	if t.Ships, err = Ships(orig.Blocks(l, "SHIPS"), im.Mem[im.DS:], e); err != nil {
		return err
	}
	t.Monsters = map[rules.MonsterID]xxvc.Monster{}
	mons := orig.Blocks(l, "MON0CHA")
	for _, id := range slices.Sorted(maps.Keys(mons)) { // in MON0 id order, never the map's: addTemplates numbers the templates as they come (bug class: an import that differed from run to run)
		rec := mons[id]
		r := make([]byte, brCharacterSize)
		copy(r, rec) // a short block reads as zero-padded, as dax.Monsters does
		items := l.BlockShared("MON0ITM", id)
		items = items[:len(items)/binary.Size(brItem{})*binary.Size(brItem{})] // a partial last record dropped, as dax.Records
		if err := addTemplates(&t, items); err != nil {
			return fmt.Errorf("MON0 %d: %w", id, err)
		}
		m, err := monster(t, r, items)
		if err != nil {
			return fmt.Errorf("MON0 %d: %w", id, err)
		}
		t.Monsters[rules.MonsterID(id)] = m
	}
	if err := adventure.WriteJSON(dst, adventure.RuleTablesFile, t); err != nil {
		return err
	}
	return importBRSaves(l, t, filepath.Join(dst, CharactersDir))
}

// SavesDir holds the shipped characters in the engine's roster format (writeRoster).
const SavesDir = "saves"

// CharactersDir holds the originals' saved characters in the folder's data
// dir, each <name>.json a rulesxxvc.Character (saves.go).
const CharactersDir = "characters"

// template is a TItemTemplate typed.
func template(r brItemTemplate) xxvc.ItemTemplate {
	return xxvc.ItemTemplate{
		Type: int32(r.ItemType), Words: [3]int32{int32(r.NameIDs[0]), int32(r.NameIDs[1]), int32(r.NameIDs[2])},
		Plus: int32(r.Plus), SaveBonus: int32(r.SaveBonus),
		NotTreasure: r.Flags&notTreasure != 0, NotJoinable: r.Flags&notJoinable != 0,
		Weight: int32(r.Weight), Quantity: int32(r.Quantity), Value: int32(r.Value),
		Special: [3]int32{int32(r.Special[0]), int32(r.Special[1]), int32(r.Special[2])},
	}
}

// addTemplates gives each item of items (TItem records) whose type and name
// words no template of t has (rulesxxvc.ConvertBRCharacter's match) a new
// template, the next ItemID past t's highest: a monster's or a save's own
// item. Stable: the import adds them in MON0 id order, item order.
func addTemplates(t *xxvc.Tables, items []byte) error {
	its, err := records[brItem](items)
	if err != nil {
		return err
	}
	for _, it := range its {
		n := template(it.Template)
		next := rules.ItemID(1)
		found := false
		for id, have := range t.Items {
			next = max(next, id+1)
			found = found || have.Type == n.Type && have.Words == n.Words
		}
		if !found {
			t.Items[next] = n
		}
	}
	return nil
}

// ItemWords is itemNameParts from the data segment ds: index 0 empty (no
// word), then words 1..e.NumWords.
func ItemWords(ds []byte, e BREXE) []string {
	w := []string{""}
	for i := range e.NumWords {
		w = append(w, pascalString(ds[e.ItemWords+i*itemWordSize:][:itemWordSize]))
	}
	return w
}

// AmmoPrices is ammoPrices from the data segment ds, by item type.
func AmmoPrices(ds []byte, e BREXE) map[int32]rules.Copper {
	m := map[int32]rules.Copper{}
	for ty := int32(firstAmmoType); ty <= lastAmmoType; ty++ {
		m[ty] = rules.Copper(binary.LittleEndian.Uint16(ds[e.AmmoPrices+int(ty-firstAmmoType)*2:]))
	}
	return m
}

// The shop lists' sizes (gdconsts.pas itemListStarts[1..16], itemListCounts[1..16]).
const shopLists = 16

// ShopLists is the shop item lists from the data segment ds: list number (1..16) -> the templates it stocks,
// in the original's order (itemListEntries from itemListStarts, itemListCounts of them; an entry is an
// ITEM0 template number, the ID of xxvc.Tables.Items).
func ShopLists(ds []byte, e BREXE) map[rules.ShopListID][]rules.ItemID {
	starts, counts, entries := ds[e.ShopLists:], ds[e.ShopLists+shopLists:], ds[e.ShopLists+2*shopLists-1:] // entries[1] is the first entry
	m := map[rules.ShopListID][]rules.ItemID{}
	for n := range shopLists {
		for i := range int(counts[n]) {
			m[rules.ShopListID(n+1)] = append(m[rules.ShopListID(n+1)], rules.ItemID(entries[int(starts[n])+i]))
		}
	}
	return m
}

// The ship tables' rows (gdconsts.pas) and records (saveload.pas).
const (
	shipWeaponKinds = 8
	shipWeaponSize  = 34 // TShipWeaponDef
	weaponShortSize = 9  // String[8]
	shipTypeSize    = 18 // String[17]
	shipTypes       = 11 // shipTypeNames[1..11]
	firstShipRecord = 40 // SaveLoad FIRST_SHIP_RECORD: SHIPS blocks below are pictures
	shipRecordSize  = 0x6C
	crewMount       = 0x80 // ShipCbt CREW_MOUNT
	kindMask        = 0x7F // C_Ship.KIND_MASK
)

// ShipWeapons is shipWeapons and weaponShortNames from the data segment ds, by kind.
func ShipWeapons(ds []byte, e BREXE) map[xxvc.ShipWeaponKind]xxvc.ShipWeapon {
	m := map[xxvc.ShipWeaponKind]xxvc.ShipWeapon{}
	for k := range shipWeaponKinds {
		r := ds[e.ShipWeapons+k*shipWeaponSize:][:shipWeaponSize]
		m[xxvc.ShipWeaponKind(k+1)] = xxvc.ShipWeapon{MaxRange: r[0], ReloadAmmo: r[2], Damage: r[3], TargetBonus: int8(r[4]),
			Name: pascalString(r[9:]), Short: pascalString(ds[e.WeaponShort+k*weaponShortSize:][:weaponShortSize])}
	}
	return m
}

// brShip is types.pas TShip as SHIPS stores it.
type brShip struct {
	Name                                     [16]byte
	Tonnage, Length, Width                   int16
	Status, Maximum                          [6]int16
	Cargo, Crew                              int16
	ArmorClass                               int8
	AlternateSkill, SystemSkill, SystemValue uint8
	BaseSpeed, Reaction                      int8
	Maneuver, ShipType, Race, RaceTail       uint8
	Mounts                                   [8]struct{ Kind, MaxStatus, Status, StatusTail, Ammo, Reloads uint8 }
}

// Ships types the SHIPS records (the $6C-byte blocks from firstShipRecord on;
// the rest are pictures); ds gives shipTypeNames.
func Ships(blocks map[int][]byte, ds []byte, e BREXE) (map[rules.ShipID]xxvc.ShipRecord, error) {
	m := map[rules.ShipID]xxvc.ShipRecord{}
	for id, b := range blocks {
		if id < firstShipRecord || len(b) != shipRecordSize { // the pictures
			continue
		}
		r, err := shipRecord(b, ds, e)
		if err != nil {
			return nil, fmt.Errorf("SHIPS %d: %w", id, err)
		}
		m[rules.ShipID(id)] = r
	}
	return m, nil
}

// shipRecord types one $6C-byte TShip (a SHIPS block, or the party's ship in a saved game's zone 0); ds gives shipTypeNames.
func shipRecord(b, ds []byte, e BREXE) (xxvc.ShipRecord, error) {
	var s brShip
	if err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &s); err != nil {
		return xxvc.ShipRecord{}, err
	}
	if s.ShipType < 1 || s.ShipType > shipTypes || int(s.Race) > len(xxvc.Races) {
		return xxvc.ShipRecord{}, fmt.Errorf("ship type %d, race %d", s.ShipType, s.Race)
	}
	r := xxvc.ShipRecord{Name: pascalString(s.Name[:]), Tonnage: s.Tonnage, Length: s.Length, Width: s.Width,
		Status: s.Status, Maximum: s.Maximum, Cargo: s.Cargo, Crew: s.Crew, ArmorClass: s.ArmorClass,
		AlternateSkill: s.AlternateSkill, SystemSkill: s.SystemSkill, SystemValue: s.SystemValue,
		BaseSpeed: s.BaseSpeed, Reaction: s.Reaction, Maneuver: s.Maneuver,
		Type: shipTypeName(ds, e, s.ShipType)}
	if s.Race > 0 { // RaceMonster: none
		r.Race = xxvc.Races[s.Race-1]
	}
	for i, w := range s.Mounts {
		r.Mounts[i] = xxvc.ShipMount{Kind: xxvc.ShipWeaponKind(w.Kind & kindMask), Crew: w.Kind&crewMount != 0,
			MaxStatus: w.MaxStatus, Status: w.Status, Ammo: w.Ammo, Reloads: w.Reloads}
	}
	return r, nil
}

// shipTypeName is shipTypeNames[shipType] (1..shipTypes) in the data segment ds.
func shipTypeName(ds []byte, e BREXE, shipType uint8) string {
	return pascalString(ds[e.ShipTypes+int(shipType-1)*shipTypeSize:][:shipTypeSize])
}

// itemTypes is the rows of GDVars.itemTypes (array[0..128] of TItemType),
// which the original reads ITEMS into, zeroed past the file.
const itemTypes = 129

// itemTypeSize is a TItemType's bytes.
const itemTypeSize = 9

// armorBase is what TItemType.weaponClass holds over an armor's class.
const armorBase = 50

// templateBlock is ITEM0's block of item templates (GDVars.itemTemplates).
const templateBlock = 1

// emptyTemplate marks an unused template (EMPTY_TEMPLATE).
const emptyTemplate = 255

// The template flags (c_item.pas).
const (
	notTreasure = 0x40
	notJoinable = 0x20
)

// Tables decodes ITEMS (TItemType rows, types.pas) and the ITEM0 templates
// (TItemTemplate, 16 bytes; template n is row n-1, as loadItemTables reads
// them) into the rule tables.
func Tables(items, templates []byte) xxvc.Tables {
	b := make([]byte, itemTypes*itemTypeSize)
	copy(b, items)
	t := xxvc.Tables{ItemTypes: make([]xxvc.ItemType, itemTypes)}
	for i := range t.ItemTypes {
		r := b[i*itemTypeSize:]
		it := xxvc.ItemType{
			Slot:           xxvc.EquipSlot(r[0]),
			AttacksDoubled: int32(r[1]),
			WeaponClass:    xxvc.WeaponClass(r[2]),
			Heavy:          r[3] != 0,
			Damage:         rules.Dice{N: int32(r[4]), Sides: int32(r[5]), Bonus: int32(int8(r[6]))},
			Range:          int32(r[7]),
			Consumable:     r[8] != 0,
		}
		if it.Slot == xxvc.SlotArmor {
			it.WeaponClass, it.ArmorClass = 0, int32(r[2])-armorBase
		}
		t.ItemTypes[i] = it
	}
	t.Items = map[rules.ItemID]xxvc.ItemTemplate{}
	tm, _ := records[brItemTemplate](templates[:len(templates)/templateSize*templateSize])
	for i, r := range tm {
		if r.ItemType != emptyTemplate {
			t.Items[rules.ItemID(i+1)] = template(r)
		}
	}
	return t
}
