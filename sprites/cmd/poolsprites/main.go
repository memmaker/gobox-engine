// Command poolsprites builds series/pool/{portraits,combat}: every portrait
// head/body and combat icon of the four Pool games, deduplicated by exact
// pixels, plus INDEX.md's tables and series/pool/origins.json.
//
//	go run . [-games ~/Projects] [-out ../..]   (run in sprites/cmd/poolsprites)
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
	"strings"

	"gobox/dax"
	"gobox/ega"
)

// source is one kind of picture: the files (per game) it is read from.
type source struct {
	dir, prefix string // output kind dir and pool name prefix
	masked      bool   // colour 0 is transparent (combat art)
	files       []string
}

var (
	games = []string{"por", "cotab", "sotsb", "pod"}
	vga   = map[string]bool{"pod": true} // games whose pictures are VGA (dax.ParseVGA) not EGA
	// The per-game file lists: the only game knowledge here.
	sources = []source{
		{"portraits", "head", false, []string{"HEAD1", "HEAD2", "HEAD3", "HEAD4", "HEAD5", "HEAD6", "HEAD7", "HEAD8"}},
		{"portraits", "body", false, []string{"BODY1", "BODY2", "BODY3", "BODY4", "BODY5", "BODY6", "BODY7", "BODY8"}},
		{"combat", "combat_head", true, []string{"CHEAD"}},
		{"combat", "combat_body", true, []string{"CBODY"}},
		{"combat", "combat_sprite", true, []string{"COMSPR"}},
		{"combat", "combat_icon", true, []string{"ICON"}},
	}
)

type origin struct{ game, file, index string }

type entry struct {
	name    string
	w, h    int
	origins []origin
}

func main() {
	root := flag.String("games", "..", "directory holding <game>-go")
	out := flag.String("out", "../..", "sprites module directory")
	flag.Parse()
	pool := filepath.Join(*out, "series", "pool")
	byDir := map[string][]*entry{}
	origins := map[string]map[string]map[string]string{}
	count := map[string]int{}             // prefix -> next number
	seen := map[string]*entry{}           // kind dir + pixels -> entry
	before := map[string]map[string]int{} // game -> dir -> pictures decoded
	after := map[string]map[string]int{}  // game -> dir -> pictures first seen there
	for _, g := range games {
		before[g], after[g], origins[g] = map[string]int{}, map[string]int{}, map[string]map[string]string{}
		for _, s := range sources {
			for _, f := range s.files {
				path := filepath.Join(*root, g+"-go", "data", "dos", f+".DAX")
				d, err := dax.Open(path)
				if err != nil {
					continue
				}
				for _, e := range d.Entries {
					blk := d.BlockShared(e.ID)
					im := dax.ParseImage(blk, s.masked, 0)
					if vga[g] {
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
						en := seen[key]
						if en == nil {
							en = &entry{name: fmt.Sprintf("%s_%03d", s.prefix, count[s.prefix]), w: im.Width * 8, h: im.Height}
							count[s.prefix]++
							seen[key] = en
							byDir[s.dir] = append(byDir[s.dir], en)
							if s.masked { // combat art keeps its palette indices (gobox recolours the icon parts by them)
								write(filepath.Join(pool, s.dir, en.name+".png"), indexed(im, fr))
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
	b, _ = json.MarshalIndent(poses, "", " ")
	check(os.WriteFile(filepath.Join(pool, "combat", "poses.json"), append(b, '\n'), 0o644))
	index(filepath.Join(*out, "INDEX.md"), byDir)
	for _, g := range games {
		for _, k := range []string{"portraits", "combat"} {
			fmt.Printf("%s %s: %d decoded, %d new\n", g, k, before[g][k], after[g][k])
		}
	}
	fmt.Println("pool portraits", len(byDir["portraits"]), "combat", len(byDir["combat"]))
}

// render is frame fr in RGBA: the image's own palette (VGA) or EGA, Clear transparent.
func render(im *dax.Image, fr int) *image.NRGBA {
	w := im.Width * 8
	out := image.NewNRGBA(image.Rect(0, 0, w, im.Height))
	for i, v := range im.Frame(fr)[:w*im.Height] {
		switch {
		case v == dax.Clear && im.Palette == nil:
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

// indexed is frame fr as a paletted image keeping the DAX indices: EGA art
// has 17 palette entries (the 16 colours, then transparent for Clear), VGA
// art its own palette.
func indexed(im *dax.Image, fr int) *image.Paletted {
	w := im.Width * 8
	var pal color.Palette
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
