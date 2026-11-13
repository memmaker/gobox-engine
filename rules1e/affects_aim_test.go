package rules1e

import "testing"

// aimScene: a monster (Party[1], computer run) next to the caster, its foe.
func aimScene(t *testing.T, handler HandlerName, typ AffectType) (*RuleState, *Character, *Character) {
	t.Helper()
	m := newChar("MONSTER", 8)
	e := cotabFirstScene(t, m)
	foe := e.Party[0]
	m.Side, m.Auto, m.Control = 1, 1, 0x80+50
	e.Tables.Affects[typ] = handler
	e.Tables.Spells[aimTouch] = e.Tables.Spells[aimGaze] // a later game's record: a chosen foe
	e.Who = 1
	e.CombatRound = 1
	return e, m, foe
}

// TestAimHandlers (T4d, M1): the aimed monster attacks reach their foe.
func TestAimHandlers(t *testing.T) {
	for _, k := range []struct {
		h   HandlerName
		typ AffectType
		hit func(foe *Character) bool
	}{
		{"breathFireSotsb", affBreathFireSotsb, func(f *Character) bool { return f.HP < f.MaxHP }},
		{"breathWeaponSotsb", AffBreathWeapon, func(f *Character) bool { return f.HP < f.MaxHP || f.Okay == 0 }},
		{"blackBreathPod", affBlackBreathPod, func(f *Character) bool { return f.HP < f.MaxHP || f.Okay == 0 }},
		{"confusingGaze", affConfusingGaze, nil},
		{"stoningGazePod", affStoneGaze, func(f *Character) bool { return f.Status == StatusStoned }},
	} {
		hit := false
		for seed := uint32(1); seed < 30 && !hit; seed++ {
			e, m, foe := aimScene(t, k.h, k.typ)
			e.RandSeed = seed
			a := &Affect{AffectRecord: AffectRecord{Type: k.typ, Data: breathsPerFight}}
			m.Affects = append(m.Affects, a)
			e.callAffect(k.typ, m, a, 0)
			if k.hit == nil {
				hit = e.FindAffect(foe, affConfused) != nil
			} else {
				hit = k.hit(foe)
			}
		}
		if !hit {
			t.Errorf("%s: never reached the foe", k.h)
		}
	}
}

// TestGazeImmune: a gaze immunity stops the stoning gaze.
func TestGazeImmune(t *testing.T) {
	for seed := uint32(1); seed < 30; seed++ {
		e, m, foe := aimScene(t, "stoningGazePod", affStoneGaze)
		e.RandSeed = seed
		e.AddAffect(foe, affGazeImmunePod, 0, 0, 0)
		a := &Affect{AffectRecord: AffectRecord{Type: affStoneGaze}}
		e.callAffect(affStoneGaze, m, a, 0)
		if foe.Status == StatusStoned {
			t.Fatalf("seed %d: gaze-immune foe stoned", seed)
		}
	}
}
