# gobox/adf
`Read(img []byte) (volume string, files []File, err error)`; `File{Path string; Data []byte}`.
Reads an 880K/1760K OFS/FFS .adf. No deps. Byte-identical copy of the three games' adf.go.

Switch (por, cotab, sotsb): in go.mod add
`require gobox/adf v0.0.0` and `replace gobox/adf => ../gobox-engine/adf`;
rewrite imports `por/adf`, `cotab/adf`, `sotsb/adf` to `gobox/adf`; no call changes; delete the game's adf/.
