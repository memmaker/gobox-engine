// Package rules1e is a shared view of a game's byte-record character: the game
// describes the record as a Layout (data, JSON-loadable) and Rec reads and
// writes it. Stats is the small interface the rules package needs, so a
// game with a Go-struct character (por) implements it directly.
package rules1e

import (
	"encoding/binary"
	"encoding/json"
	"os"
)

// Field is one named field: N elements of Size bytes (1, 2 or 4, little
// endian) starting at Off, Stride bytes apart (default Size).
type Field struct {
	Off    int  `json:"off"`
	Size   int  `json:"size"`
	N      int  `json:"n,omitempty"`
	Stride int  `json:"stride,omitempty"`
	Signed bool `json:"signed,omitempty"`
}

// CharLayout maps field names to Fields. The names rules uses: saves, saveBonus,
// thac0, ac, hp, maxHP, hpRoll, multiclassLevel, levels, oldLevels, status,
// okay. Others (score, scoreUse, str00, race, class, alignment, coins, exp,
// slots ...) are for the game; see testdata/*.json.
type CharLayout struct {
	Size   int              `json:"size"`
	Fields map[string]Field `json:"fields"`
}

// LoadLayout reads a JSON layout.
func LoadLayout(path string) (*CharLayout, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseLayout(b)
}

// ParseLayout parses JSON layout data.
func ParseLayout(b []byte) (*CharLayout, error) {
	l := &CharLayout{}
	return l, json.Unmarshal(b, l)
}

// CharStats is what rules needs of a character. Values are the raw stored
// bytes (0..255) so the rules apply the original's signed readings.
type CharStats interface {
	Classes() int // number of class levels
	Level(k int) int
	OldLevel(k int) int
	MulticlassLevel() int
	HPRoll() int
	MaxHP() int
	HP() int
	THAC0() int
	AC(i int) int
	Save(typ int) int
	SaveBonus() int
	Status() int
	SetHP(v int)
	SetMaxHP(v int)
	SetStatus(v int)
	SetOkay(v int)
}

// Rec is a record and the layout that describes it. It implements Stats.
type Rec struct {
	B []byte
	L *CharLayout
}

// New returns a zero record of the layout's size.
func New(l *CharLayout) *Rec { return &Rec{B: make([]byte, l.Size), L: l} }

// Has reports whether the layout has the field.
func (r *Rec) Has(name string) bool { _, ok := r.L.Fields[name]; return ok }

// Len is the field's element count (0 if absent).
func (r *Rec) Len(name string) int {
	f, ok := r.L.Fields[name]
	if !ok {
		return 0
	}
	return max(f.N, 1)
}

func (f Field) at(i int) int {
	s := f.Stride
	if s == 0 {
		s = f.Size
	}
	return f.Off + i*s
}

// Get reads element i (sign-extended if the field is signed). An absent
// field reads 0.
func (r *Rec) Get(name string, i int) int {
	f, ok := r.L.Fields[name]
	if !ok {
		return 0
	}
	p := f.at(i)
	switch f.Size {
	case 1:
		if f.Signed {
			return int(int8(r.B[p]))
		}
		return int(r.B[p])
	case 2:
		v := binary.LittleEndian.Uint16(r.B[p:])
		if f.Signed {
			return int(int16(v))
		}
		return int(v)
	default:
		v := binary.LittleEndian.Uint32(r.B[p:])
		if f.Signed {
			return int(int32(v))
		}
		return int(v)
	}
}

// Set writes element i, truncated to the field size. An absent field is ignored.
func (r *Rec) Set(name string, i, v int) {
	f, ok := r.L.Fields[name]
	if !ok {
		return
	}
	p := f.at(i)
	switch f.Size {
	case 1:
		r.B[p] = byte(v)
	case 2:
		binary.LittleEndian.PutUint16(r.B[p:], uint16(v))
	default:
		binary.LittleEndian.PutUint32(r.B[p:], uint32(v))
	}
}

// Bytes is the field's raw area (name, spell lists, item slot area).
func (r *Rec) Bytes(name string) []byte {
	f, ok := r.L.Fields[name]
	if !ok {
		return nil
	}
	return r.B[f.Off : f.Off+f.Size*max(f.N, 1)]
}

func (r *Rec) Classes() int         { return r.Len("levels") }
func (r *Rec) Level(k int) int      { return r.Get("levels", k) }
func (r *Rec) OldLevel(k int) int   { return r.Get("oldLevels", k) }
func (r *Rec) MulticlassLevel() int { return r.Get("multiclassLevel", 0) }
func (r *Rec) HPRoll() int          { return r.Get("hpRoll", 0) }
func (r *Rec) MaxHP() int           { return r.Get("maxHP", 0) }
func (r *Rec) HP() int              { return r.Get("hp", 0) }
func (r *Rec) THAC0() int           { return r.Get("thac0", 0) }
func (r *Rec) AC(i int) int         { return r.Get("ac", i) }
func (r *Rec) Save(t int) int       { return r.Get("saves", t) }
func (r *Rec) SaveBonus() int       { return r.Get("saveBonus", 0) }
func (r *Rec) Status() int          { return r.Get("status", 0) }
func (r *Rec) SetHP(v int)          { r.Set("hp", 0, v) }
func (r *Rec) SetMaxHP(v int)       { r.Set("maxHP", 0, v) }
func (r *Rec) SetStatus(v int)      { r.Set("status", 0, v) }
func (r *Rec) SetOkay(v int)        { r.Set("okay", 0, v) }
