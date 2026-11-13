package engine

import (
	"bytes"
	"cmp"
	"encoding/gob"
	"fmt"
	"gobox/asset"
	"gobox/char"
	"gobox/ecl/script"
	"gobox/rules"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Named save slots (extra), F10 at any prompt once a game is under way, and
// the party menu's Load (loading only). They replace the original's lettered
// saves A-J: Save is gone from camp and the party menu, and lettered saves
// in the save directory (an imported game's) become slots "GAME A".."GAME J".
// A slot is a folder <SaveDir>/saves/<name> holding the canonical save
// (GAME.JSON, savegame.go) and its CHRDATA* files; a slot holding an
// original savgamA.dat is imported once (importSlots).
// Loading unwinds whatever is running (a menu, a script, a battle) back to
// adventure, which resumes from the loaded game like the party menu's Load.
// Saving during combat is refused: the battle is not part of the format.

// endGame: the game in play is over (a script quit, the party lost, Back
// to Main Menu). extra: the original exits to DOS when the party is lost
// (start:013D-0142); the port goes back to the party menu, and the next
// party starts afresh, not where this one died.
func (e *Engine) endGame() {
	e.unwindReset()
	e.resetAdventure()
	e.loadedGame = false
}

// unwindReset closes everything open (without what follows them) and
// their input state: a load or Back to Main Menu; the call that did it
// finds them gone (closed).
func (e *Engine) unwindReset() {
	e.resets++
	x := &e.ext
	x.read = nil
	x.back, x.aoe, x.enemiesOnly = 0, 0, false
	x.panel = rules.NoChar
	e.k = nil
	x.explore, e.eclWait = false, nil
	e.do(rules.EndSession{})
	e.resting, e.menu, e.es.hold = nil, partyMenu{}, holdNone
	x.redraws, e.temps = nil, nil
	e.view.on, e.view.chain = false, false
	e.picker, e.creation, e.iconEd, x.creating = nil, nil, nil, false
	e.pending = PendNone
	e.snap.invOn, e.snap.shop, e.snap.sheetOn = false, false, false
	e.sel = 0                         // a defer on the way out may have put back the old party's index
	e.loadedGame = e.beginFlag() != 0 // DS:49F8, as the party menu's Load
}

func (e *Engine) slotRoot() string { return filepath.Join(e.Cfg.SaveDir, "saves") }

// store is the slot folders (gobox app/saves): a slot holds save A.
func (e *Engine) store() asset.Store { return asset.Store{Dir: e.slotRoot(), Marker: gameFile} }

// slots lists the saved games, newest first.
func (e *Engine) slots() []asset.Slot {
	e.importSlots()
	return e.store().List()
}

// slotLevelOf is slotLevel of sl, read again only when the slot was written again.
func (e *Engine) slotLevelOf(sl asset.Slot) int {
	if c, ok := e.ext.levels[sl.Name]; ok && c.at.Equal(sl.At) {
		return c.level
	}
	if e.ext.levels == nil {
		e.ext.levels = map[string]slotLevelAt{}
	}
	c := slotLevelAt{sl.At, e.slotLevel(sl.Name)}
	e.ext.levels[sl.Name] = c
	return c.level
}

type slotLevelAt struct {
	at    time.Time
	level int
}

// slotLevel is the highest class level of a member of the slot's party
// (0: none read).
func (e *Engine) slotLevel(name string) (lv int) {
	e.inSlot(name, func() {
		f, err := e.readSlotSave()
		if err != nil {
			return
		}
		for _, nm := range f.Party {
			if b, err := os.ReadFile(e.savePath(nm + ".sav")); err == nil {
				for _, l := range char.SheetOf(b).Levels {
					lv = max(lv, int(int8(l)))
				}
			}
		}
	})
	return lv
}

// inSlot runs f with the save directory pointed at the slot's folder.
func (e *Engine) inSlot(name string, f func()) { e.inDir(e.store().Path(name), f) }

// inDir runs f with the save directory pointed at dir.
func (e *Engine) inDir(dir string, f func()) {
	old := e.Cfg.SaveDir
	e.Cfg.SaveDir = dir
	defer func() { e.Cfg.SaveDir = old }()
	f()
}

// SlotIndex is the slot's line on the slot screen, or -1 (tests).
func (e *Engine) SlotIndex(name string) int {
	return e.store().Index(name)
}

// seenFile holds a slot's explored squares by map (extra: the original
// keeps none); oldSeenFile is what slots saved before 30 Oct 2026 hold
// them in, by script area (areaKey was script area and GEO block then).
const seenFile, oldSeenFile = "MAPS.GOB", "SEEN.GOB"

// voicedFile holds the voice-over recordings a slot has heard (voice.go).
const voicedFile = "VOICED.GOB"

// metFile holds the monster kinds a slot has met (bestiary.go).
const metFile = "BESTIARY.GOB"

// questsFile holds the quests a slot was given (quests.go).
const questsFile = "QUESTS.GOB"

// notesFile holds the journal notes of a slot (lore.go).
const notesFile = "NOTES.GOB"

// doorsFile holds the doors a slot opened for good (openDoor).
const doorsFile = "DOORS.GOB"

// loadFrom loads the slot, its explored squares and its map labels (none
// for a slot saved before they were kept).
func (e *Engine) loadFrom(name string) (ok bool) {
	e.inSlot(name, func() { ok = e.loadSlot() })
	if !ok {
		return false
	}
	e.ext.seen, e.ext.voiced, e.ext.met, e.ext.beasts, e.ext.quests, e.ext.notes, e.ext.doors = nil, nil, nil, nil, nil, nil, nil
	if f, err := os.Open(filepath.Join(e.slotRoot(), name, seenFile)); err == nil {
		gob.NewDecoder(f).Decode(&e.ext.seen)
		f.Close()
	} else if f, err := os.Open(filepath.Join(e.slotRoot(), name, oldSeenFile)); err == nil {
		var old map[int]*[256]uint8
		gob.NewDecoder(f).Decode(&old)
		f.Close()
		e.ext.seen = e.seenByMap(old)
	}
	if f, err := os.Open(filepath.Join(e.slotRoot(), name, voicedFile)); err == nil {
		gob.NewDecoder(f).Decode(&e.ext.voiced)
		f.Close()
	}
	if f, err := os.Open(filepath.Join(e.slotRoot(), name, metFile)); err == nil {
		gob.NewDecoder(f).Decode(&e.ext.met)
		f.Close()
	}
	if f, err := os.Open(filepath.Join(e.slotRoot(), name, questsFile)); err == nil {
		gob.NewDecoder(f).Decode(&e.ext.quests)
		f.Close()
	}
	if f, err := os.Open(filepath.Join(e.slotRoot(), name, notesFile)); err == nil {
		gob.NewDecoder(f).Decode(&e.ext.notes)
		f.Close()
	}
	if f, err := os.Open(filepath.Join(e.slotRoot(), name, doorsFile)); err == nil {
		gob.NewDecoder(f).Decode(&e.ext.doors)
		f.Close()
	}
	e.reopenDoors() // the map loadSlot loaded
	e.readLabels(filepath.Join(e.slotRoot(), name, labelsFile))
	return true
}

// seenByMap turns explored squares keyed by script area and GEO block
// (oldSeenFile) into ones keyed by data set and GEO block: the scripts of
// one map (names/areas.txt) pool what they saw.
func (e *Engine) seenByMap(old map[int]*[256]uint8) map[int]*[256]uint8 {
	areas := e.scriptAreas()
	seen := map[int]*[256]uint8{}
	for k, m := range old {
		k = mapOfArea(areas, k)
		nm := seen[k]
		if nm == nil {
			nm = &[256]uint8{}
			seen[k] = nm
		}
		for i, v := range m {
			nm[i] = max(nm[i], v)
		}
	}
	return seen
}

func (e *Engine) scriptAreas() []script.Area {
	areas, _ := script.ReadAreas(os.DirFS(e.scriptDir()), "names/areas.txt")
	return areas
}

// mapOfArea turns an old key (script area and GEO block) into areaKey's
// (data set and GEO block) by the scripts' areas (names/areas.txt).
func mapOfArea(areas []script.Area, k int) int {
	set := k >> 16 // unknown: the area number stands (nothing pools)
	for _, a := range areas {
		if a.ID == k>>16 {
			set = a.Set
		}
	}
	return set<<16 | k&0xFFFF
}

// writeGob writes v gob-encoded to path in one write.
func writeGob(path string, v any) error {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(v); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

func (e *Engine) saveTo(name string) error {
	// written beside the old save, which is replaced only once the new one is
	// complete (a fresh folder: no stale CHRDAT files of a bigger party; saves.Store.Write)
	err := e.store().Write(name, func(tmp string) error {
		var failed error
		e.inDir(tmp, func() { failed = e.saveSlot() })
		return cmp.Or(failed,
			writeGob(filepath.Join(tmp, seenFile), e.ext.seen),
			writeGob(filepath.Join(tmp, voicedFile), e.ext.voiced),
			writeGob(filepath.Join(tmp, metFile), e.ext.met),
			writeGob(filepath.Join(tmp, questsFile), e.ext.quests),
			writeGob(filepath.Join(tmp, notesFile), e.ext.notes),
			writeGob(filepath.Join(tmp, doorsFile), e.ext.doors),
			e.writeLabels(filepath.Join(tmp, labelsFile)))
	})
	if err != nil {
		return fmt.Errorf("Save failed: %w", err)
	}
	e.say("Game saved successfully") // extra: the note, on the slot page and the page under it
	return nil
}

// canSave says whether a save is taken now: not from the party menu's
// Load (loadOnly), and only with a game in play outside a fight and with no
// vault open (its items and the screen to leave it are not in a save).
func (e *Engine) canSave(loadOnly bool) bool {
	return !loadOnly && e.ext.inGame && len(e.partyIDs()) > 0 && e.beginFlag() != 5 && e.vault == nil
}

// SlotsView is the save slots, newest first, for the presenter's F10
// screen (nil on the title); loadOnly is the party menu's Load.
func (e *Engine) SlotsView(loadOnly bool) *rules.Slots {
	if e.ext.intro {
		return nil
	}
	v := &rules.Slots{CanSave: e.canSave(loadOnly), LoadOnly: loadOnly}
	for _, sl := range e.slots() {
		v.Names, v.Times = append(v.Names, sl.Name), append(v.Times, sl.At.Format("Jan 02 15:04"))
		v.Levels = append(v.Levels, e.slotLevel(sl.Name))
	}
	if len(v.Names) > 0 {
		v.Current = v.Names[0] // the last save (newest first), also after a restart
	}
	return v
}

// SaveNamed saves the game in slot name, over a slot of that name in any case.
func (e *Engine) SaveNamed(name string) error {
	if name = asset.Name(name); !e.canSave(false) || name == "" {
		return rules.ErrNotNow
	}
	// one slot per name in any case (the folders may not tell case apart)
	if i := slices.IndexFunc(e.slots(), func(s asset.Slot) bool { return strings.EqualFold(s.Name, name) }); i >= 0 {
		if err := os.RemoveAll(e.store().Path(e.slots()[i].Name)); err != nil {
			return err
		}
	}
	return e.saveTo(name)
}

// DeleteSlot deletes save slot name.
func (e *Engine) DeleteSlot(name string) error {
	return e.store().Delete(name)
}

// LoadNamed loads save slot name; false when it could not. From the party
// menu (loadOnly) the menu stays; elsewhere adventure resumes from the
// loaded game.
func (e *Engine) LoadNamed(name string, loadOnly bool) bool {
	if e.SlotIndex(name) < 0 {
		return false
	}
	if e.beginFlag() == 5 {
		e.combatViewLeave()
	}
	if !e.loadFrom(name) {
		return false
	}
	if loadOnly {
		if e.menu.on && e.beginFlag() != 0 { // back from the party menu's Load with a game
			e.loadedGame = true // DS:49F8
			e.closeParty()
		}
		return true
	}
	e.unwindReset()         // what is open is closed: adventure resumes from the loaded game
	if e.beginFlag() == 0 { // a save from the party menu: no area yet, so back there
		e.PartyMenu()
	} else {
		e.GameLoop()
	}
	return true
}
