package rules1e

import (
	"fmt"
	"slices"
)

// Item powers (+0x3E above 0x7F) are an item's property code, 0x80 + N in
// every game, a code space of its own: each later game's itemPowers map
// (rules.json, written by its importer) turns its codes into the unified
// powers, never through the affect map. POR's codes are the unified powers
// (its 0x80.. affect handlers); a code that means what POR's means takes
// POR's, its game's handler a variant (affectVariants). The later games'
// codes POR has no meaning for have the powers below, decoded from each
// game's itemReadyEffect, statAffects and classcalc.
const (
	itemPowerWizardry        AffectType = 0x14F // 129: ring of wizardry doubles magic-user slots (cotab the first three levels, sotsb/pod the fifth)
	itemPowerFreeActionDex   AffectType = 0x150 // cotab 130: free action (thief skills as level 4, +10) and dexterity +4/+2/+1
	itemPowerGiantStrength   AffectType = 0x151 // 133: strength by +0x3D: 0 18/00, 1..6 19..24
	itemPowerConstitution    AffectType = 0x152 // cotab 134: constitution +1, charisma -1, the dwarf's and halfling's save bonus
	itemPowerScoreUp         AffectType = 0x153 // cotab 136: the score +0x3D names +1 (under 18)
	itemPowerDexterityDown   AffectType = 0x154 // cotab 138: dexterity -2
	itemPowerThievery        AffectType = 0x155 // cotab 139: thief skills as at least level 5 / 7
	itemPowerLowIntelligence AffectType = 0x156 // cotab 140: intelligence 7
	itemPowerLowStrengthMind AffectType = 0x157 // cotab 141: strength and intelligence 3
	itemPowerNoEffect92Pod   AffectType = 0x158 // pod 220: no readied effect (itemReadyEffect and statAffects have no case 92); uncursing grants +0x3D
	itemPowerNoEffect104Pod  AffectType = 0x159 // pod 232: as 220 (case 104)
)

// PowerMap is a game's item power codes as unified item powers: rules.json
// itemPowers, written by the game's importer.
type PowerMap map[byte]ItemPower

var (
	CotabItemPowers = PowerMap{
		0x80: ItemPower(AffItem0),
		0x81: ItemPower(itemPowerWizardry),
		0x82: ItemPower(itemPowerFreeActionDex),
		0x84: ItemPower(affCursedAlignment),
		0x85: ItemPower(itemPowerGiantStrength),
		0x86: ItemPower(itemPowerConstitution),
		0x88: ItemPower(itemPowerScoreUp),
		0x89: ItemPower(affEndsSpiritualHammer),
		0x8A: ItemPower(itemPowerDexterityDown),
		0x8B: ItemPower(itemPowerThievery),
		0x8C: ItemPower(itemPowerLowIntelligence),
		0x8D: ItemPower(itemPowerLowStrengthMind),
	}
	SotsbItemPowers = PowerMap{
		0x80: ItemPower(AffItem0),
		0x81: ItemPower(itemPowerWizardry),
		0x83: ItemPower(affGiantStrengthGirdle),
		0x84: ItemPower(affCursedAlignment),
		0x85: ItemPower(itemPowerGiantStrength),
		0x89: ItemPower(affEndsSpiritualHammer),
	}
	PodItemPowers = PowerMap{
		0x80: ItemPower(AffItem0),
		0x81: ItemPower(itemPowerWizardry),
		0x83: ItemPower(affGiantStrengthGirdle),
		0x84: ItemPower(affCursedAlignment),
		0x85: ItemPower(itemPowerGiantStrength),
		0x89: ItemPower(affEndsSpiritualHammer),
		220:  ItemPower(itemPowerNoEffect92Pod),
		232:  ItemPower(itemPowerNoEffect104Pod),
	}
)

// powerID is the game's item power code b as a unified power; nil (POR): b is the power.
func (t *Tables) powerID(b byte) (ItemPower, error) {
	if t.PowerIDs == nil {
		return ItemPower(b), nil
	}
	p, ok := t.PowerIDs[b]
	if !ok {
		return 0, fmt.Errorf("rules1e: item power %d is not in the game's itemPowers", b)
	}
	return p, nil
}

// gamePower is powerID's inverse.
func (t *Tables) gamePower(p ItemPower) (byte, error) {
	if t.PowerIDs == nil {
		return byte(p), nil
	}
	for b, v := range t.PowerIDs {
		if v == p {
			return b, nil
		}
	}
	return 0, fmt.Errorf("rules1e: item power %#x has no code in the game", p)
}

// GamePower is an item power code as a game's data has it (before its map).
type GamePower uint8

// Missing is the power codes of items (a game's ITEM and MON ITM records)
// that m lacks: a game's data must have none.
func (m PowerMap) Missing(items []ItemFile) []GamePower {
	var out []GamePower
	for _, it := range items {
		b := byte(it.Spells[2])
		if _, ok := m[b]; b > 0x7F && !ok && !slices.Contains(out, GamePower(b)) {
			out = append(out, GamePower(b))
		}
	}
	return out
}
