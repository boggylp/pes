---
name: pes-faces-set
description: 'Set, map, audit, or atomically apply one or more PES 2021 faces in a live Football Life install.'
---

# Set PES faces

> The whole is greater than the sum of its parts. — Aristotle

Use the active database, preserve source archives, and leave a tested rollback path.

## 1. Resolve the live state

- Use the user-provided live-install root.
- Read its `AGENTS.md` and `README.md` when present.
- Resolve the writable custom face root from active `cpk.root` entries.
- Stop before a write when Football Life or Sider is running.
- Treat an audit as read-only.

Build the roster again for each install:

1. Use the player-ID list supplied with the active database patch.
2. Otherwise run `pesdb roster` against the active livecpk `Player.bin`.
3. Use `FL26_players.txt` only when no database patch overrides it.

Never use `faces/samples/` for a live install.

## 2. Stage the faces

- List every supplied archive before extraction.
- Reject traversal paths, unexplained numeric folders, or missing `#Win/face.fpk` data.
- Extract the complete request into one scratch `real` directory.
- Confirm each numeric face folder against exactly one player in the active roster.

Map a source pack when its roster differs:

```sh
./faces map \
  --source-csv <source-roster.csv> \
  --destination-csv <active-roster.csv> \
  --source-folder <source-faces> \
  --dest-folder <staged-real>
```

Use `faces relink --folder <face-folder> --id <new-id>` for one folder.

- Equal-length IDs may change the embedded decimal ID in `#Win/*.fpk` and `*.fpkd`.
- Different-length IDs keep package bytes unchanged and mirror `sourceimages` under the old ID.
- Never byte-rewrite a different-length ID because it shifts the FMDL string table.

## 3. Validate the staged set

```sh
./faces detect --faces-dir <staged-real> --player-csv <active-roster.csv>
```

Continue only when the staged set has no findings.

## 4. Plan replacements

- Inspect every active destination before writing.
- Require approval in the current turn for the complete live-install change.
- Ask once with every existing destination when replacement approval is missing.
- Back up each approved existing destination and record its restore path.

## 5. Install atomically

- Capture a full-root `faces detect` baseline.
- Stage incoming folders beside the active root.
- Keep replaced folders outside the scanned `real` directory.
- Activate the complete requested set.
- Rerun full-root detection.
- Roll back every change when findings differ or a requested folder is absent.
- Allow unrelated pre-existing findings only when before and after are identical.

## 6. Preserve the sources

Store archives under `%USERPROFILE%\MEGA\gaming\pes\faces\` unless the user names another location.
Verify each preserved archive with SHA-256. Write `_INSTALLED_<YYYY-MM-DD>.txt` beside the archives with:

- player names and active IDs;
- installed and skipped folders;
- archive hashes;
- replacement backups;
- validation results;
- exact rollback steps.

## Review

- Treat a textures-only alias folder as part of its mapped face.
- Keep existing Sider configuration unchanged when the configured face root is active.
- Preserve source archives until in-game verification passes.
- Remove only scratch artifacts created by this run.
- Use `pes-wiki-log` for a durable machine-specific result.

## Boundaries

- Use `pes-roster-extract` to generate a roster from game data.
- Use `pes-kit-ftex` for kit textures.
- Read `faces/relink.go` and `faces/fpk.go` for package-format details.
