package ui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestArtKeys guards the bug class "a cache misses the art switch (F12)":
// every layout memo keys on a stateBuild (the State's build and the art
// shown), and the ui reads the game's Version nowhere but in
// frontState.build, so no new cache can key on the build alone.
func TestArtKeys(t *testing.T) {
	build := reflect.TypeFor[stateBuild]()
	var holds func(reflect.Type) bool
	holds = func(k reflect.Type) bool {
		if k == build {
			return true
		}
		if k.Kind() != reflect.Struct {
			return false
		}
		for i := range k.NumField() {
			if holds(k.Field(i).Type) {
				return true
			}
		}
		return false
	}
	memoKey := func(m reflect.Type) reflect.Type { f, _ := m.FieldByName("key"); return f.Type }
	l := reflect.TypeFor[layouts]()
	for i := range l.NumField() {
		f := l.Field(i)
		if k := memoKey(f.Type); !holds(k) {
			t.Errorf("layouts.%s keys on %v: no stateBuild, the art switch misses it", f.Name, k)
		}
	}
	if f, _ := reflect.TypeFor[frontState]().FieldByName("der"); !holds(memoKey(f.Type)) {
		t.Error("frontState.der keys without stateBuild")
	}
	if f, _ := reflect.TypeFor[Presenter]().FieldByName("back"); !holds(memoKey(f.Type)) {
		t.Error("Presenter.back keys without stateBuild")
	}

	files, _ := filepath.Glob("*.go")
	fs := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fs, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if ok && fd.Name.Name == "build" && fd.Recv != nil {
				return false // frontState.build: the one read
			}
			c, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if s, sel := c.Fun.(*ast.SelectorExpr); sel && s.Sel.Name == "Version" && len(c.Args) == 0 {
				t.Errorf("%s: Version() read outside frontState.build: key on build() so the art switch reaches it", fs.Position(c.Pos()))
			}
			return true
		})
	}
}

// TestSwitchArtBuild: the art switch changes the build every derivation keys on.
func TestSwitchArtBuild(t *testing.T) {
	r := &frontState{g: still{}}
	am := daxArt.amiga
	defer switchArt(am)
	before := r.build()
	switchArt(!am)
	if r.build() == before {
		t.Error("switchArt left the build as it was: the caches keep the other art")
	}
}
