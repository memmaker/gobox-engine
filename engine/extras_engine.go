package engine

import (
	"gobox/game"
	"gobox/rules"
	"time"
)

// areaKey identifies the current map: data set and GEO block (New Phlan's
// three scripts share one map).
func (e *Engine) areaKey() int {
	return int(e.set)<<16 | int(e.world.Geo)
}

// markSeen records the party's cell as visited and the cells in view
// (straight ahead until a wall, up to 3) as seen.
func (e *Engine) markSeen() {
	if e.viewMode != 1 {
		return
	}
	if e.ext.seen == nil {
		e.ext.seen = map[int]*[256]uint8{}
	}
	k := e.areaKey()
	m := e.ext.seen[k]
	if m == nil {
		m = &[256]uint8{}
		e.ext.seen[k] = m
	}
	x, y := e.posX, e.posY
	if !rules.OnMap(x, y) {
		return
	}
	m[y<<4|x] = seenVisited
	for i := 0; i < 3; i++ {
		if e.WallAt(x, y, e.facing) != 0 {
			break
		}
		x, y = x+rules.StepX[e.facing], y+rules.StepY[e.facing]
		if !rules.OnMap(x, y) {
			break
		}
		if m[y<<4|x] == 0 {
			m[y<<4|x] = seenSeen
		}
	}
}

// mapView: the party is in the 3D view out of camp, where the game's Area
// command could show a map (in camp the panel keeps the party). Whether
// the area allows one (noMap) is gobox's "No Map Zones" option's call.
func (e *Engine) mapView() bool { return e.viewMode == 1 && e.phase != rules.PhaseCamp }

// statusName = DS:1047 + 0xD*n, the status words. extra: dead with the
// poison still in (0x37, Slow Poison can still raise) reads "Poisoned".
func (e *Engine) statusName(cID rules.CharID) string {
	return e.word(cID, rules.FactStatus)
}

// viewDrawn runs after the 3D view is drawn: remember what the party saw
// and keep the minimap current.
func (e *Engine) viewDrawn() {
	e.markSeen()
}

const (
	seenSeen    = 1
	seenVisited = 2
)

// extras holds the convenience-feature state.
type extras struct {
	hall     bool                    // the party menu is the training hall's (opProgram 0), not the title's
	seen     map[int]*[256]uint8     // per map (areaKey): 1 seen, 2 visited
	voiced   map[string]bool         // voice.go: the recordings this save has heard: the lore found (lore.go)
	again    map[string]bool         // voice.go: heard, to be spoken again (Rolf's history)
	met      map[string]metIcon      // bestiary.go: the monster kinds this save has met
	quests   []questNote             // quests.go: the quests this save was given
	notes    []noteRec               // lore.go: the journal's notes
	doors    map[int][]doorSide      // per map (areaKey): the doors opened for good (openDoor)
	secrets  map[int][]doorSide      // per map (areaKey): the secret doors found (foundSecret)
	beasts   []Beast                 // bestiary.go: their entries, as last published
	art      map[string]game.Picture // gamedata.go: the overland's art, as last built
	artKey   artKey                  // ... from these pictures
	creating bool                    // CreateCharacter's sheet is up: no level, experience, party items or status on it
	inGame   bool                    // slots.go: GameLoop has started: saving allowed
	aoe      int                     // aoe.go: radius+1 of the spell being aimed, 0 none
	aoeSpell int                     // aoe.go: the spell being aimed

	// nav.go: keys and mouse
	tableFaults []TableFault // script tables read past their field (tableFault), each once
	enemiesOnly bool         // aiming a weapon or hostile spell: Tab cycles through foes only (playerAim)

	// fullmap.go
	labels map[int]map[int]mapLabel // per area: square -> label
	pq     *playQuestion            // a play action's question up (playask.go)
	read   *Read                    // the journal text being read (Read)
	levels map[string]slotLevelAt   // the slots' levels as last read (slotLevelOf)
	screen *game.Screen             // a typed screen over the game (a ship's record outside a fight) until its key
	ending bool                     // the manifest's End is up as a title-like scene (title.end) until Continue
	endRun bool                     // the run ends with it (pod: the save, then the end); else the script goes on
	intro  bool                     // the title is up (Intro) until TitleDone
	door   *doorWait                // game.go: the locked door waiting (PendDoor)
	back   int                      // how far (0 2 4 6) round from the facing the next step goes: a step back or aside (moveDir)

	// interact.go: Explicit
	stopAt      map[int]bool // the area script's instructions where a harmless event starts
	stepping    bool         // a step's events run: they stop at stopAt
	interact    *[2]int      // the square whose event waits for RunEvent (nil: none)
	doorThrough bool         // a locked door's answer: the party goes through (lockedDoor)
	interacting bool         // RunEvent was taken: afterCommand runs the event
	pre         *preStep     // gobox's RunEvent{Before}: the pre-step script for the next Arrive (preStep)
	arriving    bool         // gobox's Arrive: the step stops after the move (stepDone)
	owed        *[2]int      // the square whose step events (events[1]) wait for gobox's RunEvent

	// fonts.go
	redraws    []redrawKind // what draws the screen up again at the page's size (resizePage), the last on top; none: it stays as drawn until its next draw
	note       string       // shown over the picture for a moment
	opened     int          // slots opened so far (opens): a search that opened none found nothing
	explore    bool         // the adventuring commands wait (GameLoop), under what opens over them
	exploreSeq int          // when exploring opened (top)
	noteAt     time.Time

	// journal.go
	printed  string       // the text window's text since it was cleared
	log      ledger       // the battle's log (battlelog.go)
	narrated string       // the rules' last line (Narrated), More joined on, as the note says it
	noteSent time.Time    // the note's noteAt when result last sent it
	panel    rules.CharID // whose line the combat status row shows
}

// preStep is a step's pre-step script run ahead of its Arrive.
type preStep struct {
	x, y, side int    // the square and the way (0 2 4 6) of the step
	ran        bool   // the script ran
	hold       uint16 // its HoldMove, for the step
}
