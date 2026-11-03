// Command sounddata writes a game's sound tables (the PC speaker's and the Tandy's
// songs, their scripts, the vibrato waveforms and, with -notes, the envelopes and note
// periods) as JSON for audio.NewPlayer, read from START.EXE's sound driver:
//
//	go run gobox/audio/cmd/sounddata -segment 0x7E4 -jumptable 354 ... START.EXE > sounddata.json
//
// Every address is an offset in the driver's segment. The per-game values are in
// audio/API.md. It checks that the scripts use only what the player implements; without
// -notes (Pool of Radiance, Curse) no notes, and the fields that only notes read are dropped.
package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// The driver's segment (R7E4) in the load image, and its data. The driver is Curse of
// the Azure Bonds' (its R699), its data 2 bytes further up and its code 8591 bytes down.
const (
	channelSize   = 48
	channels      = 4
	envelopeSteps = 8
	firstCommand  = 250
)

var (
	segment, jumpTable, speakerSet, tandySet, speakerSongs, tandySongs int
	songs, enabledFlag, speakerPeriods, tandyPeriods                   int
	notes                                                              bool
)

func init() {
	// 0x7E4*16 is given as a paragraph: -segment 0x7E4; the rest are byte offsets
	flag.IntVar(&segment, "segment", 0, "the driver's segment, in paragraphs of the load image")
	flag.IntVar(&jumpTable, "jumptable", 0, "the script commands' handler table, from command 250 on")
	flag.IntVar(&speakerSet, "speakerset", 0, "the PC speaker's four channels")
	flag.IntVar(&tandySet, "tandyset", 0, "the Tandy's four channels")
	flag.IntVar(&speakerSongs, "speakersongs", 0, "the PC speaker's song table")
	flag.IntVar(&tandySongs, "tandysongs", 0, "the Tandy's song table")
	flag.IntVar(&songs, "songs", 0, "number of songs")
	flag.IntVar(&enabledFlag, "enabled", 0, "the driver's enabled flag (starts 1)")
	flag.IntVar(&speakerPeriods, "speakerperiods", 0, "with -notes: the speaker's 12 note periods")
	flag.IntVar(&tandyPeriods, "tandyperiods", 0, "with -notes: the Tandy's 12 note periods")
	flag.BoolVar(&notes, "notes", false, "the songs play notes with envelopes (Silver Blades)")
	flag.Func("commands", "handler addresses: reset=N,return=N,call=N,switch=N,loop=N,set=N", func(v string) error {
		for _, kv := range strings.Split(v, ",") {
			k, a, ok := strings.Cut(kv, "=")
			n, err := strconv.Atoi(a)
			if !ok || err != nil {
				return errors.New("bad -commands " + kv)
			}
			commands[n] = k
		}
		return nil
	})
}

// The handlers of the commands, by their address in the segment.
var commands = map[int]string{}

// The channel fields a script may set, by their offset in a channel, and what the
// engine calls them; the fields only notes and their envelope read are dropped.
var fields = map[int]string{
	0: "sndTimer", 4: "sndPeriod", 6: "sndSlide", 10: "sndLevel", 12: "sndLevelSlide",
	28: "sndVibrato", 32: "sndVibratoSpeed", 34: "sndVibratoDepth", 36: "sndVibratoLength",
	38: "sndCounter", 40: "sndCounter2", 14: "sndTempo", 18: "sndTranspose", 22: "sndEnvelope",
}
var unread = map[int]bool{16: true}                       // nothing reads it (without -notes: tempo, transpose and envelope too)
var noteOnly = map[int]bool{20: true, 24: true, 26: true} // the release timer and the envelope position

type op struct {
	addr, next int
	kind       string
	field, arg int
	length     int // a note's
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 || segment == 0 || len(commands) == 0 {
		log.Fatal("usage: sounddata [flags] START.EXE (see audio/API.md)")
	}
	if !notes {
		for _, f := range []int{14, 18, 22} {
			unread[f] = true
		}
	}
	mem, err := load(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	seg := mem[segment*16:]
	word := func(o int) int { return int(seg[o]) | int(seg[o+1])<<8 }

	if seg[enabledFlag] != 1 {
		log.Fatalf("the driver starts with the sound off (%d)", seg[enabledFlag])
	}
	if !notes {
		for o := speakerSet; o < speakerSet+channels*channelSize; o++ {
			if seg[o] != 0 || seg[tandySet+o-speakerSet] != 0 {
				log.Fatal("the channels do not start cleared")
			}
		}
	}

	// every script the songs start, and what they call
	var todo []int
	for _, table := range []int{speakerSongs, tandySongs} {
		for s := range songs {
			for c := range channels {
				if p := word(table + s*8 + c*2); p != 0 {
					todo = append(todo, p)
				}
			}
		}
	}
	ops := map[int]op{}
	vibratos := map[int]int{} // table address: its length
	for len(todo) > 0 {
		pc := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		for {
			if _, ok := ops[pc]; ok {
				break
			}
			b := int(seg[pc])
			if b < firstCommand {
				if !notes {
					log.Fatalf("a note at %d: this game's songs should play none (-notes?)", pc)
				}
				// a note: its channel (top three bits), its pitch (bit 7: the last
				// of the event's notes; 127: a rest) and its length
				o := op{addr: pc, kind: "note", field: b >> 5, arg: int(seg[pc+1]), next: pc + 3}
				o.length = int(seg[pc+2])
				ops[pc] = o
				pc = o.next
				continue
			}
			o := op{addr: pc, kind: commands[word(jumpTable+2*(b-firstCommand))]}
			switch o.kind {
			case "return":
				o.next = pc + 1
			case "loop":
				o.field, o.arg, o.next = int(seg[pc+1]), word(pc+2), pc+4
				if fields[o.field] == "" || o.field < 38 {
					log.Fatalf("a loop at %d counts in field %d", pc, o.field)
				}
				todo = append(todo, o.arg)
			case "call", "switch":
				o.arg, o.next = word(pc+1), pc+3
				if o.kind == "call" {
					todo = append(todo, o.arg)
				} else if o.arg%channelSize != 0 || o.arg/channelSize >= channels {
					log.Fatalf("a switch at %d to channel offset %d", pc, o.arg)
				}
			case "set":
				o.field, o.arg, o.next = int(seg[pc+1]), word(pc+2), pc+4
				if noteOnly[o.field] {
					log.Fatalf("a script at %d sets field %d, which only notes use", pc, o.field)
				}
				if fields[o.field] == "" && !unread[o.field] {
					log.Fatalf("a script at %d sets the unknown field %d", pc, o.field)
				}
			default:
				log.Fatalf("command %d (%s) at %d is not implemented", b, o.kind, pc)
			}
			ops[pc] = o
			if o.kind == "return" || o.kind == "set" && o.field == 0 && o.arg == 0 {
				break // a timer of 0 frees the channel: the script ends
			}
			pc = o.next
		}
	}

	var addrs []int
	for a := range ops {
		addrs = append(addrs, a)
	}
	sort.Ints(addrs)
	// the vibrato tables lie one after the other below the scripts: each ends where the
	// next begins; no script lets its position run past the longest table it may use
	longest := 0
	for _, a := range addrs {
		o := ops[a]
		switch {
		case o.kind == "set" && o.field == 28:
			vibratos[o.arg] = 0
		case o.kind == "set" && o.field == 36:
			longest = max(longest, o.arg/16)
		}
	}
	var vibAddrs []int
	for a := range vibratos {
		vibAddrs = append(vibAddrs, a)
	}
	sort.Ints(vibAddrs)
	for i, a := range vibAddrs {
		end := addrs[0]
		if i+1 < len(vibAddrs) {
			end = vibAddrs[i+1]
		}
		vibratos[a] = end - a
	}
	if longest > vibratos[vibAddrs[len(vibAddrs)-1]] {
		log.Fatalf("a vibrato of %d entries runs past the tables", longest)
	}
	vibIndex := map[int]int{}
	for i, a := range vibAddrs {
		vibIndex[a] = i + 1
	}
	// the envelopes overlap: each is taken as its eight steps
	envIndex := map[int]int{}
	var envAddrs []int
	for _, a := range addrs {
		if o := ops[a]; o.kind == "set" && o.field == 22 && envIndex[o.arg] == 0 {
			envAddrs = append(envAddrs, o.arg)
			envIndex[o.arg] = len(envAddrs)
		}
	}

	// the program: the kept ops in address order after the empty op 0; a dropped set
	// stands for the op after it
	index := map[int]int{}
	var kept []op
	for _, a := range addrs {
		o := ops[a]
		if o.kind == "set" && unread[o.field] {
			continue
		}
		index[a] = len(kept) + 1
		kept = append(kept, o)
	}
	resolve := func(a int) int {
		for {
			if i, ok := index[a]; ok {
				return i
			}
			o, ok := ops[a]
			if !ok {
				log.Fatalf("a jump to %d, which is no op", a)
			}
			a = o.next
		}
	}
	for i := 1; i < len(kept); i++ {
		if p := kept[i-1]; p.kind != "return" && p.next != kept[i].addr && resolve(p.next) != i+1 {
			log.Fatalf("the op after %d is not the next one kept", p.addr)
		}
	}

	type jop struct {
		K string `json:"k,omitempty"`
		F string `json:"f,omitempty"`
		C int    `json:"c,omitempty"`
		A int    `json:"a,omitempty"`
		L int    `json:"l,omitempty"`
	}
	out := struct {
		Program        []jop          `json:"program"`
		Speaker        [][4]int       `json:"speaker"`
		Tandy          [][4]int       `json:"tandy"`
		Vibratos       [][]int8       `json:"vibratos"`
		Envelopes      [][8][2]uint16 `json:"envelopes,omitempty"`
		SpeakerPeriods []uint16       `json:"speakerPeriods,omitempty"`
		TandyPeriods   []uint16       `json:"tandyPeriods,omitempty"`
	}{Program: []jop{{}}, Vibratos: [][]int8{nil}, Envelopes: [][8][2]uint16{{}}}
	for _, o := range kept {
		j := jop{K: o.kind}
		switch o.kind {
		case "note":
			j.C, j.A, j.L = o.field, o.arg, o.length
		case "call":
			j.A = resolve(o.arg)
		case "switch":
			j.A = o.arg / channelSize
		case "loop":
			j.F, j.A = fields[o.field], resolve(o.arg)
		case "set":
			j.F, j.A = fields[o.field], o.arg
			switch o.field {
			case 28:
				j.A = vibIndex[o.arg]
			case 22:
				j.A = envIndex[o.arg]
			}
		}
		out.Program = append(out.Program, j)
	}
	for _, t := range []struct {
		dst  *[][4]int
		addr int
	}{{&out.Speaker, speakerSongs}, {&out.Tandy, tandySongs}} {
		for s := range songs {
			var row [4]int
			for c := range channels {
				if p := word(t.addr + s*8 + c*2); p != 0 {
					row[c] = resolve(p)
				}
			}
			*t.dst = append(*t.dst, row)
		}
	}
	for _, a := range vibAddrs {
		v := make([]int8, vibratos[a])
		for i := range v {
			v[i] = int8(seg[a+i])
		}
		out.Vibratos = append(out.Vibratos, v)
	}
	if notes {
		for _, a := range envAddrs {
			var e [8][2]uint16
			for i := range envelopeSteps {
				e[i] = [2]uint16{uint16(word(a + 4*i)), uint16(word(a + 4*i + 2))}
			}
			out.Envelopes = append(out.Envelopes, e)
		}
		for _, t := range []struct {
			dst  *[]uint16
			addr int
		}{{&out.SpeakerPeriods, speakerPeriods}, {&out.TandyPeriods, tandyPeriods}} {
			for i := range 12 {
				*t.dst = append(*t.dst, uint16(word(t.addr+2*i)))
			}
		}
	} else {
		out.Envelopes = nil
	}
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(out); err != nil {
		log.Fatal(err)
	}
}

func load(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) < 0x20 || string(b[:2]) != "MZ" {
		return nil, errors.New("not an MZ executable")
	}
	hdr := int(binary.LittleEndian.Uint16(b[8:])) * 16
	cs := int(int16(binary.LittleEndian.Uint16(b[0x16:])))
	exepackOfs := hdr + cs*16 // start of the EXEPACK variables block
	if exepackOfs+18 > len(b) || string(b[exepackOfs+16:exepackOfs+18]) != "RB" {
		return nil, errors.New("not EXEPACK packed")
	}
	v := b[exepackOfs:]
	destLen := int(binary.LittleEndian.Uint16(v[12:])) * 16 // dest_len paragraphs
	// packed data lies between the load module start and the EXEPACK block,
	// and is processed backwards (commands follow their data).
	packed := b[hdr:exepackOfs]
	// skip trailing 0xFF padding
	end := len(packed)
	for end > 0 && packed[end-1] == 0xFF {
		end--
	}
	out := make([]byte, destLen)
	dst := destLen
	i := end
	for i > 0 {
		cmd := packed[i-1]
		n := int(binary.LittleEndian.Uint16(packed[i-3:]))
		i -= 3
		switch cmd & 0xFE {
		case 0xB0: // fill
			fill := packed[i-1]
			i--
			for k := 0; k < n; k++ {
				dst--
				out[dst] = fill
			}
		case 0xB2: // copy
			for k := 0; k < n; k++ {
				dst--
				i--
				out[dst] = packed[i]
			}
		default:
			return nil, errors.New("bad EXEPACK command")
		}
		if cmd&1 == 1 {
			break
		}
	}
	// unconsumed leading bytes are stored as-is
	copy(out[dst-i:dst], packed[:i])
	return out, nil
}
