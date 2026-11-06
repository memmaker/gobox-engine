package ui

import "testing"

func TestResultsRuns(t *testing.T) {
	r := Results{Outcome: "Victory.", Exp: "Each character receives 10", Harmed: []Harm{{"Bob", "Poisoned"}}, End: "Dead."}
	runs := ResultsRuns(r, 40, 25)
	if len(runs) != 6 || runs[0].Y != 2 || runs[2].Text != "experience points." || runs[3].Role != RunCyan || runs[4].Role != RunRed || runs[4].X != 19 {
		t.Fatalf("%+v", runs)
	}
	if ResultsTitle(r) != "The END!" || ResultsTitle(Results{}) != "After Combat" {
		t.Fatal("title")
	}
	if len(ResultsRuns(r, 40, 6)) != 1 {
		t.Fatal("clip")
	}
}
