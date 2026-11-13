package rules1e

import (
	"embed"
	"encoding/json"
	"sync"
)

//go:embed treasure/*.json
var treasureFiles embed.FS

// treasureData reads a game's random treasure from treasure/<game>.json.
func treasureData(game string) func() *TreasureRolls {
	return sync.OnceValue(func() *TreasureRolls {
		b, err := treasureFiles.ReadFile("treasure/" + game + ".json")
		if err != nil {
			panic(err)
		}
		var t TreasureRolls
		if err := json.Unmarshal(b, &t); err != nil {
			panic("rules1e: treasure/" + game + ".json: " + err.Error())
		}
		return &t
	})
}

// The random treasure of each game, as data (rules.json randomTreasure is
// written from these): POR's (ovr22 RandomItemType and RandomItem), COTAB's
// (cotab-re money.pas GenerateRandomItem, eclvm.pas EclTreasure; its gems
// and jewels are POR's, so it has no Appraisal), SotSB's and PoD's
// (sotsb-re EclTreasure, GenerateRandomItem, randItemW, spellsByLevel).
var (
	PorTreasure   = treasureData("por")
	CotabTreasure = treasureData("cotab")
	SotsbTreasure = treasureData("sotsb")
)

// TreasureRolls is a game's random treasure (TREASURE $80+n): each item's
// type rolled, then the item made for it (rules.json randomTreasure; nil:
// POR's code, RandomItemType and RandomItem). SotSB and PoD: SotsbTreasure.
type TreasureRolls struct {
	Type         Roll[ItemType]                // the item type
	Bonus        Roll[byte]                    // a weapon's or armour's plus
	PlusBase     WordID                        // the name word of plus 0 (the plus word is PlusBase+plus)
	Kinds        map[ItemType]TreasureKind     // the types made with a plus
	Scrolls      map[ItemType]TreasureScroll   // the scrolls
	Magic        map[ItemType]Roll[TemplateAt] // the types made from a template alone
	Templates    []uint16                      // eight words a template: three name words, weight, price, three spells
	TemplateSave bool                          `json:",omitempty"` // a template item's Plus2 is its plus (the saveBonus the originals set)
	TemplatePlus map[TemplateAt]byte           `json:",omitempty"` // a template's plus when not 1
}

// TemplateAt is a template's first word in TreasureRolls.Templates, from 1
// (the originals' randItemW index); NoTemplate none.
type TemplateAt int

// NoTemplate: the item keeps what its kind made.
const NoTemplate TemplateAt = 0

// Roll is one die and what each result gives: the first case whose Hi the
// result is not above. Die 0 rolls nothing and gives Cases[0].
type Roll[V any] struct {
	Die   int
	Cases []RollCase[V]
}

// RollCase is a Roll's results up to Hi: Value, or Then rolled.
type RollCase[V any] struct {
	Hi     int
	Value  V        `json:",omitempty"`
	Then   *Roll[V] `json:",omitempty"`
	Reroll bool     `json:",omitempty"` // the original rolls the die again for this result
}

// rollOf rolls r.
func rollOf[V any](e *RuleState, r Roll[V]) V {
	if r.Die == 0 {
		return r.Cases[0].Value
	}
	for {
		d := e.Dice(1, r.Die)
		for _, c := range r.Cases {
			if d <= c.Hi {
				if c.Reroll {
					break
				}
				if c.Then != nil {
					return rollOf(e, *c.Then)
				}
				return c.Value
			}
		}
		if d > r.Cases[len(r.Cases)-1].Hi {
			var none V
			return none
		}
	}
}

// NoPlusWord: no name word shows the plus.
const NoPlusWord = -1

// TreasureKind is how an item of a plus is made.
type TreasureKind struct {
	Words     [3]WordID         // the name words
	PlusWord  int               // the word shown as PlusBase+plus (0..2), or NoPlusWord
	Plus      byte              `json:",omitempty"` // a fixed plus (the bonus rolled first all the same); 0: the rolled one
	SaveBonus bool              `json:",omitempty"` // the plus counts for saves (Plus2)
	Hidden    byte              // Item.Hidden
	Weight    uint16            //
	Count     byte              `json:",omitempty"`
	PriceMul  uint16            // the price a plus
	Magic     *Roll[TemplateAt] `json:",omitempty"` // a template that may replace it (rolled after the plus)
	PlusMul   byte              `json:",omitempty"` // the item's plus is the roll times PlusMul plus PlusAdd (0: the roll)
	PlusAdd   byte              `json:",omitempty"`
	PlusWords map[byte]WordID   `json:",omitempty"` // the plus word by the item's plus, in place of PlusBase+plus (none: 0)
}

// TreasureScroll is how a scroll is made: 1d3 spells, each of a level
// rolled 1..len(Levels), then a spell of that level's row.
type TreasureScroll struct {
	Words     [3]WordID   // the name words; Words[CountAt] is CountBase+the spell count
	CountAt   int         `json:",omitempty"`
	CountBase WordID      //
	Hidden    byte        //
	Weight    uint16      //
	LevelCost uint16      // the price a spell level
	Levels    [][]SpellID // the game's spell numbers of each level
}

// treasureItem = SotSB's GenerateRandomItem (money.pas): an item of type typ
// made as t says, as the game's file holds it, then converted to ids.
func (e *RuleState) treasureItem(t *TreasureRolls, typ ItemType) (*Item, error) {
	it := e.treasureRecord(t, typ)
	out, err := e.gameItems([]Item{it})
	if err != nil {
		return nil, err
	}
	return &out[0], nil
}

// treasureRecord is treasureItem's item as the game's file holds it.
func (e *RuleState) treasureRecord(t *TreasureRolls, typ ItemType) Item {
	it := Item{Type: typ, Hidden: 6}
	at := NoTemplate
	if k, ok := t.Kinds[typ]; ok {
		plus := rollOf(e, t.Bonus)
		if k.Plus != 0 {
			plus = k.Plus
		}
		if k.PlusMul != 0 {
			plus = plus*k.PlusMul + k.PlusAdd
		}
		it.Words, it.Plus = k.Words, plus
		if k.PlusWord != NoPlusWord {
			it.Words[k.PlusWord] = t.PlusBase + WordID(plus)
			if k.PlusWords != nil {
				it.Words[k.PlusWord] = k.PlusWords[plus]
			}
		}
		if k.Magic != nil {
			at = rollOf(e, *k.Magic)
		}
		if k.SaveBonus {
			it.Plus2 = plus
		}
		it.Hidden, it.Weight, it.Count, it.Price = k.Hidden, k.Weight, k.Count, uint16(plus)*k.PriceMul
	} else if s, ok := t.Scrolls[typ]; ok {
		n := e.Dice(1, 3)
		it.Words = s.Words
		it.Words[s.CountAt] = s.CountBase + WordID(n)
		it.Plus, it.Weight, it.Hidden = 1, s.Weight, s.Hidden
		for j := range n {
			lvl := e.Dice(1, len(s.Levels))
			row := s.Levels[lvl-1]
			it.Spells[j] = ItemPower(row[e.Dice(1, len(row))-1])
			it.Price += uint16(lvl) * s.LevelCost
		}
	} else if m, ok := t.Magic[typ]; ok {
		at = rollOf(e, m)
	}
	if at != NoTemplate {
		w := t.Templates[at-1 : at+7]
		for j := range 3 {
			it.Words[j] = WordID(byte(w[j]))
		}
		it.Plus, it.Plus2, it.Weight, it.Count, it.Price = 1, 0, w[3], 0, w[4]
		if p, ok := t.TemplatePlus[at]; ok {
			it.Plus = p
		}
		if t.TemplateSave {
			it.Plus2 = it.Plus
		}
		for j := range 3 {
			it.Spells[j] = ItemPower(byte(w[5+j]))
		}
	}
	return it
}

// makes: t makes items of type typ (nil: none).
func (t *TreasureRolls) makes(typ ItemType) bool {
	if t == nil {
		return false
	}
	_, k := t.Kinds[typ]
	_, s := t.Scrolls[typ]
	_, m := t.Magic[typ]
	return k || s || m
}

// treasureRolls is the game's random treasure; POR's when it has none.
func (e *RuleState) treasureRolls() *TreasureRolls {
	if t := e.T().Treasure; t != nil {
		return t
	}
	return PorTreasure()
}

// RandomItem is a random item of type typ made by the game's treasure rolls.
// It panics when the item table is not loaded (core.LoadItemTable first).
func (e *RuleState) RandomItem(typ ItemType) *Item {
	it, err := e.treasureItem(e.treasureRolls(), typ)
	if err != nil {
		panic(err)
	}
	return it
}

// devItem is a random item of POR's type t in this game (dev fixtures):
// the game's type of the same kind; false when the game has none.
func (e *RuleState) devItem(t ItemType) (*Item, bool) {
	g, ok := e.devType(t)
	if !ok {
		return nil, false
	}
	tr := e.treasureRolls()
	if e.T().Kinds == nil && !tr.makes(g) { // a type the game's rolls never make: POR's table
		tr = PorTreasure()
	}
	it, err := e.treasureItem(tr, g)
	return it, err == nil
}

// devType is this game's type of POR's type t's kind; false: none.
func (e *RuleState) devType(t ItemType) (ItemType, bool) {
	if e.T().Kinds == nil {
		return t, true
	}
	g := e.typeOf(ItemIDs[t])
	return g, g != 0
}

// devItems are devItem's of ts the game has.
func (e *RuleState) devItems(ts ...ItemType) []*Item {
	var out []*Item
	for _, t := range ts {
		if it, ok := e.devItem(t); ok {
			out = append(out, it)
		}
	}
	return out
}
