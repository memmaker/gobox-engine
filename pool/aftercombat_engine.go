package pool

import (
	"errors"
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/game/play"
	"slices"
	"strconv"
	"strings"
)

func (e *Engine) duel() bool { return e.run.duelMode != 0 }

// endFight = AfterCombat's combatResult and cleanupCombatants (ovr06:04C9,
// :11AC): core.Rules EndFight, then the removed combatants' icons freed
// and what the results page reads kept (ac).
func (e *Engine) endFight() {
	before := e.partyIDs()
	icons := make([]core.Number, len(before))
	for i, id := range before {
		icons[i] = e.fact(id, core.IconFact{}).(core.Number)
	}
	e.as(e.sel) // a duel's experience is the selected member's level
	res := e.ruleset().EndFight(core.Fight{Terms: core.FightTerms{Duel: e.duel()}}, nil)
	after := e.partyIDs()
	for i, id := range before {
		if !slices.Contains(after, id) {
			e.freeIcon(byte(icons[i])) // ovr33:014E
		}
	}
	e.run.quit = res.Verdict == core.PartyLost && !e.duel()
	e.ac.won = res.Verdict == core.PartyWon
	e.ac.fled = res.Verdict == core.PartyFled
	e.ac.monsters = res.Foes
	e.ac.shared, e.ac.sharers, e.ac.loot = res.Shared, res.Sharers, res.Items
	if e.ac.won {
		e.ac.exp = res.Exp
	}
	e.run.monCount = 0
	e.sel = 0
}

// hiredShare = ovr06:12E8's page: the hired NPCs' shares were taken by
// EndFight; their page (if any) has its key.
func (e *Engine) hiredShare() {
	if !e.ac.shared {
		e.ret()
		return
	}
	e.newPage()
	e.blank(1, 1, winRight, winBottom)
	r := &Results{}
	for _, id := range e.ac.sharers {
		r.Shares = append(r.Shares, e.ruleset().Sheet(id).Name+" takes and hides his share.")
	}
	e.snap.results = r
	e.keyWait(cont{kind: kNop}) // press <enter>/<return> to continue
}

// showResult = ovr06:0916 (exp): the page (View.Results).
func (e *Engine) showResult(exp int32) {
	e.newPage()
	e.blank(1, 1, winRight, winBottom)
	r := &Results{}
	switch {
	case e.ac.monsters || e.duel():
		switch {
		case e.duel() && !e.ac.won:
			r.Outcome = "You have lost the duel."
			exp = 0
		case e.duel():
			r.Outcome = "You have won the duel."
		default:
			r.Outcome = "The party has won."
			e.songMusic(songVictory, 0) // extra
		}
	case e.ac.fled:
		r.Outcome = "The party has fled."
		exp = 0
	default:
		r.Outcome = "The party has found Treasure!"
	}
	s := "Each character receives "
	if e.duel() {
		s = "The duelist receives "
	}
	r.Exp = goldbox.Cut(s+strconv.Itoa(int(exp)), 0xFF)
	if l := e.purse(true); !l.Empty() {
		var got []string
		if l.Gold > 0 || l.Copper > 0 {
			got = append(got, moneyText(l.Gold, l.Copper))
		}
		for _, p := range []struct {
			n    uint64
			name string
		}{{l.Gems, " gems"}, {l.Jewelry, " jewelry"}} {
			if p.n > 0 {
				got = append(got, strconv.Itoa(int(p.n))+p.name)
			}
		}
		r.Taken = "The party takes " + strings.Join(got, ", ") + "."
	}
	for _, id := range e.partyIDs() { // extra: who came out of the fight dead or worse, or with harmful effects on them
		sh := e.ruleset().Sheet(id)
		var what []string
		if st := e.statusName(id); sh.Status != core.StatusOkay && sh.Status != core.StatusRunning && st != "" {
			what = append(what, st)
		}
		if h := e.word(id, core.FactHarmed); h != "" {
			what = append(what, h)
		}
		if len(what) > 0 {
			r.Harmed = append(r.Harmed, Harm{Name: sh.Name, What: strings.Join(what, ", ")})
		}
	}
	e.snap.results = r
}

// lootMenu = ovr06:0ECD: the treasure screen, the script's wait PendLoot
// (View.Loot) until LootLeave; what follows (pushed by the caller) goes on
// once the party leaves it.
func (e *Engine) lootMenu() {
	e.pw.anim = e.animInput
	e.lootRedraw()
	e.wait(PendLoot)
}

// lootRedraw = ovr06:101F: the picture animates while it waits.
func (e *Engine) lootRedraw() {
	e.DrawMainScreen()
	e.animInput = true
}

// TakeLoot = the treasure screen's Take (ovr06:0C06), on gobox's loot
// screen: item h of the treasure (State.Loot) for the
// selected member; all: every piece in turn until one does not fit. The
// money (NoItem) is the party's already (the purse is shared), so there is
// none to take; the vault's is banked by Take.
func (e *Engine) TakeLoot(h core.ItemHandle, all bool) error {
	if !e.waiting(PendLoot) {
		return core.ErrNotNow
	}
	if e.vault != nil && (all || h == core.NoItem) { // the vault's money is left there until taken
		e.must(core.BankLoot{})
	}
	if all {
		return e.takeLoot(core.TakeLoot{Who: e.selected(), All: true})
	}
	if h == core.NoItem && e.vault != nil {
		return nil
	}
	if h == core.NoItem {
		return &play.Refusal{} // the money line: the purse is shared, nothing to take, no toast
	}
	if !slices.Contains(e.goods(), h) {
		return core.ErrNotNow
	}
	return e.takeLoot(core.TakeLoot{Who: e.selected(), Item: h})
}

func (e *Engine) takeLoot(a core.TakeLoot) error {
	_, err := e.ruleset().Resolve(core.Fight{}, a, nil)
	return err
}

// LootDetect = the treasure screen's Detect: the selected member casts
// Detect Magic on it.
func (e *Engine) LootDetect() error {
	if !e.waiting(PendLoot) {
		return core.ErrNotNow
	}
	e.as(e.sel)
	if _, err := e.ruleset().Resolve(nil, core.DetectLoot{}, nil); errors.Is(err, core.ErrNotNow) {
		return err
	} else if err != nil {
		panic(fmt.Sprintf("detect loot: %v", err))
	}
	e.waitRedrawNow()
	return nil
}

// LootLeave = the treasure screen's Exit (ovr06:109A); with treasure left
// it asks ovr06's question first (askSure) until the player is sure.
func (e *Engine) LootLeave(sure bool) error {
	if !e.waiting(PendLoot) {
		return core.ErrNotNow
	}
	if len(e.goods()) > 0 && !sure && e.vault == nil { // the vault keeps what is left
		return e.askSure("There is still treasure left.  Do you want to go back and claim your treasure?", pendingCall{kind: callLoot}) // ovr06:0E7D, 0E9D
	}
	e.resume()
	return nil
}

// pushAfterCombat = AfterCombat (ovr06:1527): the results, then the treasure screen, which ends it.
func (e *Engine) pushAfterCombat() {
	e.setWorldWord(core.WordBeginFlag, 6)
	e.endFight()
	if !e.run.quit || e.duel() {
		e.push(cont{kind: kAfterShared})
		e.hiredShare()
		return
	}
	e.setWorldWord(core.WordCombatEnd, 0x80)
	e.Music(5) // Amiga 0x4602E
	e.newPage()
	e.blank(1, 1, winRight, winBottom)
	e.endPage("The monsters rejoice for the party has been destroyed", CellRect{X1: 2, Y1: 5, X2: 37, Y2: winBottom})
	e.push(cont{kind: kAfterEnd})
	e.keyWait(cont{kind: kNop}) // was PromptRow24's any key: its pause is only a fight event, so the page never waited and the menu came at once
}

// afterShared: the hired shares are taken; the results page.
func (e *Engine) afterShared() {
	e.showResult(e.ac.exp)
	if _, err := e.ruleset().Resolve(nil, core.BankLoot{}, nil); err != nil {
		panic(fmt.Sprintf("bank loot: %v", err))
	}
	e.push(cont{kind: kAfterResults})
	e.keyWait(cont{kind: kNop}) // the results page's press <enter>/<return> (ovr06:0916), before the treasure
}

// afterResults: the results page had its key; the treasure screen.
func (e *Engine) afterResults() {
	if len(e.ac.loot) == 0 { // extra: no treasure screen (its picture, View/Exit) when there is nothing to take: the money is the party's already
		e.afterEnd()
		return
	}
	e.push(cont{kind: kLootDone})
	e.lootMenu()
}

// afterEnd ends AfterCombat; then what follows it.
func (e *Engine) afterEnd() {
	for _, w := range []core.WorldWord{core.WordMonsterHit, core.WordPartyHit, core.WordPartyMove, core.WordNoLoot, core.WordDuel} {
		e.setWorldWord(w, 0)
	}
	e.ret()
}

// endPage is the text under "The END!" (textBlock into box r, cleared):
// View.Results.End (gobox wraps it), the text cursor back at its start.
func (e *Engine) endPage(text string, r CellRect) {
	e.snap.results = &Results{End: text, EndAt: r}
	e.msgCol, e.msgRow = r.X1, r.Y1
}

// afterCombat holds the ovr06 globals.
type afterCombat struct {
	monsters bool  // DS:443A: there were monsters to beat
	fled     bool  // DS:443B
	exp      int32 // DS:84D6
	won      bool  // DS:84DA
	shared   bool  // hired NPCs took shares (EndFight): their page
	sharers  []core.CharID
	loot     []core.ItemHandle // the treasure EndFight left (FightResult.Items)
}
