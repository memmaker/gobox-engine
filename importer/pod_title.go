package importer

import "gobox/adventure"

// StartSequence is PoD's start-up, the adventure's Start (the pod import writes
// it into the manifest; the engine plays what SetStart is given): CONTROL.EXE's
// VACE INTRO animation with its tune P1, then TITLE 3 (the credits), 1
// (Forgotten Realms) and 2 (the painting), whole screen; the painting stays
// under the main menu.
func StartSequence() *adventure.Sequence {
	pic := func(b int) adventure.SequenceStep {
		return adventure.SequenceStep{Kind: adventure.StepPicture, Whole: true, Credits: b == titleCredits, Layers: []adventure.Layer{{Art: adventure.Art{File: "TITLE", Block: b}, At: wholePage}}}
	}
	return &adventure.Sequence{
		Steps:    []adventure.SequenceStep{{Kind: adventure.StepAnim, Anim: "INTRO", Song: "P1"}, pic(3), pic(1), pic(2)},
		Backdrop: pic(2).Layers,
	}
}

// EndSequence is the adventure's End: VACE FINAL with P3 (CONTROL.EXE after
// the finale's save and quit, errorlevel 23, O043:0000).
func EndSequence() *adventure.Sequence {
	return &adventure.Sequence{Steps: []adventure.SequenceStep{{Kind: adventure.StepAnim, Anim: "FINAL", Song: "P3"}}}
}

// titleCredits is the TITLE block that is the credits picture.
const titleCredits = 3

// wholePage is the 40x25 page.
var wholePage = adventure.Cells{X2: 39, Y2: 24}
