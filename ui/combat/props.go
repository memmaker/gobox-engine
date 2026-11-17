package combat

import (
	"encoding/json"
	"image"
	_ "image/png"
	"io/fs"
)

// Props (modes flat, extrude, block): a series' combat/props.json maps a tile name to how its decoration
// draws: {"mode":"flat"|"extrude"|"mesh"|"floor", "trunk":["<tile name>", ...]}.
// flat lies on the floor, extrude stands as its extruded sprite, floor draws
// the plain floor alone (a piece of the original's drawn wall a model replaces), mesh is a
// glTF from Meshes (a tile with one is mesh listed or not). A listed
// extrude crown's trunk tiles: when one is south of it, the trunk stands
// under the crown and its own square is floor. Unlisted tiles are inferred.

type propJSON struct {
	Mode  string   `json:"mode"`
	Trunk []string `json:"trunk"`
	Back  string   `json:"back"` // a model's backrest side in its own frame (n, e, s, w): it turns that side to a wall beside it
}

// propSet is props.json by tile index.
type propSet struct {
	mode  [256]string // "flat", "extrude", "block", "slab" or "floor"; "": inferred
	trunk [256][]int  // the tiles a crown stands on (nil: the inferred rule)
	top   [256]int    // a block's top atlas slot (0: its side picture)
	back  [256]string // a model's backrest side (props.json "back"); "": it keeps its heading
}

func (ps *propSet) modeOf(t int) string {
	if ps == nil || t < 0 || t >= len(ps.mode) {
		return ""
	}
	return ps.mode[t]
}

func (ps *propSet) trunkOf(t int) []int {
	if ps == nil || t < 0 || t >= len(ps.trunk) {
		return nil
	}
	return ps.trunk[t]
}

// readProps reads props.json from fsys for the tile names (nil: none, or unreadable).
func readProps(fsys fs.FS, names []string) *propSet {
	if fsys == nil {
		return nil
	}
	b, err := fs.ReadFile(fsys, "props.json")
	if err != nil {
		return nil
	}
	var m map[string]propJSON
	if json.Unmarshal(b, &m) != nil {
		return nil
	}
	byName := map[string][]int{}
	for i, n := range names {
		if i < 256 && n != "" {
			byName[n] = append(byName[n], i)
		}
	}
	ps := &propSet{}
	for n, p := range m {
		for _, i := range byName[n] {
			if p.Mode == "flat" || p.Mode == "extrude" || p.Mode == "block" || p.Mode == "slab" {
				ps.mode[i] = p.Mode
			}
			ps.back[i] = p.Back
			if p.Mode == "extrude" {
				for _, tn := range p.Trunk {
					ps.trunk[i] = append(ps.trunk[i], byName[tn]...)
				}
			}
		}
	}
	return ps
}

// readBlocks reads a block tile's own cube pictures, blocks/<name>.png
// (side) and blocks/<name>_top.png, by atlas slot: a side takes its tile's
// slot (the atlas only; the tile itself is untouched), a top the next free
// one after floorSlot, recorded in ps.top.
func readBlocks(fsys fs.FS, names []string, ps *propSet) map[int]image.Image {
	pic := func(n string) image.Image {
		f, err := fsys.Open("blocks/" + n + ".png")
		if err != nil {
			return nil
		}
		defer f.Close()
		m, _, _ := image.Decode(f)
		return m
	}
	out, slot := map[int]image.Image{}, floorSlot+1
	for i, n := range names {
		if fsys == nil || ps.modeOf(i) != "block" && ps.modeOf(i) != "slab" {
			continue
		}
		if m := pic(n); m != nil {
			out[i] = m
		}
		if m := pic(n + "_top"); m != nil && slot < atlasCols*atlasRows {
			out[slot], ps.top[i] = m, slot
			slot++
		}
	}
	return out
}
