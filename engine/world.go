package engine

import (
	"encoding/json"

	"gobox/adventure"
	"gobox/rules"
)

// savedWorld is the world as a save records it: the running area (after an
// area change the live AreaKind still names the one left; the load reloads
// the running script) in the kind word.
func (e *Engine) savedWorld() *rules.World {
	w := e.world
	w.AreaKind = uint16(e.gs.areaKind)
	return &w
}

// loadWorld takes the world a save holds; a save older than the engine's
// world has none and keeps its words in the ruleset's part.
func (e *Engine) loadWorld(g adventure.Save) {
	if g.World != nil {
		e.world = *g.World
		return
	}
	e.world = worldInOldBlob(g.Rules)
}

// worldInOldBlob reads the world words out of the ruleset's blob of a save
// written before the engine kept them (rules1e's SaveBlob: the area record's
// and the game record's words under their field names); no blob has none.
func worldInOldBlob(blob json.RawMessage) rules.World {
	if len(blob) == 0 {
		return rules.World{}
	}
	var old struct {
		Area struct {
			AreaKind, PrevX, PrevY, RiverOpen, Geo, WallSplitA, WallSplitB, NoCoords, MsgDelay, ViewFlags uint16
		}
		Game struct {
			MaxDist, HoldMove, Search, LeftMap, Portrait, Temple, Shop, SymbolSet uint16
		}
	}
	if err := json.Unmarshal(blob, &old); err != nil {
		panic("save: the ruleset's blob: " + err.Error()) // the ruleset's Load read the same bytes
	}
	a, g := old.Area, old.Game
	return rules.World{
		AreaKind: a.AreaKind, PrevX: a.PrevX, PrevY: a.PrevY, RiverOpen: a.RiverOpen, Geo: a.Geo,
		WallSplitA: a.WallSplitA, WallSplitB: a.WallSplitB, NoCoords: a.NoCoords, MsgDelay: a.MsgDelay, ViewFlags: a.ViewFlags,
		MaxDist: g.MaxDist, HoldMove: g.HoldMove, Search: g.Search, LeftMap: g.LeftMap,
		Portrait: g.Portrait, Temple: g.Temple, Shop: g.Shop, SymbolSet: g.SymbolSet,
	}
}
