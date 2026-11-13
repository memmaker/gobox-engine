#!/usr/bin/env bash
# pkggraph.sh: the import graph of this repo's own packages (no std lib, no
# third-party) as Graphviz dot, Mermaid and D2, plus index.html rendering them
# (pinch to zoom, scroll to pan), written to outdir and opened in the browser.
#   scripts/pkggraph.sh [outdir]   (default: a fresh /tmp/gobox-pkggraph.* dir)
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
out=${1:-$(mktemp -d /tmp/gobox-pkggraph.XXXXXX)}
mkdir -p "$out"

edges=$(mktemp)
trap 'rm -f "$edges"' EXIT
# the root module and the nested ones; each package's own (non-test) imports
for mod in $(cd "$root" && find . -name go.mod -not -path './.claude/*' -not -path '*/testdata/*' -exec dirname {} \;); do
	(cd "$root/$mod" && go list -e -f '{{.ImportPath}}{{range .Imports}} {{.}}{{end}}' ./...)
done | awk '{ for (i = 2; i <= NF; i++) if ($i ~ /^gobox(\/|$)/) print $1, $i; if (NF == 1 || $1 ~ /^gobox/) print $1 }' |
	grep '^gobox' | sort -u >"$edges"

# node: every package seen; edge: lines with two fields
nodes() { tr ' ' '\n' <"$edges" | sort -u; }
pairs() { awk 'NF == 2' "$edges"; }
id() { sed 's#^gobox/\{0,1\}##; s#^$#gobox#; s#[/.-]#_#g'; }

{
	echo 'digraph gobox {'
	echo '  rankdir=LR; node [shape=box, fontname="Helvetica", fontsize=10];'
	pairs | awk '{ printf "  \"%s\" -> \"%s\";\n", $1, $2 }'
	echo '}'
} >"$out/packages.dot"

{
	echo 'flowchart LR'
	nodes | while read -r n; do echo "  $(echo "$n" | id)[\"$n\"]"; done
	pairs | while read -r a b; do echo "  $(echo "$a" | id) --> $(echo "$b" | id)"; done
} >"$out/packages.mmd"

{
	echo 'direction: right'
	nodes | while read -r n; do echo "$(echo "$n" | id): \"$n\""; done
	pairs | while read -r a b; do echo "$(echo "$a" | id) -> $(echo "$b" | id)"; done
} >"$out/packages.d2"

# d2 has no browser renderer: an SVG only when the d2 CLI is installed
d2svg='<pre>install d2 (brew install d2) to render this; source in packages.d2</pre>'
if command -v d2 >/dev/null; then d2 --layout elk "$out/packages.d2" "$out/packages-d2.svg" && d2svg=$(cat "$out/packages-d2.svg"); fi
esc() { sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g' "$1"; }

cat >"$out/index.html" <<HTML
<!doctype html><meta charset="utf-8"><title>gobox packages</title>
<style>
body{margin:0;font:14px system-ui;background:#fafafa}
h2{margin:12px 16px 4px;font-size:16px}
.view{height:80vh;margin:0 16px 16px;overflow:auto;border:1px solid #ccc;background:#fff}
.view svg{display:block}
.src{display:none}
</style>
<p style="margin:12px 16px">$(nodes | wc -l | tr -d ' ') packages. Pinch to zoom, two-finger scroll to pan.</p>
<h2>Graphviz (dot)</h2><div class="view" id="dot"></div>
<h2>Mermaid</h2><div class="view" id="mmd"></div>
<h2>D2</h2><div class="view" id="d2">$d2svg</div>
<pre class="src" id="dot-src">$(esc "$out/packages.dot")</pre>
<pre class="src" id="mmd-src">$(esc "$out/packages.mmd")</pre>
<script src="https://cdn.jsdelivr.net/npm/@viz-js/viz@3.2.4/lib/viz-standalone.js"></script>
<script src="https://cdn.jsdelivr.net/npm/mermaid@10.9.1/dist/mermaid.min.js"></script>
<script>
const src = id => document.getElementById(id + '-src').textContent;
// pinch (a ctrl+wheel on macOS) scales the svg; plain scroll pans the box
function zoomable(view) {
  const svg = view.querySelector('svg'); if (!svg) return;
  const box = svg.getBBox ? svg.getBoundingClientRect() : null;
  const w = box.width, h = box.height; let k = 1;
  svg.style.maxWidth = 'none';
  view.addEventListener('wheel', e => {
    if (!e.ctrlKey) return;
    e.preventDefault();
    const r = view.getBoundingClientRect(), x = e.clientX - r.left, y = e.clientY - r.top;
    const k2 = Math.min(8, Math.max(0.1, k * Math.exp(-e.deltaY / 100))), f = k2 / k;
    k = k2; svg.setAttribute('width', w * k); svg.setAttribute('height', h * k);
    view.scrollLeft = (view.scrollLeft + x) * f - x; view.scrollTop = (view.scrollTop + y) * f - y;
  }, {passive: false});
}
Viz.instance().then(viz => {
  const v = document.getElementById('dot'); v.appendChild(viz.renderSVGElement(src('dot'))); zoomable(v);
});
mermaid.initialize({startOnLoad: false, maxEdges: 10000, maxTextSize: 1e7});
mermaid.render('mmd-svg', src('mmd')).then(({svg}) => {
  const v = document.getElementById('mmd'); v.innerHTML = svg; zoomable(v);
});
zoomable(document.getElementById('d2'));
</script>
HTML

echo "$(nodes | wc -l | tr -d ' ') packages, $(pairs | wc -l | tr -d ' ') imports"
echo "$out/index.html"
open "$out/index.html"
