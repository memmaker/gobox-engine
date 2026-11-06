package classic

import (
	"testing"

	"gobox/game"
)

// TestTraceParty: a trace line names each member's HP and status.
func TestTraceParty(t *testing.T) {
	if got := traceParty([]game.Member{{Name: "Brann", HP: 0, MaxHP: 81, Status: "Unconscious"}}); got != " party=[Brann 0/81 Unconscious]" {
		t.Fatalf("%q", got)
	}
}

// TestTraceStamp: trace lines carry the time to the millisecond.
func TestTraceStamp(t *testing.T) {
	if s := traceStamp(); len(s) != len("15:04:05.000") || s[8] != '.' {
		t.Fatalf("%q", s)
	}
}
