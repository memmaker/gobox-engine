# B1b2: what Buck Rogers' scripts do with memory (scan of 2026-11-15)

Made by an agent from the originals' byte code (`ecl.V4.Walk` over every ECL block: BR1 25 blocks,
BR2 33) and the BR engines' Go. Static only: no flow analysis, nothing run. REFACTOR.md B1b2.

- `inventory.csv`: every address the scripts use, per region, with widths, reads, writes, writers.
  BR1 396 addresses (area 251, party 39, aux 12, view 7, code tables 87), BR2 403, union 566.
- `q2q3.txt`: addresses used at more than one width (BR1 13, BR2 21) and overlapping cells
  (BR1 17 pairs, BR2 30); "live" = one area script writes one and reads the other (BR1 9, BR2 12).
- `q_text.txt`: text variables longer than the gap to the next used address (BR1 4, BR2 3).
- `q5.txt`: signedness (no compare against an immediate >= 32768; long cells can go negative in BR).
- `q6.txt`: tables. BR1 87 const tables, no SaveTable. BR2 2 SaveTable: a debug poke of any Area
  byte (area1) and a write into a code table (area114).
- `q7_br1.csv`, `q7_br2.csv`: the fixed addresses the BR engines' Go uses (BR1 54, 28 also in scripts).
- `layout-br1-draft.txt`: a full layout for BR1 in `ecl.Layout`'s format, 333 fields, parsed by
  `ecl.ParseLayout`; 37 `# CONFLICT` lines where one cell per address does not fit. Not wired in.

Where the conflicts are: the two scratch zones the original clears at each script load,
4C00..4C1F (Area[256..287]) and 7F79..7F82 (Party[889..898]), used as bytes, words and longs over
each other; and byte destinations that must keep 8 bits (InputNumber(3) into a byte, then a
compare; a byte decremented below 0 is 255). The engine's store keeps 16 bits and reads the cell
whole whatever the operand's type.
