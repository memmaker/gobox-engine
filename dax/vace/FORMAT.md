# VACE animation formats (Pools of Darkness INTRO / FINAL)

VACE.EXE (Kerry L. Bonin, SSI 1991, Turbo C++) plays `NAME.ACF` with the
frames of `NAME.BMA` through a video driver (`VGA.DRV`, `EGA.DRV`, ...) and
the songs `*.SMS` through the resident `SOUND.COM`. Sources below are
`ndisasm -b16 -e 1024 VACE.EXE` offsets (load image, segment 0x14C based at
0x14C0) and raw `VGA.DRV` offsets.

## .ACF: the control script

A stream of records. Each record starts with a little-endian word: low
byte = record length in bytes (including that word, always even), high
byte = opcode. Arguments are the following little-endian words w0, w1, ...
VACE pre-scans the file for op 5 and op 1 (1D60), then runs it (1F64; the
switch is at 1FB1, table cs:1333).

| op | len | meaning |
|----|-----|---------|
| 0  | 4  | speed: frame delay = 62 / w0 (integer) timer ticks of 18.2065 Hz; with music on (226A) 261 / w0 ticks of 72.8227 Hz: VACE counts every INT 8 (its ISR at 30A1) and SOUND.COM sets the PIT divisor to 0x4000. About 5% slower, more after the truncation. Decode uses the music clock when the script loads a song |
| 1  | 2  | end of script |
| 2  | 6  | palette entry, set at once: hi(w0) index, lo(w0) R, hi(w1) G, lo(w1) B, VGA 6-bit |
| 3  | 2  | wait until the delay since the last frame is up, then draw the next frame (frames past the last are skipped) |
| 4-7 | - | ignored while playing; op 5 is the chunk table (pre-scan): w0 chunk number, w1 its frame count, w2 its size in bytes (low word) |
| 8  | 12 | load song: lo(w0) slot, w1..w4 the name, each word high byte then low ("P1" -> P1.SMS), space padded; only with music on |
| 9  | 2  | play the loaded song |
| 10 | 2  | stop the song |
| 11 | 2  | wait out the delay, clear the screen to colour 0; the next op 2 first restores the whole saved palette |
| 12 | 2  | nothing |
| 13/14 | 2 | set / clear a player flag (DS:089E); no effect on the picture |

Palette fades are runs of op 2 between "empty" frames (op 3 on a frame
that changes nothing).

INTRO: 14 chunks, 133 frames, about 82 s with music (74 s without). FINAL: 6 chunks, 58 frames, about 33 s (29 s).
Both load and play their song (P1, P3) before the first frame.

## .BMA: the frames

The chunks of the op 5 table, back to back. A chunk:

    u32 size (this chunk, bytes)
    u16 frames
    u32 offset[frames]   from the chunk start
    frame data

A frame (decoded by the driver's entry +0x9F, VGA.DRV 0x6733, flags 0):

    u16 width, u16 height (320, 200)
    rows:
      u8 y            (0xFF ends the frame)
      spans, until one whose x has bit 15 set:
        u16 x         (bit 15: last span of this row)
        u8  pixels    (span length; not needed to decode)
        u8  bytes     (data bytes that follow)
        tokens until the bytes are used: t = u8;
          n = t>>1 + 1; t odd: one colour byte repeated n times,
          t even: n literal colour bytes

Frames are deltas: each draws over the previous screen, across chunks.
The first frame of the animation (a one-frame chunk) is a full picture.

## .SMS: songs for SOUND.COM

IFF-like chunks: 4-byte tag, big-endian u32 size counting the 8-byte head,
data (68000 big-endian words; the format is shared with SSI's other ports).

- `GLBR`: header words (0x20 = 32 voices?, a table of 32 sizes 12/18/8)
  then OPL2 instrument register sets (e.g. `31 61 1C 80 51 54 03 67 0E 00`:
  the 0x20/0x40/0x60/0x80/0xE0 pairs of modulator and carrier).
- `SLBR`: a small header with four big-endian track offsets, then the
  tracks: byte pairs note/duration (`32 04`, `80 04` a rest) and
  0x80-0x87 commands (`84 xx` most often).

`P1.SMS` and `P3.SMS` are byte-identical to `MUSIC.DAX` blocks 1 and 3:
the same song block the game hands SOUND.COM (driver fn 0 load, fn 1 play
song 0, fn 3 stop). The sequencer (note/duration pairs, the 0x81-0xFE
commands, 0xFF end, the 72.8 Hz PIT clock with events on even ticks, the
tempo divisor = a SLBR song header word / 84, instrument loading) is
reversed in pod-go `docs/MUSIC.md` and ported register for register in
`gobox/audio/adlibseq.go`. P1: divisor 198, tracks of 878/330/337 events;
P3: 167, 146/156/114.

Playback: the game's Data lists each sequence's songs (`Sequence.Music`);
`front/present` reads `NAME.SMS` from the files directory into
`audio.AdLibData.Named` (deduplicated against the DAX blocks) and plays the
cues while the animation shows: load picks the song, play starts it on the
AdLib soundtrack, stop or the animation's end (or a skip) stops it.
