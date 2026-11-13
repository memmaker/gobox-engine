package pool

import (
	"bufio"
	"gobox/core/goldbox"
	"os"
	"rules1e"
	"strconv"
	"strings"
	"testing"
)

// TestNpcShareAgainstPascal compares AddToParty's orders with CharHlp.AddCharToParty, and hiredShare with PostCbt.NpcTreasureShare
// (rules1e testdata/por-npcshare.txt, por-re vecpost). Money is copper inside
// the port (Felix's ruling): the Pascal's five coin piles are compared as their
// worth in copper; gems and jewelry are counts on both sides.
func TestNpcShareAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-npcshare.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rate := goldbox.GoldBoxRates.Coin
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
		if strings.HasPrefix(key, "join ") {
			// CharHlp.AddCharToParty: joiners with stored orders get the lowest free ones.
			e := testEngine(t)
			e.state().Party = nil
			var js []*rules1e.Character
			for _, o := range k {
				c := &rules1e.Character{}
				c.Order = byte(o)
				e.state().AddToParty(c)
				js = append(js, c)
			}
			for i, c := range js {
				if int(c.Order) != w[i] {
					t.Errorf("%s: pascal %v, ours joiner %d order %d", key, w, i, c.Order)
				}
			}
			cases++
			continue
		}
		members, npcs, share, down, pile := k[0], k[1], k[2], k[3], k[4]
		e := testEngine(t)
		e.state().Party = nil
		for i := 1; i <= members; i++ {
			c := &rules1e.Character{}
			c.Status = rules1e.StatusOkay
			if i <= npcs {
				c.Control, c.Share = 0x80, byte(share+i-1)
			}
			if i == 1 && down == 1 && npcs > 0 {
				c.Status = rules1e.StatusDead
			}
			e.state().Party = append(e.state().Party, c)
		}
		var copper uint64
		for i := 0; i < 5; i++ {
			copper += uint64(pile*(i+1)) * rate[i]
		}
		e.state().Loot = rules1e.Purse{Gold: copper / 200, Copper: copper % 200, Gems: uint64(pile * 6), Jewelry: uint64(pile * 7)}
		e.r1e().HiredShare()
		var want uint64
		for i := 0; i < 5; i++ {
			want += uint64(w[i]) * rate[i]
		}
		got := e.state().Loot.Gold*200 + e.state().Loot.Copper
		// The documented row: the port cuts one copper pile, the Pascal each coin type, so they
		// agree when every pile divides; otherwise ours is the copper pile's cut.
		if !divides(k, pile) {
			shares, parts := 0, 0
			for _, c := range e.state().Party {
				if c.Control > 0x7F && c.Status == rules1e.StatusOkay {
					shares += int(c.Share & 7)
					parts += int(c.Share & 7)
				} else {
					parts++
				}
			}
			if shares > 0 {
				want = copper - copper/uint64(parts)*uint64(shares)
			}
		}
		if got != want || e.state().Loot.Gems != uint64(w[5]) || e.state().Loot.Jewelry != uint64(w[6]) {
			t.Errorf("%s: pascal %d copper %d gems %d jewelry, ours %d %d %d", key, want, w[5], w[6], got, e.state().Loot.Gems, e.state().Loot.Jewelry)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// divides: every coin pile of the case divides into its shares.
func divides(k []int, pile int) bool {
	members, npcs, share, down := k[0], k[1], k[2], k[3]
	parts := 0
	for i := 1; i <= members; i++ {
		if i <= npcs && !(i == 1 && down == 1 && npcs > 0) {
			parts += (share + i - 1) & 7
		} else {
			parts++
		}
	}
	if parts == 0 {
		return true
	}
	for i := 0; i < 5; i++ {
		if pile*(i+1)%parts != 0 {
			return false
		}
	}
	return true
}

func atois(t *testing.T, s []string) []int {
	r := make([]int, len(s))
	for i, x := range s {
		v, err := strconv.Atoi(x)
		if err != nil {
			t.Fatal(err)
		}
		r[i] = v
	}
	return r
}
