package orig

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gobox/asset"
)

// poolGames are the Pool of Radiance series' games sharing one portrait and
// combat icon collection.
var poolGames = []string{"por", "cotab", "sotsb", "pod"}

// pictureFiles are the DAX bases holding pictures a game shows: portraits,
// combat icons, small pictures (PIC/SPRIT/CPIC) and big pictures.
var pictureFiles = []string{"HEAD", "BODY", "CHEAD", "CBODY", "COMSPR", "ICON", "PIC", "SPRIT", "CPIC", "BIGPIC"}

// pooledFiles are the pictureFiles mapped into the shared collection
// (origins/pool.json).
var pooledFiles = []string{"HEAD", "BODY", "CHEAD", "CBODY", "COMSPR", "ICON"}

const (
	poolDir    = "../sprites/series/pool"
	dataDir    = "../../data"
	maxBlockID = 256
)

func daxBase(name string) string {
	return strings.TrimRight(strings.TrimSuffix(strings.ToUpper(name), ".DAX"), "0123456789")
}

// emptyFrame reports a frame without a single non-zero pixel.
func emptyFrame(im *asset.Image) bool {
	return im == nil || !slices.ContainsFunc(im.Data, func(c byte) bool { return c != 0 })
}

// pictureCount is one game's tally.
type pictureCount struct {
	found, imported, twin, absent, broken int
}

// TestPoolPictures checks, per Pool game, every picture block of the DOS
// archives (pictureFiles): its imported sheet frames are not empty; the
// Amiga stand-in is imported where the Amiga archives have one ("absent"
// where they do not; a block they have that the sheet lost is broken);
// frame counts agree; every portrait and combat icon block of both
// platforms has an origins entry naming a pool file; every pool file is
// named and none duplicates another's pixels.
func TestPoolPictures(t *testing.T) {
	origins := asset.Origins()
	named := map[string]bool{}
	for _, g := range poolGames {
		dosDir := filepath.Join(dataDir, g, "sources", "dos")
		if _, err := os.Stat(dosDir); err != nil {
			t.Skipf("no %s archives: %v", g, err)
		}
		amigaDir := filepath.Join(dataDir, g, "sources", "amiga")
		l, err := NewLoader(dosDir)
		if err != nil {
			t.Fatal(err)
		}
		l.UseAmiga(amigaDir)
		artDir := filepath.Join(dataDir, g, "art")
		key, err := SourceKey(ArtFormat, dosDir, amigaDir)
		if err != nil {
			t.Fatal(err)
		}
		if have, _ := os.ReadFile(filepath.Join(artDir, "key.txt")); string(have) != key {
			t.Errorf("%s: art is stale: run gobox import %s %s", g, g, dosDir)
		}
		art := asset.OpenArt(artDir)
		files, err := daxNames(dosDir)
		if err != nil {
			t.Fatal(err)
		}
		var c pictureCount
		var absent []string
		for _, name := range files {
			base := daxBase(name)
			if !slices.Contains(pictureFiles, base) {
				continue
			}
			pooled := slices.Contains(pooledFiles, base)
			for id := range maxBlockID {
				b := l.File(name).BlockShared(id)
				if b == nil {
					continue
				}
				if _, ok := artBlock(name, id, b); !ok {
					continue
				}
				c.found++
				ref := fmt.Sprintf("%s %s %d", g, name, id)
				dos := art.Block(asset.DOS, name, id)
				if dos == nil || brokenBlock(t, ref+" dos", dos, b) {
					c.broken++
					continue
				}
				c.imported++
				if pooled && !hasOrigin(origins[g], name+".DAX", id) {
					t.Errorf("%s: no origins entry (origins/pool.json %q)", ref, g)
					c.broken++
				}
				if !l.HasAmiga() {
					continue
				}
				am := art.Block(asset.Amiga, name, id)
				if l.AmigaShared(name, id) == nil {
					if am != nil {
						t.Errorf("%s: an Amiga sheet block without an Amiga archive block", ref)
					}
					c.absent++
					absent = append(absent, name+":"+strconv.Itoa(id))
					continue
				}
				if am == nil {
					t.Errorf("%s: the Amiga archives have it, the import lost it", ref)
					c.broken++
					continue
				}
				if brokenBlock(t, ref+" amiga", am, l.AmigaShared(name, id)) {
					c.broken++
					continue
				}
				if am.Kind.Pictured() != dos.Kind.Pictured() {
					t.Errorf("%s: Amiga kind %s, DOS %s", ref, am.Kind, dos.Kind)
				} else if len(am.Frames) != len(dos.Frames) {
					t.Logf("%s: Amiga %d frames, DOS %d", ref, len(am.Frames), len(dos.Frames))
				}
				c.twin++
			}
		}
		t.Logf("%-5s found %4d imported %4d amiga twin %4d absent on amiga %4d broken %d", g, c.found, c.imported, c.twin, c.absent, c.broken)
		if len(absent) > 0 {
			t.Logf("%s absent on amiga: %s", g, strings.Join(absent, " "))
		}
		checkAmigaOrigins(t, g, amigaDir, origins[asset.PlatformID(g, asset.PlatformAmiga)])
		for k, dir := range map[asset.Platform]string{asset.PlatformDOS: dosDir, asset.PlatformAmiga: amigaDir} {
			pl, err := NewLoader(dir)
			if err != nil {
				t.Fatal(err)
			}
			for file, blocks := range origins[asset.PlatformID(g, k)] {
				for ref, n := range blocks {
					named[n] = true
					if pl.File(file).BlockShared(parseBlock(ref)) == nil {
						t.Errorf("origins %s %s %s (%s): no such archive block", asset.PlatformID(g, k), file, ref, n)
					}
				}
			}
		}
	}
	checkPool(t, named)
}

// parseBlock is the block of an origins key ("block" or "block.frame").
func parseBlock(ref string) int {
	n, _ := strconv.Atoi(strings.SplitN(ref, ".", 2)[0])
	return n
}

// brokenBlock reports (and fails) a sheet block with no frames, or with an
// empty frame whose archive bytes (raw) are not blank too.
func brokenBlock(t *testing.T, ref string, b *asset.Block, raw []byte) bool {
	if len(b.Frames) == 0 {
		t.Errorf("%s: no frames", ref)
		return true
	}
	for i, f := range b.Frames {
		if emptyFrame(f.Image) && !rawBlank(raw, b.Kind, i) {
			t.Errorf("%s: frame %d is empty, its archive bytes are not", ref, i)
			return true
		}
	}
	return false
}

// EGA block layouts (blockKind): a picture's header, an animation's frame
// count byte and each frame's header.
const (
	stillHeader     = 17
	animCountBytes  = 1
	animFrameHeader = 21
)

// rawBlank reports whether frame f of EGA block b (of kind k) is all zero
// bytes in the archive: a blank the original draws too. Other kinds are
// never blank.
func rawBlank(b []byte, k asset.Kind, f int) bool {
	var px []byte
	switch k {
	case asset.Still:
		px = b[stillHeader:]
	case asset.Anim:
		p := animCountBytes
		for i := 0; i <= f; i++ {
			n := le16(b, p+4) * le16(b, p+6) * 8 / packed(b[p+12:])
			px = b[p+animFrameHeader : p+animFrameHeader+n]
			p += animFrameHeader + n
		}
	default:
		return false
	}
	return !slices.ContainsFunc(px, func(c byte) bool { return c != 0 })
}

// hasOrigin reports whether file's block id (whole or any frame) is mapped.
func hasOrigin(o map[string]map[string]string, file string, id int) bool {
	for k := range o[file] {
		if k == strconv.Itoa(id) || strings.HasPrefix(k, strconv.Itoa(id)+".") {
			return true
		}
	}
	return false
}

// checkAmigaOrigins fails each portrait or combat icon block of the Amiga
// archives without an origins entry (key game-amiga).
func checkAmigaOrigins(t *testing.T, g, dir string, o map[string]map[string]string) {
	l, err := NewLoader(dir)
	if err != nil {
		return // no Amiga version
	}
	names, err := daxNames(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if !slices.Contains(pooledFiles, daxBase(name)) {
			continue
		}
		for id := range maxBlockID {
			if b := l.File(name).BlockShared(id); b != nil && !hasOrigin(o, name+".DAX", id) {
				t.Errorf("%s-amiga %s %d: no origins entry", g, name, id)
			}
		}
	}
}

// checkPool fails pool files no origins entry names and files whose pixels
// duplicate another's.
func checkPool(t *testing.T, named map[string]bool) {
	seen := map[[sha256.Size]byte]string{}
	for _, kind := range []string{"portraits", "combat"} {
		ents, err := os.ReadDir(filepath.Join(poolDir, kind))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range ents {
			n, ok := strings.CutSuffix(e.Name(), ".png")
			if !ok {
				continue
			}
			if !named[n] {
				t.Errorf("pool %s/%s: no origins entry names it", kind, n)
			}
			b, err := os.ReadFile(filepath.Join(poolDir, kind, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			im, err := asset.DecodePNG(b)
			if err != nil {
				t.Fatalf("pool %s/%s: %v", kind, n, err)
			}
			h := pixelHash(im)
			if o, dup := seen[h]; dup {
				t.Errorf("pool %s/%s: same pixels as %s", kind, n, o)
			}
			seen[h] = kind + "/" + n
		}
	}
}

func pixelHash(im image.Image) [sha256.Size]byte {
	var buf bytes.Buffer
	r := im.Bounds()
	fmt.Fprint(&buf, r.Dx(), r.Dy())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c0, c1, c2, c3 := im.At(x, y).RGBA()
			fmt.Fprint(&buf, c0, c1, c2, c3)
		}
	}
	return sha256.Sum256(buf.Bytes())
}
