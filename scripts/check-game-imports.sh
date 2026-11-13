#!/bin/bash
# Prints "game package -> forbidden import" for each package of each game
# (tests included) that directly imports UI code. gobox/app is
# A cmd/ main is the app wiring: it may import any of it (the app shell, ui, gpu).
# The games are packages of the gobox module (games/<game>); the buckrogers
# games are modules of their own, beside this repo.
here=$(cd "$(dirname "$0")/.." && pwd)
for d in "$here"/games/* "$here"/../buckrogers-go "$here"/../buckrogers2-go; do
  g=$(basename "$d")
  [ -d "$d" ] || continue
  case "$d" in
    "$here"/games/*) prefix="gobox/games/$g" ;;
    *) prefix="" ;;
  esac
  (cd "$d" && go list -test -deps -f '{{if .Module}}{{if .Module.Main}}{{.ImportPath}}|{{join .Imports " "}}{{end}}{{end}}' ./... 2>/dev/null) |
  awk -F'|' -v g="$g" -v prefix="$prefix" '
    { p=$1; sub(/ \[.*\]$/,"",p); sub(/\.test$/,"",p); n=split($2,im," ")
      if (prefix != "" && p != prefix && index(p, prefix "/") != 1) next
      for(i=1;i<=n;i++){ q=im[i]
        wiring = (q=="gobox/app")
        if (p ~ /\/cmd\//) continue
        if (q ~ /^gobox\/(ui|gpu)(\/|$)/ || q ~ /ebiten|wgpu|glfw/ || wiring)
          print g, p, "->", q } }' | sort -u
done
