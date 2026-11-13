package importer

// StartSequence is PoD's start-up, the adventure's Start (the pod import writes
// it into the manifest; the engine plays what SetStart is given): CONTROL.EXE's
// VACE INTRO animation with its tune P1, then TITLE 3 (the credits), 1
// (Forgotten Realms) and 2 (the painting), whole screen; the painting stays
// under the main menu.
func StartSequence() *Sequence {
	pic := func(b int) SequenceStep {
		return SequenceStep{Kind: StepPicture, Whole: true, Credits: b == titleCredits, Layers: []Layer{{Art: Art{File: "TITLE", Block: b}, At: wholePage}}}
	}
	return &Sequence{
		Steps:    []SequenceStep{{Kind: StepAnim, Anim: "INTRO", Song: "P1"}, pic(3), pic(1), pic(2)},
		Backdrop: pic(2).Layers,
	}
}

// EndSequence is the adventure's End: VACE FINAL with P3 (CONTROL.EXE after
// the finale's save and quit, errorlevel 23, O043:0000).
func EndSequence() *Sequence {
	return &Sequence{Steps: []SequenceStep{{Kind: StepAnim, Anim: "FINAL", Song: "P3"}}}
}

// titleCredits is the TITLE block that is the credits picture.
const titleCredits = 3

// wholePage is the 40x25 page.
var wholePage = Cells{X2: 39, Y2: 24}
