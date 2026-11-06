package present

import (
	"testing"

	"gobox/front"
)

type trainGame struct {
	front.Game
	calls []string
}

func (g *trainGame) CloseTraining() error { g.calls = append(g.calls, "close"); return nil }
func (g *trainGame) OpenTraining(n int) error {
	g.calls = append(g.calls, "open"+string(rune('0'+n)))
	return nil
}
func (g *trainGame) TrainingView() *front.Training { return &front.Training{} }

// TestTrainingSwitch: another member in the training hall is OpenTraining
// for them alone, no CloseTraining (that leaves a hall the game opened).
func TestTrainingSwitch(t *testing.T) {
	p := &Presenter{training: true}
	if !p.serviceAct(front.Action{Kind: front.ActLead, N: 2}) {
		t.Fatal("not taken")
	}
	g := &trainGame{}
	for _, c := range p.Calls() {
		c(g)
	}
	if len(g.calls) != 1 || g.calls[0] != "open2" {
		t.Fatalf("%v", g.calls)
	}
}

// TestTrainingSwitchGameHall: the switch works in a hall the game opened
// (ModeTraining: the View has Training, the front's Train never ran).
func TestTrainingSwitchGameHall(t *testing.T) {
	p := &Presenter{}
	p.shownView.Training = &front.Training{Why: "NOT ENOUGH EXPERIENCE"}
	if !p.serviceAct(front.Action{Kind: front.ActLead, N: 1}) {
		t.Fatal("not taken")
	}
	g := &trainGame{}
	for _, c := range p.Calls() {
		c(g)
	}
	if len(g.calls) != 1 || g.calls[0] != "open1" {
		t.Fatalf("%v", g.calls)
	}
}
