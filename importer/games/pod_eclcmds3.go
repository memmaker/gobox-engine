package games

// The treasure sets of TREASURE SET (DS:0002, DS:000C: each set's first
// and last index into DS:0017's item ids).
var (
	treasureSetFirst = [12]byte{0, 1, 45, 51, 62, 71, 100, 110, 130, 138, 146, 155}
	treasureSetLast  = [12]byte{0, 44, 50, 61, 70, 99, 109, 129, 137, 145, 154, 170}
)

const (
	treasureSetMin = 1
	treasureSetMax = 11
)

// treasureSetItems = DS:0017: the ITEM0 ids of the treasure sets (index 0
// is not an item).
var treasureSetItems = [171]byte{170,
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 49,
	32, 33, 34, 35, 49, 223,
	1, 4, 5, 6, 7, 9, 11, 15, 20, 22, 25,
	36, 37, 38, 39, 40, 41, 42, 43, 44,
	50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 71, 72, 73, 74, 75, 76, 77, 78, 79,
	45, 80, 82, 83, 84, 86, 89, 91, 93, 95,
	81, 85, 87, 88, 90, 92, 94, 96, 97, 98, 99, 101, 102, 103, 104, 105, 106, 107, 108, 109,
	70, 210, 225, 241, 243, 244, 250, 255,
	100, 211, 226, 238, 245, 246, 251, 255,
	237, 242, 247, 248, 250, 251, 255, 255, 255,
	222, 201, 208, 209, 204, 203, 205, 206, 207, 48, 46, 47, 145, 146, 233, 236,
}

// TreasureSets is the treasure sets as the import writes them (pool.TreasureSetsFile):
// set number -> its ITEM0 ids; [0] is no set.
func TreasureSets() [][]int32 {
	sets := make([][]int32, len(treasureSetFirst))
	for n := treasureSetMin; n <= treasureSetMax; n++ {
		for i := treasureSetFirst[n]; i <= treasureSetLast[n]; i++ {
			sets[n] = append(sets[n], int32(treasureSetItems[i]))
		}
	}
	return sets
}
