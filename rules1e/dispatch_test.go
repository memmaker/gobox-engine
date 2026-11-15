package rules1e

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// parseDir parses the non-test Go files of dir.
func parseDir(t *testing.T, dir string) []*ast.File {
	t.Helper()
	pkgs, err := parser.ParseDir(token.NewFileSet(), dir, func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	for _, p := range pkgs {
		for _, f := range p.Files {
			files = append(files, f)
		}
	}
	return files
}

// coreNames returns the core names a Rules answers, by the interface they
// satisfy (the marker methods action, check, fact): the types with that
// method, and for facts also the PlainFact constants.
func coreNames(t *testing.T) map[string]string {
	t.Helper()
	out, err := exec.Command("go", "list", "-f", "{{.Dir}}", "gobox/rules").Output()
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]string{} // name -> marker
	for _, f := range parseDir(t, strings.TrimSpace(string(out))) {
		for _, d := range f.Decls {
			if d, ok := d.(*ast.FuncDecl); ok {
				if d.Recv == nil || len(d.Recv.List) != 1 || (d.Name.Name != "action" && d.Name.Name != "check" && d.Name.Name != "fact") {
					continue
				}
				recv := d.Recv.List[0].Type
				if s, ok := recv.(*ast.StarExpr); ok {
					recv = s.X
				}
				if id, ok := recv.(*ast.Ident); ok {
					names[id.Name] = d.Name.Name
				}
			}
		}
	}
	// PlainFact constants of a group continue the first spec's type.
	for _, f := range parseDir(t, strings.TrimSpace(string(out))) {
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != token.CONST || len(g.Specs) == 0 {
				continue
			}
			first, ok := g.Specs[0].(*ast.ValueSpec)
			if !ok || first.Type == nil {
				continue
			}
			if id, ok := first.Type.(*ast.Ident); ok && id.Name == "PlainFact" {
				for _, s := range g.Specs {
					for _, n := range s.(*ast.ValueSpec).Names {
						names[n.Name] = "key"
					}
				}
			}
		}
	}
	delete(names, "_")
	delete(names, "PlainFact")
	return names
}

// caseName is the rules name a case expression mentions: rules.T, rules.T{...}.
func caseName(e ast.Expr) string {
	if c, ok := e.(*ast.CompositeLit); ok {
		e = c.Type
	}
	if s, ok := e.(*ast.SelectorExpr); ok {
		if x, ok := s.X.(*ast.Ident); ok && x.Name == "rules" {
			return s.Sel.Name
		}
	}
	return ""
}

// notAnswered are the core names this ruleset has no hook for: Resolve ends
// in ErrNotPorted, Fact in errNoFact, Check answers nothing. They belong to
// another ruleset (rulesxxvc: ships, shop and loot verbs, SkillRoll) or to
// the creation and engine paths (rules1e.Creation, the Engine's own verbs).
// A name that gains a hook must leave this list (the test says so).
var notAnswered = map[string]bool{
	"Aid":           true,
	"Appraise":      true,
	"ChooseAlign":   true,
	"ChooseClass":   true,
	"ChooseRace":    true,
	"ChooseSex":     true,
	"Dodge":         true,
	"DropDraft":     true,
	"Fix":           true,
	"Forget":        true,
	"GiveItem":      true,
	"Halve":         true,
	"Heal":          true,
	"Identify":      true,
	"Intimidate":    true,
	"Jetpack":       true,
	"JoinStack":     true,
	"KeepDraft":     true,
	"LayOnHands":    true,
	"Memorize":      true,
	"Modify":        true,
	"NewCharacter":  true,
	"PickSkill":     true,
	"PointBuy":      true,
	"Reload":        true,
	"RemoveMember":  true,
	"Rename":        true,
	"Repair":        true,
	"Reroll":        true,
	"Rest":          true,
	"ScriptGainExp": true,
	"SetScore":      true,
	"ShipBoard":     true,
	"ShipBoost":     true,
	"ShipClose":     true,
	"ShipCommand":   true,
	"ShipDodge":     true,
	"ShipFire":      true,
	"ShipJuryRig":   true,
	"ShipRam":       true,
	"ShipReload":    true,
	"ShipRetreat":   true,
	"ShipSensors":   true,
	"Shoot":         true,
	"SpringTrap":    true,
	"Sprint":        true,
	"TakeItem":      true,
	"TakeMoney":     true,
	"SkillRoll":     true,
	"KnownFact":     true,
	"FlagFact":      true, // XXVc's script flag bytes on a character
	"LevelFact":     true,
	"MemoFact":      true,
}

// TestDispatchDisjoint: every core Action, Check and Fact name is a case of
// exactly one topic hook (the checkX, resolveX and factX methods of Rules),
// so no name is answered by two hooks and none by none.
func TestDispatchDisjoint(t *testing.T) {
	names := coreNames(t)
	hook := regexp.MustCompile(`^(check|resolve|fact)[A-Z]`)
	handled := map[string]map[string]bool{} // core name -> hooks with a case for it
	for _, f := range parseDir(t, ".") {
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Body == nil || !hook.MatchString(fn.Name.Name) {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if cc, ok := n.(*ast.CaseClause); ok {
					for _, e := range cc.List {
						if nm := caseName(e); names[nm] != "" {
							if handled[nm] == nil {
								handled[nm] = map[string]bool{}
							}
							handled[nm][fn.Name.Name] = true
						}
					}
				}
				return true
			})
		}
	}
	var bad []string
	for nm, kind := range names {
		switch hooks := handled[nm]; {
		case len(hooks) == 0 && kind == "key":
			// a PlainFact key may be a WordFact's Of, answered by word(), not by a hook
		case len(hooks) == 0 && notAnswered[nm]:
		case len(hooks) == 0:
			bad = append(bad, kind+" "+nm+": no hook")
		case notAnswered[nm]:
			bad = append(bad, kind+" "+nm+": has a hook, remove it from notAnswered")
		case len(hooks) > 1:
			var hs []string
			for h := range hooks {
				hs = append(hs, h)
			}
			sort.Strings(hs)
			bad = append(bad, kind+" "+nm+": "+strings.Join(hs, ", "))
		}
	}
	for nm := range notAnswered {
		if names[nm] == "" {
			bad = append(bad, nm+": in notAnswered but not a core name")
		}
	}
	sort.Strings(bad)
	for _, b := range bad {
		t.Error(b)
	}
}
