#!/usr/bin/env bash
# pkggraph.sh: the import graph of this repo's own packages (no std lib, no
# third-party) as Graphviz dot, Mermaid and D2, written to the output dir.
#   scripts/pkggraph.sh [outdir]   (default: current dir)
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
out=${1:-.}
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

echo "$(nodes | wc -l | tr -d ' ') packages, $(pairs | wc -l | tr -d ' ') imports -> $out/packages.{dot,mmd,d2}"
