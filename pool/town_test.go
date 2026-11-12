package pool

import "testing"

// TestShopPriceFactor: the shops' factor over 16 (cotab/sotsb/pod's table of 1 2 4 8 32 64 128);
// the normal 16 and any other leave the price.
func TestShopPriceFactor(t *testing.T) {
	for _, c := range []struct{ price, factor, want int32 }{
		{160, 1, 10}, {160, 2, 20}, {160, 4, 40}, {160, 8, 80}, {160, 16, 160}, {160, 32, 320}, {160, 64, 640}, {160, 128, 1280}, {160, 0, 160}, {160, 5, 160},
	} {
		if got := scalePrice(c.price, c.factor); got != c.want {
			t.Errorf("price %d factor %d: %d, want %d", c.price, c.factor, got, c.want)
		}
	}
	if e := testEngine(t); e.shopPrice(100) != 100 { // POR's tables name no factor
		t.Error("POR's shop scaled")
	}
}
