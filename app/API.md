# gobox/app API (module gobox/app; app needs a window, app/saves is pure Go)

The app shell: the window, the frame loop, fullscreen and the window's
shape, a scripted capture's clock and screenshots, and save slots on disk.
A game's `cmd` loads its data, builds its rules engine and calls `app.Run`.
No game names: the title, shape, key codes and save marker are the game's data.

## app
- `Run(Config) error`: opens the window and runs until the game ends.
  Window = `Theme.W x Theme.H` times `Settings.Scale`, locked to that shape
  in windowed mode (a resize keeps it); `Settings.Fullscreen` takes the
  screen's own resolution and shape. `Capture.Window` (W, H points) opens a
  window of that size and any shape. Layout is in device pixels: the game
  draws the whole window every frame (no bars but its own).
- `Config{Game, Theme, Settings, Capture, TPS}`; `Theme{Title, W, H}`;
  `Settings{Scale, Fullscreen}`.
- `Game` interface: `Update()` once a tick (TPS a second; in a scripted run
  the game must ignore the user's input), `Step(k Key)` once a scripted
  frame before Draw (k = the key due, 0 none), `Draw(dst *gpu.Image)`.
- `Capture{Shot, Keys, Freeze, Quit, Window, GPU, Slot}`, `Scripted()`: a run is
  scripted when Shot, Freeze or Quit is set. A key every `Ticks` (15)
  frames; after the keys plus 4 more: Shot is written (with `%d`: every
  frame while the keys play, numbered) and the run exits; Quit exits;
  Freeze stops Step. GPU is passed to `gpu.SetConfig`.
  Slot (a save slot name, "" none) is passed to the game's `Start` if it is
  a `Starter`, once before the first tick; what it means is the game's.
- `Prefs{Dir}`: the settings file `Dir/SETTINGS.JSON`, generic JSON keys
  (their meaning is the game's). `Load(dst ...any) bool` fills each pointer
  (missing keys keep dst's defaults; false: no file), `Save(v) error` writes
  v indented, atomically (temp file then rename), creating Dir.
- `Key` (uint16): a game key code, or the shell's `Right`, `Left` (a click
  at the pointer), `Next` (the pointer to its next position), `Wait`.
- `ParseKeys(s, esc map[byte]Key) []Key`: `\M \C \P \w` are the shell's,
  esc the game's (`\r`, `\e`, arrows, F-keys: the game's codes); any other
  byte is its own code.
- `ParseSize("WxH") (w, h, err)`.

## app/saves
- `Store{Dir, Marker}`: a slot is a folder `Dir/<name>` holding `Marker`;
  folders starting with "." are saves in progress.
- `List() []Slot` newest first (`Slot{Name, At}`), `Index(name)`,
  `Path(name)`, `Delete(name)`.
- `Write(name, func(dir string)) error`: the game writes its files into a
  fresh folder beside the old slot; it replaces the old one only once
  Marker is there (`ErrIncomplete` otherwise).
- `Name(s)`: keeps letters, digits and " -_!'", trimmed.

## Gotchas
- The game keeps its input reading, the latch and its settings' keys; the
  shell only ticks and draws it.
