package ui

import (
	"image"
	"image/color"

	"gobox/game"
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
	Page                   *TextPage
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
func shownOf(r *frontState) shown {
	ps := r.g.State()
	c := game.CommonOf(ps)
	party := partyOf(ps)
	s := shown{Mode: r.modeOf(ps), Pending: r.pending(ps), Party: party, Selected: r.selectedOf(ps), Where: c.Where,
		Story: storyOf(c), Purse: c.Purse, Quests: c.Quests, Map: r.derived().automap, Combat: r.combatOf(ps), Shop: shopOf(ps),
		Training: trainingOf(ps), Loot: lootOf(ps), Door: doorOf(ps), Creation: creationOf(ps), Allowed: allowedOf(ps),
		PartyCap: c.PartyCap, Picker: c.Picker, IconEditor: c.IconEditor, Page: r.pageOf(ps), Flags: r.flagsOf(ps, party),
		Walls: r.derived().walls, Intro: introOf(ps), Sequence: sequenceOf(ps), Note: r.note, Results: resultsOf(ps), Read: r.readOf(ps),
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
	case game.Scene:
		if len(t.Text) > 0 {
			s.LastText = t.Text[len(t.Text)-1]
		}
	case game.Over:
		s.LastText = t.Text
	case game.Trading:
		if t.Kind == game.TradeVault {
			if s.Vault = t.Vault; s.Vault == nil {
				s.Vault = &game.Vault{}
			}
		}
	}
	return s
}

// A test's names for what the front does of its own, sent as send takes them.
type (
	tAnswer struct {
		N    int
		Text string
		At   image.Point
	}
	tWalkTo       struct{ At image.Point }
	tRest         struct{ Minutes int }
	tSetSpeed     struct{ N int }
	tCloseItems   struct{ All bool }
	tOpenSheet    struct{ Member int }
	tCloseSheet   struct{}
	tShowMember   struct{ Member int }
	tTitlePicture struct{ N int }
	tTitleCredits struct{ Page int }
	tTitleDone    struct{}
	tSequenceDone struct{ Name string }
	tStopWalk     struct{}
	tSave         struct {
		Slot int
		Name string
	}
	tLoad struct {
		Slot int
		From string
	}
	tBrowse struct{ Member int }
	tRoll   struct {
		Minimal  bool
		Minimums string
	}
	tItems struct{}
)

// send is d, a game.Action or one of the above, as the front does it.
func (r *frontState) send(d any) game.Result {
	switch d := d.(type) {
	case game.Action:
		return r.apply(d)
	case tAnswer:
		ps := r.g.State()
		if d.At != (image.Point{}) {
			return r.answer(ps, game.Reply{N: d.N, Text: d.Text}, d.At)
		}
		return r.apply(game.Reply{N: d.N, Text: d.Text})
	case tWalkTo:
		return r.walkTo(r.g.State(), d.At)
	case tRest:
		return r.restFor(d.Minutes)
	case tSetSpeed:
		r.setSpeed(d.N)
	case tCloseItems:
		return r.closeItems(d.All)
	case tOpenSheet:
		return r.openSheet(d.Member)
	case tCloseSheet:
		return r.closeSheet()
	case tShowMember:
		return r.showMember(d.Member)
	case tTitlePicture:
		return r.titlePicture(d.N)
	case tTitleCredits:
		return r.titleCredits(d.Page)
	case tTitleDone:
		return r.titleDone()
	case tSequenceDone:
		return r.sequenceDone()
	case tStopWalk:
		return r.stopWalk()
	case tSave:
		return r.saveNamed(d.Name)
	case tLoad:
		return r.loadSlot(d.Slot)
	case tBrowse, tItems:
		return r.openItems(-1)
	case tRoll:
		return r.apply(game.Reroll{Minimal: d.Minimal})
	}
	return game.Result{}
}

// holding is what the old front waits on in ps.
func holding(ps game.State) game.Hold {
	switch t := ps.(type) {
	case game.Exploring:
		if t.Door != nil && t.Ask == nil {
			return game.HoldDoor
		}
	case game.Trading:
		if t.Kind == game.TradeTemple || t.Kind == game.TradeShop && len(t.Services) > 0 {
			return game.HoldTemple
		}
		if t.Kind == game.TradeShop || t.Kind == game.TradeVault {
			return game.HoldShop
		}
	case game.Looting:
		return game.HoldLoot
	case game.Fighting:
		return game.HoldFight
	}
	return game.HoldNone // a story page waits on its question, not a hold (HoldStory is the finale's pictures)
}
