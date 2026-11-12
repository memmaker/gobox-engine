package pool

import (
	"testing"

	"gobox/core"
	"gobox/game"
	"gobox/game/play"
)

// valuedOffer is the state's offer with a value (nil: none).
func valuedOffer(t *testing.T, e *Engine) *game.Offer {
	t.Helper()
	tr, ok := e.State().(play.Trading)
	if !ok {
		t.Fatalf("not trading: %T", e.State())
	}
	for i := range tr.Offers {
		if tr.Offers[i].Value > 0 {
			return &tr.Offers[i]
		}
	}
	return nil
}

func purseGems(e *Engine) int { return int(e.state().Purse.Gems) }

// carriedCount is the first member's items, the party inventory among them.
func carriedCount(e *Engine) int { return len(commonOf(e.State()).Party[0].Items) }

// TestAppraiseFlow: clicking Appraise values the gem and neither sells nor
// loses it; Keep, selling and leaving then each settle it. Through Do
// (classic's Appraise and KeepAppraised are its own: front/classic).
func TestAppraiseFlow(t *testing.T) {
	type steps struct {
		value func() error
		keep  func(bool) error
	}
	for _, via := range []string{"Do"} {
		t.Run(via, func(t *testing.T) {
			e := devPlay(t, "shop")
			e.state().Purse.Gems, e.state().Purse.Jewelry = 3, 0
			st := steps{}
			st.value = func() error {
				tr := e.State().(play.Trading)
				_, err := e.Do(play.Appraise{Member: tr.Offers[0].Member, Offer: tr.Offers[0].Offer})
				return err
			}
			st.keep = func(k bool) error {
				o := valuedOffer(t, e)
				_, err := e.Do(play.Appraise{Member: o.Member, Offer: o.Offer, Keep: k})
				return err
			}
			gold, items := e.state().Purse.Gold, carriedCount(e)
			value := func() *game.Offer {
				t.Helper()
				if err := st.value(); err != nil {
					t.Fatal("value:", err)
				}
				o := valuedOffer(t, e)
				if o == nil {
					t.Fatal("no offer valued")
				}
				return o
			}
			o := value()
			if e.state().Purse.Gold != gold || carriedCount(e) != items || purseGems(e) != 2 {
				t.Fatalf("Appraise settled it: gold %d->%d items %d->%d gems %d", gold, e.state().Purse.Gold, items, carriedCount(e), purseGems(e))
			}
			if !o.CanKeep {
				t.Fatal("no room to keep")
			}
			if err := st.keep(true); err != nil {
				t.Fatal("keep:", err)
			}
			if carriedCount(e) != items+1 || e.state().Purse.Gold != gold || valuedOffer(t, e) != nil {
				t.Fatalf("keep: items %d->%d gold %d->%d", items, carriedCount(e), gold, e.state().Purse.Gold)
			}
			value()
			if err := st.keep(false); err != nil {
				t.Fatal("sell:", err)
			}
			if carriedCount(e) != items+1 || e.state().Purse.Gold <= gold || valuedOffer(t, e) != nil {
				t.Fatalf("sell: items %d gold %d->%d", carriedCount(e), gold, e.state().Purse.Gold)
			}
			gold = e.state().Purse.Gold
			value()
			if _, err := e.Do(play.Leave{}); err != nil {
				t.Fatal("leave:", err)
			}
			if e.state().Purse.Gold <= gold || e.valued != core.NoItem {
				t.Fatalf("leaving did not sell: gold %d->%d", gold, e.state().Purse.Gold)
			}
		})
	}
}
