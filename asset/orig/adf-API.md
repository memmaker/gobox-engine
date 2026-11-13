# gobox/asset/orig (adf)
`Read(img []byte) (volume string, files []File, err error)`; `File{Path string; Data []byte}`.
Reads an 880K/1760K OFS/FFS .adf. No deps. Byte-identical copy of the three games' adf.go.

Switch (por, cotab, sotsb): in go.mod add
go.mod: `require gobox v0.0.0`, `replace gobox => ../gobox-engine` and `replace gobox/rules => ../gobox-engine/rules` (one module: every package below imports as `gobox/<dir>`).
rewrite imports `por/adf`, `cotab/adf`, `sotsb/adf` to `gobox/asset/orig`; no call changes; delete the game's adf/.
