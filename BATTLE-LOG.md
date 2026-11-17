# Battle log: a typed ledger of every state change

Goal (Felix, 2026-11-16): the battle log records every state change in a
fight — who did how much damage to whom by which means, heals, saves, and
every ability or spell effect that goes off. Decision: typed ledger only.
Every entry comes from a typed rules effect (rules/effect.go); the engine
writes the line (engine/present_engine.go apply). The original's message
text (rules.Narrated) still goes to the window's note, never to the log. No
text matching anywhere.

How: the rules emit a typed effect at each single point a change happens
(rules1e: ApplyDamage, Damage's save, heals, AddAffect, RemoveAffect, the
affect calls; rulesxxvc: its attack/ground/effects paths). The engine turns
each into one log line from typed fields and names it reads from the rules.

Detect the bug class: a state change with no log line. TestLogEveryChange
(engine) plays a fight and checks each HP change and affect change of every
unit has a ledger entry.

Built (engine/battlelog.go): logEffect writes a line per Attacked (with
the weapon), Struck (by the spell or affect, and whose), Casting, Used,
Fell (with the rules' why: "Surrenders"), Fled. reconcile, after every
fight step, compares each unit's HP, status and spell effects with what
the log last told and logs any change no effect named (heals, affects
gained and ended, a gaseous escape). One owner: the engine; the ui reads
Battlefield.Log (its own r.log is gone).

## Checklist
- [x] Engine ledger (battlelog.go), Battlefield.Log, ui's r.log removed
- [x] rules: With on Attacked, From/By on Struck, Why on Fell, Used
- [x] rules1e: harmBy (spell, affect) on its Struck; Fell why
- [x] rulesxxvc: Used for grenades/missiles, Struck's From/By
- [x] TestLogEveryChange (an unnamed HP change is logged; blows and falls are)
- [x] Commit
- [ ] Later: typed Saved effect (a save is not a state change; not logged yet)
