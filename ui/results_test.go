package ui

import "testing"

func TestResultsRuns(t *testing.T) {
	r := Results{Outcome: "Victory.", Exp: "10", Harmed: []Harm{{"Bob", "Poisoned"}}, End: "Dead."}
	runs := ResultsRuns(r, 40, 25)
	if len(runs) != 5 || runs[0].Y != 2 || runs[2].Role != RunCyan || runs[3].Role != RunRed || runs[3].X != 19 {
		t.Fatalf("%+v", runs)
	}
	if ResultsTitle(r) != "The END!" || ResultsTitle(Results{}) != "After Combat" {
		t.Fatal("title")
	}
	if len(ResultsRuns(r, 40, 6)) != 1 {
		t.Fatal("clip")
	}
}
