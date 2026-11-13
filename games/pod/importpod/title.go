package importpod

import (
	"gobox/importer"
)

// StartSequence is PoD's start-up, the adventure's Start (importpod writes
// it into the manifest; the engine plays what SetStart is given): CONTROL.EXE's
// VACE INTRO animation with its tune P1, then TITLE 3 (the credits), 1
// (Forgotten Realms) and 2 (the painting), whole screen; the painting stays
// under the main menu.
func StartSequence() *importer.Sequence {
	pic := func(b int) importer.SequenceStep {
		return importer.SequenceStep{Kind: importer.StepPicture, Whole: true, Layers: []importer.Layer{{Art: importer.Art{File: "TITLE", Block: b}, At: wholePage}}}
	}
	return &importer.Sequence{
		Steps:    []importer.SequenceStep{{Kind: importer.StepAnim, Anim: "INTRO", Song: "P1"}, pic(3), pic(1), pic(2)},
		Backdrop: pic(2).Layers,
	}
}

// EndSequence is the adventure's End: VACE FINAL with P3 (CONTROL.EXE after
// the finale's save and quit, errorlevel 23, O043:0000).
func EndSequence() *importer.Sequence {
	return &importer.Sequence{Steps: []importer.SequenceStep{{Kind: importer.StepAnim, Anim: "FINAL", Song: "P3"}}}
}

// wholePage is the 40x25 page.
var wholePage = importer.Cells{X2: 39, Y2: 24}
