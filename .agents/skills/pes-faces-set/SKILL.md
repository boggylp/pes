---
name: pes-faces-set
description: 'Set, map, audit, or atomically apply one or more PES 2021 faces in a live Football Life install.'
---

# Set PES faces

> The whole is greater than the sum of its parts. — Aristotle

An audit stops after step 3.

## 1. Resolve the live state

- Resolve the writable face root from the active `cpk.root` entries.
- Stop before a write while Football Life or Sider runs.
- Build the active roster from the database patch's player-ID list, else with `pes-roster-extract`, else from `FL26_players.txt`.

## 2. Stage the faces

- List every archive before extraction, and reject traversal paths, unexplained numeric folders, or a missing `#Win/face.fpk`.
- Extract the whole request into one scratch `real` directory.
- Match each numeric folder to exactly one player in the active roster.
- Map a pack built for another roster:

  ```sh
  ./faces map --source-csv <source-roster.csv> --destination-csv <active-roster.csv> \
    --source-folder <source-faces> --dest-folder <staged-real>
  ```

- Relink one folder with `faces relink --folder <face-folder> --id <new-id>`.
  - A textures-only alias folder belongs to its mapped face.

## 3. Validate the staged set

Continue only when `./faces detect --faces-dir <staged-real> --player-csv <active-roster.csv>` reports nothing.

## 4. Install atomically

1. List every destination folder the install replaces, back each one up, and record its restore path.
2. Capture a full-root `faces detect` baseline.
3. Stage incoming folders beside the active root, and keep replaced folders outside the scanned `real` directory.
4. Activate the whole set, and rerun full-root detection.
5. Roll back every change when the findings differ from the baseline or a requested folder is absent.

## 5. Preserve the sources

- Store the archives under `%USERPROFILE%\MEGA\gaming\pes\faces\` with their SHA-256, until in-game verification passes.
- Write `_INSTALLED_<YYYY-MM-DD>.txt` beside them: player names and IDs, installed and skipped folders, archive hashes, backups, validation results, and rollback steps.
- Delete this run's scratch.
