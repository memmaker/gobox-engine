package rules

// TreatInjuries and Antidote are BR's after-fight treatment (postcbt.pas TreatInjuries): the medics' poison and
// wound passes, run once after a fight that said so (FightResult.Treat), before the ship's repair. The Outcome's
// Effects are its lines. An Outcome.Confirm is the antidote question: the engine answers with Antidote and
// the treatment goes on; an Outcome without one is the treatment's end.
type (
	// TreatInjuries begins it. Skip: the script asked for none (zone 0's skipsTreatment) or the fight ended
	// in a boarding; the pending treatment is cleared all the same.
	TreatInjuries struct{ Skip bool }
	// Antidote is the answer to the question a TreatInjuries Outcome asked.
	Antidote struct{ Yes bool }
)

func (TreatInjuries) action() {}
func (Antidote) action()      {}
