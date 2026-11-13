package rules1e

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gobox/core"
	"gobox/core/goldbox"
)

// The vectors come from the pure Pascal (por-re bugfixed-rules, vectors-pure.sh).
// A difference is a bug, or a row of gobox-engine/CHANGES-FROM-ORIGINAL.md.
var (
	vecClass = [8]core.Class{CoreClassCleric, CoreClassDruid, CoreClassFighter, CoreClassPaladin, CoreClassRanger, CoreClassMage, CoreClassThief, CoreClassMonk}
	vecRace  = [8]core.Race{CoreRaceMonster, CoreRaceDwarf, CoreRaceElf, CoreRaceGnome, CoreRaceHalfElf, CoreRaceHalfling, CoreRaceHalfOrc, CoreRaceHuman}
	vecCon   = 19 // the Pascal's stat vectors read no constitution; 19 shows the bonus every race gets
	vecSave  = [5]core.Save{SaveParalysis, SavePetrify, SaveRod, SaveBreath, SaveSpell}
)

// documented says whether a difference is a row of CHANGES-FROM-ORIGINAL.md:
// paladin saves and paladin, ranger XP from COTAB; XP past the table; the Dex 10 and 16 thief fixes.
func documented(key []string, want, got []int) bool {
	k := ints(key[1:])
	switch key[0] {
	case "save":
		return vecClass[k[0]] == CoreClassPaladin && k[1] <= 8
	case "xp":
		// Past level 9 the Pascal reads other data. Felix: past its table a class reads POD's
		// table, then adds POD's step (the thief: 160001, 220001, then +220000); a class that stops before stays at -1.
		if k[1] > 9 {
			xp := -1
			if vecClass[k[0]] == CoreClassThief {
				xp = []int{160001, 220001, 440001, 660001}[k[1]-10]
			}
			// the slots of a class without spells are other data too, never read
			spells := vecClass[k[0]] == CoreClassCleric || vecClass[k[0]] == CoreClassMage
			return got[0] == xp && (!spells || slices.Equal(want[1:], got[1:]))
		}
		return (vecClass[k[0]] == CoreClassPaladin || vecClass[k[0]] == CoreClassRanger) && want[0] == -1
	case "damage":
		// ruling 3 (cotab's DAMAGE for all): with flags 0x80 and 0x10 a made
		// save still hurts; POR's Pascal reads 0x10 as save bonus only. No
		// POR script's constant flags have both bits (scan of ECL*.DAX).
		return k[0]&0x90 == 0x90
	case "savethrow":
		// ruling 1: the Pascal's Byte total wraps below 0 and saves; ours fails.
		rng := core.RNG{State: uint32(int32(k[0]))}
		return want[0] == 1 && got[0] == 0 && want[1] == got[1] && RollDice(&rng, 1, 20)+k[3]+k[2] < 0
	case "stat":
		return statDocumented(k, want, got)
	case "careers":
		return vecRace[k[0]] == CoreRaceHalfElf && slices.Equal(got, []int{0, 2, 4, 5, 6, 8, 9, 10, 11, 13, 14, 15, 16}) ||
			vecRace[k[0]] == CoreRaceHuman && slices.Equal(got, []int{0, 2, 3, 4, 5, 6})
	case "age":
		return k[0] == 4 && k[1] == 4 // half-elf ranger takes the half-elf fighter's
	case "aligns":
		switch Career(k[0]) {
		case CareerRanger, 10: // good only, as COTAB
			return slices.Equal(got, []int{0, 3, 6})
		}
		return false
	case "thief":
		fixed := slices.Clone(want)
		switch k[2] {
		case 10:
			fixed[0] += 9 // pick pockets -19 to -10
		case 16:
			fixed[1] += 10 // open locks -5 to +5
		}
		return slices.Equal(fixed, got)
	}
	return false
}

// statDocumented applies the listed differences to the Pascal's stat vector.
// Vector: thac0, highest, attacks, 3 cleric, 3 mage, 5 saves, 8 skills.
func statDocumented(k, want, got []int) bool {
	level := func(c core.Class) int {
		switch {
		case vecClass[k[0]] == c:
			return k[1]
		case vecClass[k[2]] == c:
			return k[3]
		}
		return 0
	}
	fixed := slices.Clone(want)
	if level(CoreClassPaladin) > 6 || level(CoreClassRanger) > 7 {
		fixed[2] = 3 // COTAB's attacks
	}
	copied := level(CoreClassPaladin) > 0 && level(CoreClassPaladin) <= 8
	if copied {
		copy(fixed[9:14], got[9:14]) // COTAB's paladin saves, steps 1 and 4 included
	}
	// COTAB's superset (Felix 18:10), one by one:
	// 1. a monk's saves are also the MonkOldSaves row of its old level (0 here).
	if level(CoreClassMonk) > 0 && !copied {
		for i := range vecSave {
			fixed[9+i] = min(fixed[9+i], int(POR.MonkOldSaves[0][i]))
		}
	}
	// 2. a paladin from level 9 gains cleric spells (PaladinSpells); without cleric levels the held slots go.
	if add := rowSum(POR.PaladinSpells, level(CoreClassPaladin)); add != ([5]int{}) {
		if level(CoreClassCleric) <= 0 {
			copy(fixed[3:6], []int{0, 0, 0})
		}
		for i := range 3 {
			fixed[3+i] += add[i]
		}
	}
	// 3. a ranger from level 8 gains magic-user spells (RangerSpells columns 4, 5); without magic-user levels the held slots go.
	if add := rowSum(POR.RangerSpells, level(CoreClassRanger)); add != ([5]int{}) {
		if level(CoreClassMage) <= 0 {
			copy(fixed[6:9], []int{0, 0, 0})
		}
		fixed[6] += add[3]
		fixed[7] += add[4]
	}
	// 4. the constitution bonus is in the stored paralysis save (Felix: COTAB's, for all; POR's
	// roll-time affect 0x5A is gone): everyone's from 19 (vecCon), a dwarf's or halfling's from 4 (TestConSave).
	if !copied {
		fixed[9] -= int(goldbox.ConSaveHigh(byte(vecCon)))
	}
	return slices.Equal(fixed, got)
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func savethrow(rng *core.RNG, target, modifier, bonus int) bool {
	die := RollDice(rng, 1, 20)
	if made, fixed := SaveFixed(die); fixed {
		return made
	}
	return SaveMade(target, die+bonus+modifier)
}

func ints(s []string) []int {
	out := make([]int, len(s))
	for i, f := range s {
		n, err := strconv.Atoi(f)
		if err != nil {
			panic(f)
		}
		out[i] = n
	}
	return out
}

func ours(kind string, k []int) string {
	join := func(v ...int) string {
		return strings.Trim(fmt.Sprint(v), "[]")
	}
	switch kind {
	case "thac0":
		return join(Thac0At(vecClass[k[0]], k[1]))
	case "save":
		return join(SaveAt(vecClass[k[0]], k[1], vecSave[k[2]]))
	case "thief":
		s := ThiefSkills(k[0], vecRace[k[1]], k[2], ItemPowers{})
		return join(s[:]...)
	case "die":
		n, sides := TrainingDie(vecClass[k[0]])
		return join(n, sides)
	case "xp":
		return join(int(XPForLevel(vecClass[k[0]], k[1])), classSlot(k[0], k[1], 0), classSlot(k[0], k[1], 1), classSlot(k[0], k[1], 2))
	case "con":
		return join(ConHP(k[0]))
	case "attack":
		rng := core.RNG{State: uint32(int32(k[0]))}
		die := AttackDie(&rng)
		return join(b2i(!AttackMissed(die) && AttackHits(die, k[1], k[2], k[3])), int(int32(rng.State)))
	case "trap":
		rng := core.RNG{State: uint32(int32(k[0]))}
		die := AttackDie(&rng)
		return join(b2i(!AttackMissed(die) && TrapHits(die, k[1], k[2])), int(int32(rng.State)))
	case "savethrow":
		rng := core.RNG{State: uint32(int32(k[0]))}
		return join(b2i(savethrow(&rng, k[1], k[2], k[3])), int(int32(rng.State)))
	case "strength":
		i := StrengthIndex(k[0], k[1])
		return join(StrengthToHit(i), StrengthDamage(i))
	case "turn":
		return join(TurnNeed(k[0], k[1]))
	case "attacks":
		return join(AttacksFromRate(k[0], k[1]))
	case "turnundead":
		rng := core.RNG{State: uint32(int32(k[0]))}
		RollDice(&rng, 1, 12) // how many it can affect; one target here
		roll := RollDice(&rng, 1, 20)
		need := TurnNeed(k[2], TurnColumn(k[1]))
		success := roll >= max(need, -need)
		return join(b2i(success && need > 0), b2i(success && need <= 0), int(int32(rng.State)))
	case "dice":
		rng := core.RNG{State: uint32(int32(k[0]))}
		r := RollDice(&rng, k[1], k[2])
		return join(r, int(int32(rng.State)))
	case "stat":
		levels := map[core.Class]int{vecClass[k[0]]: k[1]}
		if k[2] != k[0] {
			levels[vecClass[k[2]]] = k[3]
		}
		st := ClassStats(Who{Levels: levels, Race: CoreRaceHuman, Dex: 12, Wis: k[4], Con: vecCon,
			ClericSlots: LevelSlots{7, 7, 7}, MageSlots: LevelSlots{7, 7, 7}})
		v := []int{st.Thac0Base, st.Highest, st.Attacks}
		v = append(v, st.ClericSlots[:3]...) // the Pascal has three levels
		v = append(v, st.MageSlots[:3]...)
		for _, sv := range vecSave {
			v = append(v, st.Saves[sv])
		}
		v = append(v, st.Skills[:]...)
		return join(v...)
	case "limit":
		r := vecRace[k[0]]
		v := []int{}
		lo, hi := AbilityRange(r, Strength)
		v = append(v, lo, hi, ExceptionalStrengthCap(r))
		for a := Intelligence; a <= Charisma; a++ {
			lo, _ := AbilityRange(r, a)
			v = append(v, lo)
		}
		for a := Intelligence; a <= Charisma; a++ {
			_, hi := AbilityRange(r, a)
			v = append(v, hi)
		}
		return join(v...)
	case "careers":
		var v []int
		for _, c := range RaceCareers(vecRace[k[0]]) {
			v = append(v, int(c))
		}
		return join(v...)
	case "age":
		b, n, s := StartingAge(vecRace[k[0]], Career(k[1]))
		return join(b, n, s)
	case "agestep":
		return join(AgeStep(vecRace[k[0]], k[1]))
	case "classmin":
		return join(CareerMinimum(Career(k[0]), Ability(k[1])))
	case "aligns":
		var v []int
		for _, a := range CareerAlignments(Career(k[0])) {
			v = append(v, int(a))
		}
		return join(v...)
	}
	panic(kind)
}

// classSlot: only cleric and mage rows are slots in ours; the Pascal has zeros elsewhere.
func classSlot(class, level, i int) int {
	switch vecClass[class] {
	case CoreClassCleric:
		return ClericSlots(level)[i]
	case CoreClassMage:
		return MageSlots(level)[i]
	}
	return 0
}

func TestAgainstPascal(t *testing.T) {
	cases, diffs := 0, map[string]int{}
	for _, file := range []string{"por-adv", "por-stat", "por-tab", "por-dice", "por-cmb", "por-range", "por-turn"} {
		f, err := os.Open("testdata/" + file + ".txt")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "#") {
				continue
			}
			key, want, ok := strings.Cut(line, " =")
			if !ok {
				t.Fatalf("bad vector %q", line)
			}
			cases++
			fs := strings.Fields(key)
			got := ours(fs[0], ints(fs[1:]))
			w := strings.Join(strings.Fields(want), " ")
			if got != w {
				diffs[fs[0]]++
				if !documented(fs, ints(strings.Fields(want)), ints(strings.Fields(got))) {
					t.Errorf("%s: pascal %s, ours %s", key, want, got)
				}
			}
		}
	}
	t.Logf("%d cases, differences by kind: %v", cases, diffs)
}
