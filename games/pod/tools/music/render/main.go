// Command render plays an OPL register log from podmusic.py ("tick reg val" lines,
// tick = 1193182/16384 Hz) through github.com/memmaker/opl2 into a mono 16-bit WAV.
//
//	go run . LOG OUT.wav
package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/memmaker/opl2"
)

func main() {
	in, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	chip := opl2.New()
	var pcm []int16
	tickHz := 1193182.0 / 16384
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		var t, r, v int
		fmt.Sscan(sc.Text(), &t, &r, &v)
		for float64(len(pcm)) < float64(t)*opl2.Rate/tickHz {
			pcm = append(pcm, int16(chip.Sample()))
		}
		chip.Write(r, byte(v))
	}
	for range opl2.Rate { // let the release ring out
		pcm = append(pcm, int16(chip.Sample()))
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	n := uint32(len(pcm) * 2)
	hdr := []any{[4]byte{'R', 'I', 'F', 'F'}, 36 + n, [4]byte{'W', 'A', 'V', 'E'}, [4]byte{'f', 'm', 't', ' '},
		uint32(16), uint16(1), uint16(1), uint32(opl2.Rate), uint32(opl2.Rate * 2), uint16(2), uint16(16),
		[4]byte{'d', 'a', 't', 'a'}, n}
	for _, h := range hdr {
		if err := binary.Write(out, binary.LittleEndian, h); err != nil {
			panic(err)
		}
	}
	if err := binary.Write(out, binary.LittleEndian, pcm); err != nil {
		panic(err)
	}
	if err := out.Close(); err != nil {
		panic(err)
	}
}
