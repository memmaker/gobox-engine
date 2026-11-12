package pool

import (
	"os"
	"strings"
	"testing"
)

// TestGbvmAgainstPascal compares the engine fields at the special addresses with EclMenu.ReadGbvm (rules1e testdata/por-gbvm.txt).
func TestGbvmAgainstPascal(t *testing.T) {
	data, err := os.ReadFile("../../rules1e/testdata/por-gbvm.txt")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "gbvm ") {
			continue
		}
		f := strings.Fields(line)
		v := atois(t, []string{f[1], f[3]})
		a, want := v[0], v[1]
		e := testEngine(t)
		e.es.specialB1, e.es.specialFB, e.es.specialFC = 33, 11, 22
		e.posX, e.posY, e.facing = 5, 7, 6
		e.es.wallAhead, e.es.cell = 2, 9
		got := 0 // no field there: no script reads it
		if r, ok := refAt(t, e, a); ok {
			got = int(e.host.Get(r))
		}
		if got != want {
			t.Errorf("gbvm %#x = %d, Pascal %d", a, got, want)
		}
		n++
	}
	if n != 30 {
		t.Fatalf("%d cases", n)
	}
}
