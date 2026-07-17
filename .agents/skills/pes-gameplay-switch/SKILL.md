---
name: pes-gameplay-switch
description: 'Invoke BEFORE proposing or running any switch of `dt13_all.cpk`, `dt18_all.cpk`, `FL_2026.exe`, or other gameplay-stack file on the live PES / Football Life install. Covers `tools/fl-gameplay.ps1 switch-dt13-vanilla` / `switch-dt18-vanilla`, the canonical gameplay archive root at `%USERPROFILE%\MEGA\gaming\pes\gameplay\`, the `SYSTEM00000000` delete ban, the no-partial-switch rule, and the explicit-user-approval gate. Triggers: "switch to vanilla dt18", "go back to stock", "restore dt13", "install F4L v5.1", any phrase that implies replacing a gameplay-stack file.'
---

# PES gameplay switch

Switches are destructive (overwrite live files in the install). Evidence first, approval before action, full-stack inventory always.

## When to use

- User asks to install, restore, or swap a `dt13_all.cpk`, `dt18_all.cpk`, or `FL_2026.exe`
- User asks to enable or disable a gameplay-related sider entry (`lua.module`, `cpk.root` tagged as gameplay)
- Reverting a failed playtest

## Steps

1. **Run pes-gameplay-status first.** Capture current hashes and the active sider config. The switch is reversible only if the prior state is recorded.

2. **Inventory the full gameplay stack** to confirm what the switch will actually change. A vanilla `dt18` swap is meaningless if a livecpk root or a `GamePlay-v2.lua` module is still routing the gameplay through someone's mod.

3. **Ask the user to approve the exact change.** State, in one block:

   - which file(s) will be replaced
   - source hash -> destination hash

4. For a vanilla switch, prefer the helper:

   ```powershell
   pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla
   pwsh -File .\tools\fl-gameplay.ps1 switch-dt18-vanilla
   ```

   The helper restores the vanilla copy (loose files under `%USERPROFILE%\MEGA\gaming\pes\gameplay\vanilla\`, falling back to `dt13 & dt18 vanilla.rar`) and updates the `; gameplay:` tracking comment in `SiderAddons\sider.ini`.

5. For a non-vanilla switch (e.g. install a new gameplay release), extract directly to the final live path; no temp dirs, no intermediate copies.

6. After the switch, re-run pes-gameplay-status to confirm the live hashes match the intended destination.

## Pitfalls

- **Never partial-switch.** Replacing `dt18` while leaving a previous combo's livecpk root or `lua.module` entry active produces a mixed state the user did not ask for. If a switch can leave the stack mixed, stop and ask.
- **Never propose, mention, or perform a `SYSTEM00000000` delete** (hard ban, repo `AGENTS.md`).
- **No file is "obviously unrelated".** If a mod readme bundles an animation pack, a hook, or a livecpk root as part of its gameplay, treat all of it as the gameplay stack. Verify from the mod's instructions before excluding anything.
- **The user-provided install path wins.** If the user names an install path, use it; do not search broader drives. The path varies per machine — do not hardcode it here; the user names it per session.
- **Do not create backups the user did not ask for.** The rollback source for a swap is the original mod archive or the `vanilla\` tree already on disk, not a fresh copy of the live file; MEGA is not a backup target. If the prior live file exists nowhere else and would be lost by the overwrite, say so and ask before proceeding. Explicit user-requested backups go to `D:\Backup\`.
- **No "temporary workaround" gameplay configs.** If a switch produces a mixed state, fix the mix; do not hide it behind a wrapper lua.

## Canonical references

- Sider config layout (section order, `lua.module` semantics, `cpk.root` semantics, cache behavior): [SOK Unleashed v9 thread](https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/).
- Gameplay archives: `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.

## Out of scope

- Reading the current state: see pes-gameplay-status.
- Researching whether a new release is worth installing: see pes-evoweb-research.
- Logging the playtest verdict: see pes-aikb-log.
