package engine

import (
	"fmt"
	"os"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// The session words (the phase, the clock, the party's count, the acting member, the effect table,
// camp and the hall): the engine's values at each moment, recorded on the code before they left the
// ruleset's generic words (WordIs, SetWord; RECORD_WORLD=1 prints them).

// The observations, by the engine's own accessors.
func obsPhase(e *Engine) int                    { return int(e.phase) }
func obsMinute(e *Engine) int                   { return e.clockPart(rules.ClockMinuteOfDay) }
func obsCount(e *Engine) int                    { return e.memberCount() }
func obsActing(e *Engine) int                   { return int(e.ruleset().Check(rules.ActingFor{}, nil).Value) }
func obsEffect(e *Engine, i int) (uint16, bool) { return e.effect(i) }

// sessionWant is each scene's recorded observations.
var sessionWant = map[string]string{
	"title":          "phase 0 camp false hall false inParty true",
	"exploring":      "phase 3 camp false hall false inParty false",
	"quickstart":     "phase 0 camp false hall false inParty false",
	"camp":           "phase 2 camp true hall false inParty false",
	"afterCamp":      "phase 0 camp false hall false inParty false",
	"fight":          "phase 5 camp false hall false inParty false",
	"winAsk":         "phase 5 camp false hall false inParty false",
	"afterFight":     "phase 6 camp false hall false inParty false",
	"afterFightSeen": "6",
	"shop":           "phase 1 camp false hall false inParty false",
	"temple":         "phase 1 camp false hall false inParty false",
	"hallMenu":       "phase 0 camp false hall false inParty true",
	"hallOpen":       "phase 0 camp false hall true inParty false",
	"hallClosed":     "phase 0 camp false hall false inParty true",
	"clock":          "start 0 written 450 hours 7 tens 3 rested 510 hours 8 tens 3",
	"count":          "quick 6 size 6 dropped 5 size 5 field 5 written 3",
	"acting":         "quick 5 asked 2 fight 5 sel 5 after 2 followed 2",
	"effects":        "0:97,true 1:26,true 64:0,true 127:0,true 128:0,false 100000:0,false ",
}

func sessionCheck(t *testing.T, name, got string) {
	t.Helper()
	if os.Getenv("RECORD_WORLD") != "" {
		fmt.Printf("RECORDSESSION %q: %q,\n", name, got)
		return
	}
	if want, ok := sessionWant[name]; !ok || got != want {
		t.Errorf("%s: %s, want %s (recorded)", name, got, want)
	}
}

func phaseLine(e *Engine) string {
	return fmt.Sprintf("phase %d camp %v hall %v inParty %v", obsPhase(e), e.campOn(), e.trainingOpen(), e.inParty())
}

func TestSessionWordsPhases(t *testing.T) {
	e := testEngine(t)
	e.Init()
	e.Start(e.Play)
	for range 3 {
		press(e)
	}
	sessionCheck(t, "title", phaseLine(e))

	e = devPlay(t, "overland")
	sessionCheck(t, "exploring", phaseLine(e))

	e = quickStarted(t)
	sessionCheck(t, "quickstart", phaseLine(e))
	e.CampMenu()
	sessionCheck(t, "camp", phaseLine(e))
	e.leaveCamp()
	sessionCheck(t, "afterCamp", phaseLine(e))

	e = devFirstTurn(t, "combat", false)
	sessionCheck(t, "fight", phaseLine(e))

	e = devFirstTurn(t, "winfight", false)
	for range 300 {
		if q := commonOf(e.build()).Ask; q != nil && q.Text == "Continue Battle:" {
			break
		}
		e.doNow(game.Quick{})
	}
	sessionCheck(t, "winAsk", phaseLine(e))
	if q := commonOf(e.build()).Ask; q != nil {
		e.doNow(no(q))
	}
	seen := ""
	for range 200 {
		seen += fmt.Sprint(obsPhase(e))
		if obsPhase(e) == 6 {
			break
		}
		press(e)
	}
	sessionCheck(t, "afterFight", phaseLine(e))
	sessionCheck(t, "afterFightSeen", seen)

	e = devPlay(t, "shop")
	sessionCheck(t, "shop", phaseLine(e))
	e = devPlay(t, "temple")
	sessionCheck(t, "temple", phaseLine(e))

	e = devPlay(t, "hall")
	sessionCheck(t, "hallMenu", phaseLine(e))
	if err := e.openTraining(0); err != nil {
		t.Fatal(err)
	}
	sessionCheck(t, "hallOpen", phaseLine(e))
	if err := e.CloseHall(); err != nil {
		t.Fatal(err)
	}
	sessionCheck(t, "hallClosed", phaseLine(e))
}

func TestSessionWordsClock(t *testing.T) {
	e := quickStarted(t)
	e.loadScript(14)
	got := fmt.Sprintf("start %d", obsMinute(e))
	e.setVar(e.field("Hours"), 7)
	e.setVar(e.field("TenMinutes"), 3)
	got += fmt.Sprintf(" written %d hours %d tens %d", obsMinute(e), e.getVar(e.field("Hours")), e.getVar(e.field("TenMinutes")))
	e.CampMenu()
	e.restFor(60, false, restThen{kind: thRested}) // a fixed hour of rest
	for range 300 {
		if e.RestStep() != nil {
			break
		}
	}
	got += fmt.Sprintf(" rested %d hours %d tens %d", obsMinute(e), e.getVar(e.field("Hours")), e.getVar(e.field("TenMinutes")))
	sessionCheck(t, "clock", got)
}

func TestSessionWordsPartyCount(t *testing.T) {
	e := quickStarted(t)
	e.loadScript(14)
	got := fmt.Sprintf("quick %d size %d", obsCount(e), e.partySize())
	e.CampMenu()
	e.leaveCamp()
	e.menu.on = true
	if err := e.removeMember(1); err != nil {
		got += " remove: " + err.Error()
	}
	got += fmt.Sprintf(" dropped %d size %d field %d", obsCount(e), e.partySize(), e.getVar(e.field("PartyCount")))
	e.setVar(e.field("PartyCount"), 3)
	got += fmt.Sprintf(" written %d", obsCount(e))
	sessionCheck(t, "count", got)
}

func TestSessionWordsActing(t *testing.T) {
	e := quickStarted(t)
	got := fmt.Sprintf("quick %d", obsActing(e))
	e.selectMember(2)
	e.as(e.sel)
	got += fmt.Sprintf(" asked %d", obsActing(e))
	e = devFirstTurn(t, "combat", false)
	got += fmt.Sprintf(" fight %d sel %d", obsActing(e), e.sel)
	e.doNow(game.EndTurn{})
	got += fmt.Sprintf(" after %d", obsActing(e))
	e.follow()
	got += fmt.Sprintf(" followed %d", e.sel)
	sessionCheck(t, "acting", got)
}

func TestSessionWordsEffects(t *testing.T) {
	e := devPlay(t, "effects")
	got := ""
	for _, i := range []int{0, 1, 64, 127, 128, 100000} {
		v, ok := obsEffect(e, i)
		got += fmt.Sprintf("%d:%d,%v ", i, v, ok)
	}
	sessionCheck(t, "effects", got)
}
