package importer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/rules"
	"gobox/rulesxxvc"
)

// Buck Rogers' original saved games (SAVGAMA.DAT..SAVGAMJ.DAT) become slots of the shared engine's own save format, once, at
// import, as the Pool games' do (saveimport.go): the engine reads only adventure.Save. The file is saveload.pas SaveGame's
// (TDatHeader, then the wall set, the count and the names). The bytes are read here and nowhere else: the script cells come out
// by their layout names, the engine's words as rules.World, the ruleset's state as rulesxxvc.OriginalGame.

// The zones of the file (types.pas TZone0, TZone1, TZone2: one byte per script address, from these bases).
const (
	brZone0Base, brZone0Size = 0x4B00, 0x400
	brZone1Base, brZone1Size = 0x7C00, 0x400
	brZone2Base, brZone2Size = 0x7A00, 0x200
	brNameLimit              = 8  // characterNames: array[1..8] of TString40
	brNameBytes              = 41 // a String[40]: its length, then the text (and whatever the stack held past it)
	brSaveBytes              = 2901
)

// brSaveFile is the file as saveload.pas LoadGame reads it, field for field.
type brSaveFile struct {
	DataSet           byte
	Zone0             [brZone0Size]byte
	Zone1             [brZone1Size]byte
	Zone2             [brZone2Size]byte
	Position          [5]byte // types.pas TPartyPosition: mapX, mapY (ShortInt), facing, wallAhead, mapEvent
	PreviousInterface byte
	Interface         byte
	WallFile          uint16 // wallSets[0].fileId
	WallSlot          uint16 // wallSets[0].slot
	Count             byte
	Names             [brNameLimit][brNameBytes]byte
}

// The interface states (c_ifc.pas interfaceState) a saved game can be in, and the area kinds EclMenu's WriteGbvm tells apart.
const (
	ifcCamping      = 2
	ifcWilderness   = 3
	ifcView3D       = 4
	ifcOrbitMap     = 9
	ifcPictureView  = 10
	areaOrbit       = 0x70 // c_area.pas ORBIT
	noGeoMap        = 127  // c_map.pas NO_GEO_MAP
	noWall          = 0xFFFF
	brFieldAreaKind = "TravelKind"  // zone0 areaKind (types.pas): the layout names it so
	brFieldBlock    = "ScriptBlock" // zone0 savedEclId: the ECL block GameLoop starts the map loop with (NO_ECL_SCRIPT = 0)
)

// brOwner says who holds a range of the file.
type brOwner string

const (
	ownRuleset  brOwner = "the ruleset's typed entry (rulesxxvc.OriginalGame)"
	ownWorld    brOwner = "an engine world word (rules.World)"
	ownNotState brOwner = "not game state"
)

// brRange is a range of script addresses [From, To) the layout does not name field by field.
type brRange struct {
	From, To int
	Owner    brOwner
	What     string // what it is, and for ownNotState why it is no state
}

// brSaveRanges is the one table of what the zones hold beside the layout's named cells (a layout field's addresses are its own:
// a game or area field's value is a script cell, by name; the member window and read-only fields are no state). Every other
// byte must be zero: the import fails on a nonzero byte no entry covers. Addresses from types.pas TZone0/TZone1.
var brSaveRanges = []brRange{
	{0x4BB6, 0x4BB7, ownRuleset, "difficulty"},
	{0x4BB9, 0x4BBA, ownWorld, "orbit.tick: World.OrbitTick"},
	{0x4BBB, 0x4BBC, ownWorld, "orbit.phase: World.OrbitStep"},
	{0x4BBD, 0x4BBE, ownWorld, "orbit.rocketIcon: World.RocketHeading, the icon past C_Orbit.FIRST_ROCKET_ICON"},
	{0x4BC6, 0x4BCD, ownRuleset, "clock, seven units"},
	{0x4D00, 0x4D6C, ownRuleset, "partyShip (TShip, $6C bytes; its two pad bytes statusTail and raceTail are read by no code)"},
	{0x4D6C, 0x4D7C, ownRuleset, "the last enemy ship: savedSystemStatus, savedShipType, savedCrew, savedSystemValue"},
	{0x4D81, 0x4D82, ownRuleset, "savedWeaponCount"},
	{0x7C00, 0x7EA8, ownNotState, "the selected member's window: its record mirrored into Party memory (EclMenu SpecialReadZ1/WriteZ1); the member's own file holds it"},
	{0x7F3E, 0x7F3F, ownNotState, "partyCount: LoadGame zeroes it and AddCharToParty counts the members it loads"},
}

// brRocketIcons: C_Orbit.FIRST_ROCKET_ICON, the icon at heading 0 (adventure.SpaceMap.ShipFrame), and the headings there are.
const brRocketHeadings = 8

// ImportBRSave turns the original saved game b into the canonical save. scripts holds the game's layout; ds and e give the ship
// type names. The members' files are named by the save (Party); the caller supplies them.
func ImportBRSave(b []byte, scripts fs.FS, ds []byte, e BREXE) (adventure.Save, error) {
	var f brSaveFile
	if len(b) != brSaveBytes {
		return adventure.Save{}, fmt.Errorf("saved game: %d bytes, want %d", len(b), brSaveBytes)
	}
	if err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &f); err != nil {
		return adventure.Save{}, err
	}
	l, err := ecl.ReadLayout(scripts, 0)
	if err != nil {
		return adventure.Save{}, err
	}
	if l == nil {
		return adventure.Save{}, errors.New("no script layout")
	}
	s := adventure.Save{Set: rules.Set(f.DataSet), Fields: map[string][]int64{}, Locals: map[string][]int64{},
		X: f.Position[0], Y: f.Position[1], Facing: f.Position[2], Pos3: f.Position[3], Pos4: f.Position[4], BeginFlag: byte(rules.PhaseExploring)}
	world := rules.World{}
	cells := world.Cells()
	m := brMemory{f: &f}
	claimed := map[int]bool{}
	for _, r := range brSaveRanges {
		for a := r.From; a < r.To; a++ {
			claimed[a] = true
		}
	}
	covered := map[int]bool{}
	for a := range claimed {
		covered[a] = true
	}
	for _, fl := range l.Fields {
		if fl.Owner != ecl.OwnerGame && fl.Owner != ecl.OwnerArea && fl.Owner != ecl.OwnerLocal && fl.Owner != ecl.OwnerMember {
			continue // an engine fact, a call or a constant table: no cell of the file
		}
		span, err := brSpan(fl)
		if err != nil {
			return s, err
		}
		in, out := 0, 0
		for a := fl.From; a < fl.From+span; a++ {
			if claimed[a] {
				in++
			} else {
				out++
			}
		}
		if in > 0 && out > 0 {
			return s, fmt.Errorf("layout field %s (%04X) lies across a range of the table", fl.Name, fl.From)
		}
		if in > 0 {
			continue // its bytes are another owner's: the table says whose
		}
		if m.zone(fl.From) == nil {
			return s, fmt.Errorf("layout field %s (%04X) is in no zone of the file", fl.Name, fl.From)
		}
		for a := fl.From; a < fl.From+span; a++ {
			covered[a] = true
		}
		if fl.Owner == ecl.OwnerMember || fl.ReadOnly {
			continue // the selected member's window (a mirror of its file), or derived by the engine: no state
		}
		c, err := m.cells(fl)
		if err != nil {
			return s, err
		}
		switch cell, isWorld := cells[fl.Name]; {
		case isWorld:
			if len(c) > 1 || len(c) == 1 && c[0] > 0xFFFF {
				return s, fmt.Errorf("layout field %s: world word of %v", fl.Name, c)
			}
			if len(c) == 1 {
				*cell = uint16(c[0])
			}
		case fl.Owner == ecl.OwnerLocal:
			if c != nil {
				s.Locals[fl.Name] = c
			}
		default:
			if c != nil {
				s.Fields[fl.Name] = c
			}
		}
	}
	if bad := m.uncovered(covered); len(bad) > 0 {
		return s, fmt.Errorf("saved game: nonzero byte at %s, in no range of the table (importer/br_origsave.go brSaveRanges, nor a layout field)", strings.Join(bad, ", "))
	}
	g := rulesxxvc.OriginalGame{}
	if g.Difficulty, err = brDifficulty(m.byteAt(0x4BB6)); err != nil {
		return s, err
	}
	g.Clock = rulesxxvc.OriginalClock{Segment: int(m.byteAt(0x4BC6)), Minute: int(m.byteAt(0x4BC7)), Turn: int(m.byteAt(0x4BC8)),
		Hour: int(m.byteAt(0x4BC9)), Day: int(m.byteAt(0x4BCA)), Month: int(m.byteAt(0x4BCB)), Year: int(m.byteAt(0x4BCC))}
	ship, err := shipRecord(m.slice(0x4D00, shipRecordSize), ds, e)
	if err != nil {
		return s, fmt.Errorf("saved game's ship: %w", err)
	}
	g.Ship = ship
	if g.Foe, err = brFoe(&m, ds, e); err != nil {
		return s, err
	}
	if err := brView(&f, &world, &s, m.byteAt(brAddr(l, brFieldAreaKind))); err != nil {
		return s, err
	}
	world.OrbitTick, world.OrbitStep = uint16(m.byteAt(0x4BB9)), uint16(m.byteAt(0x4BBB))
	if icon := int(m.byteAt(0x4BBD)); icon != 0 { // 0: a record never started (GameInit sets the icon)
		if icon < spaceShipFrame || icon >= spaceShipFrame+brRocketHeadings {
			return s, fmt.Errorf("saved game: rocket icon %d is no heading from %d", icon, spaceShipFrame)
		}
		world.RocketHeading = uint16(icon - spaceShipFrame)
	}
	g.Walls = brWalls(&f, &world)
	if r, ok := brField(l, brFieldBlock); ok {
		world.AreaKind = uint16(m.byteAt(r.From))
	}
	if s.Rules, err = g.Save(); err != nil {
		return s, err
	}
	s.World = &world
	if f.Count > brNameLimit {
		return s, fmt.Errorf("saved game: %d members, at most %d", f.Count, brNameLimit)
	}
	for i := range int(f.Count) {
		n := f.Names[i]
		if int(n[0]) > brNameBytes-1 {
			return s, fmt.Errorf("saved game: member %d's name is %d long", i+1, n[0])
		}
		s.Party = append(s.Party, adventure.MakeFileName(rules.RemoveSubstr(pascalString(n[:]), ".")))
	}
	return s, nil
}

// brDifficulty is zone0's difficulty (c_diff.pas) as the ruleset's level.
func brDifficulty(b byte) (rulesxxvc.Difficulty, error) {
	if d := rulesxxvc.Difficulty(b); d >= rulesxxvc.Novice && d <= rulesxxvc.Ace {
		return d, nil
	}
	return 0, fmt.Errorf("saved game: difficulty %d is no level (c_diff.pas)", b)
}

// brFoe is the last enemy ship zone0 keeps (savedSystemStatus, savedShipType, savedCrew, savedSystemValue, savedWeaponCount); a
// type of 0 is none (a ship lost off the scanners: all zero).
func brFoe(m *brMemory, ds []byte, e BREXE) (rulesxxvc.SavedFoe, error) {
	var foe rulesxxvc.SavedFoe
	for i := range foe.Status {
		foe.Status[i] = int16(binary.LittleEndian.Uint16(m.slice(0x4D6C+2*i, 2)))
	}
	kind := m.byteAt(0x4D78)
	switch {
	case kind == 0:
		if foe.Status != [6]int16{} || m.byteAt(0x4D79) != 0 || m.byteAt(0x4D7A) != 0 || m.byteAt(0x4D7B) != 0 || m.byteAt(0x4D81) != 0 {
			return foe, errors.New("saved game: a last enemy ship with no type")
		}
		return foe, nil
	case int(kind) > shipTypes:
		return foe, fmt.Errorf("saved game: last enemy ship type %d", kind)
	}
	foe.Type = shipTypeName(ds, e, kind)
	foe.Crew = int16(binary.LittleEndian.Uint16(m.slice(0x4D79, 2)))
	foe.SystemValue = m.byteAt(0x4D7B)
	foe.WeaponCount = m.byteAt(0x4D81)
	return foe, nil
}

// brView is what saveload.pas LoadGame makes of the screen the game was saved in (interfaceState, c_ifc.pas), as the engine's
// words: the 3D view, the overland (the wilderness), the system map, or a place its script runs (picture view). Camping is the
// view of the area the party camps in, which EclMenu's WriteGbvm derives from the area kind the same way (the wilderness kinds,
// ORBIT, any other). A state SaveGame cannot be called in (the party menu, a shop, a fight, the ship screen, skill editing)
// has no view: an error.
func brView(f *brSaveFile, w *rules.World, s *adventure.Save, areaKind byte) error {
	state := int(f.Interface)
	if state == ifcCamping {
		switch areaKind {
		case areaVenusWilderness, areaMarsWilderness:
			state = ifcWilderness
		case areaOrbit:
			state = ifcOrbitMap
		default:
			state = ifcView3D
		}
	}
	switch state {
	case ifcView3D:
		w.In3D, w.Scripted, w.Flying = 1, 0, 0
		s.ViewMode = adventure.ViewMode3D
	case ifcWilderness:
		w.In3D, w.Scripted, w.Flying = 0, 0, 0
		s.ViewMode = adventure.ViewModeOverland
	case ifcOrbitMap:
		w.In3D, w.Scripted, w.Flying = 0, 0, 1
		s.ViewMode = adventure.ViewModeOverland
	case ifcPictureView: // a place its script runs: the exploration frame without the 3D view
		w.In3D, w.Scripted, w.Flying = 0, 1, 0
		s.ViewMode = adventure.ViewMode3D
	default:
		return fmt.Errorf("saved game: saved in interface state %d, which has no view (c_ifc.pas)", f.Interface)
	}
	return nil
}

// The wilderness area kinds (c_area.pas).
const (
	areaVenusWilderness = 0xA2
	areaMarsWilderness  = 0xA8
)

// brWalls is the wall sets LoadGame loads back: wallSets[0] only, and only with a map (a wall set, a GEO map that is not
// NO_GEO_MAP, and a screen with a 3D view). Without one, no map is loaded and none is named.
func brWalls(f *brSaveFile, w *rules.World) rules.RecordView {
	v := rules.RecordView{WallsA: [3]uint16{noWall, noWall, noWall}, WallsB: [3]uint16{noWall, noWall, noWall}}
	if f.WallFile > 0 && w.Geo != noGeoMap && f.Interface != ifcPictureView {
		v.WallsA[0], v.WallsB[0] = f.WallFile, f.WallSlot
		return v
	}
	w.Geo = noGeoMap
	return v
}

// brMemory is the zones by script address.
type brMemory struct{ f *brSaveFile }

// zone is the byte at script address a, nil outside the three zones.
func (m brMemory) zone(a int) *byte {
	switch {
	case a >= brZone0Base && a < brZone0Base+brZone0Size:
		return &m.f.Zone0[a-brZone0Base]
	case a >= brZone1Base && a < brZone1Base+brZone1Size:
		return &m.f.Zone1[a-brZone1Base]
	case a >= brZone2Base && a < brZone2Base+brZone2Size:
		return &m.f.Zone2[a-brZone2Base]
	}
	return nil
}

func (m brMemory) byteAt(a int) byte {
	p := m.zone(a)
	if p == nil {
		panic(fmt.Sprintf("importer: address %04X is in no zone of a saved game", a))
	}
	return *p
}

func (m brMemory) slice(a, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = m.byteAt(a + i)
	}
	return b
}

// uncovered lists the addresses of nonzero bytes not in covered.
func (m brMemory) uncovered(covered map[int]bool) []string {
	var bad []string
	for _, z := range []struct{ base, size int }{{brZone0Base, brZone0Size}, {brZone2Base, brZone2Size}, {brZone1Base, brZone1Size}} {
		for a := z.base; a < z.base+z.size; a++ {
			if m.byteAt(a) != 0 && !covered[a] {
				bad = append(bad, fmt.Sprintf("%04X=%d", a, m.byteAt(a)))
			}
		}
	}
	return bad
}

// brSpan is the bytes a field occupies in the file. One element of a word or a dword is its own width; an array of them is no
// field of any BR layout (a layout is made of single words and bytes, bytes arrays and texts), so none is accepted.
func brSpan(f ecl.Field) (int, error) {
	switch f.Width {
	case ecl.WidthByte, ecl.WidthText, ecl.WidthBytes:
		return f.Len, nil
	case ecl.WidthWord, ecl.WidthDword:
		if f.Len != 1 {
			return 0, fmt.Errorf("layout field %s: an array of words", f.Name)
		}
		return map[ecl.Width]int{ecl.WidthWord: 2, ecl.WidthDword: 4}[f.Width], nil
	}
	return 0, fmt.Errorf("layout field %s: width %d", f.Name, f.Width)
}

// cells are the field's cells in the file, as the engine keeps a game's field (a text up to its NUL; nil when all are zero).
func (m brMemory) cells(f ecl.Field) ([]int64, error) {
	var c []int64
	switch f.Width {
	case ecl.WidthWord:
		c = []int64{int64(binary.LittleEndian.Uint16(m.slice(f.From, 2)))}
	case ecl.WidthDword:
		c = []int64{int64(binary.LittleEndian.Uint32(m.slice(f.From, 4)))}
	default:
		for i := range f.Len {
			v := m.byteAt(f.From + i)
			if f.Width == ecl.WidthText && v == 0 {
				break
			}
			c = append(c, int64(v))
		}
	}
	if !slices.ContainsFunc(c, func(v int64) bool { return v != 0 }) {
		return nil, nil
	}
	return c, nil
}

// brField is the layout's field named name.
func brField(l *ecl.Layout, name string) (ecl.Field, bool) {
	for _, f := range l.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return ecl.Field{}, false
}

// brAddr is the address of the layout's field name (0, in no zone, when it has none: the caller reads a zero).
func brAddr(l *ecl.Layout, name string) int {
	if f, ok := brField(l, name); ok {
		return f.From
	}
	return brZone0Base
}

// importBRGames writes each original saved game of dos as the slot "GAME <letter>" of dst's saves (the members' files as the
// slot's own: the characters importBRSaves wrote under data/characters, by the names the save lists).
func importBRGames(dos, dst string, ds []byte, e BREXE) error {
	found, err := filepath.Glob(filepath.Join(dos, origSaves))
	if err != nil || len(found) == 0 {
		return err
	}
	scripts := os.DirFS(filepath.Join(dst, "scripts"))
	chars := filepath.Join(dst, adventure.DataDir, CharactersDir)
	for _, path := range found {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		g, err := ImportBRSave(b, scripts, ds, e)
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		letter := strings.ToUpper(filepath.Base(path)[len("SAVGAM") : len("SAVGAM")+1])
		slot := filepath.Join(dst, SavesDir, adventure.SlotsDir, "GAME "+letter)
		if err := os.MkdirAll(slot, 0o755); err != nil {
			return err
		}
		for _, m := range g.Party {
			rec, err := os.ReadFile(filepath.Join(chars, m+".json"))
			if err != nil {
				return fmt.Errorf("%s: member %s: %w", filepath.Base(path), m, err)
			}
			if err := os.WriteFile(filepath.Join(slot, m+".sav"), rec, 0o644); err != nil {
				return err
			}
		}
		if err := adventure.WriteSave(filepath.Join(slot, adventure.GameFile), g); err != nil {
			return err
		}
	}
	return nil
}
