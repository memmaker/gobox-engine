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
		},
	}
}
