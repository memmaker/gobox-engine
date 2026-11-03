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
