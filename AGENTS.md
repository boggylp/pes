# PES repo

@README.md

## Repo-specific

- Monorepo; independent tools per subdir, own language/build.
- Before every commit, run `make all` in the changed subdir.
- Tracked changes in `git status` → ask `commit/push?`; commit+push on confirm.
- PES/Football Life mod questions: check AIKB (`~/dev/priv/ai-knowledge-base/wiki/pes/`) first, then `evoweb/data/` freshness; scrape >1 week old → rescrape before answering. Update AIKB when findings worth preserving.
- AIKB house style: lead with conclusion, tight bullets, short factual sentences, mark hypotheses **Not verified** once.
- Live-install task → read the live install's game-root `AGENTS.md` + `README.md` first (install path varies per machine; the user-given path wins).
- Canonical gameplay backup root: `%USERPROFILE%\MEGA\gaming\pes\gameplay\`; EDIT/SYSTEM saves under `%USERPROFILE%\MEGA\gaming\pes\edit\saves\`.
- Save/EDIT/SYSTEM backup filenames must embed the FL26 version, machine, and date: `fl26-<savetype>_fl26-<version>_<machine>_<YYYY-MM-DD>`. Version = patch name + exe FileVersion (e.g. `v2.2-26.2.0.3`).
- User-given install path is authoritative; don't search wider unless it fails or they ask.
- Evidence-first: verify active files (`SiderAddons\sider.ini`, `Data\dt13/dt18`, backups, archive/thread instructions) before proposing any reset/install.
- Identify installed gameplay by file hash, not filename.
- alexfe87: `dt18_v4` = May 2026 DT18 release; `GamePlay-v2.lua` = the separate required Lua module.
- No gameplay/config change to a live install without explicit user approval of the change step.
- Before any switch, inventory the full stack (dt13, dt18, gameplay livecpk roots, gameplay `lua.module` entries, exe, hooks, cache); verify each from evidence or mod instructions. A file named like an animation/visual addon counts as gameplay if a mod readme bundles it.
- No partial switch leaving a mixed state unless the user asked for that exact mix.
- **`SYSTEM00000000` deletion: user approval required, never autonomous.** Propose when its mismatch likely causes a real problem (dt13/dt18/exe switch, or EDIT-save replacement prompting "create edit data"); ask first; never bundle silently. Cost: also wipes settings cache + recent-match state.
- `tools/fl-gameplay.ps1` for recurring status checks and vanilla dt13/dt18 switches.
- Query `evoweb/data/` JSON with `duckdb`.
- Elevation: spawn elevated `pwsh` from the session, don't stop. `Start-Process (Get-Command pwsh.exe).Source -Verb RunAs -Wait -ArgumentList @('-NoProfile','-ExecutionPolicy','Bypass','-File',<script>)`.
- Sider config reference (sections, `lua.module`, livecpk roots, cache): [SOK Unleashed v9](https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/). Sider 7 Lua scripting API (events, `ctx.register`): [docs](https://mapote.com/doc/sider/sider7/scripting.html) — check before judging whether a `lua.module` runs.

## Tools

Build each with `go build .` in its subdir. Usage in @README.md; operational detail in the `pes-*` skills. Subdirs: `evoweb` (XenForo scraper; always `--output data/<name>.json`), `cpk` (CPK reader/extractor + `cpk kits` kitserver-pack assembler), `pesdb` (roster extractor; EDIT saves need ejogc327's `decrypter21.exe`), `faces` (map/detect), `uniparam` (UniColor.bin/UniformParameter.bin editor; extend a team's kit-slot count via livecpk), `tools/` (PowerShell helpers).

Gotchas beyond README:

- Player base `common/etc/pesdb/Player.bin` in `Data/dt00_x64.cpk`; `dt10_x64.cpk` + `download/dt80_*E_x64.cpk` override by load order. Not encrypted: `\xff\x10\x81WESYS` + zlib. BPB ships a real Player.bin (1,751,422 B, md5 `89938c15`) in dt00=dt10; EDIT save adds 20 BPB customs.
- pesdb stride 312 B. Player.bin: Id `+0x08`, name `+0x44`, shirt `+0x81`. EDIT `data.dat`: Id `+0x0C`, name `+0x42`, shirt `+0x7F`, records from offset 112.
- **Authoritative roster = the live DB, not `FL26_players.txt`.** UML (or any DB patch) replaces `Player.bin` via livecpk (`SiderAddons\livecpk\UML_Database\...\Player.bin`) and renumbers players (Beljo base `91287` vs UML `58035`), so the stale base export keys faces to wrong/dead IDs. Before face/ID work: use the install's provided list (`UML 2026 - Player IDs.csv`, `... - Team IDs.csv`) or extract the live `UML_Database` `Player.bin` with `pesdb`; verify Beljo's ID.
- Nationality byte = `+0x1D` (BPB `Player.bin`; UML team `Country` uses a different scheme). Balkan: Croatia 144, Serbia 94, Bosnia 140, Montenegro 97 (132 = Austria), N.Macedonia 186, Slovenia 214, Albania 126, Kosovo 110.
- Faces bind by folder ID **and** the ID embedded in `face.fpk`; both must match or it renders default (created players included: folder copy alone is not enough). Equal-length IDs: byte-replace the decimal ID in place. Length-changing IDs (created-player `0x80000000`+ = 10 digits) need an FPK repack (`faces relink --folder <dir> --id <new>`): rewrites the ID and recomputes foxfpk entry offsets/sizes. Sound without FMDL-internal fixup because the IDs sit in FMDL texture-path strings nothing references by offset. `faces map` routes length-mismatched pairs through it.
- `faces map`/`relink` walk **every** `#Win/*.fpk` and `*.fpkd` package, not just `face.fpk`, and rewrite the ID wherever the path appears (covers a separate oral/hair package). The FL26 `face.fpkd` is a 48-byte ID-less `foxfpkd` dependency stub (distinct magic; `parseFpk` rejects it) — carried intact by the folder copy, no rewrite. A length-changing rewrite of a `foxfpkd` that *did* embed the path is refused (no verified repack); an equal-length swap is the format-agnostic in-place replace.
- `faces map` matches exact normalized name, then a relaxed first+last fallback (surname exact + compatible first name, middle names ignored, collisions rejected), so BPB `Dion Drena Beljo` maps to live `Dion Beljo`.
- `faces/samples/` CSVs lag the install; never use them for an install.
- Mark every custom `sider.ini` edit with a `; [GB-CUSTOM] manual: <what>, <date> <host>` line above it (positive phrasing; never "not from UML/patch").

## Domain language

### Archives and game data

| Term | Definition | Aliases to avoid | Notes |
| ---- | ---------- | ---------------- | ----- |
| **CPK** | A CRI Middleware archive used by PES and Football Life to package game data. | archive file | |
| **EDIT save** | The encrypted user save that carries custom adds and edits on top of the archive Player.bin baseline. | edit file, option file | Do not treat it as parseable by the repo CPK tool. |
| **Inner path** | The path of a file inside a CPK archive. | internal path, archive path | Matching is case-insensitive and slash-normalized in the repo tool. |
| **Load order** | The order in which base archives, patch archives, DLC archives, livecpk roots, and Sider modules override earlier game data. | priority, precedence | State the scope when discussing it. |
| **Player.bin** | Konami's binary player database stored inside PES archive data, packaged as a WESYS+zlib envelope. | player DB, players file | BPB and FL26 ship distinct Player.bin bytes; EDIT save carries custom adds on top. Decode with the repo's `pesdb` tool. |
| **Table of contents** | The parsed CPK entry list used to list and extract archived files. | TOC | |

### Faces

| Term | Definition | Aliases to avoid | Notes |
| ---- | ---------- | ---------------- | ----- |
| **Face folder** | A numeric player-ID directory containing the face assets loaded by the game. | player folder, ID folder | |
| **Face install** | Copying or remapping a face folder into the live install's configured face root. | face import | Requires a rollback record. |
| **FPK** | A PES package file inside a face folder that embeds asset paths and the referenced player ID. | face.fpk | Length-sensitive path data makes naive ID replacement unsafe. |
| **Length mismatch** | A source and destination player ID pair whose decimal string lengths differ. | digit mismatch | Requires an FPK-aware editor, not forced install. |
| **Orphaned face** | A face folder whose ID does not exist in the active player database. | orphan | |
| **Player ID** | The numeric identifier that links player database rows to face folders and embedded FPK paths. | face ID | |

### Gameplay and live install

| Term | Definition | Aliases to avoid | Notes |
| ---- | ---------- | ---------------- | ----- |
| **dt13** | The gameplay-related CPK component commonly changed by gameplay patches. | dt13 file | Verify by hash before identifying it. |
| **dt18** | The gameplay-related CPK component commonly changed by gameplay patches. | dt18 file | Verify by hash before identifying it. |
| **Effective gameplay** | The gameplay behavior that should win after applying load order, not merely every gameplay-related component installed or wired. | installed combo | State both when they differ. |
| **EXE mod** | A modded `FL_2026.exe` / `PES2021.exe` carrying hardcoded gameplay changes, ranging from a multi-MB fork to a few-byte binary patch. | exe patch, modded exe | Identify by hash, never by label. A label can be wrong (SHA256 `9EE0C306` is vanilla FL26 26.2.0.3, not the "Holland WE2026" it was filed as; real Holland = `19740A3C`, vanilla +17 bytes). |
| **Gameplay stack** | The full set of files, livecpk roots, Sider modules, executable replacements, hooks, and caches affecting gameplay. | gameplay mod, gameplay files | Inventory the whole stack before switching. |
| **Live install** | The actual PES or Football Life installation currently used for play and verification. | game folder, install path | The user-provided path wins over search. |
| **SYSTEM cache** | The PES save cache that can preserve gameplay or settings state across file switches, and which holds a fingerprint of the EDIT save. A stale fingerprint after replacing EDIT00000000 triggers a "create edit data" prompt that wipes the new save. | system file, cache | Removal is destructive (loses settings cache, recent-match state); always ask the user before deleting. |
| **Vanilla** | A known clean baseline copy of a component from the active game or patch. | default, original | For BPB, vanilla means BPB stock unless explicitly qualified as Konami vanilla. Prefer hash evidence over filename claims. |

### Kits

| Term | Definition | Aliases to avoid | Notes |
| ---- | ---------- | ---------------- | ----- |
| **FTEX** | The PES texture container format used by kitserver for kit textures. | ftex file | The converter emits single embedded FTEX files. |
| **Kit slot** | The numbered player or goalkeeper kit variant selected by kitserver naming. | slot | Valid slot values are tool-specific. |
| **Kitserver** | The Sider-based kit loading system that maps teams and kit slots to texture folders. | kit loader | Folder scaffolding is separate from texture conversion. |
| **Pixel format** | The texture compression format stored inside an FTEX. | format | DXT5 is broad compatibility, not necessarily stock-equivalent. |

### Sider and mod loading

| Term | Definition | Aliases to avoid | Notes |
| ---- | ---------- | ---------------- | ----- |
| **cpk.root entry** | A Sider config entry that adds a livecpk root to the game data override chain. | cpk root, livecpk entry | |
| **livecpk root** | A loose-file directory that Sider presents to the game as overrideable CPK content. | livecpk folder | |
| **lua.module entry** | A Sider config entry that loads a Lua module into the game process. | Lua module, module entry | Treat bundled gameplay modules as part of the gameplay stack. Loaded is not running: a module that never registers a real Sider event via `ctx.register` is dead code (e.g. Holland's `Difficulty_Manager.lua`). |
| **Sider** | The PES mod loader that injects modules and loose-file roots into the running game. | sider.ini | `sider.ini` is the config file, not the loader. |

### Web research

| Term | Definition | Aliases to avoid | Notes |
| ---- | ---------- | ---------------- | ----- |
| **AI Knowledge Base** | The private PES research wiki checked before answering mod and community-content questions. | AIKB | Update it when findings are worth preserving. |
| **Evoweb scrape** | A persisted JSON capture of an Evoweb thread or forum listing. | scrape data, scraped JSON | Must be refreshed when stale for the question. |
| **Forum listing** | A XenForo forum index page used to discover thread URLs and metadata. | forum page | |
| **Thread** | A XenForo discussion page scraped as posts and metadata. | topic | |
