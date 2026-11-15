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

// loadWorld takes the world a save holds. A word the ruleset's part still
// carries wins: the save was written while the ruleset kept that word (the
// words moved to the engine in steps; once one moved, no ruleset part has it).
func (e *Engine) loadWorld(g adventure.Save) {
	e.world = rules.World{}
	if g.World != nil {
		e.world = *g.World
	}
	wordsInOldBlob(g.Rules, &e.world)
}

// wordsInOldBlob puts into w the world words the ruleset's blob of an older
// save holds (rules1e's SaveBlob: the area record's and the game record's
// words under their field names, which are the cells' names).
func wordsInOldBlob(blob json.RawMessage, w *rules.World) {
	if len(blob) == 0 {
		return
	}
	var old struct{ Area, Game map[string]json.RawMessage }
	if err := json.Unmarshal(blob, &old); err != nil {
		panic("save: the ruleset's blob: " + err.Error()) // the ruleset's Load read the same bytes
	}
	for name, cell := range w.Cells() {
		for _, record := range []map[string]json.RawMessage{old.Area, old.Game} {
			if raw, ok := record[name]; ok {
				if err := json.Unmarshal(raw, cell); err != nil {
					panic("save: the ruleset's " + name + ": " + err.Error())
				}
			}
		}
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
		NoMagic:      e.noMagic(),
		In3D:         e.in3D(),
		PartyAt:      image.Pt(e.posX, e.posY),
		PartyFacing:  rules.Facing(e.facing),
		Approach:     rules.Approach(e.run.monsterSides),
	}
	if p, _ := e.placeTraits(); p.Has(rules.TraitWilderness) {
		t.Area = rules.AreaWilderness
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
