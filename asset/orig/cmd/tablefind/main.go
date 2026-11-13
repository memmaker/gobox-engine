// Command tablefind looks for each table of a game's tables.json in another
// executable's data segment, by its bytes (a porting aid: where a table an
// earlier game holds sits in a later one). Table.img indexes the sources,
// as in tablegen.
//
//	tablefind -spec tables.json -in TARGET.EXE:DS SOURCE.EXE:DS [OTHER.EXE:DS ...]
//
// sotsb (Curse tables in Blades): -in BLADES-START.EXE:3554 BLADES-START.EXE:3554 CURSE-START.EXE:2750
// pod (Blades and Curse tables in POD): -in POD-GAME.EXE:2535 BLADES-START.EXE:3554 CURSE-START.EXE:2750
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"gobox/asset/orig"
)

func main() {
	spec := flag.String("spec", "", "JSON file with the table list")
	in := flag.String("in", "", "the executable searched, PATH:DS")
	flag.Parse()
	if *spec == "" || *in == "" || flag.NArg() < 1 {
		log.Fatal("usage: tablefind -spec tables.json -in TARGET.EXE:DS SOURCE.EXE:DS [OTHER.EXE:DS ...]")
	}
	target, err := orig.LoadArg(*in)
	if err != nil {
		log.Fatal(err)
	}
	var sources []*orig.ExeImage
	for _, a := range flag.Args() {
		im, err := orig.LoadArg(a)
		if err != nil {
			log.Fatal(err)
		}
		sources = append(sources, im)
	}
	raw, err := os.ReadFile(*spec)
	if err != nil {
		log.Fatal(err)
	}
	var tables []orig.TablespecTable
	if err := json.Unmarshal(raw, &tables); err != nil {
		log.Fatal(err)
	}
	data := target.Mem[target.DS:]
	for _, t := range tables {
		if t.Img >= len(sources) {
			log.Fatalf("%s: img %d, but %d sources given", t.Name, t.Img, len(sources))
		}
		size := t.Len * t.Kind.Size()
		if t.Cols > 0 {
			size *= t.Cols
		}
		if t.Stride > 0 {
			size = t.Len * t.Stride
		}
		want := sources[t.Img].Bytes(t.At, size)
		var found []int
		for at := 0; ; {
			i := bytes.Index(data[at:], want)
			if i < 0 {
				break
			}
			found = append(found, at+i)
			at += i + 1
		}
		fmt.Printf("%-24s img %d at %6d %04X size %5d  found %v\n", t.Name, t.Img, t.At, t.At, size, found)
	}
}
