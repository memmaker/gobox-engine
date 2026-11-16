package rulesxxvc

import (
	"fmt"
	"maps"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// BRSave is one original BR saved character (saveload.pas SaveChar: the
// TCharacter record, its .itm items and .spc effects) as the importer decodes
// it, typed; field names follow types.pas TCharacter. The byte layout stays
// with the importer.
type BRSave struct {
	Name                  string
	Abilities             [7]int // STR DEX CON INT WIS CHA TCH: the ones the rules read (charhlp.pas StrToHitAdjust..., combatai.pas)
	CurrentAbilities      [7]int // as they stand; only CharGen writes them
	Saves                 [8]int
	Gender                int // TGender: 0 male, 1 female
	Race                  int // TRace: RaceMonster, then xxvc.Races
	Career                int // TCareer: CareerNone, then xxvc.Careers
	Level                 int
	Age                   int // types.pas age
	Icon                  int // types.pas icon: the combat icon's body in CHARS.DAX
	CreditTotal           int64
	Experience            int64
	BaseCombat            int
	IconDimensions        int // types.pas iconDimensions
	BaseMovement          int
	BaseArmorClass        int
	BaseScale, Scale      int
	Unarmed               [2]rules.Dice // unarmedRolls, unarmedDice, unarmedModifiers
	HitPointMaximum       int
	ArmorFlags            int // C_Armor bits
	SkillLevels           [84]int
	XPAward               int64
	WeaponSpecializations [SpecializationSlots]int
	ScriptFlags           [ScriptFlagSlots]int // skillLevelsTail $A1..$A3, the bytes ECL keeps on the character
	Status                Status
	IsEnabled             bool
	Hostile               bool // side = C_Side.HOSTILE
	IsCompanion           bool
	NonPlayer             bool // npc's NON_PLAYER_FLAG
	NPCMorale             int  // npc's low seven bits
	AttackBonus           int
	ArmorClass            [2]int
	Current               [2]rules.Dice // currentRolls, currentDice, currentModifiers
	HitPoints             int
	CurrentMovement       int
	Encumbrance           int
	Items                 []BRSavedItem
	Effects               []Effect // TEffect: effectId, duration, strength, callsOnRemove
}

// BRSavedItem is one TItem of the .itm file: its TItemTemplate.
type BRSavedItem struct {
	ItemType int32
	NameIDs  [3]int32
	Plus     int32
	Quantity int32
	Readied  bool // C_Item.READIED
}

// C_Armor's armorFlags bits.
const (
	armorGogglesBit   = 0x08
	armorECMBit       = 0x10
	armorFieldsBit    = 0x20
	armorBodyBit      = 0x40
	armorBreathingBit = 0x80
)

// brGenders is TGender's order (types.pas: GenderMale, GenderFemale).
var brGenders = [...]rules.Sex{rules.Male, rules.Female}

// brStatOrder is TCharacter.abilities' order: STR DEX CON INT WIS CHA TCH.
var brStatOrder = [7]rules.Ability{rules.Strength, rules.Dexterity, rules.Constitution, rules.Intelligence,
	rules.Wisdom, rules.Charisma, xxvc.Tech}

// abilityNames are gdconsts.pas abilityAbbreviations (STR DEX CON INT WIS CHA TCH) with the names written out.
var abilityNames = map[rules.Ability]rules.AbilityName{
	rules.Strength: {Short: "STR", Long: "Strength"}, rules.Dexterity: {Short: "DEX", Long: "Dexterity"},
	rules.Constitution: {Short: "CON", Long: "Constitution"}, rules.Intelligence: {Short: "INT", Long: "Intelligence"},
	rules.Wisdom: {Short: "WIS", Long: "Wisdom"}, rules.Charisma: {Short: "CHA", Long: "Charisma"},
	xxvc.Tech: {Short: "TCH", Long: xxvc.TechLabel},
}

// ConvertBRCharacter turns one BR saved character, as the importer decoded
// it, into a typed Character. Each saved item becomes the template of
// t.Items with its type and name words (the lowest ItemID of equals), its own
// plus and quantity kept. The importer calls it once; the game never reads
// the original bytes.
func ConvertBRCharacter(t xxvc.Tables, b BRSave) (Character, error) {
	c := Character{
		Name: b.Name, Level: b.Level, Age: b.Age, Icon: b.Icon, Scores: map[rules.Ability]int{},
		HP: b.HitPoints, MaxHP: b.HitPointMaximum, BaseCombat: b.BaseCombat, IconDimensions: b.IconDimensions,
		Coins: rules.Copper(b.CreditTotal), Enabled: b.IsEnabled, Status: b.Status,
		Exp: b.Experience, XPAward: b.XPAward, AttackBonus: b.AttackBonus, ArmorClass: b.ArmorClass,
		BaseScale: b.BaseScale, Scale: b.Scale, Movement: b.CurrentMovement,
		BaseArmorClass: b.BaseArmorClass, BaseMovement: b.BaseMovement, Encumbrance: b.Encumbrance,
		Armor: armorFlagsOf(b.ArmorFlags), Side: rules.SideParty, Unarmed: b.Unarmed, Damage: b.Current,
		Effects: slices.Clone(b.Effects),
	}
	if b.Gender < 0 || b.Gender >= len(brGenders) {
		return Character{}, fmt.Errorf("ConvertBRCharacter: gender %d", b.Gender)
	}
	c.Sex = brGenders[b.Gender]
	for i, v := range b.ScriptFlags {
		c.ScriptFlags[i] = v
	}
	if b.Race > 0 && b.Race <= len(xxvc.Races) {
		c.Race = xxvc.Races[b.Race-1]
	}
	if b.Career > 0 && b.Career <= len(xxvc.Careers) {
		c.Career = xxvc.Careers[b.Career-1]
	}
	if b.Hostile {
		c.Side = rules.SideMonsters
	}
	for i, a := range brStatOrder {
		c.Scores[a] = b.Abilities[i]
	}
	c.Saves = b.Saves
	c.Companion, c.NonPlayer, c.Morale = b.IsCompanion, b.NonPlayer, b.NPCMorale
	for i, v := range b.SkillLevels {
		c.Skills.Set(SkillNo(i+1), int8(v)) // SkillLevels[0] is skill 1
	}
	for i, s := range b.WeaponSpecializations {
		c.Specializations[i] = int32(s)
	}
	var err error
	c.Items, err = convertBRItems(t, b.Items)
	return c, err
}

func armorFlagsOf(f int) ArmorFlags {
	return ArmorFlags{Body: f&armorBodyBit != 0, Breathing: f&armorBreathingBit != 0, Fields: f&armorFieldsBit != 0,
		Goggles: f&armorGogglesBit != 0, ECM: f&armorECMBit != 0}
}

func convertBRItems(t xxvc.Tables, items []BRSavedItem) ([]Carried, error) {
	ids := slices.Sorted(maps.Keys(t.Items))
	var out []Carried
	for i, it := range items {
		id, ok := rules.ItemID(0), false
		for _, cand := range ids {
			if tt := t.Items[cand]; tt.Type == it.ItemType && tt.Words == it.NameIDs {
				id, ok = cand, true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("ConvertBRCharacter: item %d (type %d, words %v) has no template", i+1, it.ItemType, it.NameIDs)
		}
		out = append(out, Carried{Slot: rules.ItemSlot(i + 1), Readied: it.Readied,
			Ref: rules.ItemRef{ID: id, Plus: it.Plus, Count: it.Quantity}})
	}
	return out, nil
}
