package main

import (
	"os"
	"reflect"
	"testing"

	"gobox/adventure"
	"gobox/asset"
	"gobox/asset/orig"
	"gobox/game"

	"gobox/engine"
)

// pictures calls f on every game.Picture reachable from v.
func pictures(v reflect.Value, f func(game.Picture)) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			pictures(v.Elem(), f)
		}
	case reflect.Struct:
		if p, ok := v.Interface().(game.Picture); ok {
			f(p)
			return
		}
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				pictures(v.Field(i), f)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			pictures(v.Index(i), f)
		}
	case reflect.Map:
		for it := v.MapRange(); it.Next(); {
			pictures(it.Value(), f)
		}
	}
}

// TestFightPicturesResolve (BUG-CLASSES 43): every Picture a fight's State
// carries resolves: pixels, or a DAX File and Block with a non-empty block,
// or a sprite Name; every tile name is in Art(). The dev fights start through
// addMonsters, the load opcode's own (opLoadMonster, eclops.go).
func TestFightPicturesResolve(t *testing.T) {
	for _, sc := range []string{"combat", "missile", "castfight", "wildcombat", "meleefight"} {
		if _, err := os.Stat("../../data/por/sources/dos/TITLE.DAX"); err != nil {
			t.Skip("no game files")
		}
		e, err := engine.NewPool("../../data/por", "../../data/por/sources")
		if err != nil {
			t.Fatal(err)
		}
		ref, err := orig.NewLoader("../../data/por/sources/" + adventure.DOSDir) // the reference decoder: a picture's block is in the archives
		if err != nil {
			t.Fatal(err)
		}
		ref.UseAmiga("../../data/por/sources/" + adventure.AmigaDir)
		e.Cfg.SaveDir = t.TempDir()
		e.Explore = true
		e.DevAction = e.DevScene(sc)
		seed := uint32(1)
		e.FixedSeed = &seed
		e.Init()
		e.Start(e.Play)
		for range 200 {
			st := e.State()
			if f, ok := st.(game.Fighting); ok {
				pictures(reflect.ValueOf(f), func(p game.Picture) {
					switch {
					case p.Pixels != nil || p.File == "" && p.Name == "":
					case p.File != "":
						if f := ref.File(p.File); f == nil || len(f.BlockShared(p.Block)) == 0 && p.Block >= 0 {
							t.Fatalf("%s: %s block %d does not resolve", sc, p.File, p.Block)
						}
					case asset.SpriteImage("pool", "combat", p.Name) == nil && asset.SpriteImage("pool", "icon", p.Name) == nil:
						t.Errorf("%s: sprite %q does not resolve", sc, p.Name)
					}
				})
				art := e.Art()
				for _, n := range f.Field.Tiles {
					if _, ok := art[n]; n != "" && !ok {
						t.Fatalf("%s: tile %q not in Art()", sc, n)
					}
				}
			}
			if _, err := e.Do(game.Reply{}); err == nil { // what it asks, first answer
				continue
			}
			if _, err := e.Do(game.Quick{}); err == nil {
				continue
			}
			if _, err := e.Do(game.ContinueAction{}); err != nil {
				e.Step()
			}
		}
	}
}
