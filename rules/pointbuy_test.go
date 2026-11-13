package rules

import "testing"

func TestPointBuy(t *testing.T) {
	want := map[int32]int32{7: -1, 8: 0, 9: 1, 10: 2, 11: 3, 12: 4, 13: 5, 14: 6, 15: 8, 16: 10, 17: 13, 18: 16, 19: -1}
	for s, c := range want {
		if got := PointBuyCost(s); got != c {
			t.Errorf("cost(%d) = %d, want %d", s, got, c)
		}
	}
	if n, ok := PointBuyTotal([]int32{18, 16, 12, 8, 8, 8}); n != 30 || !ok {
		t.Errorf("spent %d %v", n, ok)
	}
	if _, ok := PointBuyTotal([]int32{18, 18, 9, 8, 8, 8}); ok {
		t.Error("over budget taken")
	}
}
