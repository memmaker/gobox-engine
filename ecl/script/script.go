package script

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"

	"gobox/game/play"
)

type Script struct {
	Area   int
	Events [5]int         // the entry points' instruction numbers
	Code   []Instr        // the commands in file order
	Labels map[string]int // a code label's instruction number
	At     map[int]string // an instruction's first label
	Data   []byte         // the tables, from Version.Base
	Tables map[string]int // a table's address in Data (Version.Base + its offset)
	Layout *Layout        // the fields its operands' Refs name (ResolveRefs)
}

// Instr is one command.
type Instr struct {
	Op      byte
	Args    []Arg
	Line    int           // in the script file
	Comment string        // a note from the decoder, written after the command
	Exit    play.EscReply // a Menu or VMenu: the row Esc picks (exit= in the text; none: Esc does nothing)
}

// Arg is one operand in its byte-code form.
type Arg struct {
	Type  byte   // 0 byte, 1 and 3 variable, 2 word, 0x80 string, 0x81 string variable
	N     int    // the value or address (a label's instruction number)
	S     string // a string's text
	Label bool   // N is a label's instruction number (a jump's operand)
	Ref   Ref    // a variable operand's typed reference (ResolveRefs); the zero Ref before or for no variable
}

// ReadNames reads area's names from a game's script data (its names/ and
// layout/ folders) over the version's: the variables (with labels, the
// labels and tables too), the areas and the values.
func ReadNames(data fs.FS, area int, labels bool, v *Version) (*Names, error) {
	n := &Names{Version: v, Vars: map[int]string{}, Calls: map[int]string{}}
	files := []string{"global.txt", fmt.Sprintf("area%d.txt", area)}
	if labels {
		files = append(files, fmt.Sprintf("labels/area%d.txt", area))
	}
	for _, f := range files {
		if err := ReadVarNames(data, path.Join("names", f), n.Vars); err != nil {
			return nil, err
		}
	}
	l, err := ReadLayout(data, area)
	if err != nil {
		return nil, err
	}
	for _, f := range layoutFields(l) { // the layout names the words: a field its name, element i of an array Name_i
		if f.routine() {
			n.Calls[f.From] = f.Name
			continue
		}
		n.Vars[f.From] = f.Name
		for i := 1; i < f.Len && f.Width != WidthText; i++ {
			n.Vars[f.From+i] = fmt.Sprintf("%s_%d", f.Name, i)
		}
	}
	if n.Areas, err = ReadAreas(data, "names/areas.txt"); err != nil {
		return nil, err
	}
	for _, a := range n.Areas {
		if a.ID == area {
			n.Set = a.Set
		}
	}
	for _, f := range []string{"values.txt", "pictures.txt", "variables.txt"} {
		if n.Values, err = ReadValues(data, path.Join("names", f), n.Values); err != nil {
			return nil, err
		}
	}
	n.Values = slices.DeleteFunc(n.Values, func(x Value) bool { return x.Area != 0 && x.Area != area })
	return n, nil
}

// ReadVarNames adds the "XXXX Name" lines of a names file, if there is
// one (# starts a comment).
func ReadVarNames(data fs.FS, path string, names map[int]string) error {
	f, err := data.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line, _, _ := strings.Cut(sc.Text(), "#")
		fs := strings.Fields(line)
		if len(fs) == 0 {
			continue
		}
		a, err := strconv.ParseUint(fs[0], 16, 16)
		if len(fs) != 2 || err != nil {
			return fmt.Errorf("%s:%d: want \"XXXX Name\"", path, n)
		}
		names[int(a)] = fs[1]
	}
	return sc.Err()
}

// Names are the names a script's words stand for.
type Names struct {
	Version  *Version
	Vars     map[int]string // variables (and, for cmd/eclscript, labels and tables) by address
	Calls    map[int]string // engine routines by CALL's operand (the layout's call fields)
	Areas    []Area
	Values   []Value
	Set      int                    // the area's data set
	Treasure func(block int) string // an ITEM block's items, for cmd/eclscript's comments
}

// Area is one area of names/areas.txt.
type Area struct {
	ID, Set int // NewECL's number, the data set holding it
	Name    string
}

// ReadAreas reads the "ID SET Name" lines of an areas file, if
// there is one (# starts a comment).
func ReadAreas(data fs.FS, path string) ([]Area, error) {
	b, err := fs.ReadFile(data, path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var areas []Area
	for n, line := range strings.Split(string(b), "\n") {
		line, _, _ = strings.Cut(line, "#")
		fs := strings.Fields(line)
		if len(fs) == 0 {
			continue
		}
		id, err1 := strconv.Atoi(fs[0])
		set, err2 := strconv.Atoi(fs[1%len(fs)])
		if len(fs) != 3 || err1 != nil || err2 != nil {
			return nil, fmt.Errorf("%s:%d: want \"ID SET Name\"", path, n+1)
		}
		areas = append(areas, Area{id, set, fs[2]})
	}
	return areas, nil
}

// "Kind Set Number Name"; set 0 is every set.

// Value is one named number.
type Value struct {
	Kind   string
	Set, N int
	Name   string
	Area   int // a value name of one area only (Set written aN: the area's own use of a shared word); 0 every area
}

// operandKinds are the kinds of the commands' operands, "" a plain number.
var operandKinds = map[string][]string{
	"LoadMonster":   {"Monster", "", "Icon"},
	"AddNPC":        {"Monster"},
	"Picture":       {"Picture|Body"},
	"SetupMonster":  {"Sprite", "", "Picture|Body"},
	"EncounterMenu": {"Sprite", "", "Picture|Body", "", "Reaction", "Reaction", "Reaction", "Reaction", "Reaction"},
	"FindItem":      {"Item"},
	"ShowPortrait":  {"Head", "Body"},
	"LoadMap":       {"Map", "", "MapIcons"},
	"LoadWalls":     {"Walls", "Walls", "Walls"},
	"NewECL":        {"Area"},
}

// operandLabels name the commands' operands in the script, as label=value
// (LoadMonster Orc count=8 icon=Orc); "" is an operand without one, a|b
// two labels (a SetupMonster's close-up is a body picture while a
// portrait is up).
var operandLabels = map[string][]string{
	"LoadMonster":  {"", "count", "icon"},
	"SetupMonster": {"", "distance", "closeup|body"},
	"LoadMap":      {"", "", "icons"},
	"EncounterMenu": {"", "distance", "closeup|body", "result", "combat", "wait", "flee", "advance", "parlay",
		"near", "mid", "far", "escape", "strength"},
}

// operandDefaults are the values of labelled operands a command may leave
// out (EncounterMenu's empty texts).
var operandDefaults = map[string]map[int]string{
	"EncounterMenu": {9: `""`, 10: `""`, 11: `""`},
}

// ReadValues adds the values of a values file, if there is one.
func ReadValues(data fs.FS, path string, vals []Value) ([]Value, error) {
	b, err := fs.ReadFile(data, path)
	if errors.Is(err, fs.ErrNotExist) {
		return vals, nil
	} else if err != nil {
		return nil, err
	}
	for n, line := range strings.Split(string(b), "\n") {
		line, _, _ = strings.Cut(line, "#")
		fs := strings.Fields(line)
		if len(fs) == 0 {
			continue
		}
		if len(fs) != 4 {
			return nil, fmt.Errorf("%s:%d: want \"Kind Set Number Name\"", path, n+1)
		}
		var set, area int
		var err1 error
		if a, ok := strings.CutPrefix(fs[1], "a"); ok {
			area, err1 = strconv.Atoi(a)
		} else {
			set, err1 = strconv.Atoi(fs[1])
		}
		v, err2 := strconv.Atoi(fs[2])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("%s:%d: want \"Kind Set Number Name\" (Set aN: area N only)", path, n+1)
		}
		vals = append(vals, Value{Kind: fs[0], Set: set, N: v, Name: fs[3], Area: area})
	}
	return vals, nil
}

// ValueName is the name of kind's value v in set, or "".
func ValueName(vals []Value, kind string, set, v int) string {
	for _, x := range vals {
		if x.Kind == kind && x.N == v && (x.Set == set || x.Set == 0) {
			return x.Name
		}
	}
	return ""
}

// valueOf is the number kind's name stands for in set; kind may be a|b,
// either kind's.
func valueOf(vals []Value, kind string, set int, name string) (int, bool) {
	if a, b, ok := strings.Cut(kind, "|"); ok {
		if v, ok := valueOf(vals, a, set, name); ok {
			return v, true
		}
		kind = b
	}
	for _, x := range vals {
		if x.Kind == kind && x.Name == name && (x.Set == set || x.Set == 0) {
			return x.N, true
		}
	}
	return 0, false
}

// layoutFields are l's fields; none for no layout (ReadLayout).
func layoutFields(l *Layout) []Field {
	if l == nil {
		return nil
	}
	return l.Fields
}
