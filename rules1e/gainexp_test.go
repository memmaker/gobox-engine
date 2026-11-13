package rules1e

import "testing"

// TestGainExperienceSplits: the script's points split among the classes (ruling S8).
func TestGainExperienceSplits(t *testing.T) {
	c := &Character{}
	c.Levels[ClsFighter], c.Levels[ClsMage] = 3, 2
	(&RuleState{}).GainExperience(c, 1001)
	if c.Exp != 500 {
		t.Fatalf("exp %d, want 500", c.Exp)
	}
	none := &Character{}
	(&RuleState{}).GainExperience(none, 1000)
	if none.Exp != 0 {
		t.Fatalf("no class: exp %d", none.Exp)
	}
}

// TestGainExperienceTraining: the award sets the training state (pod's, for all).
func TestGainExperienceTraining(t *testing.T) {
	s := &RuleState{}
	c := &Character{}
	c.Race, c.Levels[ClsFighter] = RaceHuman, 1
	need := s.T().XPForLevel(PorClass[ClsFighter], 2)
	s.GainExperience(c, need-1)
	if c.CanTrain {
		t.Fatalf("can train at %d of %d", c.Exp, need)
	}
	s.GainExperience(c, 1)
	if !c.CanTrain {
		t.Fatalf("cannot train at %d of %d", c.Exp, need)
	}
}
