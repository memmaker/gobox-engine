package engine

import (
	"bufio"
	"fmt"
	"gobox/rules1e"
	"os"
	"strconv"
	"strings"
	"testing"
)

// The vectors come from the pure Pascal (por-re veceff). The handlers still live here;
// a difference is a bug, or a row of gobox-engine/CHANGES-FROM-ORIGINAL.md.
var vecNames = []string{"", "NONE", "KOBOLD", "ORC", "OGRE", "BUGBEAR", "GNOLL"}

func setVecName(c *rules1e.Character, name string) {
	c.SetName(name)
}

func effectVector(e *Engine, kind string, k []int) string {
	a, d := &rules1e.Character{}, &rules1e.Character{}
	a.Side, d.Side = rules1e.SideParty, rules1e.SideMonsters
	a.THAC0 = 40
	a.Combat, d.Combat = &rules1e.CombatInfo{Target: d}, &rules1e.CombatInfo{Target: a}
	e.state().Party, e.state().Who = []*rules1e.Character{a, d}, 0
	e.state().Rv.Damage, e.state().AffectVeto = 0, 0
	if kind == "effattack" {
		eff, side, seed, typ, name, size, v, st := k[0], k[1], k[2], k[3], k[4], k[5], k[6], k[7]
		holder := a
		if side == 0 {
			d.Kind, d.Size = byte(typ), byte(size)
			setVecName(d, vecNames[name])
		} else {
			holder = d
			a.Kind = byte(typ)
			setVecName(a, vecNames[name])
			setVecName(d, vecNames[name])
			d.Kind, d.Size = 1, byte(size)
		}
		e.state().AddAffect(holder, rules1e.AffectType(eff), 0, byte(st), 0)
		d.Combat.Delay = byte(v)
		e.state().CombatRound = byte(v)
		e.state().RandSeed = uint32(int32(seed))
		e.state().Rv.HitRoll = 0
		hit := e.state().AttackHits(a, d, e.state().Hand(a), 50)
		return fmt.Sprintf("%s%d %d %d %d", b2iVec(hit), e.state().Rv.HitRoll, e.state().Rv.Damage, e.state().AffectVeto, holder.Affects[0].Data)
	}
	switch kind {
	case "effweapon":
		eff, side, seed, dice, sides, bonus := k[0], k[1], k[2], k[3], k[4], k[5]
		holder := a
		if side != 0 {
			holder = d
		}
		if eff > 0 {
			e.state().AddAffect(holder, rules1e.AffectType(eff), 0, 2, 0)
		}
		d.Kind = 4
		a.Dice[0], a.Sides[0], a.Bonus[0] = rules1e.Count(dice), rules1e.Count(sides), rules1e.Modifier(bonus)
		e.state().RandSeed = uint32(int32(seed))
		e.state().DamageRoll(a, d, e.state().Hand(a), 1)
		return fmt.Sprintf("%d %d %d", e.state().Rv.Damage, e.state().Rv.DmgType, int32(e.state().RandSeed))
	case "effplain":
		eff, seed, dmg, mask, mode, saved, hp := k[0], k[1], k[2], k[3], k[4], k[5], k[6]
		if eff > 0 {
			e.state().AddAffect(d, rules1e.AffectType(eff), 0, 2, 0)
		}
		d.HP, d.Status, d.Okay = byte(hp), rules1e.StatusOkay, 1
		e.state().Rv.DmgType, e.state().Rv.DiceN = byte(mask), 2
		e.state().RandSeed = uint32(int32(seed))
		e.state().Damage(d, dmg, byte(mode), saved == 1)
		return fmt.Sprintf("%d %d %d %d %d", e.state().Rv.Damage, d.HP, d.Status, d.Okay, int32(e.state().RandSeed))
	case "effmove":
		eff, base, side, modifier := k[0], k[1], k[2], k[3]
		if eff > 0 {
			e.state().AddAffect(d, rules1e.AffectType(eff), 0, 0, 0)
		}
		d.Move, d.Side = byte(base), byte(side)
		e.state().Game.PartyMove = uint16(int16(modifier))
		return strconv.Itoa(int(byte(e.state().MoveRate(d))))
	}
	eff, seed, sv, con, al, mask, st := k[0], k[1], k[2], k[3], k[4], k[5], k[6]
	a.Alignment = byte(al)
	d.Stats[rules1e.StatCon] = rules1e.Score(byte(con))
	d.Saves[sv] = 10
	e.state().AddAffect(d, rules1e.AffectType(eff), 0, byte(st), 0)
	e.state().Rv.DmgType, e.state().Rv.Damage, e.state().Rv.DiceN, e.state().Rv.AffVal = byte(mask), 10, 2, 0
	e.state().RandSeed = uint32(int32(seed))
	made := e.state().SavingThrow(d, sv, 0)
	return fmt.Sprintf("%s%d %d", b2iVec(made), e.state().Rv.Save, e.state().Rv.Damage)
}

func b2iVec(b bool) string {
	if b {
		return "1 "
	}
	return "0 "
}

// documentedEffect: rows of CHANGES-FROM-ORIGINAL.md. Saves wrapping in the Pascal's byte
// (ruling 1).
func documentedEffect(kind string, k []int, want, got []string) bool {
	switch {
	case kind == "effsave":
		roll, _ := strconv.Atoi(want[1])
		return want[0] == "1" && got[0] == "0" && want[1] == got[1] && int8(roll) < 0
	}
	return false
}

// mirrorImage is POR's mirror image affect, the one ruled row of the vectors.
const mirrorImage = 0x1C

func TestEffectsAgainstPascal(t *testing.T) {
	e := testEngine(t)
	e.Init()
	cases, diffs := 0, map[string]int{}
	for _, file := range []string{"por-eff", "por-dmg"} {
		f, err := os.Open("../rules1e/testdata/" + file + ".txt")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if strings.HasPrefix(sc.Text(), "#") {
				continue
			}
			key, want, _ := strings.Cut(sc.Text(), " =")
			fs := strings.Fields(key)
			k := make([]int, len(fs)-1)
			for i, s := range fs[1:] {
				if k[i], err = strconv.Atoi(s); err != nil {
					t.Fatal(err)
				}
			}
			cases++
			got := strings.Join(strings.Fields(effectVector(e, fs[0], k)), " ")
			w := strings.Join(strings.Fields(want), " ")
			if fs[1] == strconv.Itoa(int(mirrorImage)) && (fs[0] == "effplain" || fs[0] == "effweapon") {
				// Ruled (Felix, all games): sotsb's mirror image, images in the data's high
				// nibble. The vectors' data 2 holds none, so the row plays as no affect,
				// but for the seed (the last field): the rule rolls its die.
				none := strings.Fields(effectVector(e, fs[0], append([]int{0}, k[1:]...)))
				ours := strings.Fields(got)
				if strings.Join(ours[:len(ours)-1], " ") != strings.Join(none[:len(none)-1], " ") {
					t.Errorf("%s: ruled mirror image (as no affect) %v, ours %v", key, none, ours)
				}
				continue
			}
			if got != w {
				diffs[fs[0]+" "+fs[1]]++
				if !documentedEffect(fs[0], k, strings.Fields(w), strings.Fields(got)) {
					t.Errorf("%s: pascal %s, ours %s", key, w, got)
				}
			}
		}
	}
	t.Logf("%d cases, differences by effect: %v", cases, diffs)
}
