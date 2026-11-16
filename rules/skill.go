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

// SkillDamage is BR's SkillDamage (eclvm.pas EclSkillDamage): each of Whom rolls Skill (one candidate
// each, as SkillRoll's) and each who fails takes Dice d Sides + Bonus, rolled anew for each, in a
// Byte. The Outcome has a Hurt per member hit; a member who is not okay is spared.
type SkillDamage struct {
	Whom               Party
	Skill              ScriptSkill
	Hardness           Hardness
	Dice, Sides, Bonus int32
}

func (SkillDamage) action() {}
