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
```
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
  -enabled 525 -commands reset=5710,return=5804,call=5811,switch=5701,loop=5678,set=5759` por embeds the JSON and drives `NewSynth` from its own PlaySound (0.1 s cut, Ctrl-S, dasp, soundtrack; Amiga via gobox/amiga).
Cotab/sotsb differences: notes, envelopes, periods are data (sotsb only); song count is the table length; volume
is Config.Volume; card/dir are Config; NewPlayer hooks the timer (sotsb never set `isr`: assumed bug).

## Migrate
- `Config.Dir` is gone (no DISK2 probing). Read the Amiga `SOUND.BIN` yourself and set `Config.Amiga` (nil keeps the DOS driver) and `Config.AmigaSongOn` (was the hard-coded 0x667d5 in `amiga`).
- The DS offsets in sound.go were comments only; no sounddata.json change.
