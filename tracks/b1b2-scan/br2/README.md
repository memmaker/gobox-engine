# BR2's script scan (L2, 2026-11-16)

The BR2 part of the B1b2 scan (README one folder up), cut out per file; no new scan was run: the scan already covered both games (`ecl.V4.Walk` over BR2's 33 ECL blocks).
Re-proved by `importer/br2_layout_test.go` (every operand of every block resolves in `importer/br2/scripts/layout/global.txt`).

- `inventory.csv`: 403 addresses (area 237, party 35, aux 22, view 5, code tables 104), widths, reads, writes, writers.
- `q2q3.txt`: addresses used at several widths (21) and overlapping cells (30). Resolved in the layout by the two byte buffers EntryBytes (4C00..4C1F)
  and NewGameBytes (7F79..7F82), by the text length rule (a text field is as long as the gap to the next used address), and by the engine tolerances the proof lists
  (byte operands on the word SelHpMax/SelHpCur; byte operands on the text words 7A00, 7B04 and the name 7C00).
- `q_text.txt`: the text variables and their lengths. `q6.txt`: the tables: 104 const tables (`ecl.ConstLayout` makes their fields), 2 SaveTable (area1: a debug poke at 4B00 + index; area114: a write into a code table).
