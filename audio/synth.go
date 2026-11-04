package audio

// Synth is the DOS driver's sequencer and its two chips without the Player's policy:
// no card switching, no Amiga, no volume, no lock. A game whose driver calls differ
// from Play's drives it itself and serializes the calls.
type Synth struct{ s *sound }

// NewSynth = the driver's state at start-up for card ('T' Tandy, else the speaker):
// unhooked, enabled, no song.
func NewSynth(data []byte, card byte) (*Synth, error) {
	d, err := parseData(data)
	if err != nil {
		return nil, err
	}
	s := &sound{d: d, enabled: true, tandy: card == 'T'}
	s.speaker = sndChannelSet{songs: d.speaker, periods: d.speakerPeriods}
	s.tandySet = sndChannelSet{songs: d.tandy, periods: d.tandyPeriods}
	s.sn.att = [4]int{snSilent, snSilent, snSilent, snSilent}
	return &Synth{s}, nil
}

// Start starts song on the card's set.
func (y *Synth) Start(song int) {
	if song < 0 || song >= len(y.s.speaker.songs) || song >= len(y.s.tandySet.songs) {
		return
	}
	if y.s.tandy {
		y.s.start(&y.s.tandySet, song)
	} else {
		y.s.start(&y.s.speaker, song)
	}
}

// Stop loads song 0, which ends every channel of both sets.
func (y *Synth) Stop() { y.s.stop() }

// Enable lets the chips sound (true) or keeps them silent while the scripts run on.
func (y *Synth) Enable(on bool) { y.s.enabled = on }

// Hook hooks the timer interrupt: the sequencer runs.
func (y *Synth) Hook() { y.s.isr = true }

// Unhook restores the timer interrupt: speaker off, all Tandy voices off.
func (y *Synth) Unhook() { y.s.unhook() }

// Song is the song playing, 0 none.
func (y *Synth) Song() int { return y.s.song }

// Render fills p (whole frames of s16le stereo, 44100 Hz) and returns its length used.
func (y *Synth) Render(p []byte) int {
	n := len(p) / 4 * 4
	y.s.render(p[:n])
	return n
}
