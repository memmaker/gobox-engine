package front

import (
	"fmt"
	"testing"
)

type titleGame struct {
	Game
	log *[]string
}

func (g titleGame) TitlePicture(n int)    { *g.log = append(*g.log, fmt.Sprint("pic", n)) }
func (g titleGame) TitleCredits(page int) { *g.log = append(*g.log, fmt.Sprint("credits", page)) }
func (g titleGame) TitleDone()            { *g.log = append(*g.log, "done") }

type titleHost struct {
	Host
	g titleGame
}

func (h titleHost) Call(f func(Game)) { f(h.g) }

// A -N step shows credits page N-1, and a picture may follow it.
func TestIntroCreditPages(t *testing.T) {
	var log []string
	h := titleHost{g: titleGame{log: &log}}
	var in IntroUI
	intro := &Intro{Pictures: []int{1, -1, -2, 4}}
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
	h := titleHost{g: titleGame{log: &log}}
	in := IntroUI{NoCredits: true}
	intro := &Intro{Pictures: []int{3, 1, 2}}
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
