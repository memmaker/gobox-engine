package importer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"gobox/adventure"
	"gobox/asset/orig"

	"gobox/rules"
	"gobox/rules/xxvc"
	"gobox/rulesxxvc"
)

// brPointer is a Turbo Pascal far pointer in a saved record: meaningless
// once saved, read only to keep the layout.
type brPointer uint32

// brCharacter is types.pas TCharacter as saveload.pas SaveChar writes it
// (BlockWrite of the $103-byte record, packed, little-endian) and as
// MON0CHA holds a monster. The field names are the Pascal's; the tails are
// the original's unused bytes.
type brCharacter struct {
	Name                  [16]byte // String[15]: length byte, then the text
	Abilities             [7]uint8
	CurrentAbilities      [7]uint8
	Saves                 [8]uint8
	Gender                uint8
	Race                  uint8 // TRace: RaceMonster, then xxvc.Races
	Career                uint8 // TCareer: CareerNone, then xxvc.Careers
	Level                 int8
	LevelTail             uint8
	CreditTotal           int32
	Experience            int32
	BaseCombat            uint8
	IconDimensions        uint8
	IconDimensionsTail    uint8
	BaseMovement          uint8
	BaseArmorClass        int8
	Age                   uint8
	AgeTail               [2]uint8
	BaseScale             uint8
	Scale                 uint8
	UnarmedRolls          [2]uint8
	UnarmedDice           [2]uint8
	UnarmedModifiers      [2]int8
	UnarmedModifiersTail  [2]uint8
	HitPointMaximum       uint16
	HitPointMaximumTail   uint8
	ArmorFlags            uint8
	ArmorFlagsTail        [3]uint8
	IsCompanion           uint8
	SkillLevels           [84]uint8
	SkillLevelsTail       [17]uint8
	XPAward               uint16
	AlternateItems        brPointer
	Icon                  uint8
	OrderNumber           uint8
	WeaponSpecializations [10]uint8
	IconSequence          [16]uint8
	Status                uint8
	IsEnabled             uint8
	Side                  int8
	IsQuickfight          uint8
	AttackBonus           uint8
	ArmorClass            [2]uint8
	Attacks               [2]uint8
	CurrentRolls          [2]uint8
	CurrentDice           [2]uint8
	CurrentModifiers      [2]int8
	HitPoints             uint16
	CurrentMovement       uint8
	CurrentMovementTail   uint8
	Npc                   uint8
	Encumbrance           uint16
	IsPilot               uint8
	ReadiedItems          [2]brPointer
	CombatState           brPointer
	ItemList              brPointer
	EffectList            brPointer
	Next                  brPointer
}

// brItemTemplate is types.pas TItemTemplate.
type brItemTemplate struct {
	ItemType         uint8
	NameIDs          [3]uint8
	Plus             int8
	SaveBonus        int8
	Flags            uint8 // C_Item.READIED, NOT_TREASURE, NOT_JOINABLE
	UnidentifiedMask uint8
	Weight           uint16
	Quantity         uint8
	Value            uint16
	Special          [3]uint8
}

// brItem is types.pas TItem: one record of a save's .itm file, of MON0ITM.
type brItem struct {
	Name     [41]byte // String[40]
	IsLabel  uint8
	Next     brPointer
	Template brItemTemplate
}

// brEffect is types.pas TEffect: one record of a save's .spc file.
type brEffect struct {
	EffectID      uint8
	Duration      uint16
	Strength      uint8
	CallsOnRemove uint8
	Next          brPointer
}

const (
	brCharacterSize = 0x103 // SaveLoad's record
	brReadied       = 0x80  // C_Item.READIED in TItemTemplate.flags
	nonPlayerFlag   = 0x80  // c_npc.pas NON_PLAYER_FLAG in TCharacter.npc; the low bits the morale
	sideHostile     = 1     // c_side.pas C_Side.HOSTILE
)

// hostile is TCharacter.side = C_Side.HOSTILE.
func hostile(side int8) bool { return side == sideHostile }

// DecodeSave decodes one saved character: the TCharacter record, its .itm
// items and its .spc effects (either may be nil).
func DecodeSave(record, items, effects []byte) (rulesxxvc.BRSave, error) {
	if len(record) != brCharacterSize {
		return rulesxxvc.BRSave{}, fmt.Errorf("importbr: record is %d bytes, want %d", len(record), brCharacterSize)
	}
	var b brCharacter
	if err := binary.Read(bytes.NewReader(record), binary.LittleEndian, &b); err != nil {
		return rulesxxvc.BRSave{}, err
	}
	s := rulesxxvc.BRSave{
		Name: pascalString(b.Name[:]), Gender: int(b.Gender), Race: int(b.Race), Career: int(b.Career), Level: int(b.Level),
		CreditTotal: int64(b.CreditTotal), Experience: int64(b.Experience), BaseCombat: int(b.BaseCombat), IconDimensions: int(b.IconDimensions),
		BaseMovement: int(b.BaseMovement), BaseArmorClass: int(b.BaseArmorClass),
		BaseScale: int(b.BaseScale), Scale: int(b.Scale), HitPointMaximum: int(b.HitPointMaximum),
		ArmorFlags: int(b.ArmorFlags), XPAward: int64(b.XPAward), Status: rulesxxvc.Status(b.Status),
		IsEnabled: b.IsEnabled != 0, Hostile: hostile(b.Side), IsCompanion: b.IsCompanion != 0,
		NonPlayer: b.Npc&nonPlayerFlag != 0, NPCMorale: int(b.Npc &^ nonPlayerFlag),
		AttackBonus: int(b.AttackBonus), ArmorClass: [2]int{int(b.ArmorClass[0]), int(b.ArmorClass[1])},
		HitPoints: int(b.HitPoints), CurrentMovement: int(b.CurrentMovement), Encumbrance: int(b.Encumbrance),
	}
	for i := range 7 {
		s.Abilities[i], s.CurrentAbilities[i] = int(b.Abilities[i]), int(b.CurrentAbilities[i])
	}
	for i, v := range b.Saves {
		s.Saves[i] = int(v)
	}
	for i := range s.ScriptFlags {
		s.ScriptFlags[i] = int(b.SkillLevelsTail[i])
	}
	for i, v := range b.SkillLevels {
		s.SkillLevels[i] = int(v)
	}
	for i, v := range b.WeaponSpecializations {
		s.WeaponSpecializations[i] = int(v)
	}
	for i := range 2 {
		s.Unarmed[i] = rules.Dice{N: int32(b.UnarmedRolls[i]), Sides: int32(b.UnarmedDice[i]), Bonus: int32(b.UnarmedModifiers[i])}
		s.Current[i] = rules.Dice{N: int32(b.CurrentRolls[i]), Sides: int32(b.CurrentDice[i]), Bonus: int32(b.CurrentModifiers[i])}
	}
	its, err := records[brItem](items)
	if err != nil {
		return rulesxxvc.BRSave{}, fmt.Errorf("importbr items: %w", err)
	}
	for _, it := range its {
		t := it.Template
		s.Items = append(s.Items, rulesxxvc.BRSavedItem{ItemType: int32(t.ItemType),
			NameIDs: [3]int32{int32(t.NameIDs[0]), int32(t.NameIDs[1]), int32(t.NameIDs[2])},
			Plus:    int32(t.Plus), Quantity: int32(t.Quantity), Readied: t.Flags&brReadied != 0})
	}
	efs, err := records[brEffect](effects)
	if err != nil {
		return rulesxxvc.BRSave{}, fmt.Errorf("importbr effects: %w", err)
	}
	for _, e := range efs {
		s.Effects = append(s.Effects, rulesxxvc.Effect{ID: rulesxxvc.EffectID(e.EffectID), Minutes: int(e.Duration),
			Strength: int(e.Strength), CallsOnRemove: e.CallsOnRemove != 0})
	}
	return s, nil
}

// records decodes b as whole T records; a partial one is an error.
func records[T any](b []byte) ([]T, error) {
	var v T
	n := binary.Size(&v)
	if len(b)%n != 0 {
		return nil, fmt.Errorf("%d bytes are no whole number of %d-byte records", len(b), n)
	}
	out := make([]T, len(b)/n)
	return out, binary.Read(bytes.NewReader(b), binary.LittleEndian, out)
}

func pascalString(b []byte) string { return string(b[1 : 1+min(int(b[0]), len(b)-1)]) }

// Monster is MON0 record id as xxvc.Monster: decoded as a save (DecodeSave)
// and converted by the ruleset (rulesxxvc.ConvertBRCharacter), so a monster
// and a saved character take one path. MON0CHA stores the base values only
// (eclmenu.pas LoadCreatureData copies the record, currents 0): ArmorClass[0],
// Damage and Movement are the record's base ones, which rulesxxvc.Spawn reads
// as base and RecalcStats/RecalcMovement turn into the currents.
func monster(t xxvc.Tables, record, items []byte) (xxvc.Monster, error) {
	s, err := DecodeSave(record, items, nil)
	if err != nil {
		return xxvc.Monster{}, err
	}
	c, err := rulesxxvc.ConvertBRCharacter(t, s)
	if err != nil {
		return xxvc.Monster{}, err
	}
	m := xxvc.Monster{Name: c.Name, Race: c.Race, Career: c.Career, Level: int32(c.Level), Scores: map[rules.Ability]int32{},
		HP: int32(c.HP), BaseCombat: int32(c.BaseCombat), IconDimensions: int32(c.IconDimensions), AttackBonus: int32(c.AttackBonus),
		ArmorClass: [2]int32{int32(c.BaseArmorClass), int32(c.ArmorClass[1])}, Damage: c.Unarmed,
		BaseScale: int32(c.BaseScale), Scale: int32(c.Scale), Movement: int32(c.BaseMovement), Skills: map[rules.Skill]int32{},
		Hostile: c.Side == rules.SideMonsters, XPAward: c.XPAward, Credits: c.Coins}
	for a, v := range c.Scores {
		m.Scores[a] = int32(v)
	}
	for i, v := range s.SkillLevels {
		if v != 0 {
			m.Skills[xxvc.Skills[i]] = int32(v)
		}
	}
	for i, v := range c.Saves {
		m.Saves[i] = int32(v)
	}
	for _, it := range c.Items {
		m.Items = append(m.Items, it.Ref)
	}
	return m, nil
}

// A saved character's files (saveload.pas SaveChar; gdconsts.pas itemFileExt,
// effectFileExt): the record, its items, its effects.
const (
	saveExt    = ".SAV"
	itemsExt   = ".STF"
	effectsExt = ".SFX"
)

// importBRSaves converts each saved character in l's directory (<stem>.SAV,
// with <stem>.STF and <stem>.SFX when present) through
// rulesxxvc.ConvertBRCharacter into dst/<stem>.json. Items no template of t
// has get none: ConvertBRCharacter refuses them (t is already written).
func importBRSaves(l *orig.Loader, t xxvc.Tables, dst string) error {
	ents, err := os.ReadDir(l.Dir)
	if err != nil {
		return err
	}
	for _, e := range ents {
		stem, ok := strings.CutSuffix(strings.ToUpper(e.Name()), saveExt)
		if !ok || e.IsDir() {
			continue
		}
		read := func(ext string) ([]byte, error) {
			if !l.Exists(stem + ext) {
				return nil, nil
			}
			return os.ReadFile(l.Path(stem + ext))
		}
		rec, err := read(saveExt)
		if err != nil {
			return err
		}
		items, err := read(itemsExt)
		if err != nil {
			return err
		}
		effects, err := read(effectsExt)
		if err != nil {
			return err
		}
		s, err := DecodeSave(rec, items, effects)
		if err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
		c, err := rulesxxvc.ConvertBRCharacter(t, s)
		if err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		if err := adventure.WriteJSON(dst, stem, c); err != nil {
			return err
		}
	}
	return nil
}
