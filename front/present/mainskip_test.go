package present

import (
	"gobox/front"
	"testing"
	"time"
)

// A key or click during the title's hold or fade ends it and is swallowed;
// once the menu is in, input goes through.
func TestMainSkip(t *testing.T) {
	p := &Presenter{main: true}
	if p.mm.fade() >= 1 {
		t.Fatal("fresh menu should be fading")
	}
	if !p.key(13, front.View{}) {
		t.Fatal("Enter during the fade should be taken")
	}
	if p.mm.fade() < 1 {
		t.Fatal("the fade should be over after the skip")
	}
	if p.mm.skip(false) {
		t.Fatal("a second key must not be swallowed")
	}
	p.mm.at = time.Now()
	if !p.Click(0, 0) || p.mm.fade() < 1 {
		t.Fatal("a click during the hold should end it and be taken")
	}
}
