"""skyrecolour.py SRC R,G,B OUT: SRC's luminance remapped onto a ramp
through one EGA colour (a third as bright, the colour, halfway to white),
so an existing pool texture takes the original's colour with no new art.
COTAB: ceiling.png -> ceiling_cotab_stone (black's roof, onto dark grey),
ceiling_sky_phlan.png -> ceiling_cotab_sky_{red,green,magenta,yellow}."""
import sys
from PIL import Image

src, rgb, out = sys.argv[1], tuple(int(v) for v in sys.argv[2].split(",")), sys.argv[3]
stops = [tuple(c // 3 for c in rgb), rgb, tuple((c + 255) // 2 for c in rgb)]
s = Image.open(src).convert("L")
px = list(s.get_flattened_data())
lo, hi = min(px), max(px)
ramp = []
for v in range(256):
    t = (v - lo) / max(hi - lo, 1) * (len(stops) - 1)
    i = max(0, min(len(stops) - 2, int(t)))
    f = min(max(t - i, 0), 1)
    a, b = stops[i], stops[i + 1]
    ramp.append(tuple(round(a[k] + (b[k] - a[k]) * f) for k in range(3)))
im = Image.new("RGB", s.size)
im.putdata([ramp[v] for v in px])
im.save(out)
