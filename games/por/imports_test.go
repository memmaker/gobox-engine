package por

import (
	"os/exec"
	"strings"
	"testing"
)

// No package but cmd/por imports gobox's UI: the game never draws, gobox
// does (the frontend is handed the rules in cmd/por's main).
func TestNoUIImports(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-f", `{{.ImportPath}}{{range .Imports}} {{.}}{{end}}`, "./...").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || !strings.HasPrefix(f[0], "por") || f[0] == "por/cmd/por" {
			continue
		}
		for _, imp := range f[1:] {
			if strings.HasPrefix(imp, "gobox/front") || strings.HasPrefix(imp, "gobox/ui") {
				t.Errorf("%s imports %s", f[0], imp)
			}
		}
	}
}
