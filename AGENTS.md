# PES repo

@README.md

## Rules

This repository contains independent PES tools and live-install workflows.

- Run `make all` in each changed tool directory before a commit.
- When restructuring documentation, preserve game and domain knowledge in its current owner or move it to another durable owner.
- Use the install path that the user gives. Search elsewhere only when that path fails or the user asks.
- Read the live install's `AGENTS.md` and `README.md` before a live-install task.
- Use `pes-evoweb-research` for PES or Football Life community research. Check the memory wiki before external research.
- Use `pes-gameplay-status` before conclusions about active gameplay.
- Identify gameplay files by hash, not by filename.
- Get explicit approval for each gameplay or configuration change to a live install.
- Inventory the full gameplay stack before a switch. Include dt13, dt18, gameplay livecpk roots, gameplay `lua.module` entries, the executable, hooks, and cache state.
- Verify each gameplay-stack component from active files or the mod instructions.
- Apply the complete requested gameplay stack. Do not leave an unrequested mixed state.
- Never propose, mention, or perform deletion of `SYSTEM00000000` until an automated clear-cache and settings-reapply tool exists.
- Store gameplay archives under `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.
- Store EDIT-save and SYSTEM-cache backups under `%USERPROFILE%\MEGA\gaming\pes\edit\saves\`.
- Name save backups `fl26-<savetype>_fl26-<version>_<machine>_<YYYY-MM-DD>`. Build `<version>` from the patch name and executable `FileVersion`.
- Use the active live database for player and face work. Do not use the sample CSV files for a live install.
- Use texture aliasing when source and destination player IDs have different lengths. Do not byte-rewrite the FMDL path.

## Domain language

These terms define the repository's recurring vocabulary.

### Archives

| Term | Definition | Aliases to avoid | Notes |
| --- | --- | --- | --- |
| **CPK** | A CRI Middleware package that contains PES game data. | archive file | |
| **EDIT save** | The encrypted user save that adds custom data and edits to the archive roster. | edit file, option file | The CPK tool does not parse it. |
| **Inner path** | The normalized path of a file inside a CPK. | internal path, archive path | Matching ignores case and slash direction. |
| **Load order** | The order in which archives, livecpk roots, and Sider modules override game data. | priority, precedence | State the applicable scope. |
| **Player.bin** | Konami's WESYS and zlib player database. | player DB, players file | Decode it with `pesdb`. |
| **Table of contents** | The parsed list of entries in a CPK. | TOC | |

### Faces

| Term | Definition | Aliases to avoid | Notes |
| --- | --- | --- | --- |
| **Face folder** | A numeric player-ID directory that contains face assets. | player folder, ID folder | |
| **Face install** | Copying or remapping a face folder into the configured livecpk face root. | face import | Record how to roll it back. |
| **FPK** | A PES package in a face folder that contains asset paths and a player ID. | face.fpk | A length-changing ID rewrite can corrupt path offsets. |
| **Length mismatch** | A source and destination player-ID pair with different decimal lengths. | digit mismatch | Use texture aliasing. |
| **Orphaned face** | A face folder whose ID is absent from the active player database. | orphan | |
| **Player ID** | The numeric key that links player records, face folders, and FPK paths. | face ID | |

### Gameplay

| Term | Definition | Aliases to avoid | Notes |
| --- | --- | --- | --- |
| **dt13** | A CPK component that gameplay patches can replace. | dt13 file | Identify it by hash. |
| **dt18** | A CPK component that gameplay patches can replace. | dt18 file | Identify it by hash. |
| **Effective gameplay** | The gameplay behavior that wins after load order is applied. | installed combo | Distinguish it from all wired components. |
| **EXE mod** | A modified game executable that contains gameplay changes. | exe patch, modded exe | Identify it by hash. |
| **Gameplay stack** | All files, livecpk roots, Sider modules, executable changes, hooks, and caches that can affect gameplay. | gameplay mod, gameplay files | |
| **Live install** | The PES or Football Life installation used for play and verification. | game folder, install path | The user-provided path is authoritative. |
| **SYSTEM cache** | The PES save cache that stores settings, recent-match state, and an EDIT-save fingerprint. | system file, cache | Removal is destructive. |
| **Vanilla** | A verified clean baseline for the active game or patch. | default, original | Qualify Konami and patch baselines when they differ. |

### Kits

| Term | Definition | Aliases to avoid | Notes |
| --- | --- | --- | --- |
| **FTEX** | The PES texture container used by kitserver. | ftex file | |
| **Kit slot** | A numbered player or goalkeeper kit variant. | slot | Valid values depend on the tool. |
| **Kitserver** | The Sider system that maps teams and kit slots to texture folders. | kit loader | Folder configuration is separate from texture conversion. |
| **Pixel format** | The texture-compression format stored in an FTEX. | format | DXT5 is compatible but not always stock-equivalent. |

### Sider

| Term | Definition | Aliases to avoid | Notes |
| --- | --- | --- | --- |
| **cpk.root entry** | A Sider setting that adds a livecpk root to game-data load order. | cpk root, livecpk entry | |
| **livecpk root** | A loose-file directory that Sider presents as CPK data. | livecpk folder | |
| **lua.module entry** | A Sider setting that loads a Lua module into the game process. | Lua module, module entry | Loaded code runs only when it registers a Sider event. |
| **Sider** | The mod loader that injects modules and loose-file roots. | sider.ini | `sider.ini` is its configuration file. |

### Research

| Term | Definition | Aliases to avoid | Notes |
| --- | --- | --- | --- |
| **Memory wiki** | The PES article set at `~/memory/priv/wiki/pes/`, searched before external research. | AIKB, AI Knowledge Base | Preserve durable findings there through `pes-wiki-log`. |
| **Evoweb scrape** | A dated JSON capture of an Evoweb thread or forum listing. | scrape data, scraped JSON | Refresh it when the question requires current state. |
| **Forum listing** | A XenForo index used to discover threads and metadata. | forum page | |
| **Thread** | A XenForo discussion and its posts. | topic | |
