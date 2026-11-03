# Font pool origins

All other `.f8` files: the 56 PC BIOS 8x8 fonts of The Ultimate Oldschool PC Font Pack v2.2 (see LICENSE.txt).

| File | Origin | sha256 |
|------|--------|--------|
| Gold_Box_8x8D1.f8 | The Gold Box games' own text font, block 201 of 8X8D1.DAX (POR, Curse of the Azure Bonds) and 8X8D5.DAX (Silver Blades). All three decode to the same bytes, so one file. Made by `cmd/fontexport`. Layout: ASCII 0x20..0x7F hold the 64 upper-case glyphs (a-z show the capitals), slots 0x80..0xF0 the 113 symbol tiles that follow them in the block (tile k at 0x80+k-64), the rest is blank. | fe84c0352f07e4f489dba359655fce3b850f756c156d7e641d5162b3c03ab686 |
| Dark_Queen_of_Krynn.f8 | por-go data/assets/fonts/ (copied unchanged) | a060513d261e91f72a46bcadfb7a5b228974788f17cc28ce99734ce620a77d93 |

Merge rule: before adding a font, `shasum -a 256 *.f8`; a repeated hash is dropped and its origin added here.
