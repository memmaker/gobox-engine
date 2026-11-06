package rules

// StrIndex is the strength as one index: 1..17 as is, 18 with its percentile
// 18..23, 19..25 as 24..30. Above 18/100 the original leaves the result
// undefined (sic): 0.
func StrIndex(str, str00 byte) byte {
	switch {
	case str <= 17:
		return str
	case str == 18:
		switch {
		case str00 == 0:
			return 18
		case str00 <= 50:
			return 19
		case str00 <= 75:
			return 20
		case str00 <= 90:
			return 21
		case str00 <= 99:
			return 22
		case str00 == 100:
			return 23
		}
		return 0
	case str <= 25:
		return str + 5
	}
	return 0
}

// StrHit is the to-hit bonus of a StrIndex.
func StrHit(v byte) int8 {
	switch {
	case v >= 1 && v <= 3:
		return -3
	case v == 4 || v == 5:
		return -2
	case v == 6 || v == 7:
		return -1
	case v >= 17 && v <= 19:
		return 1
	case v >= 20 && v <= 22:
		return 2
	case v >= 23 && v <= 25:
		return 3
	case v == 26 || v == 27:
		return 4
	case v >= 28 && v <= 30:
		return int8(v) - 23
	}
	return 0
}

// StrDamage is the damage bonus of a StrIndex.
func StrDamage(v byte) int8 {
	switch {
	case v == 1 || v == 2:
		return -2
	case v >= 3 && v <= 5:
		return -1
	case v == 16:
		return 1
	case v >= 17 && v <= 19:
		return int8(v) - 16
	case v >= 20 && v <= 29:
		return int8(v) - 17
	case v == 30:
		return 14
	}
	return 0
}

// DexAC is the armour class adjustment of a dexterity.
func DexAC(d byte) int8 {
	switch {
	case d >= 1 && d <= 3:
		return -4
	case d >= 4 && d <= 6:
		return int8(d) - 7
	case d >= 15 && d <= 18:
		return int8(d) - 14
	case d == 19 || d == 20:
		return 4
	case d >= 21 && d <= 23:
		return 5
	case d == 24 || d == 25:
		return 6
	}
	return 0
}

// DexMissile is the missile to-hit adjustment of a dexterity.
func DexMissile(d byte) int8 {
	switch {
	case d <= 2:
		return -4
	case d >= 3 && d <= 5:
		return int8(d) - 6
	case d >= 16 && d <= 18:
		return int8(d) - 15
	case d == 19 || d == 20:
		return 3
	case d >= 21 && d <= 23:
		return 4
	case d == 24 || d == 25:
		return 5
	}
	return 0
}

// ConSaveLow is the saving throw bonus of a constitution 4..18, which
// dwarves, halflings and items of the save bonus effect get.
func ConSaveLow(con byte) byte {
	switch {
	case con >= 4 && con <= 6:
		return 1
	case con >= 7 && con <= 10:
		return 2
	case con >= 11 && con <= 13:
		return 3
	case con >= 14 && con <= 17:
		return 4
	case con == 18:
		return 5
	}
	return 0
}

// ConSaveHigh is the saving throw bonus of a constitution 19..25, everyone's.
func ConSaveHigh(con byte) byte {
	switch con {
	case 19, 20:
		return 1
	case 21, 22:
		return 2
	case 23, 24:
		return 3
	case 25:
		return 4
	}
	return 0
}

// HitDiceWeight is how many hit dice a combatant counts for a sleep-like
// spell: by maxLevel, or for faerie fire by its size.
func HitDiceWeight(faerieFire bool, size, maxLevel byte) int {
	if faerieFire {
		switch size {
		case 1:
			return 1
		case 2, 3:
			return 2
		case 4:
			return 4
		}
		return 0
	}
	switch maxLevel {
	case 0, 1:
		return 1
	case 2:
		return 2
	case 3:
		return 4
	}
	return 8
}
