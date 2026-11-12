package pool

import (
	"bufio"
	"fmt"
	"os"
	"rules1e"
	"strings"
	"testing"
)

// The vectmpl states and purses (por-re src/pure/vectmpl.pas).
const (
	tmplWell = iota
	tmplAfflicted
	tmplDead
	tmplStoned
	tmplCursedItem
	tmplDeadPoisoned
)

const (
	tmplPurseChar = iota
	tmplPursePool
	tmplPurseNone
)

// tmplPlatinum is the platinum the paying purse holds.
const tmplPlatinum = 1500

// TestTempleAgainstPascal compares cureNeeded, pay and templeCure with Temple.CureNeeded and Temple.TempleCure
// (rules1e testdata/por-temple.txt, por-re vectmpl). The party holds one purse, so the character's and the pool's
// coins compare as one sum of gold.
func TestTempleAgainstPascal(t *testing.T) {
	f, err := os.Open("../../rules1e/testdata/por-temple.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		w := atois(t, strings.Fields(val))
		cure, state, purseKind, seed := k[0], k[1], k[2], k[3]
		e := quickStarted(t)
		c := fixChar(20, 0)
		c.MaxHP, c.HP, c.HPRoll = 20, 15, 12
		c.Stats[rules1e.StatCon] = 16
		c.Levels[rules1e.ClsFighter] = 3
		e.state().Purse = rules1e.Purse{}
		if purseKind != tmplPurseNone {
			e.state().Purse.Gold = tmplPlatinum * 5
		}
		switch state {
		case tmplAfflicted:
			tmplAfflict(e, c, cure)
		case tmplDead:
			c.Status, c.Okay, c.HP = rules1e.StatusDead, 0, 0
		case tmplCursedItem:
			c.Items = []*rules1e.Item{{Cursed: 1}}
		case tmplDeadPoisoned:
			c.Status, c.Okay, c.HP = rules1e.StatusDead, 0, 0
			e.state().AddAffect(c, rules1e.AffPoisoned, 0, 0, 0)
		case tmplStoned:
			c.Status, c.Okay = rules1e.StatusStoned, 0
		}
		e.state().RandSeed = uint32(int32(seed))
		needed := e.state().CureNeeded(c, cure)
		paid := e.state().Pay(uint64(e.state().T().Town.Temple[cure].Price))
		if paid {
			e.state().TempleCure(c, cure)
		}
		gold := (w[8]+w[10])*5 + w[9] + w[11]
		got := fmt.Sprintf("%d %d %d %d %d %d %d %d %d %d", rules1e.B2i(needed), rules1e.B2i(paid), c.Status, c.HP, c.MaxHP,
			c.Stats[rules1e.StatCon], c.Okay, len(c.Affects), e.state().Purse.Gold, int32(e.state().RandSeed))
		want := fmt.Sprintf("%d %d %d %d %d %d %d %d %d %d", w[0], w[1], w[2], w[3], w[4], w[5], w[6], w[7], gold, w[12])
		if got != want {
			t.Errorf("%s: Go %s, Pascal %s", key, got, want)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// tmplAfflict gives c what cure cures, as vectmpl's Afflict.
func tmplAfflict(e *Engine, c *rules1e.Character, cure int) {
	switch cure {
	case 0:
		e.state().AddAffect(c, rules1e.AffBlind, 0, 0, 0)
	case 1:
		e.state().AddAffect(c, rules1e.Diseases[1], 0, 0, 0)
	case 2, 3, 4:
		c.HP = 2
	case 5:
		e.state().AddAffect(c, rules1e.AffPoisoned, 0, 0, 0)
	case 6:
		c.Status = rules1e.StatusDead
	case 7:
		e.state().AddAffect(c, rules1e.AffBestowCurse, 0, 0, 0)
	case 8:
		c.Status = rules1e.StatusStoned
	}
}
