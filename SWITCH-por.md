# Switch por-go to gobox modules

Run in /Users/felix/Projects/por-go when no other session is editing.

1. go.mod: remove `gpu v0.0.0` and `replace gpu => ./gpu`; add
   require gobox/gpu v0.0.0, gobox/input v0.0.0, gobox/blobber v0.0.0
   replace gobox/gpu => ../gobox-engine/gpu
   replace gobox/input => ../gobox-engine/input
   replace gobox/blobber => ../gobox-engine/blobber
2. Imports (all .go, .md, .sh, .py, .json outside gobox-engine):
   grep -rlE -I '"gpu(/vector)?"|por/present/input|por/phlan-poc/blobber' . | grep -v '\.git/' |
   xargs sed -i '' 's#"gpu"#"gobox/gpu"#g; s#"gpu/vector"#"gobox/gpu/vector"#g; s#por/present/input#gobox/input#g; s#por/phlan-poc/blobber#gobox/blobber#g'
3. Fix remaining path mentions: grep -rn 'phlan-poc/blobber\|present/input\|\./gpu' (scripts, docs, comments).
4. git rm -r gpu present/input phlan-poc/blobber (also rm -rf ignored leftovers: phlan-poc/blobber/{art,cmd,.claude,settings.json}).
5. go mod tidy; go build ./... && go test -short ./present/...
Note: blobber imports no por/... packages (no cycle).

## blobber API since 5ab7d33

- `blobber.NewView()` is now `blobber.NewView(nil)` (nil keeps the embedded PoR walls).
- The package global `blobber.Night` is gone: set `v.Night` on the View instead.
- The scratch `por-switch.patch` predates this; rebuild after applying and fix these call sites.

## Tier 1/2 switch caveats (agreed with the POR session)

- dax: por's `.Block(` → `.BlockShared(` everywhere; keep the `b == nil` "missing" checks
  (engine/pic.go:32, :74, anim.go:37, files.go:317) on BlockShared.
- geo: por keeps its own LoadWalldef (no ParseWalldef). LoadGeo's wrapper: a zero-length
  block is silent (no error); a longer block is passed as b[:0x402]; keep the block-10 quirk.
- ui pager: map engine.Pages.Page onto Pager's page so the clamp holds; speak after every key.
- After the switch: por-go `go test ./...` incl. the pages_test goldens must pass.
