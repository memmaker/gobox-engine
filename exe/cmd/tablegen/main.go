// Command tablegen writes a game's data tables from its START.EXE as named
// Go values. It runs only at `go generate` time, so the game itself never
// reads the executable. The game's table list is a JSON file (see
// tablespec.Table; kind is bytes, signed, words, swords, longs or string:N).
//
//	tablegen -spec tables.json -pkg engine START.EXE:DS [OTHER.EXE:DS ...] > tables.go
//
// DS is the data-segment paragraph; Table.img indexes the executables.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"gobox/exe"
	"gobox/exe/tablespec"
)

func main() {
	spec := flag.String("spec", "", "JSON file with the table list")
	pkg := flag.String("pkg", "engine", "package name of the output")
	flag.Parse()
	if *spec == "" || flag.NArg() < 1 {
		log.Fatal("usage: tablegen -spec tables.json [-pkg engine] START.EXE:DS [OTHER.EXE:DS ...]")
	}
	var imgs []*exe.Image
	for _, a := range flag.Args() {
		i := strings.LastIndex(a, ":")
		ds, err := strconv.Atoi(a[i+1:])
		if i < 0 || err != nil {
			log.Fatalf("%q: want PATH:DS", a)
		}
		im, err := exe.Load(a[:i], ds)
		if err != nil {
			log.Fatal(err)
		}
		imgs = append(imgs, im)
	}
	raw, err := os.ReadFile(*spec)
	if err != nil {
		log.Fatal(err)
	}
	var tables []tablespec.Table
	if err := json.Unmarshal(raw, &tables); err != nil {
		log.Fatal(err)
	}
	src, err := tablespec.Generate(*pkg, tables, imgs...)
	if err != nil {
		os.Stdout.Write(src)
		log.Fatal(err)
	}
	os.Stdout.Write(src)
}
