package ui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"gobox/asset"
	"path/filepath"
	"reflect"
	"slices"
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

// TestGlyphKeys guards the bug class "a cache of drawn text misses a font
// change" (the font menu, F12's art): the Presenter keeps no bare artGen but
// for picture caches (a text cache keys on glyphGen, which holds the art and the font), the
// card and combat-row keys hold a glyphGen, and fonts.Cur is read only in
// glyphs.
func TestGlyphKeys(t *testing.T) {
	gen, glyphs := reflect.TypeFor[artGen](), reflect.TypeFor[glyphGen]()
	pr := reflect.TypeFor[Presenter]()
	pictures := map[string]bool{"viewArt": true} // the 3D view's textures: pictures, no text
	for i := range pr.NumField() {
		if f := pr.Field(i); f.Type == gen && !pictures[f.Name] {
			t.Errorf("Presenter.%s is an artGen: key text caches on glyphGen (the font too)", f.Name)
		}
	}
	for _, k := range []reflect.Type{reflect.TypeFor[cardKey](), reflect.TypeFor[rowsKey]()} {
		if !slices.ContainsFunc(reflect.VisibleFields(k), func(f reflect.StructField) bool { return f.Type == glyphs }) {
			t.Errorf("%v holds no glyphGen: a font change keeps the old words", k)
		}
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
			if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == "glyphs" {
				return false
			}
			if s, ok := n.(*ast.SelectorExpr); ok && s.Sel.Name == "Cur" {
				if x, ok := s.X.(*ast.SelectorExpr); ok && x.Sel.Name == "fonts" {
					t.Errorf("%s: fonts.Cur read outside glyphs: key on glyphs() so a font change reaches it", fs.Position(s.Pos()))
				}
			}
			return true
		})
	}
}

// TestGlyphsFont: choosing another font changes the glyphs the text caches key on.
func TestGlyphsFont(t *testing.T) {
	p := &Presenter{fonts: &asset.Chooser{}}
	before := p.glyphs()
	p.fonts.Cur++
	if p.glyphs() == before {
		t.Error("another font left the glyphs as they were: the cards keep the old font")
	}
}
