package ui

import (
	"strings"
	"sync"
	"testing"

	"gobox/asset"
)

// TestAmigaTwinsEveryGame: every pool head any game's DOS art names has its
// Amiga twin, whichever game plays (POR's cards kept the heads from the
// other games' art on F12: twins came from the playing game alone).
// noAmigaHead are the heads whose games' Amiga art has none.
func TestAmigaTwinsEveryGame(t *testing.T) {
	noAmigaHead := map[string]bool{"head_033": true, "head_035": true}
	defer func(id string) { GameID = id }(GameID)
	GameID = "por"
	twinsOnce, twins = sync.Once{}, nil
	for _, n := range asset.SpriteNames(Series, "portraits") {
		if strings.HasPrefix(n, "head_") && !noAmigaHead[n] && amigaTwin(n) == "" {
			t.Errorf("%s: no Amiga twin", n)
		}
	}
}
