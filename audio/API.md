# gobox/audio
START.EXE's speaker/Tandy sound driver, the Amiga player and oto output. Tables stay in the game (`sounddata.json`).
```go
type Config struct {
    Data   []byte                    // sounddata.json (go:embed)
    Card   byte                      // 'P' speaker, 'T' Tandy, else silent
    Amiga       []byte  // Amiga SOUND.BIN image (amiga.Base..End); nil: DOS driver
    AmigaSongOn uint32  // address of the Amiga player's "song running" byte
    Volume func(kind string) float64 // "music"|"sfx" -> 0..1; nil = 1
}
func NewPlayer(Config) (*Player, error)
func (*Player) Play(n int)   // PlaySound number 0..15, 255
func (*Player) Read([]byte)  // s16le stereo 44100 Hz
func Output(io.Reader)       // oto, shared context; no device = no-op

// The DOS sequencer and chips alone, for a game that makes the driver calls itself
// (no card switching, Amiga, volume or lock; the caller serializes).
func NewSynth(data []byte, card byte) (*Synth, error) // 'T' Tandy, else speaker; unhooked, enabled
func (*Synth) Start(song int)   // on the card's set
func (*Synth) Stop()            // song 0 on both sets
func (*Synth) Enable(bool)      // chips sound / stay silent while scripts run
func (*Synth) Hook(); func (*Synth) Unhook()
func (*Synth) Song() int
func (*Synth) Render([]byte) int // s16le stereo 44100 Hz, unscaled

// The full driver, a front.Sound (gobox/front/sound.go): PlaySound with the 0.1 s cut,
// card switching, Amiga songs, soundtrack recordings and voice ducking. Locked; Read is the mixer.
const (Tandy = iota; Speaker; None)              // SetCard values
func NewDriver(data []byte, card byte) (*Driver, error) // 44100 Hz
func (*Driver) Attach(am *amiga.Player, fx []int, t Tracks, p Pack, speaking func() bool, vol func(kind string) float64) // any nil
func (*Driver) Read([]byte) (int, error)
func (*Driver) Effect(n int, quiet bool)  // quiet: hook only
func (*Driver) CutAfter(frames int); Wait(n int); Reset(); SetCard(c int)
func (*Driver) Song(n, next int, again bool) bool; AmigaSong(n int)
func (*Driver) SetSoundtrack(on bool, n, music int); Track() (on bool, song int); Sounding() bool
type Tracks interface{ Len() int; Has(int) bool; Song() int; Reset(); Start(int); Playing() bool; Mix(m []byte, loop bool) }
type Pack interface{ Mix(p []byte) }
```
present.NewFront(g) builds a Tandy Driver from Data.Sounds (Amiga from Data.AmigaSound) when no
front.Sound is passed; NewFront(g, drv) shares the game's own. Lock order: driver, then voice, then pack.
Generator: `go -C ../gobox-engine/audio run ./cmd/sounddata <flags> START.EXE > sounddata.json` (`go run gobox/audio/...` works only from a module that requires gobox/audio).
Switch, all games: `require gobox/audio v0.0.0`, `replace gobox/audio => ../gobox-engine/audio`; delete the
game's sound.go, sounddata.go, amiga/, volume_*.go; embed the JSON; `NewPlayer`, `audio.Output(p)`.
- cotab: `Card: Cfg.Sound, Dir: *dir`, Volume = the old volume_js.go logic; EvSound n -> `p.Play(n)`, `Play(255)` on Exit.
  `-segment 0x699 -jumptable 356 -speakerset 368 -tandyset 778 -speakersongs 1186 -tandysongs 1354 -songs 21
  -enabled 336 -commands reset=5521,return=5615,call=5622,switch=5512,loop=5489,set=5570`
- sotsb: front end maps EvSound n -> `p.Play(n)`; drop Engine.snd, SoundReader, soundToggle (Ctrl-S: Play(0)/Play(1)).
  `-notes -segment 0x7E4 -jumptable 354 -speakerset 366 -tandyset 776 -speakersongs 1184 -tandysongs 1288 -songs 13
  -enabled 334 -speakerperiods 750 -tandyperiods 1160 -commands reset=14112,return=14206,call=14213,switch=14103,loop=14080,set=14161`
- por: `audio.Output(r)` replaces the oto setup (main.go:590,:687) per reader (speaker, voice share one context).
  `-segment 0x8AE -jumptable 545 -speakerset 557 -tandyset 967 -speakersongs 1375 -tandysongs 1543 -songs 21
  -enabled 525 -commands reset=5710,return=5804,call=5811,switch=5701,loop=5678,set=5759` por: `NewDriver`, `drv.Attach(eng.SoundParts())`, `eng.SetSound(drv)`, `audio.Output(drv)`, `present.NewFront(eng, drv)`.
Cotab/sotsb differences: notes, envelopes, periods are data (sotsb only); song count is the table length; volume
is Config.Volume; card/dir are Config; NewPlayer hooks the timer (sotsb never set `isr`: assumed bug).

## Migrate
- `Config.Dir` is gone (no DISK2 probing). Read the Amiga `SOUND.BIN` yourself and set `Config.Amiga` (nil keeps the DOS driver) and `Config.AmigaSongOn` (was the hard-coded 0x667d5 in `amiga`).
- The DS offsets in sound.go were comments only; no sounddata.json change.

## AdLib (OPL2) and PCM effects
```go
type AdLibData struct {
    Music, Sounds []byte      // MUSIC.DAX (song n = block n), ZOUNDS.DAX (may be nil)
    Effects map[int]int       // effect number -> Sounds block id
    Header bool; Skip []int   // 8SVX^0x80 FORMs (play BODY); Skip: block ids that are raw u8 PCM
    Rate int                  // effect rate, Hz
}
func NewAdLib(AdLibData) (*AdLib, error)  // a Tracks: Len, Has, Song, Reset, Start, Playing, Mix(m, loop)
func (*AdLib) Effect(n int) bool          // PCM over the song; the caller Resets the music if the game cuts it
```
Mix fills m (s16le stereo 44100 Hz) with OPL2 (opl2.Resampler) plus the effect. Not locked; the Driver's lock covers it.
Spec: pod-go docs/MUSIC.md; test compares the first 300 ticks with `podmusic.py regs` (testdata/pod_regs1.txt).
