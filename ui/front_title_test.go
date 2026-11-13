package ui

import (
	"fmt"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// titleHost runs the presenter's calls on a game over a title Scene and
// logs what each did: the picture or credits page it put up, or done
// (the Continue sent).
type titleHost struct {
	RuleHost
	g   *classicGame
	f   *fake
	log *[]string
}

func newTitleHost(log *[]string) titleHost {
	f := &fake{st: play.Scene{}}
	return titleHost{g: &classicGame{Game: NewGame(f, game.Data{})}, f: f, log: log}
}

func (h titleHost) Call(f func(*classicGame)) {
	sent := len(h.f.sent)
	f(h.g)
	switch t := h.g.r.title; {
	case len(h.f.sent) > sent:
		*h.log = append(*h.log, "done")
	case t.credits:
		*h.log = append(*h.log, fmt.Sprint("credits", t.page))
	default:
		*h.log = append(*h.log, fmt.Sprint("pic", t.n))
	}
}

// A -N step shows credits page N-1, and a picture may follow it.
func TestIntroCreditPages(t *testing.T) {
	var log []string
	h := newTitleHost(&log)
	var in IntroUI
	intro := &game.Intro{Pictures: []int{1, -1, -2, 4}}
	for range 5 {
		IntroStep(h, &in, intro, true)
		IntroStep(h, &in, intro, false) // a frame without a key lets the next key cut again
	}
	if got, want := fmt.Sprint(log), "[pic1 credits0 credits1 pic4 done]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

// A game without a credits page (NoCredits) goes from its last picture to
// done: never an empty credits pane (PoD's credits are picture 3).
func TestIntroNoCredits(t *testing.T) {
	var log []string
	h := newTitleHost(&log)
	in := IntroUI{NoCredits: true}
	intro := &game.Intro{Pictures: []int{3, 1, 2}}
	for range 5 {
		if len(log) > 0 && log[len(log)-1] == "done" {
			break
		}
		IntroStep(h, &in, intro, true)
		IntroStep(h, &in, intro, false)
	}
	if got, want := fmt.Sprint(log), "[pic3 pic1 pic2 done]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
