// Command poolsprites builds series/<series>/{portraits,combat}: every portrait
// and combat icon of the series' games, deduplicated by exact pixels, plus
// series/<series>/origins.json (and, for the pool series, INDEX.md's tables).
//
//	go run . [-series pool|buck-rogers] [-games ~/Projects] [-out ../..]   (run in sprites/cmd/poolsprites)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gobox/dax"
	"gobox/ega"
)

// source is one kind of picture: the files (per game) it is read from.
type source struct {
	dir, prefix string // output kind dir and pool name prefix
	masked      bool   // colour 0 is transparent (combat art)
	files       []string
	byFrame     bool // named <prefix>_<frame> with no dedup, so the game can name a tile by its frame
}

var (
	games = []string{"por", "cotab", "sotsb", "pod"}
	vga   = map[string]bool{"pod": true, "buckrogers": true} // games whose pictures are VGA (dax.ParseVGA) not EGA
	// The per-game file lists: the only game knowledge here.
	sources = []source{
		{"portraits", "head", false, []string{"HEAD1", "HEAD2", "HEAD3", "HEAD4", "HEAD5", "HEAD6", "HEAD7", "HEAD8"}, false},
		{"portraits", "body", false, []string{"BODY1", "BODY2", "BODY3", "BODY4", "BODY5", "BODY6", "BODY7", "BODY8"}, false},
		{"combat", "combat_head", true, []string{"CHEAD"}, false},
		{"combat", "combat_body", true, []string{"CBODY"}, false},
		{"combat", "combat_sprite", true, []string{"COMSPR"}, false},
		{"combat", "combat_icon", true, []string{"ICON"}, false},
	}
	// The buck-rogers series: CPIC is the character picture, CHARS the characters'
	// combat icons, COMSPR the combat sprites (buckroger-re src/O157.PAS, O0CA.PAS).
	brGames   = []string{"buckrogers"}
	brSources = []source{
		{"portraits", "portrait", false, []string{"CPIC1", "CPIC2", "CPIC3", "CPIC4", "CPIC5", "CPIC6"}, false},
		{"combat", "combat_icon", true, []string{"CHARS"}, false},
		{"combat", "combat_sprite", true, []string{"COMSPR"}, false},
		// The combat map tile sets (pO157_0022): block 1, one 24x24 tile a frame.
		{"combat", "hitek_tile", false, []string{"HITEK"}, true},
		{"combat", "lotek_tile", false, []string{"LOTEK"}, true},
		{"combat", "marscom_tile", false, []string{"MARSCOM"}, true},
		{"combat", "venuscom_tile", false, []string{"VENUSCOM"}, true},
	}
)

var games2 []string // the <game>-amiga sets found

type origin struct{ game, file, index string }

type entry struct {
	name    string
	w, h    int
	origins []origin
}

func main() {
	root := flag.String("games", "..", "directory holding <game>-go")
	out := flag.String("out", "../..", "sprites module directory")
	series := flag.String("series", "pool", "series to build: pool (por, cotab, sotsb, pod) or buck-rogers")
	flag.Parse()
	switch *series {
	case "pool":
	case "buck-rogers":
		games, sources = brGames, brSources
	default:
		check(fmt.Errorf("unknown series %q", *series))
	}
	pool := filepath.Join(*out, "series", *series)
	byDir := map[string][]*entry{}
	origins := map[string]map[string]map[string]string{}
	count := map[string]int{}             // prefix -> next number
	seen := map[string]*entry{}           // kind dir + pixels -> entry
	before := map[string]map[string]int{} // game -> dir -> pictures decoded
	after := map[string]map[string]int{}  // game -> dir -> pictures first seen there
	type run struct{ game, key, dir string }
	var runs []run
	for _, g := range games {
		runs = append(runs, run{g, g, "dos"})
	}
	for _, r := range runs {
		g := r.key
		before[g], after[g], origins[g] = map[string]int{}, map[string]int{}, map[string]map[string]string{}
		for _, s := range sources {
			files := s.files
			if r.dir == "amiga" {
				files = []string{strings.TrimRight(s.files[0], "12345678")}
			}
			for _, f := range files {
				path := filepath.Join(*root, r.game+"-go", "data", r.dir, f+".DAX")
				d, err := dax.Open(path)
				if err != nil {
					continue
				}
				for _, e := range d.Entries {
					blk := d.BlockShared(e.ID)
					im := dax.ParseImage(blk, s.masked, 0)
					if vga[r.game] {
						im = dax.ParseVGA(blk, s.masked, 0)
					}
					if im == nil || im.Count == 0 || im.Width == 0 || im.Height == 0 || len(im.Data) < im.Count*im.BPP {
						continue
					}
					for fr := 0; fr < im.Count; fr++ {
						idx := fmt.Sprint(e.ID)
						if im.Count > 1 {
							idx += fmt.Sprintf(".%d", fr)
						}
						pix := render(im, fr)
						key := s.dir + fmt.Sprint(im.Width, im.Height) + s.prefix + string(pix.Pix)
						n := count[s.prefix]
						if s.byFrame {
							key, n = s.prefix+idx, fr
						}
						en := seen[key]
						if en == nil {
							en = &entry{name: fmt.Sprintf("%s_%03d", s.prefix, n), w: im.Width * 8, h: im.Height}
							count[s.prefix]++
							seen[key] = en
							byDir[s.dir] = append(byDir[s.dir], en)
							if s.masked { // combat art keeps its palette indices (gobox recolours the icon parts by them)
								write(filepath.Join(pool, s.dir, en.name+".png"), indexed(im, fr, nil))
							} else {
								write(filepath.Join(pool, s.dir, en.name+".png"), pix)
							}
							after[g][s.dir]++
						}
						before[g][s.dir]++
						en.origins = append(en.origins, origin{g, f + ".DAX", idx})
						if origins[g][f+".DAX"] == nil {
							origins[g][f+".DAX"] = map[string]string{}
						}
						origins[g][f+".DAX"][idx] = en.name
					}
				}
				fmt.Println("decoded", g, f)
			}
		}
	}
	// The Amiga art (data/amiga, the game's -import-disk2 of its .adf disks): its own
	// amiga_-prefixed names under origins key <game>-amiga, never deduplicated
	// against the DOS art, so the F12 toggle keeps the two sets apart.
	for _, g := range games {
		dos, err := dax.NewLoader(filepath.Join(*root, g+"-go", "data", "dos"))
		if err != nil {
			continue
		}
		dos.UseAmiga(filepath.Join(*root, g+"-go", "data", "amiga"))
		if !dos.HasAmiga() {
			continue
		}
		ag := g + "-amiga"
		before[ag], after[ag], origins[ag] = map[string]int{}, map[string]int{}, map[string]map[string]string{}
		games2 = append(games2, ag)
		for _, s := range sources {
			for _, base := range slices.Compact(append([]string{strings.TrimRight(s.files[0], "0123456789")}, s.files...)) { // the Amiga keeps one archive per kind, plus a set's own (HEAD5: Curse's 549)
				d, err := dax.Open(filepath.Join(*root, g+"-go", "data", "amiga", base+".DAX"))
				if err != nil || s.byFrame {
					continue
				}
				prefix := "amiga_" + s.prefix
				for _, e := range d.Entries {
					im := dax.ParseImage(dos.BlockShared(base, e.ID), s.masked, 0) // the loader's block: HEAD centred as drawn
					if im == nil || im.Count == 0 || im.Width == 0 || im.Height == 0 || len(im.Data) < im.Count*im.BPP {
						continue
					}
					pal := dos.AmigaPalette(base, e.ID)
					for fr := 0; fr < im.Count; fr++ {
						idx := fmt.Sprint(e.ID)
						if im.Count > 1 {
							idx += fmt.Sprintf(".%d", fr)
						}
						pix := renderAmiga(im, fr, pal)
						key := s.dir + fmt.Sprint(im.Width, im.Height) + prefix + string(pix.Pix)
						en := seen[key]
						if en == nil {
							en = &entry{name: fmt.Sprintf("%s_%03d", prefix, count[prefix]), w: im.Width * 8, h: im.Height}
							count[prefix]++
							seen[key] = en
							byDir[s.dir] = append(byDir[s.dir], en)
							if s.masked { // combat art keeps its values (Amiga ones from ega.AmigaBase) for gobox's recolouring
								write(filepath.Join(pool, s.dir, en.name+".png"), indexed(im, fr, pal))
							} else {
								write(filepath.Join(pool, s.dir, en.name+".png"), pix)
							}
							after[ag][s.dir]++
						}
						before[ag][s.dir]++
						en.origins = append(en.origins, origin{ag, base + ".DAX", idx})
						if origins[ag][base+".DAX"] == nil {
							origins[ag][base+".DAX"] = map[string]string{}
						}
						origins[ag][base+".DAX"][idx] = en.name
					}
				}
				fmt.Println("decoded", ag, base)
			}
		}
	}
	games = append(games, games2...)
	b, _ := json.MarshalIndent(origins, "", " ")
	check(os.WriteFile(filepath.Join(pool, "origins.json"), append(b, '\n'), 0o644))
	poses := map[string]string{} // combat head/body frame 0 -> its pose-1 frame (the first origin's)
	for _, files := range origins {
		for _, refs := range files {
			for ref, name := range refs {
				if b, ok := strings.CutSuffix(ref, ".0"); ok && (strings.HasPrefix(name, "combat_head_") || strings.HasPrefix(name, "combat_body_")) {
					if p := refs[b+".1"]; p != "" && poses[name] == "" {
						poses[name] = p
					}
				}
			}
		}
	}
	if *series == "pool" {
		b, _ = json.MarshalIndent(poses, "", " ")
		check(os.WriteFile(filepath.Join(pool, "combat", "poses.json"), append(b, '\n'), 0o644))
		index(filepath.Join(*out, "INDEX.md"), byDir)
	}
	for _, r := range runs {
		g := r.key
		for _, k := range []string{"portraits", "combat"} {
			fmt.Printf("%s %s: %d decoded, %d new\n", g, k, before[g][k], after[g][k])
		}
	}
	fmt.Println(*series, "portraits", len(byDir["portraits"]), "combat", len(byDir["combat"]))
}

// render is frame fr in RGBA: the image's own palette (VGA) or EGA, Clear transparent.
func render(im *dax.Image, fr int) *image.NRGBA {
	w := im.Width * 8
	out := image.NewNRGBA(image.Rect(0, 0, w, im.Height))
	for i, v := range im.Frame(fr)[:w*im.Height] {
		switch {
		case v == dax.Clear && im.Palette == nil:
		case v >= ega.AmigaBase && im.Palette == nil:
			out.SetNRGBA(i%w, i/w, color.NRGBA(ega.AmigaRGBA(v)))
		case im.Palette != nil:
			if c := im.Palette[v]; c.A != 0 {
				out.SetNRGBA(i%w, i/w, color.NRGBA(c))
			}
		default:
			out.SetNRGBA(i%w, i/w, color.NRGBA(ega.RGBA(int(v))))
		}
	}
	return out
}

// renderAmiga is an Amiga frame in RGBA as gobox draws it (present.pixelRGBA):
// EGA below ega.AmigaBase, above it the picture's own palette pal, else ega.Amiga.
func renderAmiga(im *dax.Image, fr int, pal []byte) *image.NRGBA {
	w := im.Width * 8
	out := image.NewNRGBA(image.Rect(0, 0, w, im.Height))
	for i, v := range im.Frame(fr)[:w*im.Height] {
		var c color.RGBA
		switch n := len(pal) / 3; {
		case v == dax.Clear:
			continue
		case v < ega.AmigaBase:
			c = ega.RGBA(int(v))
		case n > 0:
			j := int(v-ega.AmigaBase) % n * 3
			c = color.RGBA{pal[j], pal[j+1], pal[j+2], 255}
		default:
			c = ega.AmigaRGBA(v)
		}
		out.SetNRGBA(i%w, i/w, color.NRGBA(c))
	}
	return out
}

// indexed is frame fr as a paletted image keeping the DAX indices: EGA art
// has 17 palette entries (the 16 colours, then transparent for Clear), VGA
// art its own palette, Amiga art (values from ega.AmigaBase) those 17 padded
// with transparent up to the Amiga colours, so it keeps its values too.
func indexed(im *dax.Image, fr int, apal []byte) *image.Paletted {
	w := im.Width * 8
	var pal color.Palette
	amiga := false
	for _, v := range im.Frame(fr)[:w*im.Height] {
		amiga = amiga || im.Palette == nil && v >= ega.AmigaBase && v != dax.Clear
	}
	if amiga {
		for i := 0; i < ega.AmigaBase+16*len(ega.Amiga); i++ {
			switch {
			case i < 16:
				pal = append(pal, color.NRGBA(ega.RGBA(i)))
			case i < ega.AmigaBase:
				pal = append(pal, color.NRGBA{})
			case len(apal) >= 3: // the game's own palette (its *PAL.DAX), as renderAmiga
				j := (i - ega.AmigaBase) % (len(apal) / 3) * 3
				pal = append(pal, color.NRGBA{apal[j], apal[j+1], apal[j+2], 255})
			default:
				pal = append(pal, color.NRGBA(ega.AmigaRGBA(uint8(i))))
			}
		}
		out := image.NewPaletted(image.Rect(0, 0, w, im.Height), pal)
		copy(out.Pix, im.Frame(fr)[:w*im.Height])
		return out
	}
	if im.Palette != nil {
		for _, c := range im.Palette {
			pal = append(pal, color.NRGBA(c))
		}
	} else {
		for i := 0; i < 16; i++ {
			pal = append(pal, color.NRGBA(ega.RGBA(i)))
		}
		pal = append(pal, color.NRGBA{})
	}
	out := image.NewPaletted(image.Rect(0, 0, w, im.Height), pal)
	for i, v := range im.Frame(fr)[:w*im.Height] {
		switch {
		case im.Palette != nil:
			out.Pix[i] = v
		case v == dax.Clear:
			out.Pix[i] = 16
		default:
			out.Pix[i] = v & 15
		}
	}
	return out
}

func write(path string, im image.Image) {
	check(os.MkdirAll(filepath.Dir(path), 0o755))
	f, err := os.Create(path)
	check(err)
	check(png.Encode(f, im))
	check(f.Close())
}

// index rewrites the portraits and combat tables (and pool counts) of INDEX.md.
func index(path string, byDir map[string][]*entry) {
	b, err := os.ReadFile(path)
	check(err)
	s := string(b)
	s, _, _ = strings.Cut(s, "\n## portraits") // our tables are always appended last
	s = regexp.MustCompile(`Pools: portraits \(\d+\), combat \(\d+\)`).ReplaceAllString(s,
		fmt.Sprintf("Pools: portraits (%d), combat (%d)", len(byDir["portraits"]), len(byDir["combat"])))
	s = strings.TrimRight(s, "\n") + "\n"
	for _, k := range []string{"portraits", "combat"} {
		s += "\n## " + k + "\n\nShared by por, cotab, sotsb, pod; deduplicated by exact pixels, numbered in first-seen order. Machine-readable: series/pool/origins.json.\n\n| name | size | origin | source |\n|---|---|---|---|\n"
		for _, e := range byDir[k] {
			var gs, src []string
			for _, o := range e.origins {
				if len(gs) == 0 || !strings.Contains(strings.Join(gs, ","), o.game) {
					gs = append(gs, o.game)
				}
				src = append(src, o.game+"/"+o.file+" "+o.index)
			}
			s += fmt.Sprintf("| %s | %dx%d | %s | %s |\n", e.name, e.w, e.h, strings.Join(gs, ", "), strings.Join(src, "; "))
		}
	}
	check(os.WriteFile(path, []byte(s), 0o644))
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
