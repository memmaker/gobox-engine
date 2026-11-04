#!/bin/bash
# Prints "game package -> forbidden import" for each package of each game
# module (tests included) that directly imports UI code. gobox/app is
# allowed only from cmd/ mains.
for g in por-go cotab-go sotsb-go pod-go buckrogers-go; do
  d=$HOME/Projects/$g
  [ -d "$d" ] || continue
  (cd "$d" && go list -test -deps -f '{{if .Module}}{{if .Module.Main}}{{.ImportPath}}|{{join .Imports " "}}{{end}}{{end}}' ./... 2>/dev/null) |
  awk -F'|' -v g="$g" '
    { p=$1; sub(/ \[.*\]$/,"",p); sub(/\.test$/,"",p); n=split($2,im," ")
      for(i=1;i<=n;i++){ q=im[i]
        if (q ~ /^gobox\/(front|ui|gpu|blobber|combat|input|sprites|walls)(\/|$)/ || q ~ /ebiten|wgpu|glfw/ || (q=="gobox/app" && p !~ /\/cmd\//))
          print g, p, "->", q } }' | sort -u
done
