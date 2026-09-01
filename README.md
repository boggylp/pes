# PES

Utilities for Pro Evolution Soccer and SP Football Life.

## Structure

| Path | Purpose |
| --- | --- |
| `cpk/` | Read and extract CPK files. Assemble self-contained kitserver packs. |
| `editsave/` | Decrypt, edit, encrypt, and verify EDIT saves. |
| `evoweb/` | Scrape XenForo threads and download authenticated attachments. |
| `faces/` | Map player faces and detect ID mismatches. |
| `pesdb/` | Extract player rosters from `Player.bin` and decrypted EDIT data. |
| `uniparam/` | Add kit slots to uniform databases. |
| `tools/` | Run local PowerShell workflows. |

Build a Go tool from its directory:

```sh
go build .
```

## Tools

### Football Life gameplay

`tools/fl-gameplay.ps1` checks the active gameplay stack and restores selected vanilla dt13 or dt18 sources.

```powershell
pwsh -File .\tools\fl-gameplay.ps1 status
pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla
pwsh -File .\tools\fl-gameplay.ps1 switch-dt18-vanilla
```

The default game root is `%PES_GAME_ROOT%` when that machine-level variable is set, otherwise `%ProgramFiles(x86)%\SP Football Life 2026`. The default gameplay archive root is `%USERPROFILE%\MEGA\gaming\pes\gameplay`.

- `status` prints SHA-256 hashes, gameplay tracking lines, active gameplay-related Sider entries, and save-cache presence.
- A switch command saves the current live file under `<GameRoot>\.backup\Data\`.
- A switch command restores the vanilla file and updates the gameplay tracking line.
- The helper prefers loose files under the `vanilla\` archive directory. It can fall back to `dt13 & dt18 vanilla.rar`.
- The helper reports save-cache state but does not change it.

### Sider Lua global localizer

`tools/sider-lua-localize.ps1` finds Sider Lua modules that share a global variable and rewrites the offenders to declare it file-scope local.

Lua makes an assignment without `local` a global for the whole Lua state, and Sider loads every `lua.module` into one state. Two modules that reuse a variable name overwrite each other on every livecpk lookup, which degrades a match the longer the process runs.

```powershell
pwsh -File .\tools\sider-lua-localize.ps1 report
pwsh -File .\tools\sider-lua-localize.ps1 apply
pwsh -File .\tools\sider-lua-localize.ps1 restore
pwsh -File .\tools\sider-lua-localize.ps1 report -SiderAddons "D:\Games\SP Football Life 2026\SiderAddons"
pwsh -File .\tools\sider-lua-localize.ps1 apply -AllGlobals
```

| Parameter | Effect |
| --- | --- |
| `report` | List each module, the names to localize, and the names skipped. Changes nothing. |
| `apply` | Insert a `local` declaration block at file scope in each offending module. |
| `restore` | Copy every `.pre-localize.bak` back and remove the backup. |
| `-SiderAddons <path>` | Select the install. The default is `SiderAddons` under `%PES_GAME_ROOT%`, otherwise `%ProgramFiles(x86)%\SP Football Life 2026\SiderAddons`. |
| `-AllGlobals` | Localize every module-owned global, not only the shared ones. |
| `-Force` | Localize a name another loaded module reads. |

The tool reads the loaded set from `lua.module` entries in `sider.ini`, so rerun it after adding a module. Analysis always reads `<file>.pre-localize.bak` when one exists, which makes `apply` idempotent. Assignments are never rewritten, only declared, so a value shared between functions in one file keeps working. Compiled LuaJIT modules are skipped because they cannot be parsed or patched. A name written by one module and read by another is skipped and reported instead of localized.

### Kit PNG to FTEX

`tools/kit-to-ftex.ps1` converts a PNG to a kitserver-compatible FTEX. It converts the PNG to DXT5 DDS with ImageMagick and then converts DDS to FTEX with FtexTool.

Prerequisite: ImageMagick `magick` on `PATH`.

FtexTool v0.4.0 downloads on first use to `tools/bin/FtexTool-v0.4.0/`, which Git ignores.

```powershell
pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png
pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png path\to\out\u.ftex
pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamId 2525 -Slot 3
pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamName Hajduk -Slot 3 -OutDir out\
pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamId 2525 -KitType g -Slot 1
```

| Parameter | Effect |
| --- | --- |
| `-TeamId <int>` | Name the output `u<id><p\|g><slot>.ftex`. |
| `-TeamName <text>` | Find the team with a substring match in the team list. |
| `-TeamsFile <path>` | Select another team list. |
| `-Slot <1-9>` | Select the kit slot. The default is `1`. |
| `-KitType p\|g` | Select a player or goalkeeper kit. The default is `p`. |
| `-OutDir <path>` | Select the output directory. The default is the input directory. |

The output contains one embedded DXT5 FTEX with mipmaps. Put it in an existing kitserver slot folder. The tool does not create `config.txt`, `order.ini`, `map.txt`, or partial textures. Some stock textures use a different pixel format, so verify the result in the game.

## Evoweb

The `evoweb` tool stores credentials, scrapes XenForo threads and forum listings, and downloads authenticated files.

```sh
cd evoweb
go build .
go run . login
go run . scrape --output data/example.json "https://evoweb.uk/threads/example.88633/"
go run . scrape --output data/example.json --max-pages 3 "https://evoweb.uk/threads/example.88633/"
go run . scrape --output data/example.json --last-pages 5 "https://evoweb.uk/threads/example.88633/"
go run . scrape --output data/example.json --cookie "xf_session=abc; xf_user=def" "https://evoweb.uk/threads/example.88633/"
go run . forum --output data/forum.json --max-pages 3 "https://evoweb.uk/forums/pes-2021.337/"
go run . download --output dt18_all.cpk "https://evoweb.uk/attachments/dt18_all-cpk.432685/"
```

The login command stores credentials under `~/.secrets/evoweb/credentials`. Scrape and forum commands use the stored session. Use `--cookie` only to diagnose stored-session problems.

## Faces

The `faces` tool maps face folders between player databases and detects invalid folders.

```sh
cd faces
go build .

./faces detect \
  --faces-dir "/path/to/livecpk/Asset/model/character/face/real" \
  --player-csv "/path/to/player-ids.csv"

./faces map \
  --source-csv /path/to/source-players.csv \
  --destination-csv /path/to/destination-players.csv \
  --source-folder /path/to/source/faces \
  --dest-folder /path/to/destination/faces

./faces relink --folder /path/to/face/real/100219 --id 2147483648
```

For equal-length IDs, `map` and `relink` replace the embedded decimal ID in each package. For different-length IDs, they keep package bytes unchanged and create a texture alias under the embedded ID.

## Tool documentation

- [CPK reader and kitserver-pack assembly](cpk/README.md)
- [EDIT-save operations](editsave/README.md)
- [Player roster extraction](pesdb/README.md)
- [Uniform kit-slot editing](uniparam/README.md)
