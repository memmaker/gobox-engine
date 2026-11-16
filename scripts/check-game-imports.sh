#!/bin/bash
# Prints "game package -> forbidden import" for each package of each game
# (tests included) that directly imports UI code (gobox/ui, which holds the
# host window, or gobox/gpu). A cmd/ main is the app wiring: it may import any of it.
# The games are packages of the gobox module (games/<game>).
here=$(cd "$(dirname "$0")/.." && pwd)
for d in "$here"/games/*; do
  g=$(basename "$d")
  [ -d "$d" ] || continue
  prefix="gobox/games/$g"
  (cd "$d" && go list -test -deps -f '{{if .Module}}{{if .Module.Main}}{{.ImportPath}}|{{join .Imports " "}}{{end}}{{end}}' ./... 2>/dev/null) |
  awk -F'|' -v g="$g" -v prefix="$prefix" '
    { p=$1; sub(/ \[.*\]$/,"",p); sub(/\.test$/,"",p); n=split($2,im," ")
      if (prefix != "" && p != prefix && index(p, prefix "/") != 1) next
      for(i=1;i<=n;i++){ q=im[i]
        if (p ~ /\/cmd\//) continue
        if (q ~ /^gobox\/(ui|gpu)(\/|$)/ || q ~ /ebiten|wgpu|glfw/)
          print g, p, "->", q } }' | sort -u
done
