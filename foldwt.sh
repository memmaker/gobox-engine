#!/bin/bash
# Per-track worktrees of the fold repos, replacing the shared gate lock for day-to-day work.
#   foldwt.sh new <track>   ~/Projects/fold-wt/<track>/<repo> on branch fold/<track> from main, deps/-re/data linked
#   foldwt.sh land <track>  rebase fold/<track> onto main, gate the branch's packages, fast-forward main
# A track commits freely in its own tree (gate: GATE_PROJECTS=~/Projects/fold-wt/<track> gate.sh).
# Land gates before taking the lock; under it only a main that moved meanwhile is gated again.
set -euo pipefail
projects="$HOME/Projects"
repos=(por-go rules1e rulesxxvc gobox-engine cotab-go sotsb-go pod-go buckrogers-core buckrogers-go buckrogers2-go)
cmd="${1:?new|land}"; track="${2:?track}"
tree="$projects/fold-wt/$track"
case "$cmd" in
new)
  mkdir -p "$tree"
  for r in "${repos[@]}"; do
    [ -e "$tree/$r" ] || git -C "$projects/$r" worktree add -b "fold/$track" "$tree/$r" main
    if [ -d "$projects/$r/data" ]; then # unversioned data entries, as gate.sh links them
      mkdir -p "$tree/$r/data"
      for d in "$projects/$r"/data/*; do [ -e "$tree/$r/data/$(basename "$d")" ] || ln -s "$d" "$tree/$r/data/"; done
    fi
  done
  for d in gogpu-wgpu goffi goplay gameimport gogpu-naga "$(cd "$projects" && ls -d *-re)"; do
    for x in $d; do [ -e "$tree/$x" ] || ln -s "$projects/$x" "$tree/$x"; done
  done
  echo "$tree" ;;
land)
  # Gate outside the lock; under it, only when main moved meanwhile (rebase and gate again).
  for r in "${repos[@]}"; do
    [ -z "$(git -C "$tree/$r" status --porcelain --untracked-files=no)" ] || { echo "$r: uncommitted changes in $tree/$r"; exit 1; }
    git -C "$tree/$r" rebase main
  done
  gated=(); for r in "${repos[@]}"; do gated+=("$(git -C "$projects/$r" rev-parse main)"); done
  GATE_ALL=1 GATE_BASE=main GATE_PROJECTS="$tree" "$projects/gobox-engine/gate.sh"
  "$projects/gobox-engine/gatelock.sh" take "land-$track"
  trap '"$projects/gobox-engine/gatelock.sh" drop "land-$track"' EXIT
  moved=""; i=0
  for r in "${repos[@]}"; do
    [ "$(git -C "$projects/$r" rev-parse main)" = "${gated[i]}" ] || moved=1
    i=$((i + 1))
  done
  if [ -n "$moved" ]; then
    echo "main moved since the gate: rebase and gate again"
    for r in "${repos[@]}"; do git -C "$tree/$r" rebase main; done
    GATE_ALL=1 GATE_BASE=main GATE_PROJECTS="$tree" "$projects/gobox-engine/gate.sh"
  fi
  # All or nothing: a repo that cannot fast-forward stops the land before any main moves.
  for r in "${repos[@]}"; do
    git -C "$projects/$r" merge-base --is-ancestor main "fold/$track" || { echo "$r: main is not an ancestor of fold/$track"; exit 1; }
  done
  for r in "${repos[@]}"; do
    [ "$(git -C "$projects/$r" rev-parse main)" = "$(git -C "$tree/$r" rev-parse HEAD)" ] && continue
    git -C "$projects/$r" merge --ff-only "fold/$track"
  done
  echo "landed $track" ;;
esac
