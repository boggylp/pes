---
name: pes-faces-install
description: '**Invoke this skill BEFORE installing, mapping, or detecting player faces for the live Football Life 2026 install.** Covers the `faces/` Go tool (`detect` for orphan / ID-mismatch scans, `map` for cross-version remapping, `relink` for length-changing FPK repacks), deriving the destination roster from the AUTHORITATIVE live DB (the install''s `UML 2026 - Player IDs.csv` or the livecpk `UML_Database` `Player.bin`, never the stale base `FL26_players.txt` which UML renumbers), the destination `SiderAddons\livecpk\...\face\real\<player_id>\`, the exact-then-relaxed name matcher, length-mismatch handling via `relink`, the FPK-embedded-ID-must-match rule (incl. created players), and the mandatory rollback record. Triggers: "install these faces", "map BPB faces to FL26", "remap player IDs", "find orphan faces", "fix face mismatches".'
---

# PES face install

Faces are the easiest mod to install wrong: silent drops on length mismatch, easy orphans, no SYSTEM cache step needed. Rollback record makes mistakes recoverable.

## When to use

- User has a new face pack archive to install
- Mapping faces from another game version (BPB -> FL26) by player name
- Auditing the live livecpk for orphans / mismatches

## Steps

1. **Build the destination CSV from the AUTHORITATIVE live roster, not `FL26_players.txt`.** Set `FL26` to the live install root first (`FL26=...`); never hardcode it. Critical: when UML (or any DB patch) is installed it replaces `Player.bin` via livecpk and **renumbers some players**, so `FL26_players.txt` (base export, often stale by months) yields wrong IDs and faces silently miss or clobber the wrong player. Pick the roster in this order:

    - **Provided ID list in the install** (check first): `UML 2026 - Player IDs.csv` (`Id;Name;...;Club`) and `... - Team IDs.csv` at the game root.
    - **The live `UML_Database` `Player.bin`** (full names + live IDs), via `pesdb roster --player-bin "$FL26/SiderAddons/livecpk/UML_Database/common/etc/pesdb/Player.bin" --out /tmp/uml-roster.csv`.
    - Only if no DB patch is installed: `FL26_players.txt` (`tr -d '\r' < "$FL26/FL26_players.txt" | sed '1iId;Name' | sed 's/ - /;/'`).

    Sanity-check a known renumbered player (Beljo: base `91287` vs UML `58035`) before trusting the roster. Run fresh every install; never use the `faces/samples/` snapshots.

2. **Detect first** when auditing or after any install:

    ```sh
    cd faces
    go build .
    ./faces detect \
      --faces-dir "$FL26/SiderAddons/livecpk/root/Asset/model/character/face/real" \
      --player-csv /tmp/fl26-players.csv
    ```

    Reports: orphan face folders (IDs not in the player CSV), ID mismatches (folder name vs embedded FPK ID), non-numeric folders, missing FPKs.

3. **Map between versions** when porting a pack:

    ```sh
    ./faces map \
      --source-csv samples/BPB-2023-players.csv \
      --destination-csv /tmp/fl26-players.csv \
      --source-folder /path/to/source/faces \
      --dest-folder "$FL26/SiderAddons/livecpk/root/Asset/model/character/face/real"
    ```

    Matches by normalized name; copies the face folder, renames it to the destination ID, and rewrites the embedded ID inside `face.fpk` via `strings.ReplaceAll`.

4. **Handle length mismatches.** Equal-length ID pairs are hex-swapped in place by `map`. Different-length pairs (e.g. `72284` -> `177929`, or a created-player 10-digit ID like `2147483648`) need the FPK repacked, since a naive byte replace corrupts the FPK/FMDL offsets. Use `faces relink <faceFolder> <newID>` (repacks the FPK). `map` routes length-mismatches through it; standalone for one-offs like a created player. Verify relinked faces in-game (the repack is structurally validated, not render-verified).

5. **Salvage near-misses** the matcher can't catch:

    - **Direct ID match.** If the source folder ID literally exists in the destination CSV, copy the folder as-is to `dest-folder/<id>/`. No remap needed.
    - **Relaxed name match** is now built in: `map` matches exact normalized name first, then a first+last fallback (surname exact + compatible first name, middle names ignored, ambiguous collisions rejected), so live `Dion Beljo` matches source `Dion Drena Beljo` automatically.

6. **Extract directly to the final livecpk path.** No temp dirs, no intermediate copies, no renames. One step, final destination.

7. **Write a rollback record** next to the source archives:

    ```text
    %USERPROFILE%\MEGA\gaming\pes\faces\_INSTALLED_<YYYY-MM-DD>.txt
    ```

    List every destination folder ID created and, for each not-installed entry, the reason (length mismatch, no name match, …). The user cleans up by deleting the listed numeric folders.

## Pitfalls

- **Length mismatches are silent.** Always cross-check `map`'s reported installed count against the source folder count. The difference is the silent-drop set.
- **Live DB is authoritative.** `faces/samples/FL26_players.csv` is a snapshot and lags the live install by weeks; do not use it for an install.
- **Never `rm` source archives** before the install is verified in-game.
- **SYSTEM cache delete is not warranted for faces.** Faces load via livecpk; the `SYSTEM00000000` cache is not the gating mechanism. Don't propose the delete for a face install. (Propose-and-ask is the universal SYSTEM-cache rule; here the answer should be don't propose it.)
- **Sider is already configured** (`cpk.root = .\livecpk\root`, `livecpk.enabled = 1`). Dropping ID folders into the face root is enough; do not edit `sider.ini` for a face install.
- **Hostname matters for any logged finding.** If you record a face install outcome in the AIKB or diary, tag the machine.

## Out of scope

- Roster / player ID extraction from cpk + Player.bin: see pes-roster-extract.
- Kit textures: see pes-kit-ftex.
- Internal FPK/FMDL structure details for the length-changing repack: see `faces/relink.go`.
