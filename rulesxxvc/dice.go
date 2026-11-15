// Package rulesxxvc is the Buck Rogers XXVc ruleset of BR1 and BR2 behind
// gobox core's Rules. Source of truth: the reconstructions buckroger-re and
// buckrogers2-re; each function names its Pascal.
package rulesxxvc

import "gobox/rules"

// PercentileRoll is charhlp.pas PercentileRoll: percent > Random(100).
func PercentileRoll(r *rules.RNG, percent int) bool { return percent > r.N(100) }
