package rulesxxvc

// SkillGroup is a skill's group on the skill screen (types.pas TSkillGroup).
type SkillGroup int

const (
	GroupCareer SkillGroup = iota
	GroupDexterity
	GroupConstitution // no skill is in it
	GroupIntelligence
	GroupWisdom
	GroupCharisma
	GroupTechnical
)

// Group is skillDefinitions' group of skill s (gdconsts.pas; BR1 = BR2):
// 1..23 technical, 24..50 intelligence, 51..69 dexterity, 70..79 charisma,
// 80..84 wisdom.
func Group(s SkillNo) SkillGroup {
	switch {
	case s <= 23:
		return GroupTechnical
	case s <= 50:
		return GroupIntelligence
	case s <= 69:
		return GroupDexterity
	case s <= 79:
		return GroupCharisma
	}
	return GroupWisdom
}

// notOffered are the skills skillDefinitions' isOffered (BR2: enabled) is 0 for.
var notOffered = map[SkillNo]bool{
	12: true, 15: true, 24: true, 25: true, 26: true, 28: true, 29: true, 32: true, 33: true, 34: true,
	35: true, 37: true, 38: true, 39: true, 40: true, 41: true, 42: true, 44: true, 48: true, 49: true,
	55: true, 57: true, 58: true, 60: true, 66: true, 68: true, 69: true, 70: true, 75: true, 82: true,
}

// Offered is skillDefinitions' isOffered for skill s: the general list offers it.
func Offered(s SkillNo) bool { return !notOffered[s] }

// Career is a career's index in the original (CareerRocketJock..CareerScout).
type Career int

const (
	CareerRocketJock Career = iota
	CareerMedic
	CareerWarrior
	CareerEngineer
	CareerRogue
	CareerScout
)

// CareerSkills are each career's eight skills (gdconsts.pas careerSkills).
// The same in BR1 and BR2 (BR2 names 53 DRIVE_GROUND_CAR, 11 DEMOLITIONS,
// 60 ANIMAL_RIDING, 80 TRACKING).
var CareerSkills = [6][8]SkillNo{
	{83, 67, 56, 51, 52, 53, 54, 59},
	{17, 18, 19, 23, 20, 21, 22, 16},
	{83, 73, 31, 67, 56, 62, 11, 14},
	{83, 1, 2, 3, 4, 5, 6, 67},
	{83, 7, 8, 61, 62, 63, 65, 76},
	{83, 60, 74, 62, 65, 80, 84, 30},
}
