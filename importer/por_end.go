package importer

// songEnding is the finale's song (pool songEnding).
const songEnding = "15"

// end is POR's finale (ovr19:02A4, the manifest's End): the FINAL pictures in the order the
// original draws them, a key after each. The dragon's death (FINAL 1, its animation block 2),
// Tyranthraxus (3 with the party 4..6 and his animation 7), the pool (3 again, animation 8).
// Each step says its text over the pictures; the party pictures show only for as many members
// as the party has (Layer.MinParty), the animations' frames last as the message delay says
// (Layer.MsgDelayed).
func porEnd() *Sequence {
	at := func(x, y int) Cells { return Cells{X1: x, Y1: y, X2: 39, Y2: 24} }
	page := func(block, x, y int) Layer {
		return Layer{Art: Art{File: finalFile, Block: block}, At: at(x, y)}
	}
	anim := func(block, x, y, frames int) Layer {
		l := page(block, x, y)
		for f := range frames {
			l.Frames = append(l.Frames, f)
		}
		l.FrameMs = animFrameMs
		l.MsgDelayed = true
		return l
	}
	member := func(block, x, y, of int) Layer { // shown with at least of members
		l := page(block, x, y)
		l.MinParty = of
		return l
	}
	party := []Layer{page(3, 0, 0), member(4, 0, 0xB, 2), member(5, 0xD, 9, 3), member(6, 0, 9, 4)}
	return &Sequence{Steps: []SequenceStep{
		{Kind: StepPicture, Layers: []Layer{page(1, 0, 0), anim(2, 2, 4, 4)}, Song: songEnding,
			Text: []string{"Mortally wounded, the dragon roars! The spirit of Tyranthraxus flares up from the dragon's body."}},
		{Kind: StepPicture, Layers: append(append([]Layer{}, party...), anim(7, 2, 0, 6)),
			Text: []string{"\"Fools, you have but slain the body I possessed.  I cannot be defeated!\" With the power of the pool of radiance which I moved here, to my lair, I will still rule, by possessing one of you!\""}},
		{Kind: StepPicture, Layers: append(append([]Layer{}, party...), anim(8, 2, 0, 6)),
			Text: []string{"\"No Lord Bane!  I can still rule here!  I have not failed.  Do not call me back through the pool!\" Noooo..."}},
	}}
}

const (
	finalFile   = "FINAL5" // the finale's pictures: FINAL<set> of the set the game ends in
	animFrameMs = 150      // one frame of an animation (the original: its Delay by the message delay)
)
