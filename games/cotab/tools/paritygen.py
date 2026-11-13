#!/usr/bin/env python3
"""Regenerate PARITY_CHECKLIST.md from engine doc comments.

A marker is a doc comment `// name = SEG:OFF[, SEG:OFF...] ...`; one Go
function may cover several originals. Ticks and notes of the existing
checklist are kept, matched by SEG:OFF and name (then by SEG:OFF alone).
Usage: python3 tools/paritygen.py  (from the repo root)
"""
import glob
import re

MARK = re.compile(r'^// (\w+) = ((?:[OR][0-9A-F]{3}:[0-9A-F]{4}(?:, )?)+)')
LINE = re.compile(r'^- \[([ x])\] (\S+) `(\w+)` \((\S+)\)(.*)$')
PATH = 'PARITY_CHECKLIST.md'

old = open(PATH).read().split('\n')
head = []
for l in old:
    if l.startswith('## '):
        break
    head.append(l)
byname, byoff = {}, {}
for l in old:
    m = LINE.match(l)
    if m:
        byname[(m[2], m[3])] = (m[1], m[5])
        byoff.setdefault(m[2], (m[1], m[5]))

rows = set()
for f in sorted(glob.glob('rules/*.go')):
    if f.endswith('_test.go'):
        continue
    for l in open(f):
        m = MARK.match(l)
        if m:
            for off in re.findall(r'[OR][0-9A-F]{3}:[0-9A-F]{4}', m[2]):
                rows.add((off, m[1], f.split('/')[-1]))

out, seg = list(head), None
for off, name, f in sorted(rows):
    if off[:4] != seg:
        if seg:
            out.append('')
        seg = off[:4]
        out.append('## ' + seg)
    tick, note = byname.get((off, name)) or byoff.get(off) or (' ', '')
    out.append(f'- [{tick}] {off} `{name}` ({f}){note}')
open(PATH, 'w').write('\n'.join(out) + '\n')
