// Package engine is the Pool of Radiance game engine, ported routine by
// routine from the DOS START.EXE/GAME.OVR disassembly. It reads an imported
// adventure folder (New), never the original files.
package engine

import (
	"gobox/adventure"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"image"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"gobox/asset"
)

// Config is what POOL.CFG set (New keeps its defaults, E P F): graphics letter, sound letter, game dir, save dir,
// and the 5th line ('F' = full intro with second title picture + credits).
type Config struct {
	Graphics, Sound byte
	SaveDir         string
	Foreign         func(from string) ([]rules.CharacterFile, error) // another game's saved characters by its id (game.Data.Imports), as canonical files (char.ReadRoster); nil: none
	Param           string                                           // ParamStr(1): "STING" unlocked the original's 'J' debug key (not on the party menu's bar, so never picked)
}

// Engine holds the global state the DOS program kept in its data segment.
type Engine struct {
	tiles              map[tileKey]Picture    // the combat ground tiles converted (tilePix)
	tileGen            int                    // counts the in-place writes to the tile and icon frames (tilesChanged): a cache keyed by those images keys it too
	grounds            map[*asset.Image]uint8 // each combat tile set's ground colour (groundOf)
	newRules           Ruleset                // builds rules (New; bareRuleset when nil)
	sequence           string                 // the full-screen animation playing (adventure.SeqEnding; "": none)
	vault              *vault                 // the open vault (Vault)
	given              *adventure.GameTables  // the game's GameTables (New; bareData when nil)
	eclVersion         *ecl.Version           // the game's ECL dialect (initVM)
	rules              rules.Rules            // the ruleset; it owns the rules' state (state) (rulesetseam.go)
	create             rules.CreationRules    // the ruleset's creation, made with rules
	posX, posY, facing int                    // the party's map square and facing 0..7 (DS:6AAD, 6AAE, 6AAF): world navigation, the Engine's; saves keep them (files_engine.go)
	campSavedFlag      int                    // beginFlag as camp found it
	campSavedAnim      bool                   // the picture's animation as camp found it
	campInterrupted    bool                   // a rest was interrupted (camp closes)
	campSeq            int                    // when camp opened (top); the rules keep camp up (core.WordCampOn)
	fx                 []rules.Effect         // the effects queued and not yet shown (emit, present)
	resting            *restRun               // the camp rest going on (RestStep), nil none
	adv                string                 // New's adventure folder: scripts, art, data
	data               string                 // New's host data directory: AssetsDir, AmigaDir, SaveDir

	pics     asset.Pictures // the adventure's picture table (data/pictures.json): DAX picture -> collection entry
	arts     *asset.Art     // the imported art of data (artFolder)
	gameData game.Data      // the game's static data (GameDataFile)
	gd       *gameData      // the imported data of data (gdata)
	sounds   []byte         // the imported sound tables (SoundData)
	listed   []string       // the roster list migrateRoster last saw: each new list is numbered once (it reads every .cha)

	pal         asset.EGARegs // EGA palette registers (SetPaletteColor)
	Cfg         Config
	FixedSeed   *uint32                        // if set, Randomize uses it instead of the clock (-seed)
	Dasp        atomic.Bool                    // extra: the Dubtrain sound pack's effects (dasp.go), the Esc menu's "DASP Sfx"
	NoSteps     atomic.Bool                    // extra: no step and turn sound (sndStep), the Esc menu's "Step sounds" off
	Explicit    atomic.Bool                    // extra: a square's harmless event waits for RunEvent (interact.go), the Esc menu's "Explicit Interactions"
	QuickSearch atomic.Bool                    // option (Gameplay "Quick search", ruling U2, default off): a sweeping step or a Look takes 3 minutes instead of the original's turn
	FullIntro   atomic.Bool                    // the host's "Full intro" (its default the adventure's Start.Full): TITLE 2 with its lightning after TITLE 1 (DS:49D1)
	DaySky      atomic.Bool                    // option (Gameplay "Day sky by clock", ruling U3, default off): the scripts' hour < 14 sky check follows day 6:00-20:00 (isDay)
	handedIn    atomic.Pointer[rules.Settings] // the rules' settings from the frontend, taken at the next decision point (HandIn)
	Deskew      atomic.Bool                    // extra: dungeon combat maps on a square grid instead of the original's skewed perspective (combatsetup.go); read at each fight's setup

	sym   [5]*asset.Image // DS:5E3A symbol sets
	set   rules.Set       // DS:5376: the data set the current area reads
	start time.Time

	restLine  bool          // DS:6DDA (ovr21 rest): the rest time is on row 17, messages go below it
	restShown int           // steps since the rest time was last drawn (RestStep)
	tempSel   rules.CharID  // selectTemp: selected but not in the party
	temps     []tempSel     // the selections selectTemp replaced
	creation  *creation     // the creation sheet up (Creation)
	sel       int           // the selected member: whom the screens show and the player's commands name, handed to the rules as their actor (as)
	aiKey     rules.AutoKey // the key pressed for the next computer turn (takeKey): game.Manual, game.QuickMagic
	soundNow  bool          // PlayEventSound: the sound plays now, not as an event
	msgDelay  byte          // DS:49E2, message delay factor (ovr12 sets 1)

	rosterHi   rules.CharID // SelectCharacter: 00BA:0767 highlights this one (nil: sel)
	rosterNone bool         // ... or no one (who = nil)
	loadedGame bool         // DS:49F8

	engineFields       map[string]engineField // the script fields the engine gives a meaning (scriptvars.go)
	vars               fieldCells             // the other fields' cells: the game's saved state (DS:49D2/49D6/49DA records)
	locals             fieldCells             // and the area's scratch, cleared by a fresh area
	savedFields        []ecl.Field            // every area's saved fields (savecodec.go)
	loadedLocals       fieldCells             // a loaded game's locals until its area's script is in
	vm                 *ecl.VM                // the script machine (DS:49ED its PC, here an instruction number)
	host               eclHost                // what vm asks of the engine
	native             map[int]func(*Engine)  // the native scripts at instructions of script
	purseWho           map[int]bool           // the Who instructions of script that ask who spends the party purse (opWho)
	viewMode           int                    // DS:49FA (1 = 3D view)
	prevViewMode       int                    // DS:49F9
	sqrpaci            [2]*asset.Image
	pos3, pos4         byte   // DS:6AB0, 6AB1
	viewBit0, viewBit1 byte   // DS:49F5/49F6
	mapPairA, mapPairB [3]int // DS:6AAE+4i / 6AB0+4i
	// viewAnim = area+0x3FE: nonzero cycles the view picture on every poll
	// (ovr26:03BD), cleared by displayInput (ovr26:06D8). No script sets it;
	// in the original only save/load wrote it, packing the last wall slot
	// there (an original bug the port leaves out: nothing sets it).
	viewAnim       uint16
	loadingMonster bool // DS:557C

	es             eclState                   // ovr08/ovr04 interpreter state
	msgRow, msgCol int                        // DS:5E25 / 5E24 text-block cursor
	gs             geoState                   // ovr30 map, wall sets and view
	run            eclRun                     // ovr04 globals
	pic            animPic                    // DS:6ABE view-window picture
	port           portrait                   // DS:6B00.. HEAD/BODY portrait
	eclFile        string                     // DS:547C
	ext            extras                     // convenience features (extras.go)
	title          title                      // the adventure's title sequence (SetStart)
	snap           snapState                  // the View published for other presenters (snapshot.go)
	pg             pageState                  // what is on the screen (page.go)
	ac             afterCombat                // ovr06 globals
	k              []cont                     // what follows the flows under way (cont.go)
	resets         int                        // unwindReset calls: a continuation dropped the game
	cs             combatState                // the fight's rounds and turn (combatflow.go)
	ps             game.State                 // State: built after each change (rebuild)
	psDOS          bool                       // DOSArt when ps was built
	lootSeen       bool                       // the treasure's "You found some loot!" was acknowledged (Continue); cleared when the treasure goes
	StateBuilds    int                        // the States built (the frame tests count them)
	events         []game.Event               // the facts not yet handed out (Do's outcome)
	hitTo          int                        // the unit the last hit struck, its harm still to come (0 none): Attacked carries it
	looking, over  bool                       // GameLoop: in its look and camp loop; a script quit the game
	addFrom        string                     // the game whose characters the PickAddFrom picker lists
	picker         *picker                    // the grid picker up (Picker)
	iconEd         *iconEdit                  // the combat icon designed (EditIcon)
	valued         rules.ItemHandle           // an appraised gem or jewel, sold or kept next (game.Keep)
	pending        Pending                    // what the script's wait screen waits on (Pending)
	pw             waitData                   // what the wait screen keeps (pending.go)
	waitSeq        int                        // when the wait opened (top)
	readSeq        int                        // the last Read's Seq: up by one each new text, for the engine's life
	animInput      bool                       // animate the view picture while waiting for a key
	view           viewState                  // ovr20 View screen
	menu           partyMenu                  // ovr17:0161 the party menu
	icons          [iconSlots][2]*asset.Image // DS:5E5A, ready and action pose per slot
	spr            *asset.Image               // DS:6E36, 4 frame missile/hit sprite

	// combat overlays (ovr10/13/14/23)
	spellList      []*ListNode            // DS:4766 spell menu lines (ovr23)
	spellIDs       []byte                 // DS:476A spell of each non-header line
	spellItems     [0x31]rules.ItemHandle // DS:477F scroll of each line
	spellItemCount int                    // DS:4843
	eclWait        *eclWait               // the script instruction waiting for its answer (eclask.go)

	// H
	areaIconTick byte       // DS:6DDC area map party icon pose counter
	snd          sound      // PC speaker driver (08AE)
	vox          game.Voice // voice.go: the voice-over (nil: silent)

	entryTable map[ecl.EntryID]ecl.Entry    // the long texts by id (entries)
	books      map[string]*game.JournalBook // the game's journal books by file name (book); nil: none
	jtexts     map[string][]string          // journalTexts, once

	DevAction    func() // run once at the next adventuring command prompt (-dev)
	DevMonster   int    // the combat dev scene's monster (-fight; 0: the area's first)
	FreeTraining bool   // the hall trains everyone free (-freetrain: the debug flag, POR's STING 'J')
	Explore      bool   // -explore: no intro, a Quick Start party when there is none, begin adventuring at once
	DevPortrait  []byte // the PNG the portrait dev scenes put on the sheet (-png)
	ExploreSlot  string // with Explore: the save to start from (-slot) instead of a Quick Start
}

// New opens the adventure folder adv (its scripts, art and data, written by
// `gobox import por`) and the host's data directory (the port's own assets
// under AssetsDir, the Amiga sound and title under AmigaDir, the saves). It
// reads no original file: the config is the original's POOL.CFG defaults.
// rules builds the ruleset and d is the game's GameTables.
func New(ruleset Ruleset, d adventure.GameTables, adv, data string) (*Engine, error) {
	e := &Engine{newRules: ruleset, given: &d, adv: adv, data: data, pal: asset.EGAIdentity, set: 1, start: time.Now()}
	var err error
	e.initVM()
	e.Cfg = Config{Graphics: 'E', Sound: 'P', SaveDir: filepath.Join(data, adventure.SaveDir)}
	e.FullIntro.Store(true) // POOL.CFG's default (DS:49D1) until the host passes its Full intro option
	e.pic.block = -1
	if e.gameData, err = adventure.ReadGameData(adv); err != nil {
		return nil, err
	}
	e.port = portrait{headID: 0xFF, bodyID: 0xFF}
	pics, err := asset.ReadPictures(os.DirFS(adv))
	if err != nil {
		return nil, err
	}
	e.pics = pics
	if err := e.SetStart(nil); err != nil { // no title steps until the host passes the adventure's Start
		return nil, err
	}
	return e, nil
}

// Exit = 0899:0000: back to DOS. The engine never ends the process: it
// emits game.Ended and the front ends the run (as ProgramQuit does).
func (e *Engine) Exit() {
	e.event(game.Ended{})
}

// iconSlots is how many combat icon slots DS:5E5A holds: one per combatant
// the board takes (0x48, the rules' MaxCombatants).
const iconSlots = 0x48

// snapState is what the engine keeps of the screen it shows: the pages
// and boxes up, the fight's marks and reach, the art it took once.
type snapState struct {
	quiet    bool                         // a play action is running: no scene is published for the events
	invOn    bool                         // ItemsMenu waits: Inventory is up
	shop     bool                         // ... as the shop's Browse: Inventory.Wares
	identN   int                          // ... counted: a new reveal is a new number
	identIt  rules.ItemHandle             // ... the item itself (State's Shop.Identified)
	saves    bool                         // there are saved games (the party menu's Load)
	win      Window                       // set by whatever drew into the view window last
	heads    map[int]Picture              // HEAD blocks' pixels by id
	tiles    [3]Picture                   // DrawBorder style 0's corner, side and top cells, once taken (BorderPix)
	cmapOn   bool                         // the combat map is set up (cmapShown) and shown
	sheetOn  bool                         // a character sheet is up over the combat map
	portrait CellRect                     // the sheet's portrait box while a sheet is up
	sheet    *SheetPage                   // the character sheet's page while it is up (SheetPage)
	results  *Results                     // the after-combat page while it is up (Results)
	full     bool                         // a whole-screen picture is up (Full)
	iconDrop func([]byte)                 // the icon editor, or the creation sheet, takes a dropped PNG while it is up (nil elsewhere)
	iconBox  CellRect                     // the creation sheet's icon: the cells that take the drop (zero: the whole screen)
	dungeon  bool                         // the combat tiles are DungCom's stone walls (else WildCom's outdoors)
	marks    []image.Point                // squares markReach framed
	reach    []image.Point                // squares a click on the main combat menu walks to
	paths    map[image.Point][]rules.Step // the walk to each of them
	hits     map[int]int                  // each of targets' chance to be hit, in percent
	area     []image.Point                // squares aoeMarks covers
	skipped  []image.Point                // squares of allies in the area the spell passes over (Bless: an enemy next to them)
	aimWhy   string                       // aiming: why the aimed square cannot be taken ("" when it can)
	turn     *Turn                        // the player's turn waits in act
	targets  []int                        // combatants verbAttack takes; aiming: the foes to cycle through
	strike   Strike                       // the last weapon attack (From/To filled in by battle)
	striker  [2]rules.CharID              // its attacker and target
}
