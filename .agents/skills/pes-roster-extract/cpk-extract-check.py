#!/usr/bin/env python3
"""Diagnose PES/FL cpk extraction integrity (offset / ContentOffset bugs).

Two modes:
  cpk-extract-check.py header  <cpk>            # dump ContentOffset/TocOffset + offset scheme
  cpk-extract-check.py face    <face.fpk>...    # verify an extracted face.fpk is valid

Background: the repo `cpk` tool has a rebase heuristic (reader.go, "Some PES
season packs ... FileOffsets already absolute") that decides whether to add
ContentOffset to each TOC FileOffset. It is RIGHT for dt-pack `Player.bin`
(dt10) but WRONG for relative-offset face packs like `download/fa26_bpb.cpk`:
it reads every entry one Align block (2048 B) too early, yielding a 2 KB garbage
prefix + a tail truncated by 2048 B. The game wants `foxfpk` at byte 0; a
corrupt face has `foxfpk` at offset 2048 -> crash on load.

Ground truth (verified 2026-06-07, BogambeDesktop): in BOTH dt10 and fa26_bpb
the real payload sits at rawFileOffset + ContentOffset (neither should be
rebased). dt10 Player.bin WESYS magic is at raw+CO; fa26_bpb 17003 foxfpk is at
raw+CO. A correct standalone extraction reads each entry at
rawFileOffset + ContentOffset for FileSize bytes.
"""
import struct, sys, os

def _utf(buf):
    assert buf[:4] == b'@UTF', buf[:4]
    t = 8
    rowsOff, strOff, datOff, _ = struct.unpack_from('>IIII', buf, t)
    numCols, rowWidth, numRows = struct.unpack_from('>HHI', buf, t + 0x10)
    strings = buf[t + strOff: t + datOff]
    def s(o):
        e = strings.find(b'\x00', o); return strings[o:e].decode('utf-8', 'replace')
    def rv(p, typ):
        if typ in (0, 1): return buf[p], p + 1
        if typ in (2, 3): return struct.unpack_from('>H', buf, p)[0], p + 2
        if typ in (4, 5): return struct.unpack_from('>I', buf, p)[0], p + 4
        if typ in (6, 7): return struct.unpack_from('>Q', buf, p)[0], p + 8
        if typ == 8: return struct.unpack_from('>f', buf, p)[0], p + 4
        if typ == 0xA: return s(struct.unpack_from('>I', buf, p)[0]), p + 4
        if typ == 0xB: return struct.unpack_from('>II', buf, p), p + 8
        raise ValueError(typ)
    cols, p = [], t + 0x18
    for _ in range(numCols):
        flags = buf[p]; p += 1
        nameO = struct.unpack_from('>I', buf, p)[0]; p += 4
        const = None
        if flags & 0xf0 == 0x30: const, p = rv(p, flags & 0x0f)
        cols.append((s(nameO), flags, const))
    rows, rp = [], t + rowsOff
    for _ in range(numRows):
        cp, row = rp, {}
        for (name, flags, const) in cols:
            st = flags & 0xf0
            if st == 0x30: row[name] = const
            elif st == 0x10: row[name] = 0
            else: row[name], cp = rv(cp, flags & 0x0f)
        rows.append(row); rp += rowWidth
    return rows

def header(fp):
    f = open(fp, 'rb'); f.read(16)  # "CPK " container
    f.seek(0); tsz = struct.unpack_from('<I', f.read(16), 8)[0]
    h = _utf(f.read(tsz))[0]
    co, to = h['ContentOffset'], h['TocOffset']
    f.seek(to); ttsz = struct.unpack_from('<I', f.read(16), 8)[0]
    toc = _utf(f.read(ttsz))
    offs = [r['FileOffset'] for r in toc]
    scheme = 'relative (add ContentOffset; do NOT rebase)' if min(offs) == 0 else 'absolute'
    print(f"{os.path.basename(fp)}: ContentOffset={co} TocOffset={to} rows={len(toc)} "
          f"minFileOffset={min(offs)} -> {scheme}")

def face(fp):
    d = open(fp, 'rb').read()
    off = d.find(b'foxfpk')
    if off == 0:
        size = struct.unpack_from('<I', d, 0x0a)[0]
        ok = 'OK' if size == len(d) else f'TRUNCATED (hdr says {size}, file {len(d)})'
        print(f"{os.path.basename(fp)}: foxfpk@0  {ok}")
    elif off > 0:
        print(f"{os.path.basename(fp)}: CORRUPT - foxfpk@{off} (extraction read {off}B too early; re-extract at rawFileOffset+ContentOffset)")
    else:
        print(f"{os.path.basename(fp)}: no foxfpk magic (header {d[:6].hex(' ')})")

if __name__ == '__main__':
    mode = sys.argv[1] if len(sys.argv) > 1 else ''
    for p in sys.argv[2:]:
        (header if mode == 'header' else face)(p)
