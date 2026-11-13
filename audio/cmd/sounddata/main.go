// Command sounddata writes a game's sound tables as JSON for audio.NewDriver
// (audio.DecodeSoundData), read from START.EXE's sound driver:
//
//	go run gobox/audio/cmd/sounddata -segment 0x7E4 -jumptable 354 ... START.EXE > sounddata.json
//
// The per-game values are in audio/API.md. A game's import calls audio.DecodeSoundData itself.
package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"gobox/audio"
)

func main() {
	l := audio.SoundDataLayout{Commands: map[int]string{}}
	// 0x7E4*16 is given as a paragraph: -segment 0x7E4; the rest are byte offsets
	flag.IntVar(&l.Segment, "segment", 0, "the driver's segment, in paragraphs of the load image")
	flag.IntVar(&l.JumpTable, "jumptable", 0, "the script commands' handler table, from command 250 on")
	flag.IntVar(&l.SpeakerSet, "speakerset", 0, "the PC speaker's four channels")
	flag.IntVar(&l.TandySet, "tandyset", 0, "the Tandy's four channels")
	flag.IntVar(&l.SpeakerSongs, "speakersongs", 0, "the PC speaker's song table")
	flag.IntVar(&l.TandySongs, "tandysongs", 0, "the Tandy's song table")
	flag.IntVar(&l.Songs, "songs", 0, "number of songs")
	flag.IntVar(&l.Enabled, "enabled", 0, "the driver's enabled flag (starts 1)")
	flag.IntVar(&l.SpeakerPeriods, "speakerperiods", 0, "with -notes: the speaker's 12 note periods")
	flag.IntVar(&l.TandyPeriods, "tandyperiods", 0, "with -notes: the Tandy's 12 note periods")
	flag.BoolVar(&l.Notes, "notes", false, "the songs play notes with envelopes (a game with notes)")
	flag.Func("commands", "handler addresses: reset=N,return=N,call=N,switch=N,loop=N,set=N", func(v string) error {
		for _, kv := range strings.Split(v, ",") {
			k, a, ok := strings.Cut(kv, "=")
			n, err := strconv.Atoi(a)
			if !ok || err != nil {
				return errors.New("bad -commands " + kv)
			}
			l.Commands[n] = k
		}
		return nil
	})
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: sounddata [flags] START.EXE (see audio/API.md)")
	}
	exe, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	b, err := audio.DecodeSoundData(exe, l)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(b)
}
