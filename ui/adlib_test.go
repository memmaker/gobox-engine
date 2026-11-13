package ui

import (
	"testing"

	"gobox/audio"
	"gobox/game"
)

// fakeSnd records what the presenter asks of the driver.
type fakeSnd struct {
	game.Sound
	on      bool
	songs   []int
	effects []int
}

func (f *fakeSnd) Song(n, _ int, _ bool) bool { f.songs = append(f.songs, n); return true }
func (f *fakeSnd) Effect(n int, _ bool)       { f.effects = append(f.effects, n) }
func (f *fakeSnd) Track() (bool, int)         { return f.on, 0 }

func TestAdLibRouting(t *testing.T) {
	f := &fakeSnd{on: true}
	p := &Presenter{Snd: f, AdLib: &audio.AdLib{}}
	p.Play([]Moment{{Report: game.Report{Kind: game.EvSong, N: 3}}, {Report: game.Report{Kind: game.EvSound, N: 9}}})
	p.playing()
	if len(f.songs) != 1 || f.songs[0] != 3 || len(f.effects) != 1 || f.effects[0] != 9 { // an effect the player lacks goes to the driver
		t.Fatalf("songs %v effects %v", f.songs, f.effects)
	}
	f = &fakeSnd{}
	p = &Presenter{Snd: f, AdLib: &audio.AdLib{}}
	p.Play([]Moment{{Report: game.Report{Kind: game.EvSound, N: 9}}})
	p.playing()
	if len(f.effects) != 0 { // the sound setting is off
		t.Fatalf("played %v", f.effects)
	}
}
