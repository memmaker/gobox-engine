#!/bin/sh
# check_replaces.sh GAMEDIR...: fail if a game's go.mod lacks the replace lines gobox/gpu needs (see gpu/API.md).
st=0
for d in "$@"; do
	for r in 'github.com/gogpu/wgpu => ../gogpu-wgpu' 'github.com/go-webgpu/goffi => ../goffi' 'github.com/gogpu/gogpu => ../goplay/third_party/gogpu'; do
		grep -q "$r" "$d/go.mod" || { echo "$d/go.mod: missing: replace $r" >&2; st=1; }
	done
done
exit $st
