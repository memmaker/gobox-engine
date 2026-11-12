package pool

import (
	"testing"

	"rules1e"
)

// The fight-based bug classes ported from cotab's own tests (bugclass-cotab: 17, 20,
// the fight half of 36, 39, 40, 41, 47, 53, 55, 60) run POR's pool test on each
// game's imported folder (POOL_ADV_<GAME>; they skip when it is unset).
// Not yet ported, each an open bug its port shows (agentFold.log Pg):
// FightEventsOut COTAB (a hit's damage shown twice: a Struck after its Attacked),
// FightEventsOut, PartyWipeEnds, LastBlowBody SOTSB (rules1e lays POR's body
// ground 0x1F, past sotsb's 31 grounds: aiCanStep panics), PartyWipeEnds POD (a
// lost fight never reaches the party menu).

// usageThrown is the item table's Usage bit of a weapon thrown from the hand (POR's javelin, dart).
const usageThrown = 0x10

// weaponFor is a one-handed weapon type c can use, melee (Range 0) or thrown:
// prefer when it is one (POR's), else the first in the game's item table (sotsb's
// types are numbered apart from POR's; a cotab mage cannot use a long sword).
func weaponFor(t *testing.T, e *Engine, c *rules1e.Character, prefer rules1e.ItemType, thrown bool) rules1e.ItemType {
	t.Helper()
	fits := func(ty rules1e.ItemType) bool {
		b := e.state().Base(ty)
		return b.Slot == rules1e.SlotWeapon && b.Hands == 1 && (b.Range > 0) == thrown && (b.Usage&usageThrown != 0) == thrown && b.Classes&c.ClassMask != 0
	}
	if fits(prefer) {
		return prefer
	}
	for ty := range rules1e.ItemType(0x80) {
		if fits(ty) {
			return ty
		}
	}
	t.Skipf("no weapon %s can use (thrown %v)", c.Name(), thrown)
	return 0
}

// adv is an engine maker for env's game.
func adv(env string) func(*testing.T) *Engine {
	return func(t *testing.T) *Engine { t.Helper(); return advEngine(t, env) }
}

func TestTurnWeaponCOTAB(t *testing.T)        { testTurnWeapon(t, adv(advEnvCOTAB)) }
func TestReadyInFightCOTAB(t *testing.T)      { testReadyInFight(t, adv(advEnvCOTAB)) }
func TestAllyBumpCOTAB(t *testing.T)          { testAllyBump(t, adv(advEnvCOTAB)) }
func TestDropOnceCOTAB(t *testing.T)          { testDropOnce(t, adv(advEnvCOTAB)) }
func TestViewSheetCOTAB(t *testing.T)         { testViewSheet(t, adv(advEnvCOTAB)) }
func TestAimOffersAcceptedCOTAB(t *testing.T) { testAimOffersAccepted(t, adv(advEnvCOTAB)) }
func TestPartyWipeEndsCOTAB(t *testing.T)     { testPartyWipeEnds(t, adv(advEnvCOTAB)) }
func TestLastBlowBodyCOTAB(t *testing.T)      { testLastBlowBody(t, adv(advEnvCOTAB)) }
func TestLootEscLeavesCOTAB(t *testing.T)     { testLootEscLeaves(t, adv(advEnvCOTAB)) }

func TestTurnWeaponSOTSB(t *testing.T)        { testTurnWeapon(t, adv(advEnvSOTSB)) }
func TestReadyInFightSOTSB(t *testing.T)      { testReadyInFight(t, adv(advEnvSOTSB)) }
func TestAllyBumpSOTSB(t *testing.T)          { testAllyBump(t, adv(advEnvSOTSB)) }
func TestDropOnceSOTSB(t *testing.T)          { testDropOnce(t, adv(advEnvSOTSB)) }
func TestViewSheetSOTSB(t *testing.T)         { testViewSheet(t, adv(advEnvSOTSB)) }
func TestAimOffersAcceptedSOTSB(t *testing.T) { testAimOffersAccepted(t, adv(advEnvSOTSB)) }
func TestLootEscLeavesSOTSB(t *testing.T)     { testLootEscLeaves(t, adv(advEnvSOTSB)) }

func TestTurnWeaponPOD(t *testing.T)        { testTurnWeapon(t, adv(advEnvPOD)) }
func TestReadyInFightPOD(t *testing.T)      { testReadyInFight(t, adv(advEnvPOD)) }
func TestAllyBumpPOD(t *testing.T)          { testAllyBump(t, adv(advEnvPOD)) }
func TestDropOncePOD(t *testing.T)          { testDropOnce(t, adv(advEnvPOD)) }
func TestViewSheetPOD(t *testing.T)         { testViewSheet(t, adv(advEnvPOD)) }
func TestAimOffersAcceptedPOD(t *testing.T) { testAimOffersAccepted(t, adv(advEnvPOD)) }
func TestFightEventsOutPOD(t *testing.T)    { testFightEventsOut(t, adv(advEnvPOD)) }
func TestLastBlowBodyPOD(t *testing.T)      { testLastBlowBody(t, adv(advEnvPOD)) }
func TestLootEscLeavesPOD(t *testing.T)     { testLootEscLeaves(t, adv(advEnvPOD)) }
