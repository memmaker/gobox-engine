#!/usr/bin/env python3
"""Pools of Darkness AdLib music: DAX unpacker, song dumper and driver emulator.

Spec: docs/MUSIC.md. This follows SOUND.COM (= MUSIC.DAX block 0) instruction by
instruction for the AdLib device, so its register log is what the real driver writes.

  podmusic.py dump  GOGDIR            print every song block and its events, check all
  podmusic.py regs  GOGDIR BLOCK OUT  write the OPL register log of MUSIC block BLOCK
                                      (lines "tick reg val", tick = 1193182/16384 Hz)
  podmusic.py sidregs GOGDIR BLOCK OUT  same for the Innovation SSI-2001 (device 4): SID
                                      registers 0x00..0x18 relative to the port base
  podmusic.py zounds GOGDIR           list the ZOUNDS.DAX digitized sounds
  podmusic.py test  GOGDIR            self-check (asserts)
"""
import struct
import sys

PIT_HZ = 1193182 / 16384  # INT 1Ch rate after the driver reprograms PIT channel 0


def dax(path):
    """{id: bytes} of a .DAX file (u16 index size, 9-byte entries, RLE blocks)."""
    d = open(path, 'rb').read()
    n = struct.unpack_from('<H', d)[0]
    out = {}
    for e in range(0, n, 9):
        bid, off, raw, comp = struct.unpack_from('<BIHH', d, 2 + e)
        src, dst, p = d[2 + n + off:2 + n + off + comp], bytearray(), 0
        while len(dst) < raw:
            c = src[p]; p += 1
            if c < 0x80:
                dst += src[p:p + c + 1]; p += c + 1
            else:
                dst += bytes([src[p]]) * (256 - c); p += 1
        out[bid] = bytes(dst[:raw])
    return out


def be16(b, o):
    return struct.unpack_from('>H', b, o)[0]


def parts(b):
    """GLBR parts file at the start of b: list of part byte strings."""
    assert b[:4] == b'GLBR', 'Invalid Parts File'
    n = be16(b, 10)
    sizes = [be16(b, 12 + 2 * i) for i in range(n)]
    p, out = 12 + 2 * n, []
    for s in sizes:
        out.append(b[p:p + s]); p += s
    return out


SLBR = 0x1bc  # the driver takes the song section at this fixed offset


def songs(b):
    """SLBR section: list of (tempo, velocity, [track event bytes...]) - AdLib tracks only
    (header offsets 1..n-1); track 0 is the single-voice (PC speaker) track."""
    s = b[SLBR:]
    assert s[:4] == b'SLBR', s[:4]
    cnt = s[8]
    offs = [be16(s, 10 + 2 * i) for i in range(cnt + 1)]
    base = 10 + 2 * (cnt + 1)
    out = []
    for k in range(cnt):
        h = base + offs[k]
        tempo, vel, ntr = be16(s, h), s[h + 2], s[h + 3]
        toffs = [be16(s, h + 4 + 2 * t) for t in range(ntr)]
        trs = []
        for t in range(ntr):
            p = h + 14 + toffs[t]
            q = p
            while s[q] != 0xFF:
                q += 2
            trs.append(s[p:q + 2])
        out.append((tempo, vel, trs))
    return out


def dur(code):
    """fn 0x662: length of a note in 1/div ticks."""
    x = 0x6900 >> (code & 7)
    if code & 8:
        x = (x >> 1) * 3
    if code & 0x30:
        n = (code & 0x30) >> 3
        x = x // (n + 1) * n
    return x


# ---- the AdLib half of SOUND.COM -----------------------------------------
OPSLOT = [0, 0, 1, 2, 3, 4, 5, 8, 9, 10, 11, 12, 13, 16, 17, 18, 19, 20, 21]  # DS:00DA
MAP = [1, 2, 3, 7, 8, 9, 13, 17, 15, 18, 14]   # DS:0109 (rhythm mode, DS:00BC = 0)
FNUM = [343, 365, 385, 408, 433, 459, 486, 515, 546, 579, 614, 650]  # DS:0114, C..B


class Driver:
    def __init__(self, glbr):
        self.log, self.now = [], 0
        self.inst = [bytes(p[:12]) for p in parts(glbr)[:8]]  # parts 0..7 -> 0x358
        self.mask = 0          # 0x3B8 OPL channels taken
        self.bd = 0            # 0xDB rhythm key bits
        self.oplch = [0] * 6   # 0xBD
        self.ins = [0] * 6     # 0xC9
        # 0x1302 init: 0x127D then reg 1 = 0
        self.w(0xBD, 0x20); self.w(0x08, 0x40)
        for k in range(1, 19):
            self.w(0xE0 + OPSLOT[k], 0)
        self.w(0x01, 0)

    def w(self, r, v):
        self.log.append((self.now, r, v & 0xFF))

    def alloc(self, i):               # 0xFF1
        fixed = self.inst[i][1]
        if fixed:
            self.mask |= 1 << fixed
            return fixed
        n = 0
        while self.mask & (1 << n) and n <= 6:
            n += 1
        self.mask |= 1 << n
        return n

    def program(self, ch, i):          # 0x10FE
        r = self.inst[i]
        if ch <= 6:
            ops = [OPSLOT[MAP[ch]], OPSLOT[MAP[ch] + 3]]
        else:
            ch = r[1] if r[1] else ch
            ops = [OPSLOT[MAP[ch]]]
        for k, o in enumerate(ops):
            self.w(0xBD, 0x20); self.w(0x08, 0x40)
            self.w(0x40 + o, r[4 + k])
            if k == 0:
                self.w(0xC0 + ch, r[10])
            self.w(0x60 + o, r[6 + k]); self.w(0x80 + o, r[8 + k]); self.w(0x20 + o, r[2 + k])

    def set_inst(self, t, i):          # 0x15F2
        self.mask &= ~(1 << self.oplch[t])
        c = self.alloc(i)
        self.program(c, i)
        self.oplch[t], self.ins[t] = c, i

    def voice(self, t, note, vel):     # 0x1365 (detune is always 0 here)
        c, r = self.oplch[t], self.inst[self.ins[t]]
        melodic = c <= 6
        tl = r[5] if melodic else r[4]
        lvl = (0x3F - ((0x3F - (tl & 0x3F)) * vel >> 7)) & 0xFF | (tl & 0xC0)
        if melodic:
            op, fc = OPSLOT[MAP[c] + 3], c
            self.w(0xA0 + c, 0x4A); self.w(0xB0 + c, 0x0E)
        else:
            fc = 7 if c <= 8 else 8
            self.bd &= ~(1 << (10 - c)) & 0xFF
            op = OPSLOT[MAP[c]]
            self.w(0xBD, self.bd | 0x20)
        if vel < 0:
            return
        self.w(0x40 + op, lvl)
        oc = min(max(note // 12 - 1, 0), 7)
        f = FNUM[note % 12]
        self.w(0xA0 + fc, f & 0xFF)
        b = (f >> 8) & 3 | oc << 2
        if melodic:
            self.w(0xB0 + fc, b | 0x20)
        else:
            self.w(0xB0 + fc, b)
            self.bd |= 1 << (10 - c)
            self.w(0xBD, self.bd | 0x20)

    def note(self, t, note, vel):      # 0x158A
        self.voice(t, note, vel if t < 3 or vel < 0 else vel * 100 >> 7)

    def silence(self):                 # 0x12C8
        for c in range(9):
            self.w(0xB0 + c, 0x0E)
        self.w(0xBD, 0x20); self.bd = 0


# ---- the Innovation SSI-2001 (SID) half of SOUND.COM, device 4 ------------
SIDF = [274, 291, 308, 326, 346, 366, 388, 411, 435, 461, 489, 518]  # DS:019C


class SidDriver:
    """Registers are relative to the card's port base (DS:03FC); voice v at 7*v."""
    def __init__(self, glbr):
        self.log, self.now = [], 0
        # 0x17DB copies 8 bytes of parts 16..23 to DS:03C2 + 7*i: byte 7 is overwritten
        # by the next record, the last one lands on the low byte of the voice mask.
        self.inst = [bytes(p[:8]) for p in parts(glbr)[16:24]]
        self.mask = self.inst[7][7]    # 0x3FA (high byte never matters)
        self.voice = [0, 1, 2, 3, 4, 5]  # 0x180, file data
        self.ins = [0] * 6               # 0x18C
        self.ctrl = [0, 0, 0, 0]         # 0x198..0x19B (0x19B doubles as the filter routing)
        self.silence()                   # 0x2131 -> 0x2460 -> 0x16F5
        for v in range(3):               # 0x1723
            for r, x in ((0, 0), (1, 0), (2, 0), (3, 8)):
                self.w(7 * v + r, x)
            self.ctrl[v] = 0x10
            self.w(7 * v + 4, 0x10); self.w(7 * v + 5, 0); self.w(7 * v + 6, 0)
        self.w(0x15, 7); self.w(0x16, 0xBF); self.w(0x17, 0); self.w(0x18, 0x1F)

    def w(self, r, v):                   # 0x16C6
        assert 0 <= v <= 0xFF, (r, v)    # the driver prints an error here
        self.log.append((self.now, r, v))

    def set_inst(self, t, i):            # 0x19F7
        self.mask &= ~(1 << self.voice[t])          # 0x16B3
        v = 0                                        # 0x1665
        while self.mask & (1 << v) and v <= 2:
            v += 1
        assert not self.mask & (1 << v), 'no free SID voice'
        self.mask |= 1 << v
        r = self.inst[i]                             # 0x1817
        self.ctrl[v] = r[4] << 4 & 0xFF
        self.w(7 * v + 4, self.ctrl[v])
        self.w(7 * v + 5, (r[0] << 4) + r[1])
        self.w(7 * v + 6, (r[2] << 4) + r[3])
        self.voice[t], self.ins[t] = v, i
        if r[6]:
            self.ctrl[3] |= 1 << v
        else:
            self.ctrl[3] &= ~(1 << v) & 0xFF
        self.w(0x17, self.ctrl[3])

    def note(self, t, note, vel):        # 0x18C0
        v, r = self.voice[t], self.inst[self.ins[t]]
        note += 3
        self.w(7 * v + 4, self.ctrl[v])              # gate off
        if not 0 <= v <= 2 or vel < 0:
            return
        oc = min(max(note // 12 - 1, 1), 8)
        f = SIDF[note % 12] << oc & 0xFFFF
        self.w(7 * v, f & 0xFF); self.w(7 * v + 1, f >> 8)
        vel = min(max(vel - (oc - 4) * r[5], 0), 0x7F)  # 0x162D
        s = (r[2] * vel >> 7) + 1
        self.w(7 * v + 6, (s << 4) + r[3])
        self.w(7 * v + 4, self.ctrl[v] | 1)          # gate on

    def silence(self):                   # 0x16F5
        for v in range(3):
            self.w(7 * v + 4, self.ctrl[v])


def play(glbr, song=0, maxticks=72 * 60 * 10, dev=Driver):
    """Run the driver on a song block; returns (register log, end tick)."""
    tempo, vel0, trs = songs(glbr)[song]
    drv = dev(glbr)
    div = tempo // 84            # 0x649: tempo / (0x1110 / 0x34)
    start = 1                    # 0x76B waits for a non-zero tick count
    T = [dict(next=start, ev=tr, p=0, rem=0, vel=vel0, st=3) for tr in trs[1:6]]
    now = 0
    while any(t['next'] for t in T) and now < maxticks:
        now += 2                 # the ISR calls the sequencer on even ticks only
        drv.now = now
        for i, t in enumerate(T):  # 0x9BD
            if not t['next'] or t['next'] > now:
                continue
            if t['st'] == 0:
                drv.note(i, 0, -1); t['st'] = 3; t['next'] += 1; continue
            ev = t['ev']
            while 0x80 < ev[t['p']] < 0xFF:
                b0 = ev[t['p']]
                if b0 == 0x81:
                    if t['st'] < 3:
                        drv.note(i, 0, -1)
                    t['st'] = 3
                    drv.set_inst(i, i if i < 4 else 1)
                elif b0 == 0x84:
                    t['vel'] = 0x7F
                t['p'] += 2
            b0, b1 = ev[t['p']], ev[t['p'] + 1]
            if b0 == 0xFF:
                t['next'] = 0
                if t['st'] < 3:
                    drv.note(i, 0, -1)
                t['st'] = 3
                continue
            q, r = divmod(dur(b1), div)
            t['next'] += q
            t['rem'] += r
            if t['rem'] >= div:
                t['next'] += 1; t['rem'] -= div
            if b0 == 0x80:
                if t['st'] < 3:
                    drv.note(i, 0, -1)
                t['st'] = 3
            else:
                if t['st'] != 2:
                    if t['st'] < 3:
                        drv.note(i, 0, -1)
                    drv.note(i, b0, t['vel'])
                if b1 & 0x40:
                    t['st'] = 2
                elif ev[t['p'] + 2] == b0:
                    t['next'] -= 1; t['st'] = 0
                else:
                    t['st'] = 1
            t['p'] += 2
    drv.silence()
    return drv.log, now


NAMES = {0x81: 'INST', 0x82: 'ctl82', 0x83: 'ctl83', 0x84: 'VEL7F', 0x87: 'ctl87', 0xFF: 'END'}
NOTE = 'C C# D D# E F F# G G# A A# B'.split()


def dump(gog):
    m = dax(gog + '/MUSIC.DAX')
    seen, hi = {}, set()
    for bid in sorted(m):
        b = m[bid]
        if b[:4] != b'GLBR':
            print(f'MUSIC block {bid}: {len(b)} bytes, not a song (SOUND.COM driver)')
            continue
        ss = songs(b)
        for k, (tempo, vel, trs) in enumerate(ss):
            log, end = play(b, k)
            print(f'MUSIC block {bid} song {k}/{len(ss)}: tempo {tempo} div {tempo // 84} '
                  f'vel {vel} tracks {len(trs)} (adlib {len(trs) - 1}) '
                  f'length {end / PIT_HZ:.1f}s')
            for i in range(8):
                print(f'  inst {i}: {parts(b)[i].hex(" ")}')
            for t, tr in enumerate(trs):
                evs = []
                for p in range(0, len(tr), 2):
                    b0, b1 = tr[p], tr[p + 1]
                    seen[b0 if b0 > 0x80 else 'note'] = seen.get(b0 if b0 > 0x80 else 'note', 0) + 1
                    if b0 > 0x80:
                        evs.append(f'{NAMES.get(b0, hex(b0))}:{b1:02x}')
                        assert b0 in NAMES, (bid, k, t, hex(b0))
                    else:
                        n = 'R' if b0 == 0x80 else f'{NOTE[b0 % 12]}{b0 // 12 - 1}'
                        evs.append(f'{n}/{dur(b1)}{"~" if b1 & 0x40 else ""}')
                        if b1 & 0x80: hi.add((bid, t))
                print(f'  track {t}{" (speaker)" if t == 0 else ""}: ' + ' '.join(evs))
    print("b1 bit 0x80 in (block, track):", sorted(hi))
    print('event counts:', {(hex(k) if isinstance(k, int) else k): v for k, v in seen.items()})


def zounds(gog):
    z = dax(gog + '/ZOUNDS.DAX')
    for bid in sorted(z):
        b = bytes(c ^ 0x80 for c in z[bid])
        if b[:4] == b'FORM' and b[8:12] == b'8SVX':
            rate = be16(b, 32)
            body = b.index(b'BODY')
            n = struct.unpack_from('>I', b, body + 4)[0]
            print(f'ZOUNDS {bid} (digitized sound {bid + 1}): 8SVX ^0x80, {rate} Hz, {n} samples')
        else:
            print(f'ZOUNDS {bid} (digitized sound {bid + 1}): raw u8 PCM, {len(b)} bytes')


def test(gog):
    m = dax(gog + '/MUSIC.DAX')
    assert m[0] == open(gog + '/SOUND.COM', 'rb').read()   # block 0 is the driver
    assert dur(0) == 26880 and dur(2) == 6720 and dur(0x0A) == 10080 and dur(0x12) == 4480
    for bid in range(1, 9):
        log, end = play(m[bid])
        assert log and end > 0
        slog, send = play(m[bid], dev=SidDriver)
        assert send == end and all(0 <= r <= 0x18 for _, r, _ in slog)
        on = sum(r % 7 == 4 and r < 21 and v & 1 for _, r, v in slog)
        assert on == sum(1 for t, r, v in log if 0xB0 <= r <= 0xB8 and v & 0x20)  # one gate per key-on
    s = play(m[2], dev=SidDriver)[0]
    assert s[:3] == [(0, 4, 0), (0, 11, 0), (0, 18, 0)] and s[24:28] == [(0, 0x15, 7), (0, 0x16, 0xBF), (0, 0x17, 0), (0, 0x18, 0x1F)]
    assert SIDF[60 % 12] << 4 == 4384 and SidDriver(m[2]).inst[0][:7] == bytes.fromhex('00030103020000')
    print('ok')


if __name__ == '__main__':
    cmd, gog = sys.argv[1], sys.argv[2]
    if cmd == 'dump':
        dump(gog)
    elif cmd == 'zounds':
        zounds(gog)
    elif cmd == 'test':
        test(gog)
    elif cmd in ('regs', 'sidregs'):
        log, _ = play(dax(gog + '/MUSIC.DAX')[int(sys.argv[3])], dev=SidDriver if cmd == 'sidregs' else Driver)
        with open(sys.argv[4], 'w') as f:
            for t, r, v in log:
                f.write(f'{t} {r} {v}\n')
