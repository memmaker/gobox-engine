package engine

// Oracle tests: the original functions are copied here (from cotab-go,
// sotsb-go and por-go) on plain structs, run on the same inputs and the same
// scripted RNG as rules on a char.Rec, and compared.

import (
	"gobox/rules"
	"math/rand"
	"reflect"
	"testing"

	"gobox/rules1e"
)

// seqRNG replays a seeded sequence and records the calls, so the order of
// RNG calls is compared too.
type seqRNG struct {
	r     *rand.Rand
	calls [][2]int
}

func newSeq(seed int64) *seqRNG { return &seqRNG{r: rand.New(rand.NewSource(seed))} }
func (s *seqRNG) Dice(n, sides int) int {
	s.calls = append(s.calls, [2]int{n, sides})
	return n + s.r.Intn(n*sides-n+1)
}

// tables copied from the games
var cotabThac0 = [8][13]byte{
	{40, 40, 40, 40, 42, 42, 42, 44, 44, 44, 46, 46, 46},
	{40, 40, 40, 40, 42, 42, 42, 44, 44, 44, 46, 46, 46},
	{39, 40, 40, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51},
	{40, 40, 40, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51},
	{40, 40, 40, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51},
	{39, 39, 39, 39, 39, 39, 41, 41, 41, 41, 41, 43, 43},
	{40, 40, 40, 40, 40, 41, 41, 41, 41, 44, 44, 44, 44},
	{40, 40, 40, 40, 42, 42, 42, 44, 44, 44, 46, 46, 46},
}
var sotsbThac0 = [7][19]byte{
	{40, 40, 40, 40, 42, 42, 42, 44, 44, 44, 46, 46, 46, 48, 48, 48, 50, 50, 50},
	{40, 40, 40, 40, 42, 42, 42, 44, 44, 44, 46, 46, 46, 48, 48, 48, 50, 50, 50},
	{39, 40, 40, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57},
	{40, 40, 40, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57},
	{40, 40, 40, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57},
	{39, 39, 39, 39, 39, 39, 41, 41, 41, 41, 41, 43, 43, 43, 43, 43, 46, 46, 46},
	{40, 40, 40, 40, 40, 41, 41, 41, 41, 44, 44, 44, 44, 46, 46, 46, 46, 48, 48},
}

// oc is the part of a game's Character the oracles touch.
type oc struct {
	SaveBonus                              byte
	Saves                                  [5]byte
	THAC0                                  byte
	AC                                     [2]byte
	Levels, OldLevels                      []byte
	Multi, HPRoll, MaxHP, HP, Status, Okay byte
}

func load(r *rules1e.Rec) *oc {
	c := &oc{SaveBonus: byte(r.SaveBonus()), THAC0: byte(r.THAC0()), Multi: byte(r.MulticlassLevel()),
		HPRoll: byte(r.HPRoll()), MaxHP: byte(r.MaxHP()), HP: byte(r.HP()), Status: byte(r.Status()), Okay: byte(r.Get("okay", 0))}
	for i := range c.Saves {
		c.Saves[i] = byte(r.Save(i))
	}
	for i := range c.AC {
		c.AC[i] = byte(r.AC(i))
	}
	for k := 0; k < r.Classes(); k++ {
		c.Levels = append(c.Levels, byte(r.Level(k)))
		c.OldLevels = append(c.OldLevels, byte(r.OldLevel(k)))
	}
	return c
}

func (c *oc) same(t *testing.T, r *rules1e.Rec, what string) {
	t.Helper()
	if !reflect.DeepEqual(c, load(r)) {
		t.Fatalf("%s: oracle %+v rules %+v", what, c, load(r))
	}
}

// ---- originals ----

// cotab-go/engine/rules.go:335 and sotsb-go/engine/rules.go (identical): unsigned compare.
func oracleSaveUnsigned(rv *rules.Roll, rng rules.DiceSource, c *oc, typ, bonus int) bool {
	rv.SaveMade = true
	rv.Save = byte(rng.Dice(1, 20))
	switch rv.Save {
	case 1:
		rv.SaveMade = false
	case 20:
		rv.SaveMade = true
	default:
		rv.Save = byte(int(rv.Save) + int(int8(c.SaveBonus)) + int(int8(bonus)))
		rv.SaveType = byte(typ)
		rv.SaveMade = c.Saves[int(int8(typ))] <= rv.Save
	}
	return rv.SaveMade
}

// por-go/engine/rules.go:253 (signed compare).
func oracleSavePor(rv *rules.Roll, rng rules.DiceSource, c *oc, typ, bonus int) bool {
	rv.Save = byte(rng.Dice(1, 20))
	switch rv.Save {
	case 1:
		return false
	case 20:
		return true
	}
	rv.Save = byte(int(rv.Save) + int(int8(c.SaveBonus)) + int(int8(bonus)))
	rv.SaveType = byte(typ)
	return int(c.Saves[int(int8(typ))]) <= int(int8(rv.Save))
}

// cotab-go/engine/rules.go:314 (side adjustment passed in).
func oracleAttackCotab(rv *rules.Roll, rng rules.DiceSource, att *oc, adj int, need int) bool {
	rv.HitRoll = int8(rng.Dice(1, 20))
	if rv.HitRoll <= 1 {
		return false
	}
	if rv.HitRoll == 20 {
		rv.HitRoll = 100
	}
	return rv.HitRoll >= 0 && int(rv.HitRoll)+int(att.THAC0)+adj >= int(byte(need))
}

// por-go/engine/rules.go:231 with the rules4e Hits hook.
func oracleAttackPor(rv *rules.Roll, rng rules.DiceSource, att *oc, adj int, need int, hits func(int) bool) bool {
	rv.HitRoll = int8(rng.Dice(1, 20))
	if rv.HitRoll <= 1 {
		return false
	}
	if rv.HitRoll == 20 {
		rv.HitRoll = 100
	}
	over := int(rv.HitRoll) + int(att.THAC0) + adj - int(byte(need))
	return rv.HitRoll >= 0 && over >= 0 && (rv.HitRoll == 100 || hits(over))
}

// cotab-go/engine/damage.go:21 trapHits; por RollHit (rules.go:217) is the same with byte(bonus).
func oracleTrapHits(rv *rules.Roll, rng rules.DiceSource, c *oc, bonus byte) bool {
	rv.HitRoll = int8(rng.Dice(1, 20))
	if rv.HitRoll <= 1 {
		return false
	}
	if rv.HitRoll == 20 {
		rv.HitRoll = 100
	}
	return rv.HitRoll >= 0 && int(rv.HitRoll)+int(bonus) > int(c.AC[0])
}

// cotab-go/engine/classcalc.go:122
func oracleThac0Cotab(class int, level int8) byte {
	i := class*len(cotabThac0[0]) + int(level)
	return cotabThac0[i/len(cotabThac0[0])][i%len(cotabThac0[0])]
}

// sotsb-go/engine/classcalc.go:162
func oracleThac0Sotsb(class int, level int8) byte {
	width := len(sotsbThac0[0])
	i := class*width + int(level)
	return sotsbThac0[i/width][i%width]
}

// cotab-go/engine/stats.go:36
func oracleStrengthCode(str, str00 byte) byte {
	if str == 18 {
		return str00 + 1
	}
	return str + 100
}

// cotab-go/engine/damage.go:68 damagePlayer, combat bookkeeping returned.
func oracleDamage(c *oc, dmg byte) (over byte, dying, down bool) {
	var left byte
	if c.HP >= dmg {
		left = c.HP - dmg
	} else {
		over = dmg - c.HP
	}
	switch {
	case over > 9 || left == 0 && c.Status == 1:
		c.Status = 6
	case over > 0:
		c.Status = 5
		dying = true
	case left == 0:
		c.Status = 4
	}
	if c.Status == 0 || c.Status == 1 {
		c.HP = left
		return over, false, false
	}
	c.Okay = 0
	c.HP = 0
	return over, dying, true
}

// ---- tests ----

func recs(t *testing.T, file string, n int) []*rules1e.Rec {
	l, err := rules1e.LoadLayout(file)
	if err != nil {
		t.Fatal(err)
	}
	rnd := rand.New(rand.NewSource(7))
	var out []*rules1e.Rec
	for i := 0; i < n; i++ {
		r := rules1e.New(l)
		rnd.Read(r.B)
		switch i % 3 { // sane levels / small numbers as well as raw noise
		case 1:
			for k := 0; k < r.Classes(); k++ {
				r.Set("levels", k, rnd.Intn(14))
				r.Set("oldLevels", k, rnd.Intn(8))
			}
			r.Set("multiclassLevel", 0, rnd.Intn(4))
			r.Set("status", 0, rnd.Intn(8))
		case 2:
			r.Set("saveBonus", 0, rnd.Intn(21)-10)
			r.Set("hp", 0, rnd.Intn(40))
			r.Set("status", 0, rnd.Intn(2))
		}
		out = append(out, r)
	}
	return out
}

var layouts = []string{"../rules1e/testdata/cotab.json", "../rules1e/testdata/sotsb.json"}

func TestSavingThrow(t *testing.T) {
	signedDiffers := 0
	for _, f := range layouts {
		for i, r := range recs(t, f, 600) {
			typ, bonus := i%5, i%61-30
			a, b, c := newSeq(int64(i)), newSeq(int64(i)), newSeq(int64(i))
			var rv, ov, up rules.Roll
			groups := 0
			got := SavingThrow(&rv, a, r, typ, bonus, Hooks{Group: func() { groups++ }})
			want := oracleSavePor(&ov, b, load(r), typ, bonus)
			uns := oracleSaveUnsigned(&up, c, load(r), typ, bonus)
			if got != want || rv.Save != ov.Save || !reflect.DeepEqual(a.calls, b.calls) {
				t.Fatalf("%s %d: rules %v por %v", f, i, got, want)
			}
			if groups > 1 || groups == 0 && ov.Save != 1 && ov.Save != 20 {
				t.Fatalf("group hook called %d times", groups)
			}
			if got != uns {
				// documented: only where the unsigned compare bit, a total below 0
				// (or above 127) wrapping through the byte
				total := int(int8(rv.Save))
				if total >= 0 || !uns {
					t.Fatalf("%s %d: unexpected unsigned difference total %d", f, i, total)
				}
				signedDiffers++
			}
		}
	}
	if signedDiffers == 0 {
		t.Fatal("no case exercised the unsigned bug")
	}
	t.Logf("%d cases where the games' unsigned compare saved wrongly", signedDiffers)
}

func TestAttackAndTrapHits(t *testing.T) {
	for _, f := range layouts {
		for i, r := range recs(t, f, 600) {
			adj, need := i%21-10, i%40
			for _, hit := range []func(int) bool{nil, func(o int) bool { return o%2 == 0 }} {
				a, b := newSeq(int64(i)), newSeq(int64(i))
				var rv, ov rules.Roll
				got := AttackHits(&rv, a, r, adj, need, Hooks{Hit: hit})
				var want bool
				if hit == nil {
					want = oracleAttackCotab(&ov, b, load(r), adj, need)
				} else {
					want = oracleAttackPor(&ov, b, load(r), adj, need, hit)
				}
				if got != want || rv != ov || !reflect.DeepEqual(a.calls, b.calls) {
					t.Fatalf("attack %d: %v %v", i, got, want)
				}
			}
			a, b := newSeq(int64(i)), newSeq(int64(i))
			var rv, ov rules.Roll
			if got, want := TrapHits(&rv, a, r, byte(i), Hooks{}), oracleTrapHits(&ov, b, load(r), byte(i)); got != want || rv != ov {
				t.Fatalf("trap %d", i)
			}
		}
	}
}

func TestTablesAndStrength(t *testing.T) {
	var ct [][]byte
	for i := range cotabThac0 {
		ct = append(ct, cotabThac0[i][:])
	}
	var st [][]byte
	for i := range sotsbThac0 {
		st = append(st, sotsbThac0[i][:])
	}
	for class := 0; class < 7; class++ {
		for lvl := 0; lvl < 13; lvl++ {
			if got, want := Thac0At(ct, class, int8(lvl)), oracleThac0Cotab(class, int8(lvl)); got != want {
				t.Fatalf("cotab %d %d", class, lvl)
			}
		}
		for lvl := 0; lvl < 19; lvl++ {
			if got, want := Thac0At(st, class, int8(lvl)), oracleThac0Sotsb(class, int8(lvl)); got != want {
				t.Fatalf("sotsb %d %d", class, lvl)
			}
		}
	}
	for s := 0; s < 256; s++ {
		for s0 := 0; s0 < 256; s0 += 5 {
			if StrengthCode(byte(s), byte(s0)) != oracleStrengthCode(byte(s), byte(s0)) {
				t.Fatal("strength")
			}
		}
	}
}

func TestDamagePlayer(t *testing.T) {
	for i, r := range recs(t, layouts[0], 900) {
		o := load(r)
		dmg := byte(i * 7)
		wo, wd, wdn := oracleDamage(o, dmg)
		go1, gd, gdn := DamagePlayer(r, dmg)
		if wo != go1 || wd != gd || wdn != gdn {
			t.Fatalf("%d", i)
		}
		o.same(t, r, "damage")
	}
}
