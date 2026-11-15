package asset

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// brSeries is the Buck Rogers art series.
const brSeries = "buck-rogers"

// brTileFrames are the frames each BR1 combat tile set holds, so the tile
// names a combat map can produce: <set>_tile_000 .. <frames-1>
// (buckrogers-go rules/cmap.go tileSet; REFACTOR.md B1c moves it to the importer).
var brTileFrames = map[string]int{"hitek": 53, "lotek": 52, "marscom": 42, "venuscom": 44}

// brMeshes are the hand-made models of BR1's combat maps (Felix's
// decorations and walls): REFACTOR.md B1a pins them through the move of
// Buck Rogers into the engine. A model is found by its tile's name.
var brMeshes = []string{
	"hitek_tile_019", "hitek_tile_029", "hitek_tile_035", "hitek_tile_036", "hitek_tile_037", "hitek_tile_038",
	"hitek_tile_039", "hitek_tile_040", "hitek_tile_041", "hitek_tile_042", "hitek_tile_043", "hitek_tile_044",
	"lotek_tile_029", "lotek_tile_035", "lotek_tile_036", "lotek_tile_037", "lotek_tile_039", "lotek_tile_040", "lotek_tile_041",
}

var brTileName = regexp.MustCompile(`^([a-z]+)_tile_(\d{3})$`)

// brTile: name is a tile a BR combat map can produce.
func brTile(name string) error {
	m := brTileName.FindStringSubmatch(name)
	if m == nil {
		return fmt.Errorf("%s is no tile name", name)
	}
	n, _ := strconv.Atoi(m[2])
	if frames, ok := brTileFrames[m[1]]; !ok || n >= frames {
		return fmt.Errorf("%s: no such tile (set %s has %d frames)", name, m[1], frames)
	}
	return nil
}

// TestBuckRogersMeshesKept: every one of the 19 models is there with its
// model, its Blockbench source and its texture, no other model is, each is
// named for a tile a combat map can produce, and that tile's own picture
// exists; props.json names only such tiles.
func TestBuckRogersMeshesKept(t *testing.T) {
	meshes, combat := Meshes(brSeries), Combat(brSeries)
	var models []string
	entries, err := fs.ReadDir(meshes, ".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if name, ok := strings.CutSuffix(e.Name(), ".gltf"); ok {
			models = append(models, name)
		}
	}
	slices.Sort(models)
	if !slices.Equal(models, brMeshes) {
		t.Fatalf("models %v,\nwant %v", models, brMeshes)
	}
	for _, name := range brMeshes {
		if err := brTile(name); err != nil {
			t.Error(err)
		}
		for _, ext := range []string{".gltf", ".bbmodel", ".png"} {
			if b, err := fs.ReadFile(meshes, name+ext); err != nil || len(b) == 0 {
				t.Errorf("%s%s: missing or empty (%v)", name, ext, err)
			}
		}
		if _, err := fs.Stat(combat, name+".png"); err != nil {
			t.Errorf("%s: the tile's own picture is missing: %v", name, err)
		}
	}
	b, err := fs.ReadFile(combat, "props.json")
	if err != nil {
		t.Fatal(err)
	}
	var props map[string]json.RawMessage
	if err := json.Unmarshal(b, &props); err != nil {
		t.Fatal(err)
	}
	if len(props) == 0 {
		t.Fatal("props.json names no tile")
	}
	for name := range props {
		if err := brTile(name); err != nil {
			t.Errorf("props.json: %v", err)
		}
	}
}
