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
// (dodging .. using a jetpack) are okay.
var coreStatus = map[Status]rules.Status{
	StNoStatus: rules.StatusOkay, StOkay: rules.StatusOkay, StGone: rules.StatusGone, StDead: rules.StatusDead,
	StDying: rules.StatusDying, StUnconscious: rules.StatusUnconscious, StFleeing: rules.StatusRunning,
	StPoisoned: rules.StatusPoisoned, StComatose: rules.StatusComatose, StNotHere: rules.StatusNotHere,
	StDodging: rules.StatusOkay, StSprinting: rules.StatusOkay, StFailedZeroG: rules.StatusOkay,
	StUsedJetpack: rules.StatusOkay, StUsingJetpack: rules.StatusOkay,
}

// fromCore is the BR status a script's write of s means.
var fromCore = map[rules.Status]Status{
	rules.StatusOkay: StOkay, rules.StatusGone: StGone, rules.StatusDead: StDead, rules.StatusDying: StDying,
	rules.StatusUnconscious: StUnconscious, rules.StatusRunning: StFleeing,
	rules.StatusPoisoned: StPoisoned, rules.StatusComatose: StComatose, rules.StatusNotHere: StNotHere,
}

// statusNames are gdconsts.pas statusNames, which the sheet and the cards show.
var statusNames = map[Status]string{
	StNoStatus: "No status", StOkay: "Okay", StGone: "Gone", StDead: "Dead", StDying: "Dying", StUnconscious: "Unconc",
	StFleeing: "Fleeing", StPoisoned: "Poisoned", StComatose: "Comatose", StNotHere: "Not here", StDodging: "Dodging",
	StSprinting: "Sprinting", StFailedZeroG: "Failed ZeroG", StUsedJetpack: "Used Jetpack", StUsingJetpack: "Using Jetpac",
}
