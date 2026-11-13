package rules




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

