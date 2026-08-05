---
name: pes-faces-install
description: 'Install, map, or audit player faces for a live Football Life 2026 install.'
---

# PES face install

## Steps

1. Set `FL26` to the user-provided live-install root.

2. Build the destination CSV from the active database. Use this order:

   - Use the player-ID list supplied with the installed database patch.
   - Otherwise, run `pesdb roster` against the active livecpk `Player.bin`.
   - Use `FL26_players.txt` only when no database patch overrides it.

   Build the roster again for each install. Do not use `faces/samples/` for a live install.

3. Run detection before and after an install.

   ```sh
   cd faces
   go build .
   ./faces detect \
     --faces-dir "$FL26/SiderAddons/livecpk/root/Asset/model/character/face/real" \
     --player-csv /tmp/fl26-players.csv
   ```

4. Map a source pack to the active roster.

   ```sh
   ./faces map \
     --source-csv <source-roster.csv> \
     --destination-csv /tmp/fl26-players.csv \
     --source-folder <source-faces> \
     --dest-folder "$FL26/SiderAddons/livecpk/root/Asset/model/character/face/real"
   ```

   The mapper tries an exact normalized name and then a compatible first-and-last-name match. It rejects ambiguous matches.

5. Apply the ID mapping rule:

   - For equal-length IDs, the tool changes the embedded decimal ID in each `#Win/*.fpk` and `*.fpkd` package.
   - For different-length IDs, the tool keeps package bytes unchanged and mirrors `sourceimages` under the embedded old ID.

   Use `faces relink --folder <face-folder> --id <new-id>` for one folder. Do not byte-rewrite a different-length ID because it shifts the FMDL string table.

6. Install directly into the configured livecpk face root.

7. Reconcile installed, skipped, missing-source, and error counts. Run `faces detect` again.

8. Write a rollback record beside the source archives:

   ```text
   %USERPROFILE%\MEGA\gaming\pes\faces\_INSTALLED_<YYYY-MM-DD>.txt
   ```

   List each destination folder that the operation created. List each skipped source and its reason.

## Review

- Preserve the source archives until in-game verification passes.
- Treat a textures-only alias folder as part of its mapped face.
- Keep existing Sider configuration unchanged when the configured face root is already active.
- Use `pes-wiki-log` for a durable machine-specific result.

## Boundaries

- Use `pes-roster-extract` to generate a roster from game data.
- Use `pes-kit-ftex` for kit textures.
- Read `faces/relink.go` and `faces/fpk.go` for package-format details.
