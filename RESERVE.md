# Reserve and reinforce: checklist

## Goal
A generic engine feature any adventure can use: a fight has a **reserve** (combatants
that belong to it but wait off the map) and an effect **reinforces** a side from it.
First user: POD (pod-re CharFightsAlone: ten of the area's monster, status 9; Monster
Summoning gates up to two in next to the caster, on the caster's side). Until this
works every scripted POD fight panics ("rules1e: fact not held").

## Rulings (Felix, 2026-11-15)
- Terms: reserve / reinforce.
- Triggers in this version: a spell effect only (POD's Monster Summoning).
- Declaration: an adventure table, as POD's (monster and icon by area set, a count).

## Reference (pod-re/src)
- eclmenu.pas:871 CharFightsAlone: the reserve added before a fight (status 9, npc |= $32)
- combatst.pas:822: a reserve unit takes a slot but is out of play, no size, no body
- cbtacts.pas:1326 UpdateHostileHpPercent: the reserve does not count for morale
- magic.pas:2870: the summoning: up to 2, next to the combatants in list order, 'gates in'
- postcbt.pas:97: the reserve gives no experience, money or items

## Steps
- [x] A data: adventure.Reserve table (was AreaMonsters), no dialect flag; pod's gametables.json regenerated
- [x] B rules: rules.StatusReserve, rules.HoldInReserve
- [x] C rules1e (reserve.go): the reserve at setup, in the monsters' health, after the fight; Reinforce; Monster Summoning calls it
- [x] D engine: opCombat adds the adventure's reserve (addReserve); dev scene "reservefight"
- [x] E tests: TestPodReserve (engine), TestNoUnseenQuestion walks reservefight in all four games
- [x] gate (passed)

## Open / unverified
- data/pod is unversioned: another checkout needs POD re-imported (the table's key is "Reserve" now).
- Monster Summoning is wired (spells_podfirst.go) but casting it in a fight is not played through:
  whether pod's spell 126 can be memorized and aimed here is unchecked.
- The reserve's monster is marked met in the bestiary when the fight starts (addMonsters), seen or not.
- Reinforcements arrive beside the first combatant of the fight's list (the party's first member),
  not beside the caster, as the original has it.
