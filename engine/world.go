package engine

import (
	"encoding/json"
	"image"

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
			MaxDist, HoldMove, Search, LeftMap, Portrait, Temple, Shop, SymbolSet                                         uint16
			Morale, Surprise, MonsterHit, PartyHit, PartyMove, NoLoot, Duel, MonsterDist, CombatEnd, Beaten, AttackedAlly uint16
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
		Morale: g.Morale, Surprise: g.Surprise, MonsterHit: g.MonsterHit, PartyHit: g.PartyHit, PartyMove: g.PartyMove,
		NoLoot: g.NoLoot, Duel: g.Duel, MonsterDist: g.MonsterDist, CombatEnd: g.CombatEnd, Beaten: g.Beaten,
		AttackedAlly: g.AttackedAlly,
	}
}

// The fight's words as the scripts see them.
const (
	moraleCap     = 100  // a fight's start caps the Morale word (ovr11:1E31)
	combatEndDied = 0x80 // CombatEnd: the party died (ovr06:1527)
	combatEndFled = 0x81 // CombatEnd: the party fled
)

// surpriseBits are the Surprise word's bits: bit side+1 set means that side is surprised.
var surpriseBits = [...]struct {
	side rules.Side
	bit  uint16
}{{rules.SideParty, 1 << 0}, {rules.SideMonsters, 1 << 1}}

// fightTerms are the world's words a fight reads, in the one form every fight call gets them
// (the ruleset keeps what it needs from BeginFight's and EndFight's): the to-hit and movement
// bonuses are the words' signed low bytes.
func (e *Engine) fightTerms() rules.FightTerms {
	w := &e.world
	t := rules.FightTerms{
		Morale:       int32(w.Morale),
		MonsterHit:   int32(int8(byte(w.MonsterHit))),
		PartyHit:     int32(int8(byte(w.PartyHit))),
		PartyMove:    int32(int8(byte(w.PartyMove))),
		Duel:         e.duel(),
		DuelOpponent: int32(w.Duel),
		NoItemLoot:   w.NoLoot == 1,
		MonsterDist:  int32(w.MonsterDist),
		PartyAt:      image.Pt(e.posX, e.posY),
		PartyFacing:  rules.Facing(e.facing),
	}
	for _, s := range surpriseBits {
		if byte(w.Surprise)&byte(s.bit) != 0 { // only the low byte is tested
			t.Surprised = append(t.Surprised, s.side)
		}
	}
	return t
}

// fight is a fight call with no actor of its own; fightBy one acting for actor.
func (e *Engine) fight() rules.Fight { return rules.Fight{Terms: e.fightTerms()} }

func (e *Engine) fightBy(actor rules.CharID) rules.Fight {
	f := e.fight()
	f.Actor = actor
	return f
}
