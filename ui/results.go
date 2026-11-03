package ui

// Results is the after-combat page's data: the outcome and experience, the
// hirelings' shares, what the party takes, the members harmed and, when
// the party is destroyed, the end.
type Results struct {
	Outcome, Exp string
	Shares       []string
	Taken        string
	Harmed       []Harm
	End          string
}

// Harm is a member harmed: Name, and What.
type Harm struct{ Name, What string }

// Roles of a ResultsRun's text: the game maps them to its colours (the
// original's: green, cyan, light red).
const (
	RunGreen = iota
	RunCyan
	RunRed
)

// Run is Text at cell X, Y in role Role.
type Run struct {
	Text string
	X, Y int
	Role int
}

// ResultsTitle is the page's title.
func ResultsTitle(r Results) string {
	if r.End != "" {
		return "The END!"
	}
	return "After Combat"
}

// ResultsRuns lays r out on a canvas of cw x ch cells: the runs to write.
func ResultsRuns(r Results, cw, ch int) []Run {
	var out []Run
	w, y := cw-4, 2
	at := func(s string, x, role int) {
		if y < ch-2 {
			out = append(out, Run{s, x, y, role})
		}
	}
	para := func(s string) {
		for _, l := range WrapWords(s, w) {
			at(l, 2, RunGreen)
			y++
		}
		y++
	}
	if r.Outcome != "" {
		para(r.Outcome)
		para(r.Exp + " experience points.")
	}
	for _, s := range r.Shares {
		para(s)
	}
	if r.Taken != "" {
		para(r.Taken)
	}
	for _, h := range r.Harmed {
		at(h.Name[:min(len(h.Name), 15)], 2, RunCyan)
		at(h.What[:min(len(h.What), max(0, w-17))], 19, RunRed)
		y++
	}
	if r.End != "" {
		y++
		para(r.End)
	}
	return out
}
