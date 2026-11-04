package app

import (
	"reflect"

	"gobox/blobber"
	"gobox/gpu"
	"gobox/input"
)

// Menu is the Esc menu (blobber's settings) as the game fills it: its own
// options kept beside blobber.Settings in the settings file, its rows, and
// what the menu's lines do.
type Menu struct {
	Prefs Prefs
	Extra any                      // a pointer to the game's own options struct: loaded and saved beside blobber.Settings (its fields at the file's top level)
	Rows  map[string][]blobber.Row // blobber.ExtraRows
	Apply func()                   // the options changed (once after loading too): the game takes them up

	SaveLoad, MainMenu func() // blobber.OnSaveLoad, OnMainMenu

	// blobber's Font list (FontNames, FontAt, SetFont, FontGlyph); nil keeps blobber's
	FontNames func() []string
	FontAt    func() int
	SetFont   func(int)
	FontGlyph func(font int, ch byte) [8]byte
}

// setup loads the settings file and hands blobber the menu; held is the
// latch (no key held through a screen change repeats), tps the tick rate.
func (m *Menu) setup(held func(any) bool, tps int) {
	if m.FontNames != nil {
		blobber.FontNames, blobber.FontAt, blobber.SetFont, blobber.FontGlyph = m.FontNames, m.FontAt, m.SetFont, m.FontGlyph
	}
	dst := []any{&blobber.Settings}
	if m.Extra != nil {
		dst = append(dst, m.Extra)
	}
	m.Prefs.Load(dst...)
	if m.Apply != nil {
		m.Apply()
	}
	blobber.OnSettings = func(o blobber.Options) {
		if m.Apply != nil {
			m.Apply()
		}
		m.Prefs.Save(m.both(o))
	}
	blobber.ExtraRows = m.Rows
	blobber.Repeats = func(d int) bool { return input.Repeat(d, tps) } // held keys and buttons repeat at the game's one rate
	blobber.OnMainMenu, blobber.OnSaveLoad = m.MainMenu, m.SaveLoad
	blobber.Held = func(k gpu.Key) bool { return held(k) }
}

// both is o and *Extra as one struct embedding both, so the file holds
// their fields at its top level in their order.
func (m *Menu) both(o blobber.Options) any {
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
