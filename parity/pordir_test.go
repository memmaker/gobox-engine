// Package parity checks gobox against the games: no private copies of its
// plumbing (nocopies_test.go) and the shared map, border and page art.
package parity

import (
	"os/exec"
	"strings"
	"testing"
)

// porDir is por's checkout (the replace in go.mod).
func porDir(t *testing.T) string {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "por").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}
