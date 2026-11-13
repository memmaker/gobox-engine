// Package playsotsb starts Secret of the Silver Blades on an imported
// adventure folder (`gobox import sotsb`) for cmd/gobox through the shared
// pool host (gobox/engine/playpool).
package playsotsb

import (
	"gobox/app"
	"gobox/engine/playpool"
	"gobox/rules"
)

// Play runs the game of s.Adventure until the window closes; the saves are
// the user's (playpool.HomeSaves), Add from Curse reads Curse's.
func Play(s app.Start) error {
	saves, err := playpool.HomeSaves("sotsb", s.Adventure)
	if err != nil {
		return err
	}
	return playpool.Play(s, playpool.Game{
		Title: "Secret of the Silver Blades", W: 80 * 8, H: 50 * 8, Record: rules.Layout439,
		SaveDir: saves,
		Foreign: playpool.HomeRosters(map[string]rules.RecordLayout{"cotab": rules.Layout422}),
	})
}
