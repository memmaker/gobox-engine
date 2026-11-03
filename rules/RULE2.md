# Rule 2 decisions and open differences

## Decided by the user
- (a) Saving throws: one path, por's signed compare (por-go/engine/rules.go:253), for all games.
  cotab-go/engine/rules.go:335 and sotsb-go rules.go compared the byte unsigned, so a total below 0
  wrapped to 128..255 and saved. `TestSavingThrow` counts these cases (only totals below 0 differ).
- (b) Trap damage stays per game: `Config.TrapSkipStoned` (sotsb damage.go:36),
  `Config.TrapRollEach` (sotsb damage.go:154). Zero value = cotab.
- (c) Encumbrance is cut (no weight rule in por, which has none: por-go/engine/pool.go:10). Not in rules.
  Games can delete: `encumbrance` (cotab charcalc.go:236, sotsb charcalc.go:274), its call in `ApplyItems`
  (cotab :124, sotsb :204), `strCarry` (cotab charcalc.go:374, sotsb :411), `moveLight/Medium/Heavy`
  limits for weight, `coinWeight`, `readiedWeight`, `carryLimit` (cotab coins.go:13, sotsb coins.go:11)
  and its checks (cotab coins.go:18,98,174, shop.go:86; sotsb coins.go:16,129,205, shop.go:86),
  and the weight sums in `ApplyItems`. `armourMove` (movement by armour weight) is a separate rule: decide.
- (d) XP split stays per game for now (`giveExp`). Not in rules.
- (e) Shop pricing stays per game; never unified. Not in rules.

## Differences found (for the record)
- giveExp: sotsb skips when the area's HideExperience word is set (sotsb aftercombat.go:117, absent in cotab);
  sotsb cleric prime bonus is Wis >= 15 (aftercombat.go:136, `> primeScore-1`) where cotab (:129) and the other
  classes need > 15; por (aftercombat.go:86) keeps single-byte scores and its own multi-class constants.
- expThreshold: cotab train.go:338 reads little-endian bytes from 99-byte rows and runs into the next
  class's row; sotsb train.go:531 reads an int32 table and returns 0 past the end. Not unified (data format and
  past-the-end behaviour differ); stays per game.
- shopPrice: cotab shop.go:13 and sotsb shop.go:13 are identical; only the list width differs (30 vs 38). por has none.
- Encumbrance (cut): cotab int16-wraps `Weight - carry` (charcalc.go:238), sotsb uses plain int (:275).

## Not moved yet
Ability-score effects (`statAffects`, `strengthDecode`, `strongerThan`: need affect and item lists), `ApplyItems`,
`weaponStats`, `itemAC`, `armourMove`, coins, `computeExp`, the `cmdDamage` target loop (all tied to Engine
state, items, affects and UI). They need an Item/Affect interface; do them when the item model is shared.
Neither game had unit tests for the moved functions (only fixture tests), so none were ported; the oracle tests replace them.
