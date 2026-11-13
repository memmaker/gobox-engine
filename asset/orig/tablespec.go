// tablespec.go describes data tables to copy out of a game's
// unpacked START.EXE (name, place in the data segment, length, kind) and
// generates them as named Go values. The game owns its list of Tables and
// passes it to Generate; nothing here knows which game it is.

package orig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Kind is the element type of a table.
type Kind int

const (
	Bytes Kind = iota
	Signed
	Words
	SignedWords
	Longs // signed 32-bit
	// Strings of a fixed slot size: String(n) holds a length byte and n-1 characters.
	stringBase
)

// String is a table of Pascal strings in slots of size bytes.
func String(size int) Kind { return stringBase + Kind(size) }

// Size is the bytes per element.
func (k Kind) Size() int {
	switch k {
	case Bytes, Signed:
		return 1
	case Words, SignedWords:
		return 2
	case Longs:
		return 4
	}
	return int(k - stringBase)
}

// UnmarshalText reads a Kind from JSON/TOML text: bytes, signed, words,
// swords, longs or string:N.
func (k *Kind) UnmarshalText(b []byte) error {
	s := string(b)
	switch s {
	case "bytes":
		*k = Bytes
	case "signed":
		*k = Signed
	case "words":
		*k = Words
	case "swords":
		*k = SignedWords
	case "longs":
		*k = Longs
	default:
		n, err := strconv.Atoi(strings.TrimPrefix(s, "string:"))
		if err != nil || !strings.HasPrefix(s, "string:") || n < 1 {
			return fmt.Errorf("tablespec: unknown kind %q", s)
		}
		*k = String(n)
	}
	return nil
}

// TableSpec is one generated Go value. The JSON field names are the lower-case
// ones, so a game can keep its list in a file.
type TableSpec struct {
	Name string `json:"name"`
	At   int    `json:"at"`  // offset in the data segment
	Len  int    `json:"len"` // elements (rows when Cols is set)
	Kind Kind   `json:"kind"`
	Doc  string `json:"doc"`
	Cols int    `json:"cols"` // set for a two-dimensional table: Len rows of Cols elements
	// Img selects which executable the table is read from (index into the
	// images given to Generate; 0 is the main one). A game that still reads
	// some tables from another game's executable sets it.
	Img int `json:"img"`
	// Stride is the bytes from one row to the next when the rows carry
	// padding (0: Cols elements exactly).
	Stride int `json:"stride"`
}

// ReadTableSpec is t's elements from imgs: []int for a number table, [][]int for a
// two-dimensional one (rows of Cols), []string for a string table ("" for a
// slot that holds no string).
func ReadTableSpec(t TableSpec, imgs ...*ExeImage) (any, error) {
	if t.Img < 0 || t.Img >= len(imgs) {
		return nil, fmt.Errorf("table %s: no image %d", t.Name, t.Img)
	}
	img := imgs[t.Img]
	size := t.Kind.Size()
	rowSize := t.Cols * size
	if t.Stride > 0 {
		rowSize = t.Stride
	}
	var raw []byte
	if t.Cols > 0 {
		raw = img.Bytes(t.At, t.Len*rowSize)
	} else {
		raw = img.Bytes(t.At, t.Len*size)
	}
	if raw == nil && t.Len > 0 {
		return nil, fmt.Errorf("table %s: out of range", t.Name)
	}
	signed := t.Kind == Signed || t.Kind == SignedWords || t.Kind == Longs
	if t.Cols > 0 {
		rows := make([][]int, t.Len)
		for r := range rows {
			rows[r] = ints(raw[r*rowSize:r*rowSize+t.Cols*size], size, signed)
		}
		return rows, nil
	}
	if t.Kind < stringBase {
		return ints(raw, size, signed), nil
	}
	strs := make([]string, t.Len) // Len Pascal strings of Kind.Size() bytes each
	for i := range strs {
		s := raw[i*size:]
		if int(s[0]) < size { // else not a string (an unused slot holding other data)
			strs[i] = string(s[1 : 1+int(s[0])])
		}
	}
	return strs, nil
}

// Extra is a rules.json entry the importer writes from the game's code,
// not read from the executable's tables (rules1e's creationStart).
type Extra struct {
	Name  string
	Value any
}

// jsonWith is JSON with extra's entries after the tables.
func jsonWith(tables []TableSpec, extra []Extra, imgs ...*ExeImage) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("{")
	n := 0
	entry := func(name string, v any) error {
		key, _ := json.Marshal(name)
		val, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if n > 0 {
			b.WriteString(",")
		}
		n++
		fmt.Fprintf(&b, "\n %s: %s", key, val)
		return nil
	}
	for _, t := range tables {
		v, err := ReadTableSpec(t, imgs...)
		if err != nil {
			return nil, err
		}
		if err := entry(t.Name, v); err != nil {
			return nil, err
		}
	}
	for _, x := range extra {
		if err := entry(x.Name, x.Value); err != nil {
			return nil, err
		}
	}
	b.WriteString("\n}\n")
	return b.Bytes(), nil
}

// RulesFile is the ruleset's tables in an imported game's data folder.
const RulesFile = "rules.json"

// WriteRulesWith is WriteRules with extra's entries after the tables.
func WriteRulesWith(spec []byte, dir string, extra []Extra, imgs ...*ExeImage) error {
	var all, tables []TableSpec
	if err := json.Unmarshal(spec, &all); err != nil {
		return err
	}
	for _, t := range all {
		if t.Img < len(imgs) {
			tables = append(tables, t)
		}
	}
	b, err := jsonWith(tables, extra, imgs...)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, RulesFile), b, 0o644)
}

// Generate writes the Go source of package pkg declaring every table, read
// from imgs.
func Generate(pkg string, tables []TableSpec, imgs ...*ExeImage) ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by cmd/tablegen from START.EXE; DO NOT EDIT.\n\npackage %s\n\n", pkg)
	for _, t := range tables {
		v, err := ReadTableSpec(t, imgs...)
		if err != nil {
			return nil, err
		}
		if t.Doc != "" {
			for _, l := range strings.Split(t.Doc, "\n") {
				fmt.Fprintf(&b, "// %s\n", l)
			}
		}
		typ := map[Kind]string{Bytes: "byte", Signed: "int8", Words: "uint16", SignedWords: "int16", Longs: "int32"}[t.Kind]
		switch v := v.(type) {
		case [][]int:
			fmt.Fprintf(&b, "var %s = [%d][%d]%s{\n", t.Name, t.Len, t.Cols, typ)
			for _, row := range v {
				fmt.Fprintf(&b, "{%s},\n", list(row))
			}
			b.WriteString("}\n\n")
		case []int:
			fmt.Fprintf(&b, "var %s = [%d]%s{%s}\n\n", t.Name, t.Len, typ, list(v))
		case []string:
			fmt.Fprintf(&b, "var %s = [%d]string{", t.Name, t.Len)
			for _, s := range v {
				fmt.Fprintf(&b, "%q, ", s)
			}
			b.WriteString("}\n\n")
		}
	}
	src, err := format.Source(b.Bytes())
	if err != nil {
		return b.Bytes(), err
	}
	return src, nil
}

// ints is raw read as little-endian elements of size bytes.
func ints(raw []byte, size int, signed bool) []int {
	var out []int
	for i := 0; i+size <= len(raw); i += size {
		v := 0
		for k := size - 1; k >= 0; k-- {
			v = v<<8 | int(raw[i+k])
		}
		if signed && v >= 1<<(8*size-1) {
			v -= 1 << (8 * size)
		}
		out = append(out, v)
	}
	return out
}

func list(v []int) string {
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = fmt.Sprint(x)
	}
	return strings.Join(parts, ", ")
}
