package ui

import "image"

// KeyWindowFit is where a text window (an Esc menu, the options, a prompt)
// of title over lines goes on b at the UI's K: its scale k, K as far as the
// box fits, and its top row, centred or (bottom) one cell above the bottom.
func KeyWindowFit(b image.Rectangle, K float64, title string, lines []string, bottom bool) (k float64, top int) {
	w := 0
	for _, l := range TitleLines(title) {
		w = max(w, VisLen(l))
	}
	for _, l := range lines {
		w = max(w, len(l))
	}
	k0 := float64(max(1, min(int(K), b.Dx()/((w+4)*8))))
	rows := len(lines) + 2
	if title != "" {
		rows += len(TitleLines(title)) + 1
	}
	k = FitScale(k0, 1, func(k float64) bool { // K, as far as the box fits
		return k*float64((w+4)*8) <= float64(b.Dx()) && k*float64((rows+1)*8) <= float64(b.Dy())
	})
	h := int(float64(rows*8) * k)
	top = (b.Dy() - h) / 2
	if bottom {
		top = b.Dy() - h - int(8*k)
	}
	return k, top
}
