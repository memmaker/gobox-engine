package importer

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gobox/adventure"
	"gobox/game"
)

// GameData is Silver Blades' declarative data for gobox (game.Data) over
// the imported adventure folder adv (its art and sound/; the Amiga's songs
// and effects when the import found them): the spell tips, and the other games' rosters it reads (Curse of the Azure
// Bonds' characters, rules/charimport.go), its font and window border.
func GameData(adv string) game.Data {
	d := game.Data{Series: "pool", ID: "sotsb", Adventure: adv, SpellRules: sotsbSpellTips(), Imports: []string{"cotab"}, QuickParty: sotsbQuickParty,
		MusicPlaces: SotsbMusicPlaces(), Scans: book().Picture,
		Font: "Gold_Box_8x8D1",
		Confirms: map[string]game.Confirm{ // the original's words: O100:0994, O100:2C77, O0EF:23DA (both on No)
			"Drop":         {Asks: []string{"Your %s\nwill be gone forever\n\nDrop It?"}},
			"Sell":         {Asks: []string{"I'll give you %p\nfor your %s\n\nWill you sell it?"}},
			"SellFree":     {Asks: []string{"Your %s is worthless to me.\n\nShall I dispose of it for you?"}}, // O100:2D4E
			"SellFreeMany": {Asks: []string{"Your %s are worthless.\n\nWould you like to discard them?"}},     // O100:2DA1
			"DropMember":   {Asks: []string{"Drop %s forever?", "Are you sure?"}},
		},
		// the screen frame as R19D:0000 draws it: symbol 311 (frame 55) on the top and bottom rows,
		// corners included, 309 (frame 53) down the sides; symbol set 4 from 256 is 8X8D1 block 202
		Commands:     game.CommandEntries(sotsbPartyCommands),
		Border:       &game.Border{File: "8X8D1", Block: 202, Pieces: [8]int{55, 55, 55, 55, 55, 55, 53, 53}, Caps: []int{52, 54, 52, 54}},
		SoundTables:  soundData,
		AdLib:        &game.AdLib{Music: "SOUND.DAX", Tempo: 0x76},
		DungeonTiles: adventure.DungeonTiles,
		// camp's rests: the timed one in days and hours (Rest), and Fix (CampRest Kind restFix)
		Rests:     []game.RestKind{{Text: "Rest", Key: 'R', N: 0, Duration: true}, {Text: "Fix", Key: 'F', N: restFix, Fix: true}},
		IconParts: IconPartNames[:],
		// O089's turn menu Move View Aim Use Cast Quick Done, Done's Guard Delay Quit Bandage: gobox's verbs by the original's names and letters
		AimsShots:  true,
		Texts:      map[string]string{"Ranged": "Aim", "Items": "Use", "End": "Done"},
		VerbKeys:   map[string]rune{"Ranged": 'A', "Items": 'U', "Move": 'M', "Cast": 'C', "Quick": 'Q', "End": 'D', "View": 'V'},
		VerbChords: map[string]string{"Quick All": "alt+q"},
		VerbOrder:  []string{"Move", "View", "Aim", "Use", "Cast", "Quick", "Done"},
	}
	if _, err := os.Stat(filepath.Join(adv, adventure.SoundDir, SotsbAmigaMusic.Songs)); err == nil {
		d.AdLib, d.AmigaMusic = nil, SotsbAmigaMusic
	}
	return d
}

// StaticData is GameData without a folder, with both the AdLib and the Amiga
// music: the import writes it as data/gamedata.json (pool.GameDataFile), which
// keeps the music the folder's sound/ has.
func SotsbStaticData() game.Data {
	d := GameData("")
	d.AmigaMusic = SotsbAmigaMusic
	return d
}

// soundData is START.EXE's sound tables (gobox/audio API.md, sotsb flags).
//
//go:embed sotsb/sounddata.json
var soundData []byte

// spellEffects is what a spell does beyond the numbers of its table record,
// where the cast routine (spells1.go, spells2.go: fx...) fixes it.
var spellEffects = map[int]string{
	23: "Holds (paralyses) man-sized humanoids. Lasts 4 rounds +1 per level.", 49: "Holds (paralyses) man-sized humanoids. Lasts 2 rounds per level.", 94: "Holds (paralyses) creatures. Lasts 1 round per level.", // Hold: also what a former Energy Drain casts
	3: "Heals 1d8 hit points.", 4: "1d8 damage.", 9: "Fire damage equal to the caster's level.",
	14: "Raises charisma by 2d4.", 15: "Damage of ((level+1)/2)d4 + (level+1)/2.",
	20: "Electrical damage of level + 1d8.", 21: "Puts 4d4 hit dice of creatures to sleep, weakest first.",
	35: "Raises strength.", 36: "Heals to maximum hit points less 1d4.", 90: "Heals to maximum hit points less 1d4.",
	47: "Fire damage of 1d6 per caster level.", 51: "Electrical damage of 1d6 per caster level down a line of 7, bouncing.",
	56: "The target is diseased and left with 1d4 hit points.", 58: "Heals 2d8+1 hit points.",
	59: "Giant strength 21 for 1d4*10+40 rounds.", 60: "Electrical damage of 1d6+20 on the square, 20 along the line.",
	62: "Heals 2d4+2 hit points.", 63: "The caster becomes 1d4+1 years younger (not under 18).",
	64: "Fire damage of 1d6 for each of 1d3*2+1 levels.", 65: "Damage of 2d4+2.", 66: "Damage of 2d8+1.",
	71: "Heals 3d8+3 hit points.", 72: "Damage of 3d8+3.", 74: "Fire damage of 6d8.",
	87: "Cold damage of 3d10.", 92: "Cold damage of level + 1d4 per level.",
	115: "Fire damage of level + 1d6 per level, when the delay is out.",
}

var tipClassNames = map[byte]string{sotsbSpellClassCleric: "Cleric", sotsbSpellClassDruid: "Druid", sotsbSpellClassMage: "Mage"}

// spellTips is the tip of every named spell, written from its table record
// (tables.go spellTableData: class, level, range, amount, targeting, save,
// camp or combat) and spellEffects.
func sotsbSpellTips() map[int]string {
	m := map[int]string{}
	for id := 1; id < len(sotsbSpellNames); id++ {
		if spellName(id) == "" || strings.HasPrefix(spellName(id), "spell ") {
			continue
		}
		r := spellRecord(id)
		var p []string
		class := tipClassNames[r[spellFieldClass]]
		if class == "" {
			class = "Innate"
		}
		p = append(p, fmt.Sprintf("%s level %d.", class, r[spellFieldLevel]))
		switch rng := r[spellFieldRange]; {
		case rng == 0xFF:
			p = append(p, "Touch, needs a hit roll.")
		case rng == 0 && r[spellFieldRangePerLevel] == 0 && r[spellFieldLasts] == 0:
			p = append(p, "Touch.")
		default:
			s := fmt.Sprintf("Range %d", max(int(rng), 1))
			if r[spellFieldRangePerLevel] > 0 {
				s += fmt.Sprintf(", +%d per 4 levels", r[spellFieldRangePerLevel])
			}
			p = append(p, s+".")
		}
		if n, area := spellAimTargets(id); area > 0 {
			p = append(p, fmt.Sprintf("Area of radius %d.", area))
		} else if n > 1 {
			p = append(p, fmt.Sprintf("Up to %d targets.", n))
		}
		if e, ok := spellEffects[id]; ok {
			p = append(p, e)
		} else if r[spellFieldAffect] > 0 && r[spellFieldAmount] > 0 {
			s := fmt.Sprintf("Lasts %d rounds", r[spellFieldAmount])
			if r[spellFieldAmountPerLvl] > 0 {
				s += fmt.Sprintf(" +%d per level", r[spellFieldAmountPerLvl])
			}
			p = append(p, s+".")
		}
		switch {
		case r[spellCombatField] == 0:
			p = append(p, "Camp only.")
		case r[spellFieldTargets] == 0:
			p = append(p, "Combat only.")
		}
		if t := r[spellCastTimeField]; t >= 3 {
			p = append(p, fmt.Sprintf("Takes %d round(s) to cast.", t/3))
		}
		switch r[spellFieldSaveMode] {
		case 0:
			p = append(p, "No save.")
		case 1:
			p = append(p, "Save negates.")
		default:
			p = append(p, "Save for half.")
		}
		m[id] = strings.Join(p, " ")
	}
	return m
}

// spellAimTargets is how many squares a combat spell is aimed at and the
// radius it hits round each (the targeting nibble of the record,
// combatflow's spellTargets, and its radius field).
func spellAimTargets(id int) (targets, area int) {
	r := spellRecord(id)
	k := int(r[spellTargetsField] & 0xF)
	switch {
	case k == spellTargetsCaster:
		return 0, 0
	case k >= spellTargetsAreaFirst && k <= spellTargetsAreaLast:
		return 1, k & 7
	case k == spellTargetsOneOrArea:
		return 1, int(r[spellAIRadiusField])
	case k == spellTargetsHitDice:
		return 1, int(r[spellAIRadiusField])
	}
	return k&spellTargetsCountMask + 1, 0
}

// The song events (EvSong N). titleSongN is the title's song: SOUND.DAX
// block 2 (block 1 is the driver) and SECRET.MX's song 0. The others are
// the Amiga's songs 1..6 only (DOS has no music for them: its AdLib file
// has no such blocks), at the moments the Amiga program starts them
// (Sound(14+n) calls, see songAt).
const (
	titleSongN    = 2
	songCombat    = 3 // SECRET.MX song 1
	songTreasure  = 4 // song 2
	songTraining  = 5 // song 3
	songPartyDead = 6 // song 4
	songPicture75 = 7 // song 5
	songPicture76 = 8 // song 6
)

// AmigaMusic is the Amiga's Music-X songs and DIG8 effects (Sound(cmd) at
// 0x41c2c: cmd 3..13 play sounds.glb entries 0..10, 14..20 songs 0..6).
// EffectOf takes PlaySound n to entry n-2, as the Amiga's cmd n+1 (its cmd
// 2 waits for the song): inferred, the effects' meanings were not traced.
var SotsbAmigaMusic = &game.AmigaMusic{Songs: "SECRET.MX", Effects: "sounds.glb",
	SongOf: map[int]int{titleSongN: 0, songCombat: 1, songTreasure: 2, songTraining: 3, songPartyDead: 4, songPicture75: 5, songPicture76: 6}, EffectOf: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}

// partyCommands are the game's own party commands.
var sotsbPartyCommands = []game.Option{{Label: "Add from Curse", Key: 'U', Do: game.AddCharacter{Roster: -1, From: "cotab"}}}
