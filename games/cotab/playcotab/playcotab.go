// Package playcotab starts Curse of the Azure Bonds on an imported adventure
// folder (`gobox import cotab`) for cmd/gobox through the shared pool host
// (gobox/engine/playpool).
package playcotab

import (
	"gobox/app"
	"gobox/engine/playpool"
	"gobox/rules"
)

// Play runs the game of s.Adventure until the window closes; the saves are
// the user's (playpool.HomeSaves), Add from Pool of Radiance reads POR's.
func Play(s app.Start) error {
	saves, err := playpool.HomeSaves("cotab", s.Adventure)
	if err != nil {
		return err
	}
	return playpool.Play(s, playpool.Game{
		Title: "Curse of the Azure Bonds", W: 320, H: 200, Record: rules.Layout422,
		SaveDir: saves,
		Foreign: playpool.HomeRosters(map[string]rules.RecordLayout{"por": rules.Layout285}),
	})
}
