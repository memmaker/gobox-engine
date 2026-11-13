package importer

import "gobox/adventure"

// The ending's pictures (the old rules/ending.go): the pool shatters, Tyranthraxus crumbles
// (PIC6 anims) and the feast in Shadowdale (BIGPIC6).
const (
	endPoolShatters       = 74
	endTyranthraxusGone   = 75
	endShadowdalePic      = 122
	poolShattersFrames    = 7 // the anims' frames, in art/dos/PIC6.txt
	tyranthraxusFrames    = 5
	endAnimFrameMs        = 150
	endPicFile, endBigPic = "PIC6", "BIGPIC6"
)

// endNote is what the ending says once it is over (programMenu's question, O03C:30FE).
const endNote = "You've won."

// end is COTAB's ending (O0AC end_game_text, the manifest's End): its five pages of text, two of them
// over the animated pictures and the last over Shadowdale's big one, a key after each. Once it is over the
// party is healed, the hall teaches (training 255) and the game is won (the old rules/ending.go).
func cotabEnd() *adventure.Sequence {
	pic := func(block, frames int, text string) adventure.SequenceStep {
		l := adventure.Layer{Art: adventure.Art{File: endPicFile, Block: block}, At: cotabWhole, FrameMs: endAnimFrameMs}
		for f := range frames {
			l.Frames = append(l.Frames, f)
		}
		return adventure.SequenceStep{Kind: adventure.StepPicture, Layers: []adventure.Layer{l}, Text: []string{text}}
	}
	return &adventure.Sequence{Steps: []adventure.SequenceStep{
		{Kind: adventure.StepText, Text: []string{"Tyranthraxus' spirit coalesces over the slain storm giant. 'You have defeated me. Were it not for the Amulet of Lythander, I could possess you and rob you of your victory. Still I can escape through the pool."}},
		pic(endPoolShatters, poolShattersFrames, "As you reach for the Pool of Radiance, he cries out, 'Keep the Gauntlet of Moander away from there, you will unleash dangerous energies. Stay back!' As the gauntlet contacts the pool, it contracts and shatters it."),
		pic(endTyranthraxusGone, tyranthraxusFrames, "'I am trapped without escape, you have succeeded where armies have not. Gloat while you may, Tyranthraxus is slain this day.' Before your eyes he crumbles into nothingness."),
		{Kind: adventure.StepText, Text: []string{
			"You are certain he is destroyed because your final bond fades away. The Curse of the Azure Bonds has finally been lifted from you! You are free at last!",
			"The Knights of Myth Drannor rush in, 'Congratulations, you have destroyed the Flamed One. With the power of Elminster, let us take you from this  foul place, to a fine feast.'"}},
		{Kind: adventure.StepPicture, Layers: []adventure.Layer{{Art: adventure.Art{File: endBigPic, Block: endShadowdalePic}, At: cotabWhole}},
			Text: []string{"You are teleported to Shadowdale, where festivities have already begun. A huge cheer goes up at your arrival. Gharri and Nacacia, arm in arm, yell congratulations from the nearby stands. 'You have won!'"}},
	}, After: []adventure.Effect{adventure.EffectHealParty, adventure.EffectOpenHall}, Note: endNote}
}
