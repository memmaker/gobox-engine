package ecl

import "testing"

// TestSkipCounts: where a dialect's Skip differs from its commands' fixed
// operands, the difference is the original's and listed here; anything else
// is a typo in one of the tables.
func TestSkipCounts(t *testing.T) {
	// counted commands: parse_command reads none of a counted list
	counted := []byte{21, 37, 38, 43}
	// 52 Clock, 54 AddNPC: the handlers read 2, parse_command steps over 1.
	// 66 Quest, 67 CastKnock, 68 TavernGame: the port's, unknown to the originals.
	v2 := byteSet(append(counted, 52, 54, 66, 67, 68,
		31)...) // the port's pager (1 operand) where the original's table has 2
	v3 := byteSet(append(counted, 52, 54, 66, 67, 68,
		39, 49)...) // TreasureOf, MenuKeys: counted in POD
	known := map[VersionID]map[byte]bool{
		1: byteSet(append(counted, 52, 54, 66, 67, 68, 31, 12)...), // 12 SetupMonster: SkipEclCmd steps over 2 of its 3
		2: v2,
		3: v3,
		// BR's skip() (buckrogers-go/rules/ecl.go): 12, 31, 52, 54 differ from
		// Fixed, and the counted 21, 37, 38, 39, 43, 49 step over none
		// of their operands. (59 and 60 are OpNothing: none, see V4.)
		4: byteSet(12, 31, 52, 54, 21, 37, 38, 39, 43, 49),
	}
	for id, ok := range known {
		if ok == nil {
			continue
		}
		v := Versions[id]
		for b := range 256 {
			c := v.Cmds[b]
			want := c.Fixed
			if !c.Known {
				want = 0
			}
			if int(v.Skip[b]) != want && !ok[byte(b)] {
				t.Errorf("V%d command %d (%s): Skip %d, Fixed %d", id, b, OpName(c.Op), v.Skip[b], want)
			}
		}
	}
}

func byteSet(bs ...byte) map[byte]bool {
	m := map[byte]bool{}
	for _, b := range bs {
		m[b] = true
	}
	return m
}
