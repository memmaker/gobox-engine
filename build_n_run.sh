#!/bin/sh
# Build gobox and run a game: ./build_n_run.sh [game] [flags...]   (game: por, cotab, ...; default por)
set -e
cd "$(dirname "$0")"
# debug build: no optimisation or inlining (-N -l), so dlv sees every variable; attach with dlv attach $(pgrep -n gobox)
go build -gcflags=all="-N -l" -o bin/gobox ./cmd/gobox
game=${1:-por}
[ $# -gt 0 ] && shift
exec bin/gobox "$@" "$game"
