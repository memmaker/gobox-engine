# Pools of Darkness: AdLib music and digitized sounds

This spec is for writing an OPL2 player in Go with `github.com/memmaker/opl2`. It
covers the file formats, the game's calls into the driver, and the AdLib driver
(`SOUND.COM`) step by step. The reference implementation is
`tools/music/podmusic.py`, which follows the driver's code line for line. It can
dump every song (`dump`), write the driver's OPL register log for a block
(`regs`), and run a self-check (`test`). `tools/music/render` is a small Go main
that plays such a log through opl2 into a WAV file.

Addresses: `SOUND.COM:xxxx` means the offset in the COM image, with origin 0x100
(see `sound.asm`, made by `ndisasm -o 0x100`). The driver's data segment is
`DS = CS + 0x2D1`, so `DS:x` is file offset `0x2C10 + x`. Anything past the end
of the file is BSS and starts at zero.

Confidence labels:
- **[verified]**: checked against the data or by running the reference code.
- **[disasm]**: read from the disassembly but not compared with a DOSBox
  capture.

## 1. Which driver, which files

- **[verified]** `MUSIC.DAX` block 0 is byte-identical to `SOUND.COM`. The game
  never runs `SOUND.COM` from DOS: it loads block 0 itself (R5A0:00CC, error text
  "unable to load com file"). `SOUND.COM` is Paragon's multi-device song driver.
  Its device names are psnd=0, tsnd=1, adlib=2, cms=3, innov=4 and covox=5. With
  no command line it probes for the devices in the order adlib, cms, innov,
  covox, tandy, speaker, so on an AdLib it picks device 2. Only device 2 matters
  here.
- **[verified]** `SND.COM` is identical to `SBLASTER.COM`: John Ratcliff's
  **DigPak** Sound Blaster driver. `START.BAT` loads it as a TSR on INT 66h. It
  plays the digitized sounds. `ADLIB.COM`, `IBMSND.COM` and the `TAN*.COM` files
  are the other DigPak builds that `INSTALL` can copy over `SND.COM`. They don't
  matter for an AdLib player.
- **[verified]** `MUSIC.DAX` blocks 1–8 are songs. `P1.SMS` is the same file as
  block 1, and `P3.SMS` is the same as block 3.
- **[verified]** `ZOUNDS.DAX` blocks 1–11 are digitized sounds (section 7).

### DAX container [verified]
- File layout: a u16 LE `n` (the index size in bytes), then `n/9` entries, each
  `{id u8, offset u32 LE, rawSize u16 LE, compSize u16 LE}`. The block data
  starts at `2 + n + offset`.
- RLE: read a control byte `c`.
  - `c < 0x80`: copy the next `c+1` bytes.
  - Otherwise: repeat the next byte `256-c` times.
- Stop when `rawSize` bytes are out.

## 2. Game side: N → block, and the driver calls

The engine emits two events: `EvMusic N` (from `rules/music.go`) and `EvSound N`
(from `rules/digitized.go` and the ECL `SOUND` command in
`rules/eclcmds2.go`).

| Event | Game routine | Data |
|---|---|---|
| `EvMusic N` | playMusic R5A0:0031(N) | `MUSIC.DAX` block **N** (1–8), song 0 inside it |
| `EvSound N` (N = 2..12) | playDigitized R5A0:0000(N) → R6C9:0000 | `ZOUNDS.DAX` block **N−1** (1–11) |

- **[verified]** At startup O079:0121 loads ZOUNDS blocks 1..11 into `tab[1..11]`
  (DS:FB1A, 6-byte entries). R6C9 plays `tab[N-1]` for `1 < N < 13`.
- ECL `SOUND n`: an `n` of 128 or more starts music block `n-128`. An `n` from 2
  to 14 is a digitized sound, but only 2..12 do anything.
- `stepSound` is 10, which is ZOUNDS block 9.
- Nothing plays when the sound setting is "none" (`b92A8 == 5`).

### playMusic(N) [disasm, R5A0/R6B9]
Music state is `b92C0`: 0 = idle, 2 = music playing.

1. If music is playing and `N` differs from the last block (`b92CA`), stop it
   (driver function 3).
2. If the state is idle:
   1. Remember `N`.
   2. Load `MUSIC.DAX` block N into a 6144-byte buffer.
   3. Call driver function 0 with that buffer.
3. Always call driver function 1 with argument 0. This (re)starts song 0 from
   the beginning, so asking for the song that is already playing restarts it.

### playDigitized(N) [disasm]
1. If music is playing, **stop the music** (driver function 3). A digitized
   sound cuts the music, and the music does not come back by itself.
2. Hand the sound to DigPak (section 7).

### Driver call interface [disasm]
- **Entry point.** The driver image starts on a paragraph boundary. The game far
  calls `(imageSeg − 0x10):0x0100`, the COM entry point.
- **Arguments.** The game pushes `seg`, then `off`, then `fn`, Pascal style. The driver returns with a plain RETF (it does not pop them), and the
  result comes back in AX.

| fn | Arguments | Effect |
|---|---|---|
| 0 | `seg:off` of the song block | Init: detect the device, init the OPL, program the PIT and hook INT 1Ch, load instruments, parse SLBR |
| 1 | `n` in 0..2 | `n` selects a detune preset (`DS:0000` = {0,2,3} → section 5.6; the game always passes 0 → no detune). Then start song 0 |
| 2 | – | Return 1 if any track is still playing |
| 3 | – | Stop: restore the PIT divisor to 0 (65536), unhook INT 1Ch, key off everything |

## 3. Song block layout (GLBR + SLBR) [verified on all 8 blocks]

All multi-byte numbers in the block are **big-endian**, unlike the DAX index.

```
0x000  GLBR parts file, always 0x1BC bytes
  0x00 "GLBR"
  0x04 u32  size of the GLBR part (0x1BC)
  0x08 u16  0xFFFF
  0x0A u16  part count (32)
  0x0C u16[32] part sizes
  0x4C parts, in order:
       0..7   12 bytes each: AdLib instruments 0..7 (section 5.2)
       8..15  18 bytes each: CMS instruments (ignore)
       16..23 8 bytes each: Innovation (ignore)
       24..31 8 bytes each: Covox (ignore)
0x1BC  SLBR song section (the driver uses this fixed offset, SOUND.COM:03A0)
  0x00 "SLBR"
  0x04 u32  size of the SLBR section
  0x08 u8   song count S (1 in every block)
  0x09 u8   (0)
  0x0A u16[S+1] song offsets, relative to songBase = 0x0A + 2*(S+1)
song header at songBase + off[s]:
  0x00 u16  tempo
  0x02 u8   initial velocity (127 everywhere)
  0x03 u8   track count T (4 everywhere)
  0x04 u16[5] track offsets, relative to song + 0x0E
  0x0E track data
```

The driver finds part `i` at `12 + 2*count + sum(size[0..i-1])` (SOUND.COM:0136).

### Tracks
- **Track 0** is the single-voice track. It is used only for the PC speaker
  (device 0, flag `DS:0260`). **An AdLib player skips track 0.**
- Header tracks `1..T-1` go to driver track slots `0..T-2` (SOUND.COM:076B). In
  every PoD block that means 3 AdLib tracks.
- The driver has 6 slots of 14 bytes each at DS:029E:

| Offset | Field |
|---|---|
| +0 | u32 tick of the next event (0 = inactive) |
| +4 | far pointer to the event stream |
| +8 | u16 `div` |
| +A | u16 fractional remainder |
| +C | u8 velocity |
| +D | u8 state |

### Events: 2 bytes `b0 b1` [verified: every byte of every track is one of these]

| b0 | Meaning in SOUND.COM |
|---|---|
| 0x00–0x7F | Note (MIDI number, 60 = C4) with duration code `b1` |
| 0x80 | Rest with duration code `b1` |
| 0x81 | "Program": key off if a note is sounding, then **set instrument = slot index (slot < 4 ? slot : 1)**. `b1` holds the original program number but is **ignored**. Each track starts with two of these. |
| 0x84 | Velocity: **sets velocity to 0x7F**. `b1` (the authored velocity) is **ignored**. |
| 0x82, 0x83, 0x87 | **Ignored**: skipped without effect. In the data they hold 0x82 = 0x1A, 0x83 = 0, and 0x87 = 0/1, all at the start of each track. They look like leftovers of a MIDI conversion. |
| 0xFF | End of track |
| other > 0x80 | Skipped (none appear in the data) |

Any number of control events (0x81–0xFE) are processed at the same tick, up to
the next note, rest or end.

Duration code `b1` (fn 0x662) **[verified arithmetic]**:
- bits 0–2: base length `L = 0x6900 >> (b1 & 7)`. 0x6900 (26880) is a whole
  note, so 1/2, 1/4 … 1/128.
- bit 3: dotted, `L = (L >> 1) * 3`.
- bits 4–5: tuplet, with `n = (b1 & 0x30) >> 3`, so 2, 4 or 6, and
  `L = L / (n+1) * n` (integer division). For example 0x10 is a triplet (×2/3).
- bit 6 (0x40): tie or legato into the next note.
- bit 7: ignored. It appears in a few tracks and has no effect.

The data uses note numbers from about C2 to C6.

Counts over all 8 blocks: 4049 notes and rests, 428 × 0x84, 56 × 0x81 and
0x87, 32 × 0x82, 0x83 and 0xFF.

## 4. Timing [disasm]

- **PIT.** fn 0 programs PIT channel 0 with divisor **0x4000**: `out 43h,36h;
  out 40h,00h; out 40h,40h`. That gives `1193182/16384 = 72.8227 Hz`.
- **Tick counter.** It hooks **INT 1Ch**. The ISR (SOUND.COM:2BA9) adds 1 to a
  32-bit counter `now` on every IRQ, but **runs the sequencer only when `now` is
  even**, that is at 36.41 Hz. All durations are counted in 72.82 Hz ticks.
- **Tempo divisor.** `div = tempo / 84`, integer division. 84 =
  `0x1110 / 0x34` (fn 0x649 and 0x33F). The PoD tempos are 7680–18944, giving
  `div` from 91 to 225.
- **Event length in ticks.** `L(b1) / div`. The remainder `L % div` adds to the
  slot's `+A` field. When `+A >= div`, add 1 tick and subtract `div` from `+A`.
- **Start tick.** At song start every slot gets `next = now` (the driver waits
  until `now != 0`), `rem = 0`, `vel = header velocity`, `state = 3`, and its
  stream pointer.
- **Example.** A 1/16 note (1680) at div 198 is 8.48 ticks = 0.116 s.

Song lengths from the reference code: block 1 134 s, block 2 24 s, block 3 35 s,
block 4 14 s, block 5 22 s, block 6 28 s, block 7 25 s, block 8 15 s.

**No looping.** When every slot is inactive, the driver checks the loop song
`DS:0320`. fn 0 and fn 1 always leave it at −1, so the driver keys everything
off and the song ends. `SOUND.COM:073C` would set a loop, but nothing calls it.

### Sequencer step (SOUND.COM:09BD), run on each even `now`, for slots 0..5
Slot states: 0 = early note-off pending, 1 = sounding, 2 = tied, 3 = silent.

```
if slot.next == 0 or slot.next > now: continue
if state == 0:                      // early release before a repeated note
    noteOff(slot); state = 3; next += 1; continue
while 0x80 < b0 < 0xFF:              // control events
    0x81: if state<3: noteOff; state=3; setInstrument(slot, slot<4 ? slot : 1)
    0x84: vel = 0x7F
    p += 2
if b0 == 0xFF: next = 0; if state<3: noteOff; state = 3; continue
next += L(b1)/div (with the remainder carry above)
if b0 == 0x80: if state<3: noteOff; state = 3
else:
    if state != 2:                   // not tied: (re)strike
        if state < 3: noteOff
        noteOn(slot, b0, vel)
    if b1 & 0x40:            state = 2
    elif next event's b0 == b0: next -= 1; state = 0   // same pitch follows: release 1 tick early
    else:                    state = 1
p += 2
```
After the loop: if a slot ended this step and no slot is active, key everything
off (the "no loop" case above). A slot handles at most one note per step, so a
note shorter than 2 ticks makes that track lag a little. The real driver has the
same lag.

## 5. AdLib output (device 2)

### 5.1 Init: fn 0 → 0x1302, 0x127D [disasm]
The driver writes `BD=0x20` and `08=0x40`, then `E0+op = 0` for all 18
operators, then `01=0x00`. Notes:
- **Rhythm mode is on**: `BD` bit 5 stays set from here on.
- The waveform select is off, so every operator is a sine.
- All 9 channels are free (channel mask `DS:03B8 = 0`).

**Stop or silence** (0x12C8): `B0..B8 = 0x0E` (key off), `BD = 0x20`, and the
rhythm key bits (`DS:00DB`) are cleared.

### 5.2 Instrument record: 12 bytes, GLBR parts 0..7 [verified layout, disasm semantics]

| Byte | Register | Meaning |
|---|---|---|
| 0 | – | unused |
| 1 | – | fixed OPL channel (0 = allocate freely; a channel above 6 means a rhythm voice) |
| 2, 3 | 0x20 | modulator, carrier: AM/VIB/EG/KSR/MULT |
| 4, 5 | 0x40 | modulator, carrier: KSL/TL. The carrier TL is replaced at note-on (5.4) |
| 6, 7 | 0x60 | modulator, carrier: AR/DR |
| 8, 9 | 0x80 | modulator, carrier: SL/RR |
| 10 | 0xC0 + channel | feedback/connection |
| 11 | – | unused |

There is no waveform byte (registers 0xE0 stay 0).

Every PoD instrument has byte 1 = 0, so the rhythm-voice code (5.5) never runs.

Example, block 2 instrument 1: `00 00 b1 61 1c 80 41 92 1f 3b 0e 00`.

### 5.3 Instrument set (0x15F2 → 0xFF1 → 0x10FE) [disasm]

**Allocate a channel.**
1. Free the slot's old channel: clear its bit in the mask. The bit starts as
   channel 0 for every slot, because the field is BSS.
2. If the instrument has a fixed channel `f`, set its bit and use `f`.
3. Otherwise take the lowest `n` from 0 up whose bit is clear (stop at 7), and
   set its bit. With 3 tracks this gives slot 0 → channel 0, slot 1 → 1,
   slot 2 → 2.

**Operator offsets.**
- Word table `DS:00DA` = `0,0,1,2,3,4,5,8,9,10,11,12,13,16,17,18,19,20,21`.
- Byte map `DS:0109` (used because `DS:00BC = 0`, rhythm mode) =
  `1,2,3,7,8,9,13,17,15,18,14`.
- For channel `c`, the modulator is at `T[M[c]]` and the carrier at
  `T[M[c]+3]`. These are the standard OPL operator offsets.

**Program.** For each operator `k` (0 = modulator, 1 = carrier), in this order:
1. `BD=0x20`
2. `08=0x40`
3. `40+op = rec[4+k]`
4. only for `k=0`: `C0+c = rec[10]`
5. `60+op = rec[6+k]`
6. `80+op = rec[8+k]`
7. `20+op = rec[2+k]`

### 5.4 Note on / off (0x158A → 0x1365) [disasm; pitches verified in render]
`noteOn(slot, note, vel)`: slots 0–2 use `vel`. Slots 3–5 use `vel*100>>7`, but
they are unused in PoD.

**Melodic channel** (`c ≤ 6`). Let `r` be the slot's instrument.

1. **Key off.** Write `A0+c = 0x4A` and `B0+c = 0x0E`. This happens on every
   note-on and every note-off.
2. **Note-off** is `vel = -1`: stop here.
3. **Level.** Write `40+carrier = (0x3F − ((0x3F − (r[5]&0x3F)) * vel >> 7)) | (r[5]&0xC0)`.
   At vel 127 this is close to the instrument's TL.
4. **Octave.** `oct = clamp(note/12 − 1, 0, 7)`.
5. **F-number.** `fnum = F[note%12]`, with `F` = `343,365,385,408,433,459,486,515,546,579,614,650`
   (DS:0114, C..B), plus `detune*G[note%12]` (G = DS:012C = `1,1,1,1,2,2,2,2,2,2,2,2`).
   In PoD the detune is always 0.
6. **Key on.** Write `A0+c = fnum & 0xFF` and `B0+c = 0x20 | oct<<2 | fnum>>8`.

So MIDI 60 is block 4, fnum 343, which is 260 Hz.

opl2 applies one write per sample, so the key off and the key on land in
different samples and the note restarts, as on real hardware.

### 5.5 Rhythm voice (`c > 6`; unused by PoD data, kept for completeness) [disasm]
- **Channel.** `fc = (c ≤ 8) ? 7 : 8`.
- **Key bit.** `bit = 1 << (10−c)`: c=7 SD, 8 TT, 9 CY, 10 HH.
- **Operator.** A single operator, `T[M[c]]`.
- **Programming.** Only that one operator is programmed. `C0+c` gets `rec[10]`.

Note-on and note-off:
1. Clear `bit` and write `BD = bits|0x20`.
2. A note-off stops here.
3. Write the level from `r[4]`, the same way as the melodic carrier.
4. Write `A0+fc`, then `B0+fc` without 0x20.
5. Set `bit` and write `BD = bits|0x20`.

### 5.6 Detune / doubling (fn 1 presets) [disasm]
Bytes `DS:00D5..D7` hold a detune value for each slot 0..2. When a value is
non-zero, slot `s` also sounds on slot `s+3`'s channel, at velocity
`vel*100>>7`, with that detune. fn 1 with `n=0`, the only value the game uses,
sets all three to 0. `DS:00D8..` (the detune for slots 3–5) is 0 as well.
**A PoD player can ignore detune.**

## 6. Player recipe for gobox

1. **Parse.** Load the block. Read instruments `parts[0..7][:12]` and song 0.
   Take tracks `1..T-1` as slots `0..T-2`.
2. **Init.** Write the init registers from 5.1. Keep the mask and the slot →
   channel table.
3. **Clock.** Advance a 72.8227 Hz tick counter from the audio clock
   (`opl2.Rate / 72.8227` ≈ 682.7 chip samples per tick). On every even tick,
   run the step from section 4.
4. **Stop.** When the song ends, or on a stop or digitized sound, key off with
   0x12C8. A digitized `EvSound` stops the music. A new `EvMusic` restarts from
   song 0, even when it is the same block.

`podmusic.py regs GOG BLOCK OUT` prints the exact register stream
(`tick reg val`). Compare a Go player against it register for register.

## 7. Digitized sounds (ZOUNDS.DAX + DigPak) [verified format; playback disasm]

- **Format.** Each ZOUNDS block is an IFF `FORM/8SVX` file (VHDR, NAME, ANNO,
  BODY). Only the header reads ^0x80 (`FORM` is stored as C6 CF D2 CD); the
  BODY is already **unsigned 8-bit PCM** centred on 0x80, so the block plays
  as stored. Block 4 is headerless raw u8 PCM. XORing the BODY too turns it
  into wrapped signed values, heard as heavy distortion.
- **Playback.** R6C9:0000 first waits while DigPak reports a sound playing
  (INT 66h AX=0689h). Then it copies the **whole block**, header included, into
  a 9000-byte buffer and calls INT 66h AX=0688h (DigPlay) with
  `SNDSTRUC {far ptr, len = block size, IsPlaying = nil, frequency = 9000}`.
  Notes:
  - Every sound plays as **unsigned 8-bit mono at 9000 Hz**. The VHDR rates
    (1772–10980 Hz) are ignored.
  - The ~100 header bytes play too: a few milliseconds of noise at the start.
    It is accurate to play them, and harmless to skip to BODY.
  - The driver is found by its "KERN" signature 6 bytes before the INT 66h
    vector.

`podmusic.py zounds GOG` lists them.

## 8. What is verified

- **Formats: verified.** This covers the DAX unpacking, the GLBR/SLBR layout,
  the instrument records, the event grammar and the 8SVX^0x80 layout.
  `podmusic.py dump` parses all 8 song blocks, accepts only the event codes
  listed, finds 0xFF ending every track, and reproduces the section sizes in the
  headers. `podmusic.py test` checks that block 0 equals SOUND.COM and that the
  duration codes come out right.
- **Driver behaviour: disasm.** This covers the sequencer, timing, allocation,
  register writes and the ignored 0x82/0x83/0x87 and 0x81/0x84 arguments.
  `podmusic.py` was transcribed from it. It has not been compared with a DOSBox
  OPL capture.
- **Rendering: checked.** Block 2 rendered through opl2 (`tools/music/render`)
  gives a 25 s piece. FFT peaks fall on the expected notes (C4 261 Hz,
  A#3 231 Hz, G#3 206 Hz) in the expected arpeggio rhythm. The judgement that it
  sounds like music rests on these peaks; nobody has listened to it.
- **Game side.** The N → block mapping and the call order come from the pod-re
  decompilation (R5A0, R6B9, R6C9, O079:0121), confidence high. The 9000 Hz
  DigPak rate is `dFB74 := 9000` in R6C9.

## 9. Innovation SSI-2001 (SID)

Device 4 (`innov`) drives a MOS 6581 at a port base found by the probe
(SOUND.COM:2933, ports 0x280..0x2E0 step 0x20, stored by 0x165A into `DS:03FC`).
Every write goes through 0x16C6 `(reg, val)`: `out base+reg, val`, with an error
message when `val` is outside 0..255. Voice `v` (0..2) uses registers `7v+0..6`.
`podmusic.py sidregs GOG BLOCK OUT` writes this log (`tick reg val`, reg relative
to the base, same ticks as `regs`).

The sequencer, timing and track layout are exactly sections 3 and 4: the
device-4 entries of the dispatch tables are init 0x1723, silence 0x16F5 (via
0x2460), stop 0x17D3 (= 0x16F5), instrument load 0x17DB, set instrument 0x19F7
(via 0x240F) and note 0x18C0 (via 0x2569, no velocity scaling and no detune
doubling, unlike AdLib's 0x158A). [disasm]

### 9.1 Init (fn 0 → 0x2131) [disasm]
1. 0x2460 → 0x16F5: `7v+4 = ctrl[v]` for v=0..2 (`ctrl` = `DS:0198..019A`, 0 in
   the file), so `04=0B=12=0`.
2. 0x1723, for v=0..2: `7v+0=0, 7v+1=0, 7v+2=0, 7v+3=0x08` (pulse width 0x800,
   50%, never changed afterwards), `ctrl[v]=0x10, 7v+4=0x10, 7v+5=0, 7v+6=0`.
3. `15=0x07, 16=0xBF` (cutoff 0x5FF), `17=0x00` (resonance 0, no voice
   filtered), `18=0x1F` (low-pass, volume 15).

### 9.2 Instrument record: 8 bytes, GLBR parts 16..23 [verified layout, disasm semantics]
0x225C → 0x21CD → 0x17DB copies 8 bytes of part `16+i` to `DS:03C2 + 7*i`, for
i = 0..7 in order. The stride is 7, so byte 7 of each record is overwritten by
byte 0 of the next one, and byte 7 of record 7 lands on the low byte of the voice
mask `DS:03FA` (it is 1 in blocks 1, 3 and 8; harmless, see 9.3).

| Byte | Meaning |
|---|---|
| 0 | attack (high nibble of `7v+5`) |
| 1 | decay (low nibble of `7v+5`) |
| 2 | sustain, scaled by velocity at note-on |
| 3 | release (low nibble of `7v+6`) |
| 4 | waveform: `ctrl = b4<<4` (1 triangle, 2 saw, 4 pulse, 8 noise) |
| 5 | key scaling of velocity per octave (0 in all PoD data) |
| 6 | filter: non-zero routes the voice through the filter |
| 7 | unused (overwritten) |

Instruments used (slot `s` gets instrument `s`, so only records 0..2 sound),
example block 2: `00 03 01 03 02 00 00`, `05 08 08 0c 04 00 00`, `00 06 00 01 01 00 00`.
**[verified]** Every record 0..2 in PoD has byte 6 = 0, so nothing is filtered
and reg 0x17 is always 0; only records 4..7 of some blocks set it.

### 9.3 Set instrument (0x19F7) and voice allocation [disasm]
1. Free the slot's voice: clear bit `voice[s]` in the mask (0x16B3). `voice[]`
   is `DS:0180`, initialised in the file to 0,1,2,3,4,5.
2. Allocate (0x1665): the lowest `v` from 0 whose bit is clear, stopping at 3.
   An already-set bit there prints an error; v=3 would write into the filter
   registers. Set the bit. With PoD's 3 tracks: slot 0 → voice 0, 1 → 1, 2 → 2.
3. Program (0x1817): `ctrl[v] = b4<<4`, `7v+4 = ctrl[v]` (gate off),
   `7v+5 = b0<<4 + b1`, `7v+6 = b2<<4 + b3`.
4. Store `voice[s]=v`, `ins[s]=i` (`DS:018C`).
5. Filter routing byte `DS:019B` (= `ctrl[3]`): set bit `v` if `b6 != 0`, else
   clear it. Write `17 = DS:019B`.

### 9.4 Note on / off (0x18C0) [disasm; transposition verified against `regs`]
`note(s, n, vel)`, `v = voice[s]`, `r` = instrument `ins[s]`:
1. `n += 3`.
2. Write `7v+4 = ctrl[v]` (gate off). This is the whole note-off (`vel = -1`),
   and also happens before every note-on.
3. Stop if `v` is not 0..2 or `vel < 0`.
4. `oct = clamp(n/12 − 1, 1, 8)`, `F = (T[n%12] << oct) & 0xFFFF`,
   `T` = `274,291,308,326,346,366,388,411,435,461,489,518` (DS:019C).
   Write `7v+0 = F & 0xFF`, `7v+1 = F >> 8`.
5. `vel = clamp(vel − (oct−4)*b5, 0, 127)` (0x162D), `S = (b2*vel >> 7) + 1`.
   Write `7v+6 = S<<4 + b3`. With vel 127 (always in PoD): b2 = 0 → 1,
   8 → 8, 1 → 1.
6. Write `7v+4 = ctrl[v] | 1` (gate on).

**Clock and pitch.** `SID Hz = F * clk / 2^24`. `T[0]<<4 = 4384` is C4
(261.3 Hz vs 261.63) within 2 cents at **clk = 1.000 MHz**, so the table is a C-based scale
for a 1 MHz SID. But the index is `(n+3)%12`, so **every note sounds a minor
third (3 semitones) higher than on AdLib** (block 2's first bass note: AdLib F2
87 Hz, SID F=1740 = G#2 104 Hz). On a real SSI-2001 (clock 14.31818/14 =
1.0227 MHz) everything is a further +39 cents sharp. Note ranges in PoD give
oct 2..6, so the 16-bit shift never overflows.

### 9.5 Stop / end of song [disasm]
0x16F5: `7v+4 = ctrl[v]` for v=0..2 (gates off, waveform kept).

**[verified]** `sidregs` on blocks 1..8: 6979, 1633, 1597, 187, 337, 571, 1501,
253 writes; gate-ons 1154, 263, 257, 22, 47, 86, 241, 33, equal to the AdLib
key-ons; same song lengths. Not compared with a capture or rendered.
