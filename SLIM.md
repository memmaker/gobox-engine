# Slim-down checklist (ponytail audit, 2026-11-13)

**Goal.** Cut ~3.7k Go lines and 5 packages without behaviour change. **How.** One agent per track in its own git worktree; each commit builds (`go build ./... && go vet ./...`, targeted tests only); the lead merges tracks into main. Full test suite once at the end.

- [x] Stale worktree agent-a372c521a57fd0f1e removed (no unmerged commits)
- [ ] D: dead code (deadcode -test list, rules1e stats accessors, unused spell-ID consts, engine.lead/viewMember, xinput packet), tracked poolsprites binary + sprites/, empty blobber/
- [ ] M: package merges gpu/vector→gpu, game/play+game/journal→game, ecl/script→ecl, audio/sounddata→audio
- [x] T: dev dump cmds removed (keep importstale: gate.sh), fixture cmds folded into one
- [ ] L: per-game launchers → cmd/gobox `import` subcommand + manifest Amiga specs
- [ ] Docs: stale root docs folded into ARCHITECTURE.md
- [ ] Final: full build/vet/test on main
