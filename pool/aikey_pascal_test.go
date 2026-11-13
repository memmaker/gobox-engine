package pool

import (
	"bufio"
	"fmt"
	"os"
	"rules1e"
	"strings"
	"testing"

	"gobox/core"
	"gobox/game/play"
)

// TestAiKeysAgainstPascal compares aiKeys with CombatAi.CheckKeystroke
// (rules1e testdata/por-aikey.txt, por-re vecakey): no key, Alt-M, Space or
// Alt-X for a player character, an NPC and an animated member, all on Quick,
// the turn the first's or the NPC's, quick magic off or on. quickfight is Auto,
// quickMagic magicOn, initiative Combat.Delay.
func TestAiKeysAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-aikey.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := quickStarted(t)
	if len(e.state().Party) < 3 {
		t.Fatal("the quick start has fewer than three members")
	}
	m := e.state().Party[:3]
	e.state().Party = m
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		for _, c := range m {
			c.Auto, c.Control, c.Status = 1, 0, rules1e.StatusOkay
			c.Combat = &rules1e.CombatInfo{Delay: 5}
		}
		m[1].Control = 0x80
		m[2].Status = rules1e.StatusAnimated
		e.state().MagicOn = byte(k[2])
		e.state().Fight.Key = rules1e.AiKey(k[0])
		c := m[k[1]-1]
		r := e.state().AiKeys(c)
		got := fmt.Sprintf("%d %d %d %d %d %d %d", rules1e.B2i(r), e.state().MagicOn, m[0].Auto, m[1].Auto, m[2].Auto, c.Combat.Delay, e.state().Fight.Key)
		if got != val {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// play.Manual and play.QuickMagic set the key the computer turn polls; Quick lasts across battles until taken back.
func TestAiKeyActions(t *testing.T) {
	e := quickStarted(t)
	run := func(f func() error) error { return f() }
	for _, c := range []struct {
		a    play.Action
		want core.AutoKey
	}{{play.Manual{}, core.AutoKeyTakeOver}, {play.QuickMagic{}, core.AutoKeyQuickMagic}} {
		e.aiKey = core.AutoKeyNone
		if ok, err := e.combatDo(c.a, play.Allowed{}, run); !ok || err != nil || e.aiKey != c.want {
			t.Errorf("%T: taken %v, err %v, key %d", c.a, ok, err, e.aiKey)
		}
	}
	e.state().Party[0].Auto = 1
	if !e.state().PartyAuto() {
		t.Error("a member on Quick is not seen")
	}
}
