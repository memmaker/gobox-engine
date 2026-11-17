package engine

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestNoUnseenMoves: in every dev fight a unit stands where the fight's
// events (Engaged, then each Walked) left it; a move the presenter is not
// told of plays the next blow as a lunge from the old square, across the
// map and through walls (meleefight's DevPlace after Engaged). Bug class:
// a board change with no event.
func TestNoUnseenMoves(t *testing.T) {
	for _, scene := range []string{"combat", "meleefight", "wildcombat", "castfight", "fleefight", "missile"} {
		for seed := uint32(1); seed <= 2; seed++ {
			e := testEngine(t)
			e.Explore = true
			e.DevAction = e.DevScene(scene)
			e.FixedSeed = &seed
			e.Init()
			e.Start(e.Play)
			at := map[int]rules.BoardPos{}
			sync := func() {
				if b := e.fightState(); b != nil {
					for _, u := range b.Units {
						at[u.ID] = u.At
					}
				}
			}
			for i := 0; i < 400; i++ {
				s := e.build()
				if _, ok := s.(game.Fighting); !ok {
					if i > 5 {
						break
					}
					press(e)
					sync()
					continue
				}
				if len(at) == 0 {
					sync()
				}
				var o game.Outcome
				var err error
				if myTurn(s) {
					o, err = e.doNow(game.Quick{All: true})
				} else if q := commonOf(s).Ask; q != nil {
					o, err = e.doNow(game.Reply{})
				} else {
					err = e.stepOn(&o, nil)
				}
				if err != nil {
					break
				}
				for _, ev := range o.Events {
					switch v := ev.(type) {
					case game.Engaged:
						for _, u := range v.At.Field.Units {
							at[u.ID] = u.At
						}
					case game.Walked:
						if len(v.Path) > 0 {
							at[v.Unit] = v.Path[len(v.Path)-1].At
						}
					}
				}
				if b := e.fightState(); b != nil {
					for _, u := range b.Units {
						if p, ok := at[u.ID]; ok && p != u.At && u.HP > 0 {
							t.Errorf("%s seed %d: unit %d %s at %v, its events left it at %v", scene, seed, u.ID, u.Name, u.At, p)
						}
						at[u.ID] = u.At
					}
				}
			}
		}
	}
}
