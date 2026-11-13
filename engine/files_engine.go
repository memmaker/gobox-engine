package engine

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"gobox/adventure"
	"gobox/asset"
	"gobox/rules"
	"gobox/rules1e"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// savePath resolves a file name inside the save directory case-insensitively
// (DOS was case-insensitive; the shipped files are upper case).
func (e *Engine) savePath(name string) string {
	dir := e.Cfg.SaveDir
	if ents, err := os.ReadDir(dir); err == nil {
		for _, en := range ents {
			if strings.EqualFold(en.Name(), name) {
				return filepath.Join(dir, en.Name())
			}
		}
	}
	return filepath.Join(dir, strings.ToUpper(name))
}

func (e *Engine) fileExists(path string) bool { // 0802:02EB
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// fileError reports a save file that could not be read or written, once.
// (DOS asked for the save disk here and retried until it worked.)
func (e *Engine) fileError(verb, path string) {
	s := "Cannot " + verb + " " + filepath.Base(path)
	if len(s) > 40 {
		s = s[:40]
	}
	e.PromptRow24(asset.EGAYellow, s)
}

// writeFile writes data to path (its folder made first). A failure shows
// "Cannot write" and is the error returned.
func (e *Engine) writeFile(path string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err == nil {
		err = os.WriteFile(path, data, 0o644)
	}
	if err != nil {
		e.fileError("write", path)
	}
	return err
}

// removeSaved removes the save file at path; one that was never there is no
// error. A failure shows "Cannot remove" and is the error returned.
func (e *Engine) removeSaved(path string) error {
	err := os.Remove(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		e.fileError("remove", path)
		return err
	}
	return nil
}

// BuildCharList = ovr18:0010: names from charlist.txt, "* " for party
// members.
func (e *Engine) BuildCharList() []string {
	p := e.savePath("charlist.txt")
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	var list []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		name := strings.TrimRight(sc.Text(), "\r")
		if len(name) > 40 {
			name = name[:40]
		}
		for _, id := range e.partyIDs() {
			if e.ruleset().Sheet(id).Name == name {
				name = "* " + name
				if len(name) > 40 {
					name = name[:40]
				}
				break
			}
		}
		list = append(list, name)
	}
	if !slices.Equal(list, e.listed) {
		e.migrateRoster(list)
		e.listed = slices.Clone(list)
	}
	return list
}

// WriteCharList = ovr18:01B8: rewrite charlist.txt without "*" entries. A
// write that fails shows "Cannot write" and is the error returned.
func (e *Engine) WriteCharList(list []string) error {
	var b strings.Builder
	for _, n := range list {
		if len(n) > 0 && n[0] == '*' {
			continue
		}
		b.WriteString(n + "\r\n")
	}
	return e.writeFile(e.savePath("charlist.txt"), []byte(b.String()))
}

// SaveCharacter = ovr18:069E (c, name): name "" = roster save (.cha, listed
// in charlist.txt), else <name>.sav as used by SaveGame. A write that fails
// shows the game's "Cannot write" and is the error returned (the first one).
func (e *Engine) SaveCharacter(cID rules.CharID, name string) error {
	var base, ext string
	var listFailed error
	if name == "" {
		list := append(e.BuildCharList(), e.ruleset().Sheet(cID).Name)
		listFailed = e.WriteCharList(list)
		if rules1e.ReadPoolIDs(e.exportFiles(cID, false).Record).Roster == rules.NoRoster {
			id, err := e.newRosterID()
			listFailed = cmp.Or(listFailed, err)
			e.do(rules.SetRoster{Who: cID, Roster: id})
		}
		base, ext = e.charBase(cID), ".cha"
	} else {
		if len(name) > 8 {
			name = name[:8]
		}
		base, ext = name, ".sav"
	}
	f := e.exportFiles(cID, name != "")
	failed := e.writeFile(e.savePath(base+ext), f.Record)
	for _, x := range []struct {
		ext string
		b   []byte
	}{{".itm", f.Items}, {".spc", f.Affects}, {".mem", f.Plan}, {".dual", f.Dual}, {".pow", f.Powers}, {".png", f.Portrait}, {".icon.png", f.Icon}} {
		p := e.savePath(base + x.ext)
		failed = cmp.Or(failed, e.removeSaved(p))
		if len(x.b) > 0 {
			failed = cmp.Or(failed, e.writeFile(p, x.b))
		}
	}
	return cmp.Or(listFailed, failed)
}

// exportFiles are cID's files (game: a game save's).
func (e *Engine) exportFiles(cID rules.CharID, game bool) rules1e.Files {
	b, err := e.ruleset().Export(cID)
	if err != nil {
		panic(fmt.Sprintf("save: %v", err)) // a member always exports
	}
	f, err := rules1e.Unpack285(b, game)
	if err != nil {
		panic(fmt.Sprintf("save: %v", err)) // Export's own JSON
	}
	return f
}

// LoadCharacter = ovr18:0AEC (name, c, useSav): the new character, false
// when the file is missing.
func (e *Engine) LoadCharacter(name string, useSav bool) (rules.CharID, bool) {
	nm := adventure.MakeFileName(name)
	ext := ".cha"
	if useSav {
		ext = ".sav"
	}
	path := e.savePath(nm + ext)
	var f rules1e.Files
	var err error
	if f.Record, err = os.ReadFile(path); err != nil {
		e.fileError("read", path)
		return rules.NoChar, false
	}
	read := func(ext string) []byte {
		b, err := os.ReadFile(e.savePath(nm + ext))
		if err != nil {
			return nil
		}
		return b
	}
	f.Items, f.Affects, f.Dual, f.Plan, f.Powers = read(".itm"), read(".spc"), read(".dual"), read(".mem"), read(".pow")
	// extra: the dropped portrait and combat icon
	if f.Portrait, err = asset.ReadPNG(e.savePath(nm + ".png")); err != nil {
		e.fileError("read", e.savePath(nm+".png"))
	}
	if f.Icon, err = asset.ReadPNG(e.savePath(nm + ".icon.png")); err == nil && f.Icon != nil {
		err = asset.IconShape(f.Icon)
	}
	if err != nil {
		f.Icon = nil
		e.fileError("read", e.savePath(nm+".icon.png"))
	}
	b, err := rules1e.Pack285(f)
	if err == nil {
		e.setWorldWord(rules.WordWho, e.sel) // ImportCharacter's member
		var id rules.CharID
		if id, err = e.ruleset().Import(b); err == nil {
			return id, true
		}
	}
	panic(fmt.Sprintf("load %s: %v", path, err)) // Pack285's own JSON
}

// LoadMonster = ovr18:0EC4 (index): MON<n>CHA/SPC/ITM.DAX block index as a
// new character.
func (e *Engine) LoadMonster(index int) (rules.CharID, bool) {
	if _, ok := e.gdata().monster(e.monsterSet().File("MON"), index); !ok {
		e.PromptRow24(asset.EGAWhite, "Unable to load monster")
		e.Exit()
		return rules.NoChar, false
	}
	e.loadingMonster = true // DS:557C
	o, err := e.ruleset().Resolve(nil, rules.MakeMonster{Of: rules.MonsterID(index)}, nil)
	e.loadingMonster = false
	if err != nil {
		panic(fmt.Sprintf("LoadMonster: %v", err)) // the data has it
	}
	return o.Effects[0].(rules.MonsterMade).Who, true
}

// AddMonsterToParty = ovr18:1262 (index): NPC joins the party.
func (e *Engine) AddMonsterToParty(index int) {
	if e.worldWord(rules.WordPartyCount) > 7 {
		return
	}
	id, _ := e.LoadMonster(index) // a failed load dresses a blank one, as the original
	o, err := e.ruleset().Resolve(nil, rules.DressNPC{Who: id}, nil)
	if err != nil {
		panic(fmt.Sprintf("AddMonsterToParty: %v", err)) // never refused
	}
	e.join(o.Effects[0].(rules.NPCDressed).Who)
}

// eraseRosterCopies deletes <name>.cha/.itm/.spc (nested check as in LoadGame
// when nested is true, independent checks as in SaveGame otherwise).
func (e *Engine) eraseRosterCopies(cID rules.CharID, nested bool) error {
	base := e.charBase(cID)
	cha := e.savePath(base + ".cha")
	var failed error
	if e.fileExists(cha) {
		failed = cmp.Or(e.removeSaved(cha), e.removeSaved(e.savePath(base+".itm")), e.removeSaved(e.savePath(base+".spc")))
	} else if !nested {
		failed = cmp.Or(e.removeSaved(e.savePath(base+".itm")), e.removeSaved(e.savePath(base+".spc")))
	}
	return failed
}

// loadSlot is the original's LoadGame (ovr18:1602) after the prompt: a
// slot's canonical save (adventure.Save) and its party (CHRDATA1..).
func (e *Engine) loadSlot() bool {
	f, err := e.readSlotSave()
	if err != nil {
		log.Printf("load: %v", err)
		e.fileError("read", e.savePath(adventure.GameFile))
		return false
	}
	e.clearPrompt()
	e.prompt(asset.EGALightGreen, "Loading...Please Wait")
	e.set = f.Set
	if err := e.ruleset().Load(f.Rules); err != nil {
		log.Printf("load: %v", err)
		e.fileError("read", e.savePath(adventure.GameFile))
		return false
	}
	taken, err := e.ruleset().Resolve(nil, rules.TakeSavedWalls{}, nil)
	if err != nil {
		panic(fmt.Sprintf("load: %v", err)) // the action never refuses
	}
	walls := taken.Effects[0].(rules.SavedWalls)
	e.vars, e.loadedLocals = f.Fields, f.Locals
	if e.vars == nil {
		e.vars = fieldCells{}
	}
	if e.loadedLocals == nil {
		e.loadedLocals = fieldCells{}
	}
	e.posX, e.posY, e.facing = int(f.X), int(f.Y), int(f.Facing)
	e.run.doorTry = [3]byte{1, 1, 1} // extra: else a game loaded at a locked door offers only Exit, which walls it up
	e.pos3, e.pos4 = f.Pos3, f.Pos4
	e.viewMode = int(f.ViewMode)
	e.setWorldWord(rules.WordBeginFlag, int(f.BeginFlag))
	w := uint16(e.worldWord(rules.WordViewFlags))
	e.viewBit1, e.viewBit0 = byte(w>>1), byte(w&1)
	e.msgDelay = byte(e.worldWord(rules.WordMsgDelay))
	e.do(rules.EmptyParty{})
	e.do(rules.ForgetParty{})
	e.sel = 0
	for _, nm := range f.Party {
		if e.fileExists(e.savePath(nm + ".sav")) {
			if id, ok := e.LoadCharacter(nm, true); ok {
				e.join(id)
			}
		}
	}
	for _, id := range e.partyIDs() {
		e.eraseRosterCopies(id, true)
	}
	e.sel = 0
	e.set = rules.Set(e.worldWord(rules.WordSymbolSet))
	if e.viewMode < 2 {
		e.LoadGeo(int(byte(e.worldWord(rules.WordGeo)))) // ovr30:10C7
	}
	if e.viewMode > 1 {
		e.loadAreaIcon(0, 0xB) // ovr33:01C8
	} else {
		for i := 1; i <= 3; i++ {
			e.mapPairA[i-1] = int(int16(walls.A[i-1]))
			e.mapPairB[i-1] = int(int16(walls.B[i-1]))
			if e.mapPairA[i-1] >= 0 {
				e.LoadWalldef(e.mapPairA[i-1]&0xFF, e.mapPairB[i-1]&0xFF) // ovr30:0E08
			}
		}
	}
	if e.worldWord(rules.WordRiverOpen) != 0 {
		e.EndgameHook(0) // ovr19:06B3
	}
	e.clearPrompt()
	return true
}

// saveSlot is the original's SaveGame (ovr18:1D58) after the prompt, into
// the slot folder (saveTo).
func (e *Engine) saveSlot() error {
	e.clearPrompt()
	e.prompt(asset.EGALightGreen, "Saving...Please Wait")
	e.setWorldWord(rules.WordMsgDelay, int(e.msgDelay))
	e.setWorldWord(rules.WordViewFlags, int(e.viewBit1)*2+int(e.viewBit0))
	e.setWorldWord(rules.WordSymbolSet, int(e.set))
	// the running area: after an area change the area block still names the
	// one left (opNewECL; the game loop copies the new one after the next
	// command), but the load reloads the running script (GameLoop)
	v := rules.RecordView{AreaKind: uint16(e.gs.areaKind)}
	for i := range 3 {
		v.WallsA[i] = uint16(e.mapPairA[i])
		v.WallsB[i] = uint16(e.mapPairB[i])
	}
	e.do(v)
	blob, err := e.ruleset().Save()
	if err != nil {
		panic(err) // the ruleset's part always marshals
	}
	f := adventure.Save{Set: e.set, Rules: blob, Fields: e.vars, Locals: e.locals}
	f.X, f.Y, f.Facing, f.Pos3, f.Pos4 = byte(e.posX), byte(e.posY), byte(e.facing), e.pos3, e.pos4
	f.ViewMode, f.BeginFlag = byte(e.viewMode), byte(e.beginFlag())
	for i := range e.partyIDs() {
		f.Party = append(f.Party, "CHRDATA"+strconv.Itoa(i+1))
	}
	path := e.savePath(adventure.GameFile)
	failed := os.MkdirAll(e.Cfg.SaveDir, 0o755)
	if failed == nil {
		failed = adventure.WriteSave(path, f)
	}
	if failed != nil {
		log.Printf("save: %v", failed)
		e.fileError("write", path)
	}
	for i, id := range e.partyIDs() {
		failed = cmp.Or(failed, e.SaveCharacter(id, "CHRDATA"+strconv.Itoa(i+1)), e.eraseRosterCopies(id, false))
	}
	e.clearPrompt()
	return failed
}

// charBase = savePath + adventure.MakeFileName(RemoveSubstr(name,'.')) for roster files.
func (e *Engine) charBase(cID rules.CharID) string {
	return adventure.MakeFileName(rules.RemoveSubstr(e.ruleset().Sheet(cID).Name, "."))
}

const charRecSize = int(adventure.RecordLayout)

// fileBytes is v as the bytes of a save file (.mem, .dual) or of a byte field.
func fileBytes[T ~uint8](v []T) []byte {
	b := make([]byte, len(v))
	for i, x := range v {
		b[i] = byte(x)
	}
	return b
}
