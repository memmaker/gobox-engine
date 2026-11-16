package engine

import (
	"fmt"
	"slices"

	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
)

// The ship fight (BR shipcbt.pas ShipCombat, script command SpaceFight). The ruleset owns the ships, the foe, the
// rolls and the fight's order (rules.ShipRules); the engine owns the screen, the pending question and the script's
// continuation. It keeps no ship numbers: every state it publishes is read from the ruleset again.
//
//	opSpaceFight: BeginShips; then, until ShipVerdict ends it: ShipRound, the crew's menu (Allowed, Resolve per
//	answer), ShipFoe; then EndShips, the last page, the script's FightState and its resume.

const fieldFightState = "FightState" // the layout's zone 0 shipCombatOutcome: 0 the foe destroyed or boarded, 1 the party's ship lost or the foe gone

// The values the script reads from FightState.
const (
	fightStateWon  = 0 // shipcbt.pas: the enemy ship destroyed or boarded
	fightStateLost = 1 // the team's ship destroyed, or the enemy out of range
)

// shipStage is where the round stands.
type shipStage int

const (
	stageRound shipStage = iota // the round begins (ShipRound)
	stageCrew                   // the crew's members take their turns
	stageFoe                    // the foe's turn and the round's end (ShipFoe)
	stageOver                   // the verdict is in, EndShips ran: the last page is up
)

// shipRun is what the engine keeps of the fight: the screen's, not the ruleset's.
type shipRun struct {
	stage shipStage
	log   []string     // the message window's lines since the round began
	sheet *shipSheetUp // a ship's record shown over the fight
	ram   rules.Action // the action a yes/no question confirms
}

// shipSheetUp is the record shown: whose, and whether its figures are known.
type shipSheetUp struct{ foe, known bool }

// ships is the ruleset's ship fight; nil when it has none.
func (e *Engine) ships() rules.ShipRules { return e.ruleset().Ships() }

func init() { eclOps[ecl.OpSpaceFight] = (*Engine).opSpaceFight }

// opSpaceFight = BR eclvm.pas EclSpaceFight -> shipcbt.pas ShipCombat (the foe's ship number, a spare value, the foe's
// pilot skill, its to-hit number): the fight holds the script until it is over.
func (e *Engine) opSpaceFight() {
	e.vm.ReadArgs(4)
	ships := e.ships()
	if ships == nil {
		panic("script: SpaceFight in a ruleset without ship combat")
	}
	enc := rules.ShipEncounter{
		Foe:        rules.ShipID(byte(e.vm.Arg(1, e.host)) - 1),
		PilotSkill: int32(byte(e.vm.Arg(3, e.host))),
		ToHit:      int32(byte(e.vm.Arg(4, e.host))),
	}
	if _, err := ships.BeginShips(enc, nil); err != nil {
		panic(fmt.Sprintf("script: SpaceFight: %v", err))
	}
	e.takeover()
	e.push(cont{kind: kOpCombatEnd})
	e.shipRun = &shipRun{}
	e.shipOn()
	e.openAsk()
}

// shipOn plays the fight on until it asks the player or is over.
func (e *Engine) shipOn() {
	run, ships := e.shipRun, e.ships()
	for {
		if run.stage != stageOver && ships.ShipVerdict(ships.ShipState()) != rules.FightOn {
			e.shipOver()
			return
		}
		switch run.stage {
		case stageRound:
			run.log = nil
			_, o := ships.ShipRound(ships.ShipState(), nil)
			run.stage = stageCrew
			if e.shipOutcome(o, dShipDodge) {
				return
			}
		case stageCrew:
			al := e.ruleset().Allowed(ships.ShipState())
			if len(al.Offers) == 0 {
				run.stage = stageFoe
				continue
			}
			e.askShipMenu(al)
			return
		case stageFoe:
			_, o := ships.ShipFoe(ships.ShipState(), nil)
			run.stage = stageRound
			if e.shipOutcome(o, dShipDodge) {
				return
			}
		case stageOver:
			return
		}
	}
}

// shipOutcome shows what an action did; true when it asks a question (a confirmation) or shows a record.
func (e *Engine) shipOutcome(o rules.Outcome, d doneKind) bool {
	run := e.shipRun
	for _, f := range o.Effects {
		switch f := f.(type) {
		case rules.Narrated:
			run.log = append(run.log, f.Text)
		case rules.ShipSheetShown:
			run.sheet = &shipSheetUp{foe: f.Foe, known: f.Known}
		default:
			panic(fmt.Sprintf("ship fight: an effect with no presentation: %T", f))
		}
	}
	switch {
	case o.Confirm != "":
		q := yesNo(o.Confirm, false)
		if d == dShipRam {
			q.Default = 0 // YesNoMenu(.., C_Menu.DEFAULT_NO)
		}
		e.eclWait = &eclWait{ask: q, shape: shYesNo, done: askDone{kind: d}}
		return true
	case run.sheet != nil:
		e.eclWait = &eclWait{ask: &game.Question{Continue: true}, shape: shPause, done: askDone{kind: dShipSheet}}
		return true
	}
	return false
}

// shipEntry is one place of the crew menu: an offer (offer is its place in Allowed.Offers) or a submenu (group, offer -1).
type shipEntry struct {
	label, group string
	offer        int
}

// shipEntries are the crew menu's places in the ruleset's order: its offers, those of one Group as one place.
func shipEntries(al rules.Allowed) (out []shipEntry) {
	for i, o := range al.Offers {
		switch {
		case o.Group == "":
			out = append(out, shipEntry{label: o.Name, offer: i})
		case !slices.ContainsFunc(out, func(en shipEntry) bool { return en.group == o.Group }):
			out = append(out, shipEntry{label: o.Group, group: o.Group, offer: -1})
		}
	}
	return out
}

// askShipMenu asks the crew member's menu: the ruleset's offers, in its order (the Notice first).
func (e *Engine) askShipMenu(al rules.Allowed) {
	if al.Notice != "" {
		e.shipRun.log = append(e.shipRun.log, al.Notice)
	}
	q := &game.Question{Text: al.Prompt, NoEsc: true}
	entries := shipEntries(al)
	ids := make([]int, len(entries))
	for i, en := range entries {
		q.Replies = append(q.Replies, en.label)
		ids[i] = i
	}
	e.eclWait = &eclWait{ask: q, shape: shChoice, ids: ids, done: askDone{kind: dShipMenu}}
}

// askShipGroup asks a submenu of the crew menu: the offers of the group, by their place in Allowed.Offers.
func (e *Engine) askShipGroup(al rules.Allowed, group string) {
	q := &game.Question{}
	var ids []int
	var esc *ansv // shipcbt.pas 1137: Esc (C_Keys.NONE) is no system chosen, the volley goes untargeted
	for i, o := range al.Offers {
		if o.Group == group {
			q.Text = o.Ask
			q.Replies = append(q.Replies, o.Name)
			ids = append(ids, i)
		}
		if o.EscFor == group {
			esc = &ansv{n: i}
		}
	}
	q.NoEsc = esc == nil
	e.eclWait = &eclWait{ask: q, shape: shChoice, ids: ids, esc: esc, done: askDone{kind: dShipGroup}}
}

// shipAnswered takes an answer of the fight's questions and plays on.
func (e *Engine) shipAnswered(kind doneKind, v ansv) {
	run, ships := e.shipRun, e.ships()
	switch kind {
	case dShipMenu:
		al := e.ruleset().Allowed(ships.ShipState())
		entries := shipEntries(al)
		if v.n < 0 || v.n >= len(entries) {
			panic(fmt.Sprintf("ship fight: menu place %d of %d", v.n, len(entries)))
		}
		if en := entries[v.n]; en.group != "" {
			e.askShipGroup(al, en.group)
			return
		}
		e.shipResolve(al.Offers[entries[v.n].offer].Action)
		return
	case dShipGroup:
		al := e.ruleset().Allowed(ships.ShipState())
		if v.n < 0 || v.n >= len(al.Offers) {
			panic(fmt.Sprintf("ship fight: submenu offer %d of %d", v.n, len(al.Offers)))
		}
		e.shipResolve(al.Offers[v.n].Action)
		return
	case dShipDodge:
		e.shipResolve(rules.ShipDodge{Yes: v.n == 1})
		return
	case dShipRam:
		if v.n == 1 {
			e.shipResolve(rules.Confirmed{Action: run.ram})
			return
		}
	case dShipSheet:
		run.sheet = nil
	case dShipEnd:
		e.shipRun = nil
		return
	}
	e.shipOn()
}

// shipResolve gives the ruleset the crew's action (or the dodge answer) and plays on.
func (e *Engine) shipResolve(a rules.Action) {
	o, err := e.ruleset().Resolve(e.ships().ShipState(), a, nil)
	if err != nil {
		panic(fmt.Sprintf("ship fight: %T: %v", a, err))
	}
	if _, isRam := a.(rules.ShipRam); isRam && o.Confirm != "" {
		e.shipRun.ram = a
		e.shipOutcome(o, dShipRam)
		return
	}
	if !e.shipOutcome(o, dShipDodge) {
		e.shipOn()
	}
}

// shipOver ends the fight: the ruleset's end (repairs, experience), the script's FightState, the last page.
func (e *Engine) shipOver() {
	run, ships := e.shipRun, e.ships()
	verdict := ships.ShipVerdict(ships.ShipState())
	res := ships.EndShips(ships.ShipState(), nil)
	for _, f := range res.Effects {
		n, ok := f.(rules.Narrated)
		if !ok {
			panic(fmt.Sprintf("ship fight: an end effect with no presentation: %T", f))
		}
		run.log = append(run.log, n.Text)
	}
	if r, ok := e.layoutField(fieldFightState); ok {
		state := int64(fightStateLost)
		if verdict == rules.PartyWon {
			state = fightStateWon
		}
		e.host.Set(r, state)
	}
	run.stage = stageOver
	e.eclWait = &eclWait{ask: &game.Question{Continue: true}, shape: shPause, done: askDone{kind: dShipEnd}}
}

// shipFightState is the state the front draws while the fight is on.
func (e *Engine) shipFightState(cm game.Common) game.ShipFight {
	cm.NoSave = true // the original has no save in a ship fight (the ruleset's ship fight is not in its save)
	cm.Ask = e.question()
	st := game.ShipFight{Common: cm, Fight: e.spaceFightState()}
	if up := e.shipRun.sheet; up != nil {
		sh, _ := e.ShipSheetState(!up.foe, up.known)
		st.Sheet = &sh
	}
	return st
}

// panelSystems is the sensors' figure among the panel's six (game.ShipSystems: Hull Sens Ctrl Life Fuel Engn).
const panelSensors = 1

// spaceFightState is game.SpaceFight read from the ruleset's ships.
func (e *Engine) spaceFightState() game.SpaceFight {
	st := e.ships().ShipState()
	f := game.SpaceFight{Range: max(int(st.Range), 0), Status: st.Status}
	if file := e.artNames().ShipPictures; file != "" {
		f.Foe = game.Picture{File: file, Block: st.Foe.Picture}
	}
	fill := func(s rules.Ship) (sys game.ShipSystems) {
		for i, id := range s.Order[:min(len(s.Order), len(sys.Now))] {
			sys.Now[i], sys.Max[i] = int(s.Systems[id].Now), int(s.Systems[id].Max)
		}
		return sys
	}
	f.Own, f.Enemy = fill(st.Own), fill(st.Foe)
	f.SensorsOut = f.Own.Now[panelSensors] == 0
	for _, m := range st.Own.Mounts[:min(len(st.Own.Mounts), ownPanelMounts)] {
		f.Weapons = append(f.Weapons, weaponSlot(m))
	}
	for _, l := range e.shipRun.log {
		f.Log = append(f.Log, game.Message{Tone: game.ToneText, Text: l})
	}
	return f
}

// The panel shows the own ship's first three mounts, the sheet the first four that hold a weapon (game.SpaceFight, game.ShipSheet).
const (
	ownPanelMounts = 3
	sheetMounts    = 4
)

func weaponSlot(m rules.ShipMount) game.WeaponSlot {
	return game.WeaponSlot{Short: m.Short, Mounts: int(m.Count), Working: int(m.Working), Ammo: int(m.Ammo), Reloads: int(m.Reloads)}
}

// rangeBands are the foe's range bands by their upper bounds (CharHlp.DrawShipDetails).
var rangeBands = []struct {
	below int32
	band  game.RangeBand
}{{1, game.RangeRamming}, {3, game.RangeShort}, {5, game.RangeMedium}, {7, game.RangeLong}}

// ShipSheetState is a ship's record as the sheet shows it, built from the ruleset's ship: the party's (own), or
// the fight's foe; known: the foe's armour class, crew, speed, reaction, ammunition and reloads are known (the
// sensor roll made). ok is false when the ruleset has no ships, or no foe is on.
func (e *Engine) ShipSheetState(own, known bool) (game.ShipSheet, bool) {
	ships := e.ships()
	if ships == nil {
		return game.ShipSheet{}, false
	}
	r := ships.ShipSheetOf(!own)
	if r.Name == "" {
		return game.ShipSheet{}, false
	}
	h := game.ShipSheet{Own: own, Known: known || own, Name: r.Name, Type: r.Type, Race: r.Race, Tonnage: int(r.Tonnage),
		Length: int(r.Length), Width: int(r.Width), Cargo: int(r.Cargo), AC: int(r.Armor), Crew: int(r.Crew),
		Speed: int(r.Speed), React: int(r.Reaction)}
	if !own {
		rng := ships.ShipState().Range
		h.Range = game.RangeExtreme
		for _, b := range rangeBands {
			if rng < b.below {
				h.Range = b.band
				break
			}
		}
	}
	for _, m := range r.Mounts {
		if m.Short != "" && m.Count > 0 && len(h.Weapons) < sheetMounts {
			h.Weapons = append(h.Weapons, weaponSlot(m))
		}
	}
	return h, true
}

// shipAllowed is what the front may send while the fight asks: a reply by its place, or a key.
func (e *Engine) shipAllowed() game.Allowed {
	var a game.Allowed
	q := e.question()
	switch {
	case q == nil:
	case q.Continue:
		a.Plain = append(a.Plain, game.ContinueAction{})
	default:
		for i := range q.Replies {
			a.Plain = append(a.Plain, game.Reply{N: i})
		}
		if w := e.eclWait; w != nil && w.esc != nil && !q.NoEsc { // Esc answers a reply the menu does not list (the Target menu: untargeted)
			a.Plain = append(a.Plain, game.Leave{})
		}
	}
	return a
}
