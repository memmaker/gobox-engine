// hostrows.go: the Esc menu rows (blobber.ExtraRows) of the presenter's
// Options, and the row makers a game builds its own settings' rows with.

package present

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"gobox/blobber"
)

// OnOff is b as the settings show it.
func OnOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

// Choice is no or yes by b.
func Choice(b bool, no, yes string) string {
	if b {
		return yes
	}
	return no
}

// step moves v by s within lo..hi: Left down, Right up, Enter up and round to lo past hi.
func step(v, lo, hi, s float64, d int) float64 {
	switch {
	case d < 0:
		v -= s
	case v+s > hi+s/2:
		if d == 0 {
			return lo
		}
	default:
		v += s
	}
	return math.Round(math.Max(lo, math.Min(hi, v))/s) * s
}

// bar draws v in lo..hi (step s) as a slider.
func bar(v, lo, hi, s float64) string {
	i, n := int(math.Round((v-lo)/s)), int(math.Round((hi-lo)/s))
	return strings.Repeat("#", i+1) + strings.Repeat("-", n-i)
}

// FlipRow flips *p, shown as show says.
func FlipRow(label string, at int, p *bool, show func() string) blobber.Row {
	return blobber.Row{Label: label, At: at, Get: show, Change: func(int) { *p = !*p }}
}

// SwitchRow flips *p, shown ON or OFF.
func SwitchRow(label string, at int, p *bool) blobber.Row {
	return FlipRow(label, at, p, func() string { return OnOff(*p) })
}

// SliderRow steps *v within lo..hi by s; fmtv is the text after the bar
// (pct: the value as a percentage instead).
func SliderRow(label string, at int, v *float64, lo, hi, s float64, fmtv string, pct bool) blobber.Row {
	return blobber.Row{Label: label, At: at, Get: func() string {
		if pct {
			return bar(*v, lo, hi, s) + fmt.Sprintf(" %3.0f%%", *v*100)
		}
		return bar(*v, lo, hi, s) + fmt.Sprintf(fmtv, *v)
	}, Change: func(d int) { *v = step(*v, lo, hi, s, d) }}
}

// SizeRow steps *v through UISteps (Enter: round to the smallest past the largest).
func SizeRow(label string, at int, v *float64) blobber.Row {
	return blobber.Row{Label: label, At: at, Get: func() string { return fmt.Sprintf("%.0fx", *v) }, Change: func(d int) {
		i := max(slices.Index(UISteps, *v), 0)
		switch {
		case d < 0:
			i = max(i-1, 0)
		case d > 0:
			i = min(i+1, len(UISteps)-1)
		default:
			i = (i + 1) % len(UISteps)
		}
		*v = UISteps[i]
	}}
}

// OptionRows are o's Esc menu rows by page: the Graphics rows go before
// its built-in row 3 (Font), Zoom speed before Gameplay's row 4.
func OptionRows(o *Options) map[string][]blobber.Row {
	return map[string][]blobber.Row{
		"Graphics": {
			SizeRow("Portraits", 3, &o.Cards),
			FlipRow("Party layout", 3, &o.PartyRow, func() string { return Choice(o.PartyRow, "Sides", "Row") }),
			FlipRow("Minimap", 3, &o.MapOff, func() string { return OnOff(!o.MapOff) }),
			SliderRow("Minimap opacity", 3, &o.MapAlpha, 0.2, 1, 0.1, "", true),
			SizeRow("Minimap size", 3, &o.Map),
			SizeRow("Picture box", 3, &o.Pic),
			SwitchRow("Damage Numbers", 3, &o.DamageNumbers),
		},
		"Gameplay": {
			SliderRow("Zoom speed", 4, &o.Zoom, 0.25, 2, 0.25, " %.2fx", false),
			SwitchRow("No Map Zones", 4, &o.NoMapZones),
		},
	}
}

// HostRow is a row of the Esc menu for a setting of the game's own: on Page
// before the built-in row At, editing Flag (ON/OFF, or Names off and on) or
// Level (a 0 to 1 slider), or shown by Get and changed by Change.
type HostRow struct {
	Page, Label string
	At          int
	Flag        *bool
	Names       [2]string
	Level       *float64
	Get         func() string
	Change      func(int)
}

// HostRows are rows and o's own rows (OptionRows) by page for blobber.ExtraRows:
// where they share a built-in row, rows come first.
func HostRows(o *Options, rows ...HostRow) map[string][]blobber.Row {
	out := map[string][]blobber.Row{}
	for _, r := range rows {
		var b blobber.Row
		switch {
		case r.Flag != nil && r.Names != [2]string{}:
			b = FlipRow(r.Label, r.At, r.Flag, func() string { return Choice(*r.Flag, r.Names[0], r.Names[1]) })
		case r.Flag != nil:
			b = SwitchRow(r.Label, r.At, r.Flag)
		case r.Level != nil:
			b = SliderRow(r.Label, r.At, r.Level, 0, 1, 0.1, "", true)
		default:
			b = blobber.Row{Label: r.Label, At: r.At, Get: r.Get, Change: r.Change}
		}
		out[r.Page] = append(out[r.Page], b)
	}
	if adLib != nil { // the game's AdLib songs: on the AdLib or the SSI-2001's SID
		out["Audio"] = append(out["Audio"], FlipRow("Music card", 0, &o.SIDMusic, func() string { return Choice(o.SIDMusic, "AdLib", "SSI-2001 (SID)") }))
	}
	for page, rs := range OptionRows(o) {
		out[page] = append(out[page], rs...)
	}
	return out
}
