package adventure

import (
	"encoding/json"
	"errors"
	"fmt"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules1e"
	"io/fs"
	"os"
	"path/filepath"
)

// SoundFile is the driver's songs, scripts and vibrato waveforms, decoded from
// START.EXE at import (ImportData, audio.DecodeSoundData at porSoundLayout) and
// synthesized at run time.
const SoundFile = "sound.json"

// The data directory (New's dir) holds one folder per kind of file.
const (
	DOSDir    = "dos"    // the DOS game files the engine reads (*.DAX, ITEMS, POOL.CFG)
	AmigaDir  = "amiga"  // the Amiga graphics and sound, converted from its disks
	AssetsDir = "assets" // the host fonts/ (and, for the tests, an adventure folder: scripts/, art/, data/, sound/, speech/, soundtrack/)
	SaveDir   = "save"   // what the game writes: the saves, FONT.CFG
)

// CharRecord is the 0x11D-byte DOS character record (gobox/char's shared
// codec): the .SAV files and the MON*CHA blocks (and the imported data's
// JSON) in their layout. The ruleset holds the record as its Sheet
// (char.SheetOf, char.SheetBytes).
type CharRecord = rules1e.Record285

// RecordLayout is the layout of the game's character records (its
// adventure's manifest names it).
const RecordLayout = rules1e.Layout285

// Rest choices (CampRest).
const (
	RestFix   = iota // until everyone is healed (healParty) and the planned spells are learned
	RestWait         // an hour
	RestUntil        // until day (at night) or night (by day)
)

// DungeonTiles is how many tiles DungCom puts at the start of the set.
const DungeonTiles = 0x18 + 1

// IconParts are the parts in IconColour order (the hair and the face
// share a part).
var IconParts = [6]string{"Body", "Arm", "Leg", "Hair/Face", "Shield", "Weapon"}

// JournalDir is the game's journal books and pictures in the adventure folder.
const JournalDir = "journal"

// DataDir is the imported game data under the assets (ImportData): the
// maps as a binary file and a metadata text file per GEO archive, the
// monsters and items as JSON. Nothing at run time reads the archives' data.
const DataDir = "data"

// ArtDir is the imported art under the assets (gobox/asset sheets): dos/
// and, after the Amiga import, amiga/.
const ArtDir = "art"

// LogoFile is the menu lettering, finished, in the adventure's ArtDir (PNG,
// transparent round the letters), written by the game's import; absent: the
// game has none (its title backdrop carries its own).
const LogoFile = "LOGO.png"

// SoundDir is the adventure's Amiga sound (SOUND.BIN) and the sound pack (daspDir).
const SoundDir = "sound"

// PortraitsFile is the portrait tables in the adventure's data folder
// (DataDir), written by the game's import.
const PortraitsFile = "portraits.json"

// GameTablesFile is the game's GameTables other than its portraits in the adventure's
// data folder (DataDir), written by the game's import.
const GameTablesFile = "gametables.json"

// ReadGameTables reads the game's tables from the adventure folder adv, as the
// game's import wrote them: GameTablesFile and the portrait tables (PortraitsFile);
// a game writes either or both. Script is the Dialect's.
func ReadGameTables(adv string) (GameTables, error) {
	var t GameTables
	found := false
	for _, f := range []string{GameTablesFile, PortraitsFile} {
		b, err := os.ReadFile(filepath.Join(adv, DataDir, f))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return GameTables{}, err
		}
		if err := json.Unmarshal(b, &t); err != nil {
			return GameTables{}, fmt.Errorf("%s: %w", f, err)
		}
		found = true
	}
	if !found {
		return GameTables{}, fmt.Errorf("%s and %s: not found (run the game's import)", GameTablesFile, PortraitsFile)
	}
	if t.Dialect != 0 {
		t.Script = ecl.Versions[ecl.VersionID(t.Dialect)]
	}
	return t, nil
}

// TreasureSetsFile is GameTables.TreasureSets in the adventure's data folder (DataDir),
// written by the game's import (pod).
const TreasureSetsFile = "treasuresets.json"

// GameDataFile is the game's declarative data (game.Data without what the
// engine builds at run time) in the adventure's data folder (DataDir), written by the
// game's import: its id, party menu, verbs, Quick Start party, font, border,
// rests, spell tips and the rest of game.Data's static fields.
const GameDataFile = "gamedata.json"

// WriteGameData writes d (its static fields) as GameDataFile into the data folder dir.
func WriteGameData(dir string, d game.Data) error {
	b, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, GameDataFile), b, 0o644)
}

// ReadGameData reads the game's data from the adventure folder adv, as its
// import wrote it. An Amiga music set (AmigaMusic) whose files the folder's sound/
// lacks is dropped; with the files the game has both the DOS release's music
// (AdLib) and the Amiga's, and the player picks (the Music option).
func ReadGameData(adv string) (game.Data, error) {
	b, err := os.ReadFile(filepath.Join(adv, DataDir, GameDataFile))
	if err != nil {
		return game.Data{}, fmt.Errorf("%w (run the game's import)", err)
	}
	var d game.Data
	if err := json.Unmarshal(b, &d); err != nil {
		return game.Data{}, fmt.Errorf("%s: %w", GameDataFile, err)
	}
	if am := d.AmigaMusic; am != nil {
		have := true
		for _, f := range []string{am.Songs, am.Effects} {
			if f == "" {
				continue
			}
			if _, err := os.Stat(filepath.Join(adv, SoundDir, f)); err != nil {
				have = false
			}
		}
		if !have {
			d.AmigaMusic = nil
		}
	}
	return d, nil
}
