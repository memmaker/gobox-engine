package ui

import (
	"testing"

	"gobox/game"
)

// TestFightBoardFollowsArt: the drawn fight board is made again when the
// art changes (F12: ran), so the units' pictures show the new art; before,
// its key held no art and a fight kept the old pictures (sotsb).
func TestFightBoardFollowsArt(t *testing.T) {
	g := NewGame(&fake{st: game.Exploring{}}, game.Data{})
	p := presenterOf(g)
	p.board = &game.Combat{}
	p.drawnBoard()
	before := p.fightBoard
	p.drawnBoard()
	if p.fightBoard != before {
		t.Fatal("the board was made again with nothing changed")
	}
	p.ran++ // the art changed
	p.drawnBoard()
	if p.fightBoard == before {
		t.Fatal("the board kept the old art's pictures")
	}
}
