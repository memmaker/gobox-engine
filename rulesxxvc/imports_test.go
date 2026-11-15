package rulesxxvc

import (
	"os/exec"
	"strings"
	"testing"
)

// TestImportsOnlyCore: a ruleset leaf imports only the standard library and
// gobox/rules (rules/xxvc included), directly or not (ARCHITECTURE.md Dependencies).
func TestImportsOnlyCore(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-f", "{{if not .Standard}}{{.Module.Path}} {{.ImportPath}}{{end}}", ".").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		f := strings.Fields(l) // module, package: gobox/rules is a package of the gobox module
		if f[1] != "gobox/rulesxxvc" && f[1] != "gobox/rules" && !strings.HasPrefix(f[1], "gobox/rules/") {
			t.Errorf("imports %s", l)
		}
	}
}
