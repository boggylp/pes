# PES

Utilities for Pro Evolution Soccer 2021 and SP Football Life. Build a Go tool with `go build .` in its directory.

| Path | Purpose |
| --- | --- |
| `cpk/` | Read and extract CPK files, and assemble self-contained kitserver packs, per [its README](cpk/README.md). |
| `editsave/` | Decrypt, edit, encrypt, and verify EDIT saves, per [its README](editsave/README.md). |
| `evoweb/` | Scrape and search Evoweb threads, and download authenticated attachments. |
| `faces/` | Map player faces between databases and detect ID mismatches. |
| `pesdb/` | Extract player rosters from `Player.bin` and decrypted EDIT data, per [its README](pesdb/README.md). |
| `uniparam/` | Add kit slots to uniform databases, per [its README](uniparam/README.md). |
| `tools/` | PowerShell workflows for the live install. |

The PowerShell tools default to the install at `%PES_GAME_ROOT%`, else `%ProgramFiles(x86)%\SP Football Life 2026`.

## Football Life gameplay

`tools/fl-gameplay.ps1` reports the active gameplay stack and restores vanilla dt13 or dt18 from `%USERPROFILE%\MEGA\gaming\pes\gameplay`.

```powershell
pwsh -File .\tools\fl-gameplay.ps1 status
pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla
pwsh -File .\tools\fl-gameplay.ps1 switch-dt18-vanilla
```

- `status` prints the dt13, dt18, and executable hashes, the `; gameplay:` tracking line, gameplay-related Sider entries, and SYSTEM-cache presence.
- `status` resolves each of the 9 constant bins by Sider load order, first `cpk.root` then `dt18_all.cpk`, and names a release only when every winner matches its known hashes.
- A switch restores the file from the archive's `vanilla\` directory, else from `dt13 & dt18 vanilla.rar`, and updates the tracking line.
- Neither command changes the SYSTEM cache.

## Sider Lua global localizer

Sider loads every `lua.module` into one Lua state, so two modules assigning the same global without `local` overwrite each other on every livecpk lookup. `tools/sider-lua-localize.ps1` finds those shared globals and declares them file-scope `local` in each offending module.

```powershell
pwsh -File .\tools\sider-lua-localize.ps1 report
pwsh -File .\tools\sider-lua-localize.ps1 apply
pwsh -File .\tools\sider-lua-localize.ps1 restore
```

| Parameter | Effect |
| --- | --- |
| `report` | List each module, the names to localize, and the names skipped. |
| `apply` | Insert a file-scope `local` block in each offending module, keeping a `.pre-localize.bak`. |
| `restore` | Copy every `.pre-localize.bak` back and remove it. |
| `-SiderAddons <path>` | Select another install's `SiderAddons`. |
| `-AllGlobals` | Localize every module-owned global, not only the shared ones. |
| `-Force` | Localize a name another loaded module reads. |

- The loaded set comes from the `lua.module` entries in `sider.ini`, so rerun it after adding a module.
- Analysis reads the `.pre-localize.bak` when one exists, which makes `apply` idempotent.
- It only declares names and never rewrites an assignment, so a value shared within one file still works.
- It skips compiled LuaJIT modules, and reports instead of localizing a name one module writes and another reads.

## Kit PNG to FTEX

`tools/kit-to-ftex.ps1` converts a PNG to a DXT5 FTEX with mipmaps for an existing kitserver slot folder. It needs ImageMagick `magick` on `PATH`, and downloads FtexTool v0.4.0 to the ignored `tools/bin/` on first use.

```powershell
pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png [path\to\out\u.ftex]
pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamName Hajduk -KitType g -Slot 1 -OutDir out\
```

| Parameter | Effect |
| --- | --- |
| `-TeamId <int>` | Name the output `u<id><p\|g><slot>.ftex`. |
| `-TeamName <text>` | Find the team by substring in the team list. |
| `-TeamsFile <path>` | Select the team list. The default is the installed `FL26_teams.txt`. |
| `-Slot <1-9>` | Select the kit slot. The default is `1`. |
| `-KitType p\|g` | Select a player or goalkeeper kit. The default is `p`. |
| `-OutDir <path>` | Select the output directory. The default is the input directory. |

It writes no `config.txt`, `order.ini`, `map.txt`, or partial textures. Some stock textures use another pixel format, so check the result in game.

## Evoweb

`evoweb` keeps credentials under `~/.secrets/evoweb/credentials` after `login`, and every other command uses that session.

```sh
go run . login
go run . scrape --output data/example.json [--max-pages 3 | --last-pages 5] "https://evoweb.uk/threads/example.88633/"
go run . forum --output data/forum.json --max-pages 3 "https://evoweb.uk/forums/pes-2021.337/"
go run . search -q "<query>" [--user <name>] --output data/search.json
go run . download --output dt18_all.cpk "https://evoweb.uk/attachments/dt18_all-cpk.432685/"
```

`--cookie "xf_session=…; xf_user=…"` overrides the stored session, for diagnosing login problems.

## Faces

`faces` maps face folders between player databases and detects invalid ones.

```sh
./faces detect --faces-dir <livecpk>/Asset/model/character/face/real --player-csv <player-ids.csv>
./faces map --source-csv <source.csv> --destination-csv <destination.csv> --source-folder <source-faces> --dest-folder <destination-faces>
./faces relink --folder <face-folder> --id <new-id>
```

For an equal-length ID, `map` and `relink` replace the embedded decimal ID in each package. For a different-length ID, they leave the package bytes unchanged and add a texture alias under the embedded ID.
