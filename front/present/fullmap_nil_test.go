package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// No 16x16 map in the State: no full map, never one without its squares.
func TestFullMapNoMap(t *testing.T) {
	if f := FullMap(&front.FullMap{}, game.Automap{}, game.Where{}); f != nil {
		t.Fatalf("FullMap without a map = %+v, want nil", f)
	}
}
