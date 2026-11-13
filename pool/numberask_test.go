package pool

import (
	"gobox/game/play"
	"testing"
)

// TestNumberAskHuge: a "How much?" with a huge Max publishes a bounded
// state (no option or Allowed entry per amount) and takes its answer
// (BUG-CLASSES 51).
func TestNumberAskHuge(t *testing.T) { numberAskHuge(t, testEngine(t)) }

// TestNumberAskHugeCOTAB: the same on cotab's imported folder.
func TestNumberAskHugeCOTAB(t *testing.T) { numberAskHuge(t, advEngine(t, advEnvCOTAB)) }

// numberAskHuge is TestNumberAskHuge on e.
func numberAskHuge(t *testing.T, e *Engine) {
	t.Helper()
	e.Init()
	e.Start(e.Play)
	v := testField(t, e, fieldSurprise)
	got := func() int { return int(e.host.Get(v)) }
	e.host.Set(v, 0xFFFF)
	e.askNumber(askDone{kind: dNumber, dst: v})
	e.eclWait.ask.Max = 1 << 30
	e.openAsk()
	cm := commonOf(e.build())
	if q := cm.Ask; q == nil || !q.Number || len(q.Replies) > 16 || len(cm.Allowed.Plain) > 32 {
		t.Fatalf("published %+v, %d allowed", q, len(cm.Allowed.Plain))
	}
	if _, err := e.doNow(play.Reply{N: 1234}); err != nil || got() != 1234 {
		t.Fatalf("answer: %v, got %d", err, got())
	}
	e.askNumber(askDone{kind: dNumber, dst: v}) // class 52: the typed Text, N left 0
	e.openAsk()
	if _, err := e.doNow(play.Reply{Text: "100"}); err != nil || got() != 100 {
		t.Fatalf("typed 100: %v, got %d", err, got())
	}
}
