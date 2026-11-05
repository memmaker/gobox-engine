"""Pixel variants of the 3D floors, ceilings and skies, made by PixelLab only.
  python3 make_pixel.py GAMEDIR [--force] [NAME...]
For each remaster.json "floor" entry with "pixel": path, sends the hi-res twin ("out") to PixelLab
/image-to-pixelart (the gobox pixel-wall recipe: init_image_strength 200, seed 7) at 256 wide, the twin's
aspect kept. Pixel variants are never derived from the twin by shrinking or colour quantizing."""
import json, subprocess, sys, tempfile
from pathlib import Path
from PIL import Image
import areas

CLI = Path.home() / ".claude/skills/pixellab/pixellab_cli.py"
CFG = areas.load(sys.argv[1])
F = CFG["floor"]
names = [a for a in sys.argv[2:] if not a.startswith("--")]
for F in (F if isinstance(F, list) else [F]):
    if "pixel" not in F or (names and F["raw"] not in names):
        continue
    out = areas.GAME / F["pixel"]
    if out.exists() and "--force" not in sys.argv:
        print(f"{out} exists; skipped (--force to overwrite)")
        continue
    twin = areas.GAME / F["out"]
    w, h = Image.open(twin).size
    with tempfile.TemporaryDirectory() as d:
        subprocess.run([sys.executable, CLI, "call", "POST", "/image-to-pixelart", f"image=@{twin}",
                        "image_size=" + json.dumps({"width": w, "height": h}),
                        "output_size=" + json.dumps({"width": 256, "height": h * 256 // w}),
                        "init_image_strength=200", "seed=7", "-o", d], check=True, stdout=subprocess.DEVNULL)
        png, = Path(d).glob("*.png")
        out.parent.mkdir(parents=True, exist_ok=True)
        Image.open(png).convert("RGBA").save(out)
    print(F["raw"], "ok")
