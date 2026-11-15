package rulesxxvc

import "gobox/rules"

// Status is types.pas TStatus, the original's order.
type Status uint8

const (
	StNoStatus Status = iota
	StOkay
	StGone
	StDead
	StDying
	StUnconscious
	StFleeing
	StPoisoned
	StComatose
	StNotHere
	StDodging
	StSprinting
	StFailedZeroG
	StUsedJetpack
	StUsingJetpack
)

// coreStatus is how a sheet and a script see s. The fight's own states
// (dodging .. using a jetpack) are okay; poisoned and comatose are helpless.
// ponytail: poisoned and comatose shown as unconscious; core has no names for them.
var coreStatus = map[Status]rules.Status{
	StNoStatus: rules.StatusOkay, StOkay: rules.StatusOkay, StGone: rules.StatusGone, StDead: rules.StatusDead,
	StDying: rules.StatusDying, StUnconscious: rules.StatusUnconscious, StFleeing: rules.StatusRunning,
	StPoisoned: rules.StatusUnconscious, StComatose: rules.StatusUnconscious, StNotHere: rules.StatusGone,
	StDodging: rules.StatusOkay, StSprinting: rules.StatusOkay, StFailedZeroG: rules.StatusOkay,
	StUsedJetpack: rules.StatusOkay, StUsingJetpack: rules.StatusOkay,
}

// fromCore is the BR status a script's write of s means.
var fromCore = map[rules.Status]Status{
	rules.StatusOkay: StOkay, rules.StatusGone: StGone, rules.StatusDead: StDead, rules.StatusDying: StDying,
	rules.StatusUnconscious: StUnconscious, rules.StatusRunning: StFleeing,
}
