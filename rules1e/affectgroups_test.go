package rules1e

import (
	"maps"
	"slices"
	"testing"
)

// TestAffectGroupsCoverHandlers (fold T4e): every id of a game that has a
// handler is in one of the game's affect groups, or is on directCallOnly
// (its handler runs only from callAffect/RemoveAffect, never from a group).
// A handler outside both is never dispatched in play.
func TestAffectGroupsCoverHandlers(t *testing.T) {
	for name, g := range map[string]struct {
		ids      AffectMap
		variants map[AffectType]HandlerName
		groups   [][]AffectType
	}{"por": {nil, nil, affectGroups}, "cotab": {CotabAffectIDs, CotabAffectVariants, CotabAffectGroups}, "sotsb": {SotsbAffectIDs, SotsbAffectVariants, SotsbAffectGroups}, "pod": {PodAffectIDs, PodAffectVariants, PodAffectGroups}} {
		a := maps.Clone(POR.Affects)
		maps.Copy(a, g.variants)
		ids := slices.Collect(maps.Values(g.ids))
		if g.ids == nil {
			ids = slices.Collect(maps.Keys(POR.Affects))
		}
		slices.Sort(ids)
		for _, id := range slices.Compact(ids) {
			h := a[id]
			if affectHandler[h] == nil || directCallOnly[h] {
				continue
			}
			if !slices.ContainsFunc(g.groups, func(l []AffectType) bool { return slices.Contains(l, id) }) {
				t.Errorf("%s: affect %#x (%s) has a handler but is in no group", name, id, h)
			}
		}
	}
}

// directCallOnly: handlers no game's own group lists name. They run from
// RemoveAffect (a timed affect's end, Call != 0), from item readying, from
// another handler's callAffect, or never (noEffect, and the handlers the
// original tables fill but no group or call site reaches: PoD's charm and
// paralysing gazes).
var directCallOnly = map[HandlerName]bool{"addAffect62": true, "berserk": true, "berserkFoe": true, "berserkFoeCotab": true, "bestStrength": true, "charismaCotab": true, "charmGazePod": true, "clearCombatBlock": true, "clearStinkingCloud": true, "cloudkill": true, "collapses": true, "CursedAlignment": true, "fallsDead": true, "feeblemindedSotsb": true, "friendlyPod": true, "friendlySotsb": true, "frightened": true, "getUpOrWait": true, "giantStrengthCotab": true, "giantStrengthGirdle": true, "giantStrengthSpell": true, "gnomeVsFoes": true, "healOverTime": true, "huggingCotab": true, "IsHardier": true, "isWeakened": true, "itemGrantsAffect": true, "itemGrantsAffectCotab": true, "itemGrantsAffectLater": true, "magicalStone": true, "mirrorPod": true, "mirrorSotsb": true, "needsGiantStrength": true, "noEffect": true, "paladinCures": true, "paralysingGazePod": true, "plusOneSaveHit": true, "pod87": true, "poisonDamage": true, "poisonEnds": true, "regenerate": true, "removeEnds17": true, "reviveFull": true, "rots": true, "SaveBonusVsType42": true, "setCharisma": true, "sickDamage": true, "slowPoisonPod": true, "sotsb101": true, "sotsb102": true, "sotsb84": true, "sotsb85": true, "sotsb87": true, "startsToTrain": true, "weakAndSick": true, "weakenedPod": true}
