package rules

// Hardness is how hard a skill roll is (BR opSkill's hardness, charhlp.pas
// SelectSkillCharacter): the skill value is scaled before the d100.
type Hardness int32

const (
	VeryEasy Hardness = iota // x4
	Easy                     // x2
	Normal                   // x1
	Hard                     // /2
	VeryHard                 // /4
)

// Grade is how a skill roll went, SkillRoll's Answer.
type Grade int32

const (
	FailedBadly Grade = iota // missed by more than the value
	Failed
	Made
	MadeWell // by half the value or more
)
