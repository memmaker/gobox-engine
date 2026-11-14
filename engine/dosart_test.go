package engine

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDOSArt: an Amiga picture carries its DOS twin, and the published
// picture carries both, whatever art shows (DOSArt): the front picks.
func TestDOSArt(t *testing.T) {
	if _, err := os.Stat("../data/por/sources/amiga/PIC.DAX"); err != nil {
		t.Skip("no Amiga import")
	}
	e := testEngine(t)
	e.LoadAnimPic("PIC", 1, false)
	if len(e.pic.frames) == 0 || e.pic.frames[0].Image.Alt == nil {
		t.Fatal("PIC 1 has no DOS twin")
	}
	im := e.pic.frames[0].Image
	for _, dos := range []bool{false, true} {
		DOSArt.Store(dos)
		g := gamePicture(e.picPix(im, 0, false, nil))
		if g.Alt == nil || g.Alt.W != im.Alt.Width*8 || bytes.Equal(g.Pixels, g.Alt.Pixels) {
			t.Fatalf("DOSArt %v: the picture carries no DOS twin", dos)
		}
	}
	DOSArt.Store(false)
}

// TestDOSArtReads guards the bug class "the engine bakes the art shown into
// State" (a menu's frozen backdrop then keeps the old art after F12): DOSArt
// is read in amigaUp alone, whose callers (the title's steps, the icon
// designer's colours with no picture) choose what runs, not which pixels show.
func TestDOSArtReads(t *testing.T) {
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
			if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == "amigaUp" {
				return false
			}
			if s, ok := n.(*ast.SelectorExpr); ok {
				if id, ok := s.X.(*ast.Ident); ok && id.Name == "DOSArt" {
					t.Errorf("%s: DOSArt read outside amigaUp: publish both twins (Picture.Alt), the front picks", fs.Position(id.Pos()))
				}
			}
			return true
		})
	}
}
