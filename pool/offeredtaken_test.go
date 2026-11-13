package pool

import (
	"reflect"
	"testing"

	"gobox/core"
	"gobox/game/play"
)

// TestOfferedTaken (bug class: offered, then refused "not now"): every
// action a building's State allows (Allowed.Plain), done on a fresh scene,
// is never refused core.ErrNotNow (Do's refusal of it); the temple's Leave
// was (65ce639). Run the same over any dev scene to scan for the class.
// (The shop's UseItem was, for items with no use: allowed offers Use only for a use.)
func TestOfferedTaken(t *testing.T) { offeredTaken(t, testEngine) }

// TestOfferedTakenAdv: the same on the imported cotab, sotsb and pod folders
// (each skips without its POOL_ADV_ variable).
func TestOfferedTakenAdv(t *testing.T) {
	for _, env := range []string{advEnvCOTAB, advEnvSOTSB, advEnvPOD} {
		t.Run(env, func(t *testing.T) {
			offeredTaken(t, func(t *testing.T) *Engine { return advEngine(t, env) })
		})
	}
}

func offeredTaken(t *testing.T, base func(*testing.T) *Engine) {
	notNow := error(refusal(core.ErrNotNow))
	for _, name := range []string{"shop", "temple", "hall", "hallpoor", "appraise"} {
		acts := commonOf(devPlayOn(t, base(t), name).build()).Allowed.Plain
		if len(acts) == 0 {
			t.Errorf("%s: nothing offered", name)
		}
		for _, a := range acts {
			if _, quit := a.(play.Quit); quit {
				continue
			}
			if _, err := devPlayOn(t, base(t), name).doNow(a); reflect.DeepEqual(err, notNow) {
				t.Errorf("%s: %T%+v offered but refused: %v", name, a, a, err)
			}
		}
	}
}
