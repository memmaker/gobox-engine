package pool

import (
	"strconv"
)

// moneyText is gold and copper as "12 gold" / "12 gold and 50 copper".
func moneyText(gold, copper uint64) string {
	switch {
	case copper == 0:
		return strconv.Itoa(int(gold)) + " gold"
	case gold == 0:
		return strconv.Itoa(int(copper)) + " copper"
	}
	return strconv.Itoa(int(gold)) + " gold and " + strconv.Itoa(int(copper)) + " copper"
}
