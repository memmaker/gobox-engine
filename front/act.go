package front

// Actions: the remastered screens act on the game through typed actions
// instead of keys; the game publishes what each screen shows and takes the
// decisions (por: engine.Apply).

// ActKind is what an Action does.
type ActKind int

const (
	ActNone       ActKind = iota
	ActClose              // leave the screen
	ActMember             // show party member N
	ActLead               // make party member N the selected one (the command bar's member): F1-F8, the member keys; in a member pick the candidate
	ActReady              // inventory: ready or unready item N
	ActUse                // inventory: use item N
	ActRanged             // combat: throw the Missiles slot's weapon (aims as Attack)
	ActDrop               // inventory: drop item N (the player already confirmed)
	ActSell               // inventory: sell item N (confirmed)
	ActWalk               // combat: walk to square X, Y (one of Battle.Reach)
	ActAttack             // combat: attack unit N (one of Battle.Targets; 0: aim first)
	ActGuard              // combat: end the turn on guard
	ActBandage            // combat: bandage the first dying member
	ActWait               // combat: Done's Delay
	ActEnd                // combat: Done's Quit: end the turn
	ActQuick              // combat: the computer fights for the member (N 1: the whole party)
	ActTurn               // combat: turn undead
	ActCast               // cast spell N (one of Turn.Spells in combat, of View.Ask (AskSpell) elsewhere)
	ActItems              // combat: the inventory
	ActAim                // combat, aiming: take square X, Y
	ActPoint              // the cursor is on square X, Y: combat's aim (the preview follows), the full map's (View.MapCursor)
	ActYes                // yes to View.Ask (AskConfirm)
	ActNo                 // no to View.Ask (AskConfirm); ActClose too
	ActFlee               // combat: step off the map's edge and try to run away (Turn offers it at the edge)
	ActChoose             // choose option N of View.Ask (X: the line of its List)
	ActOK                 // Enter: the highlighted choice, or "press a key"
	ActText               // the line being typed (View.Ask) now reads Text
	ActInventory          // member N's items (a card's left click)
	ActSheet              // the inventory: back to the sheet under it
	ActMinimap            // the minimap over the roster (Tab)
	ActMap                // the full map (M)
	ActOverview           // the party overview, on the View page the items (P)
	ActHelp               // the help pages (?)
	ActFont               // the next text font (N -1: the one before)
	ActSound              // sound on or off (Ctrl-S)
	ActResize             // the window shows X by Y text cells (40x25 at least): the page takes that size and the screen draws itself again
	ActSpells             // inventory: the member's memorised spells
	ActMagic              // inventory: the spell effects on the member
	ActHeal               // inventory: a paladin lays on hands
	ActCure               // inventory: a paladin cures disease
	ActMainMenu           // drop the game in play for the party menu (the pause menu's Back to Main Menu)
	ActBuy                // shop inventory: buy ware N (Inventory.Wares)
	ActTakeAll            // loot inventory: take every piece of treasure (Inventory.Loot)
	ActHead               // member N's next paper-doll head (N < 0: the selected member); Shift-click on the sheet's portrait
	ActPortrait           // the PNG at Text is member N's portrait (N < 0: the selected member); a file dropped on the sheet
	ActIcon               // the PNG at Text is the member's combat icon; a file dropped on the icon editor (View.IconDrop)
	ActOrder              // member N takes place X in the party (View.Ordering): a card dragged onto another's
	ActNote               // the text pane's text goes into the journal's notes (the text box's "> Notes")
	ActNoteRemove         // the journal's note N (Lore.ID) is removed

	// Adventuring decisions (View.Exploring)
	ActForward     // a step ahead (3D view)
	ActBackward    // a step back, facing kept (3D view)
	ActTurnLeft    // face left (3D view)
	ActTurnRight   // face right (3D view)
	ActStrafeLeft  // a step left, facing kept (3D view)
	ActStrafeRight // a step right, facing kept (3D view)
	ActGo          // a step towards Dir N (North..West) on the area map
	ActCamp        // make camp
	ActSearch      // search this square once
	ActSweep       // the slow, searching pace on or off
	ActCastPick    // member N (< 0: the selected one) picks a memorised spell to cast (View.Ask (AskSpell))

	// Camp decisions (View.Camp); N is the member (< 0: the selected one)
	ActRest       // rest: N is RestFix .. RestUntil
	ActPlan       // member N memorizes spell X on every rest Y (+1 or -1) times more
	ActScribe     // member N scribes spell X from a scroll on the next rest, or no longer
	ActDropMember // member N leaves the party (the player confirmed)
	ActIconEdit   // member N's combat icon editor

	// Temple, shop and training decisions (View.Temple, View.Shop, View.Training)
	ActKeep // keep (N 1) or sell (N 0) the one appraised

	// Party menu decisions (View.MainMenu); N is the member (< 0: the
	// selected one) where one is meant; ActMember views, ActDropMember drops
	ActCreate     // create a character
	ActModify     // member N's sheet to change
	ActAdd        // add saved characters to the party
	ActForget     // forget a saved character
	ActRemove     // member N goes back to the saved characters
	ActTraining   // the training hall for member N
	ActNewClass   // (extra) member N, a human, changes class (dualclass.go)
	ActLoad       // load a saved game
	ActQuickStart // a ready-made party
	ActBegin      // begin adventuring
	ActQuit       // quit the game (the player confirmed)

	// Creation and Modify decisions (View.Creation); the pickers and the
	// icon editor are the presenter's (OpenPicker, EditIcon)
	ActRoll     // roll the abilities (N 1: until the minimums Text, "STR INT WIS DEX CON CHA", met)
	ActSetField // creation field N (FieldSex ..) takes its option X (Creation.Fields[N].Options)
	ActName     // the character's name is Text
	ActStat     // ability N is Text (STR as 18/xx)
	ActPage     // a pager (help, the journal) goes N (+1 or -1) on: a picture page scrolls first

	ActManeuver // combat: maneuver N (one of Turn.Maneuvers; the rule set's, combatrules.go)

	ActInteract // (extra) run the square's event that waits for it (View.Interact, interact.go)

	ActDir    // a direction key answers View.Ask (Ask.Dirs, AskMove): N is the keypad's 1..9 (8 up, 6 right ...)
	ActRaise  // AskModify: row N's value goes X (+1 or -1)
	ActImport // the party menu: import characters from another game's saves (a Data.Entries entry's Act)

	ActStep // combat: one step, N the keypad's direction 1..9 (the bar's Move button sends it with N 0)
	ActView // combat: the bar's View button (what it shows is the game's)

	ActAlter // camp: setting N (Camp.Settings) takes the value X; the game answers it

	ActTrade // inventory: trade item N to another member (offered with Inventory.Trade; the game asks whom, or refuses)
	ActHalve // inventory: split item N's stack in two (Inventory.Halve; the game answers or refuses)
	ActJoin  // inventory: join item N with the same items of the member (Inventory.Join; the game answers or refuses)

)

// Action is one thing the player did on a remastered screen.
type Action struct {
	Kind ActKind
	N    int
	X, Y int    // a combat map square, a text cell (ActSpot)
	Text string // ActText
}
