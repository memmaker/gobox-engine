#!/bin/bash
# The agents' commit lock: gatelock.sh take <name> | drop <name>.
# take waits until the lock is free; drop removes it only if <name> holds it (never another agent's).
set -euo pipefail
lock=/private/tmp/por-gate.lock
case "${1:-}" in
take)
  until mkdir "$lock" 2>/dev/null; do sleep 5; done
  echo "$2" > "$lock/owner" ;;
drop)
  [ "$(cat "$lock/owner" 2>/dev/null)" = "$2" ] || { echo "gatelock: not held by $2" >&2; exit 1; }
  rm "$lock/owner"; rmdir "$lock" ;;
*) echo "usage: gatelock.sh take|drop <name>" >&2; exit 2 ;;
esac
