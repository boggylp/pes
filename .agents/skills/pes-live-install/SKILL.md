---
name: pes-live-install
description: 'Invoke BEFORE installing or restoring a full PES / Football Life patch stack on a machine: a fresh setup from staged parts or a MEGA backup, a UML/BPB version bump, or writing an EDIT save / custom player into the live install. Covers the ordered patch-apply discipline, the canonical live save path under Documents KONAMI, the per-machine journal rule, EDIT-save handling via the ejogc editor, and the Windows prereqs. Triggers: "install my pes stuff on this machine", "set up FL26 from my backup", "install UML v3", "apply my EDIT save", "restore my pes install", "add boggy to <team>".'
metadata:
  trusted_sources:
    - https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/
---

# PES live install / restore

Installing or restoring a patch stack overwrites the live install and the live save. Evidence first, approval before each write, and the live save path is fixed — never guessed.

## When to use

- Fresh setup on a machine from staged parts or a MEGA backup
- A UML / BPB version bump over an existing install
- Writing an EDIT save or a custom player (e.g. boggy) into the live install
- Restoring a broken install to a known-good stack

## Steps

1. **Fix the machine and journal it.** Run `hostname`; known machines are `BogambeDesktop` and `BogambeLegion5`. Open the session journal and record which machine this install runs on — install state is per-machine, and an unlabelled note becomes fake news on the other machine. Every commit, AIKB entry, and journal line names the host.

2. **Inventory current state.** Run pes-gameplay-status to capture the live stack (hashes, sider config, `SYSTEM00000000` presence) before any write. Confirm the FL26 install root — the user-provided path wins; do not search wider unless it fails.

3. **Locate the staged parts.** The user names the staging folder (parts are typically under `D:\syncthing\<patch>\`, e.g. `D:\syncthing\uml v3\`, or a `%USERPROFILE%\MEGA\gaming\pes\` backup root). Use the folder the user names; do not search other drives.

4. **Apply parts in the patch's stated order, one at a time.** Read the patch's own install instructions (its README / evoweb thread) and follow that order exactly — base first, then point releases, then the latest, then any fix last (e.g. `FL2026 v2.0` → `v2.2` → `v3` → the WC v3 fix). Verify each part landed before applying the next. **If a fix already bundles the EDIT file, skip the standalone EDIT.**

5. **Write the EDIT save only to the live save path.** The live save is always:

   ```text
   %USERPROFILE%\Documents\KONAMI\eFootball PES 2021 SEASON UPDATE\2026\save\EDIT00000000
   ```

   Never write to a MEGA backup copy (`%USERPROFILE%\MEGA\gaming\pes\edit\`) and never to smoke's backup (`%USERPROFILE%\MEGA\gaming\pes\smoke\`) — those are sources, not targets. State the exact destination file before writing it. `SYSTEM00000000` deletion stays hard-banned (repo `AGENTS.md`).

6. **For a custom player, do the full swap — not name + kit only.** Applying boggy (Bogambe) means attributes, physicals, and traits as well as name and kit. Use the ejogc editor at `%USERPROFILE%\MEGA\gaming\pes\Pes 2020 Editor V0.12.10 by Ejogc327\...\PES2020Editor.exe` (decrypts the EDIT save via its bundled `Lib\decrypter21.exe`); import the full CSV, save back to the Documents save path. For live-DB / face-folder work behind an ID, defer to pes-roster-extract and pes-faces-install.

7. **Set Windows prereqs.** HAGS (Hardware-accelerated GPU scheduling) off, regardless of laptop-only display.

8. **Verify.** Re-run pes-gameplay-status to confirm the live hashes match the intended stack, then launch the game and confirm the save loads without a "create edit data" prompt.

## Pitfalls

- **The live save is the Documents KONAMI path, always.** Writing to a MEGA backup or to smoke's backup instead of `Documents\KONAMI\...\save\` is the recurring failure. Backups are read-only sources; confirm the destination is the Documents save before any write.
- **Journal the machine.** State is per-machine; never carry one laptop's install facts onto another host.
- **Ordered apply, no skipping.** Later parts assume earlier ones. A fix may bundle EDIT — then skip the standalone EDIT file.
- **Full player swap, not partial.** A name-and-kit-only edit silently drops attributes, physicals, and traits.
- **A stale `SYSTEM00000000` fingerprint** after replacing `EDIT00000000` triggers a "create edit data" prompt that wipes the new save. Do not delete `SYSTEM00000000` to fix it (hard-banned); confirm the save loads and stop if it prompts.
- **No unrequested backups.** The rollback source is the staged parts or the existing backup tree, not a fresh copy of live files. Explicit user-requested backups go to `D:\Backup\`.

## Canonical references

- Repo `AGENTS.md`: canonical paths, filename conventions, `SYSTEM00000000` ban, elevation pattern.
- Sider config layout: [SOK Unleashed v9 thread](https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/).

## Out of scope

- Reading current live state → pes-gameplay-status.
- Swapping a `dt13` / `dt18` / exe gameplay file → pes-gameplay-switch.
- Face folder install / remap → pes-faces-install.
- Roster / Player.bin extraction → pes-roster-extract.
- Researching whether a release is worth installing → pes-evoweb-research.
- Logging the verdict → pes-aikb-log.
