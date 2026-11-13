#!/bin/bash
# Prints "game package -> forbidden import" for each package of each game
# module (tests included) that directly imports UI code. gobox/app is
# A cmd/ main is the app wiring: it may import any of it (the app shell,
# front, present, classic, playrules, walls, gpu, fonts).
here=$(cd "$(dirname "$0")/.." && pwd)
for d in "$here"/games/* "$here"/../buckrogers-go "$here"/../buckrogers2-go; do
  g=$(basename "$d")
  [ -d "$d" ] || continue
  (cd "$d" && go list -test -deps -f '{{if .Module}}{{if .Module.Main}}{{.ImportPath}}|{{join .Imports " "}}{{end}}{{end}}' ./... 2>/dev/null) |
  awk -F'|' -v g="$g" '
    { p=$1; sub(/ \[.*\]$/,"",p); sub(/\.test$/,"",p); n=split($2,im," ")
      for(i=1;i<=n;i++){ q=im[i]
        wiring = (q=="gobox/app" || q=="gobox/front/present" || q=="gobox/front/classic")
        if (p ~ /\/cmd\//) continue
        if (q ~ /^gobox\/(front|ui|gpu|blobber|combat|input|sprites|walls)(\/|$)/ || q ~ /ebiten|wgpu|glfw/ || wiring)
          print g, p, "->", q } }' | sort -u
done
