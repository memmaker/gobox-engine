package classic

import (
	"image"
	"image/color"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// shown is a test's bag of what the screens read: what each reader of
// view_state.go says, and the builders' arguments in the table tests.
type shown struct {
	Mode                   game.Mode
	Pending                *game.Ask
	Party                  []game.Member
	Selected               int
	Where                  game.Where
	Story                  []game.Picture
	Picture                game.Picture
	Saves                  []game.SaveSlot
	LastText               string
	Purse                  map[string]int
	Quests                 []game.Quest
	Map                    game.Automap
	Combat                 *game.Combat
	Shop                   *game.Shop
	Temple                 []game.Service
	Training               *game.Training
	Loot                   *game.Loot
	Door                   *game.Door
	Vault                  *game.Vault
	Creation               *game.Creation
	Allowed                map[string]bool
	RestHeals, RestStudies bool
	TempleOffers           []game.Offer
	PartyCap               int
	Bestiary               []game.Beast
	Picker                 *game.Picker
	IconEditor             *game.IconEditor
	IconPalette            []color.RGBA
	Page                   *game.Page
	Flags                  *game.Flags
	Walls                  *game.Walls
	Intro                  *game.Intro
	Sequence               string
	Note                   game.Note
	Results                *game.Results
	Read                   *game.Read
	Cells                  image.Point
	Ink                    [16]color.RGBA
	Pacing                 *game.Pacing
	Window                 game.Window
	Holding                game.Hold
	Lore                   []game.Lore
	PowerPick              *game.PowerPick
	PoolItems              int
	Sheet                  *game.Sheet
	Items                  game.ItemsUp
	SheetPage              *game.SheetPage
	Portrait               image.Rectangle
	IconBox                image.Rectangle
	Screen                 *game.Screen
}

// shownOf reads every reader once, as Game.View does.
func shownOf(r *rules) shown {
	ps := r.g.State()
	c := commonOf(ps)
	party := partyOf(ps)
	s := shown{Mode: r.modeOf(ps), Pending: r.pending(ps), Party: party, Selected: r.selectedOf(ps), Where: c.Where,
		Story: storyOf(c), Purse: c.Purse, Quests: c.Quests, Map: r.automapOf(ps), Combat: r.combatOf(ps), Shop: shopOf(ps),
		Training: trainingOf(ps), Loot: lootOf(ps), Door: doorOf(ps), Creation: creationOf(ps), Allowed: allowedOf(ps),
		PartyCap: c.PartyCap, Picker: c.Picker, IconEditor: c.IconEditor, Page: r.pageOf(ps), Flags: r.flagsOf(ps, party),
		Walls: r.wallsOf(ps), Intro: introOf(ps), Sequence: sequenceOf(ps), Note: r.note, Results: resultsOf(ps), Read: r.readOf(ps),
		Pacing: r.pacingOf(ps), Window: windowOf(ps), Holding: holding(ps), Lore: c.Lore, PowerPick: c.PowerPick, PoolItems: c.PoolItems,
		Sheet: r.sheetOf(c), Items: r.itemsUp(ps), SheetPage: r.sheetPageOf(ps, party), Screen: screenOf(ps)}
	s.Temple, s.TempleOffers = templeOf(ps)
	s.Saves = r.saves()
	if c.Picture != nil {
		s.Picture = *c.Picture
	}
	s.RestHeals, s.RestStudies = restOf(ps)
	if len(c.Text) > 0 {
		s.LastText = c.Text[len(c.Text)-1]
	}
	switch t := ps.(type) {
	case play.Scene:
		if len(t.Text) > 0 {
			s.LastText = t.Text[len(t.Text)-1]
		}
	case play.Over:
		s.LastText = t.Text
	case play.Trading:
		if t.Kind == play.TradeVault {
			if s.Vault = t.Vault; s.Vault == nil {
				s.Vault = &game.Vault{}
			}
		}
	}
	return s
}

// menusOf is the menus and question s puts up, as Game.View builds them.
func menusOf(s shown) front.View {
	var v front.View
	v.MainMenu, v.Hall = PartyMenu(s.Mode, s.Training, s.Pending, s.Where.Hall)
	return askOver(v, s.Mode, s.Pending, s.Loot != nil || s.Training != nil)
}
