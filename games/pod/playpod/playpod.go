// Package playpod starts Pools of Darkness on an imported adventure folder
// (`gobox import pod`) for cmd/gobox and cmd/pod through the shared pool host
// (gobox/pool/playpool).
package playpod

import (
	"gobox/app"
	"gobox/core"
	"gobox/pool/playpool"
)

// Play runs the game of s.Adventure until the window closes; the saves are
// the user's (playpool.HomeSaves), Add from Secret reads Silver Blades's.
func Play(s app.Start) error {
	saves, err := playpool.HomeSaves("pod", s.Adventure)
	if err != nil {
		return err
	}
	return playpool.Play(s, playpool.Game{
		Title: "Pools of Darkness", W: 80 * 8, H: 50 * 8, Record: core.Layout510,
		SaveDir: saves,
		Foreign: playpool.HomeRosters(map[string]core.RecordLayout{"sotsb": core.Layout439}),
	})
}
