package ui

import (
	"log"
	"reflect"

	"gobox/gpu"
)

// shellMenu is the Esc menu (the Esc menu's settings) as the game fills it: its own
// options kept beside Settings in the settings file, its rows, and
// what the menu's lines do.
type shellMenu struct {
	Prefs prefs
	Extra any              // a pointer to the game's own options struct: loaded and saved beside Settings (its fields at the file's top level)
	Rows  map[string][]Row // ExtraRows
	Apply func()           // the options changed (once after loading too): the game takes them up

	SaveLoad, MainMenu func() // OnSaveLoad, OnMainMenu

	// blobber's Font list (FontNames, FontAt, SetFont, FontGlyph); nil keeps blobber's
	FontNames func() []string
	FontAt    func() int
	SetFont   func(int)
	FontGlyph func(font int, ch byte) [8]byte
}

// fontHost is a Front with text fonts of its own (present.Front): the
// Font list comes from it unless the Menu fills it.
type fontHost interface {
	KeepFonts(dir, title string)
	FontMenu() (func() []string, func() int, func(int), func(int, byte) [8]byte)
}

// setup loads the settings file and hands blobber the menu; held is the
// latch (no key held through a screen change repeats), tps the tick rate.
func (m *shellMenu) setup(held func(any) bool, tps int) {
	if m.FontNames != nil {
		FontNames, FontAt, SetFont, FontGlyph = m.FontNames, m.FontAt, m.SetFont, m.FontGlyph
	}
	dst := []any{&Settings}
	if m.Extra != nil {
		dst = append(dst, m.Extra)
	}
	m.Prefs.Load(dst...)
	if m.Apply != nil {
		m.Apply()
	}
	OnSettings = func(o MenuOptions) {
		if m.Apply != nil {
			m.Apply()
		}
		if err := m.Prefs.Save(m.both(o)); err != nil { // the menu has no place for an error: the settings stand, the failed keeping is logged
			log.Printf("settings not kept: %v", err)
		}
	}
	ExtraRows = m.Rows
	Repeats = func(d int) bool { return Repeat(d, tps) } // held keys and buttons repeat at the game's one rate
	OnMainMenu, OnSaveLoad = m.MainMenu, m.SaveLoad
	Held = func(k gpu.Key) bool { return held(k) }
}

// both is o and *Extra as one struct embedding both, so the file holds
// their fields at its top level in their order.
func (m *shellMenu) both(o MenuOptions) any {
	if m.Extra == nil {
		return o
	}
	e := reflect.ValueOf(m.Extra).Elem()
	t := reflect.StructOf([]reflect.StructField{
		{Name: "Options", Type: reflect.TypeOf(o), Anonymous: true},
		{Name: e.Type().Name(), Type: e.Type(), Anonymous: true},
	})
	v := reflect.New(t).Elem()
	v.Field(0).Set(reflect.ValueOf(o))
	v.Field(1).Set(e)
	return v.Interface()
}
