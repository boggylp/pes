---
name: pes-faces-install
description: '**Invoke this skill BEFORE installing, mapping, or detecting player faces for the live Football Life 2026 install.** Covers the `faces/` Go tool (`detect` for orphan / ID-mismatch scans, `map` for cross-version remapping), the live player database at `C:\Program Files (x86)\SP Football Life 2026\FL2621_players.txt` and the CSV conversion formula, the destination at `C:\Program Files (x86)\SP Football Life 2026\SiderAddons\livecpk\root\Asset\model\character\face\real\<player_id>\`, the silently-dropped length-mismatch behavior in `map.go:81`, the salvage techniques for near-misses (direct-ID match, source folder rename), and the mandatory rollback record. Triggers: "install these faces", "map BPB faces to FL26", "remap player IDs", "find orphan faces", "fix face mismatches".'
---

# PES face install

Faces are the easiest mod to install wrong: silent drops on length mismatch, easy orphans, no SYSTEM cache step needed. Rollback record makes mistakes recoverable.

## When to use

- User has a new face pack archive to install
- Mapping faces from another game version (BPB -> FL26) by player name
- Auditing the live livecpk for orphans / mismatches

## Steps

1. **Convert the live player DB to the tool's CSV format.** The live file is `<ID> - <Name>` with CRLF; the tool expects semicolon CSV with header `Id;Name`:

    ```sh
    tr -d '\r' < "/c/Program Files (x86)/SP Football Life 2026/FL2621_players.txt" \
      | sed '1iId;Name' \
      | sed 's/ - /;/' > /tmp/fl26-players.csv
    ```

    Run this fresh every install; the live file is the authoritative source, not the snapshot in `faces/samples/`.

2. **Detect first** when auditing or after any install:

    ```sh
    cd faces
    go build .
    ./faces detect \
      --faces-dir "/c/Program Files (x86)/SP Football Life 2026/SiderAddons/livecpk/root/Asset/model/character/face/real" \
      --player-csv /tmp/fl26-players.csv
    ```

    Reports: orphan face folders (IDs not in the player CSV), ID mismatches (folder name vs embedded FPK ID), non-numeric folders, missing FPKs.

3. **Map between versions** when porting a pack:

    ```sh
    ./faces map \
      --source-csv samples/BPB-2023-players.csv \
      --destination-csv /tmp/fl26-players.csv \
      --source-folder /path/to/source/faces \
      --dest-folder "/c/Program Files (x86)/SP Football Life 2026/SiderAddons/livecpk/root/Asset/model/character/face/real"
    ```

    Matches by normalized name; copies the face folder, renames it to the destination ID, and rewrites the embedded ID inside `face.fpk` via `strings.ReplaceAll`.

4. **Handle length mismatches.** `map.go:81` silently drops any pair where source and destination IDs have different decimal-string lengths (e.g. `72284` -> `177929`). The hex replace would corrupt FPK length-prefixed offsets. These need a dedicated FPK-aware editor; do not force-install them. Log them in the rollback record as **not installed**.

5. **Salvage near-misses** the matcher can't catch:

    - **Direct ID match.** If the source folder ID literally exists in the destination CSV, copy the folder as-is to `dest-folder/<id>/`. No remap needed.
    - **Source folder rename.** When the live name has fewer parts than the source name (e.g. live `Dion Beljo` vs source `Dion Drena Beljo`), the matcher fails on `len(targetParts) != len(candidateParts)` at `normalize.go:82`. Rename the source folder to match the live name's part count and rerun. The embedded FPK ID still gets remapped correctly.

6. **Extract directly to the final livecpk path.** No temp dirs, no intermediate copies, no renames. One step, final destination. See [[feedback_extract_directly]].

7. **Write a rollback record** next to the source archives:

    ```text
    %USERPROFILE%\MEGA\gaming\pes\faces\_INSTALLED_<YYYY-MM-DD>.txt
    ```

    List every destination folder ID created and, for each not-installed entry, the reason (length mismatch, no name match, …). The user cleans up by deleting the listed numeric folders.

## Pitfalls

- **Length mismatches are silent.** Always cross-check `map`'s reported installed count against the source folder count. The difference is the silent-drop set.
- **Live DB is authoritative.** `faces/samples/FL26_players.csv` is a snapshot and lags the live install by weeks; do not use it for an install.
- **Never `rm` source archives** before the install is verified in-game. Backup first. See [[feedback_never_rm_user_data]].
- **No SYSTEM cache delete for faces.** Faces load via livecpk; the `SYSTEM00000000` cache is not the gating mechanism. Delete only for dt13 / dt18 / exe changes. See [[feedback_system_cache_scope]].
- **Sider is already configured** (`cpk.root = .\livecpk\root`, `livecpk.enabled = 1`). Dropping ID folders into the face root is enough; do not edit `sider.ini` for a face install.
- **Hostname matters for any logged finding.** If you record a face install outcome in the AIKB or diary, tag the machine. See [[feedback_pes_log_hostname]].

## Out of scope

- Roster / player ID extraction from cpk + Player.bin: see [[pes-roster-extract]].
- Kit textures: see [[pes-kit-ftex]].
- Modifying the embedded FPK structure for cross-length-ID remapping. The `map` tool is intentionally length-safe; that work needs a different tool.
