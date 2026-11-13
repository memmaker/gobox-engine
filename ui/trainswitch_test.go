package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestTrainingSwitch: another member in the training hall is OpenTraining
// for them alone, no CloseTraining (that leaves a hall the game opened).
func TestTrainingSwitch(t *testing.T) {
	p := &Presenter{training: true}
	if !p.serviceAct(Action{Kind: ActLead, N: 2}) {
		t.Fatal("not taken")
	}
	g, f := playGame(play.Trading{})
	for _, c := range p.Calls() {
		c(g)
	}
	if len(f.sent) != 1 || f.sent[0] != (play.OpenTraining{Member: 2}) {
		t.Fatalf("%v", f.sent)
	}
}

// TestTrainingSwitchGameHall: the switch works in a hall the game opened
// (ModeTraining: the View has Training, the front's Train never ran).
func TestTrainingSwitchGameHall(t *testing.T) {
	g, f := playGame(play.Trading{Kind: play.TradeTraining, Training: &game.Training{Why: map[int]string{1: "NOT ENOUGH EXPERIENCE"}}})
	p := &Presenter{}
	p.SetEngine(g)
	if !p.serviceAct(Action{Kind: ActLead, N: 1}) {
		t.Fatal("not taken")
	}
	for _, c := range p.Calls() {
		c(g)
	}
	if len(f.sent) != 1 || f.sent[0] != (play.OpenTraining{Member: 1}) {
		t.Fatalf("%v", f.sent)
	}
}
