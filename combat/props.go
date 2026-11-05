package combat

import (
	"encoding/json"
	"io/fs"
)

// Props: a series' combat/props.json maps a tile name to how its decoration
// draws: {"mode":"flat"|"extrude"|"mesh", "trunk":["<tile name>", ...]}.
// flat lies on the floor, extrude stands as its extruded sprite, mesh is a
// glTF from Meshes (a tile with one is mesh listed or not). A listed
// extrude crown's trunk tiles: when one is south of it, the trunk stands
// under the crown and its own square is floor. Unlisted tiles are inferred.

type propJSON struct {
	Mode  string   `json:"mode"`
	Trunk []string `json:"trunk"`
}

// propSet is props.json by tile index.
type propSet struct {
	mode  [256]string // "flat" or "extrude"; "": inferred
	trunk [256][]int  // the tiles a crown stands on (nil: the inferred rule)
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
			if p.Mode == "flat" || p.Mode == "extrude" {
				ps.mode[i] = p.Mode
			}
			if p.Mode == "extrude" {
				for _, tn := range p.Trunk {
					ps.trunk[i] = append(ps.trunk[i], byName[tn]...)
				}
			}
		}
	}
	return ps
}
