// Command ecl lists one ECL script block: `ecl DISK BLOCK` reads block
// BLOCK of ECL<DISK>.DAX from data/dos; run it from the repo root.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gobox/dax"
	"gobox/ecl/script"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: ecl DISK BLOCK")
	}
	disk, _ := strconv.Atoi(os.Args[1])
	block, _ := strconv.Atoi(os.Args[2])
	l, err := dax.NewLoader("../../data/por/sources/dos")
	if err != nil {
		log.Fatal(err)
	}
	b := l.BlockShared(fmt.Sprintf("ECL%d", disk), block)
	if len(b) < 2 {
		log.Fatalf("ECL%d has no block %d", disk, block)
	}
	script.V1.List(b[2:], os.Stdout)
}
