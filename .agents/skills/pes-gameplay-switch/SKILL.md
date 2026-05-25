---
name: pes-gameplay-switch
description: '**Invoke this skill BEFORE proposing or running any switch of `dt13_all.cpk`, `dt18_all.cpk`, `FL_2026.exe`, or other gameplay-stack file on the live PES / Football Life install.** Covers `tools/fl-gameplay.ps1 switch-dt13-vanilla` / `switch-dt18-vanilla`, the canonical gameplay backup root at `%USERPROFILE%\MEGA\gaming\pes\gameplay\`, the SYSTEM cache propose-and-ask rule (propose `SYSTEM00000000` deletion when warranted — dt13/dt18/exe or EDIT-save replacements that prompt "create edit data" — never for kits/cameras/luas; always ask before running), the no-partial-switch rule, and the explicit-user-approval gate. Triggers: "switch to vanilla dt18", "go back to stock", "restore dt13", "swap in alexfe87", "install F4L v5.1", any phrase that implies replacing a gameplay-stack file.'
---

# PES gameplay switch

Switches are destructive (overwrite live files in the install). Evidence first, approval before action, full-stack inventory always.

## When to use

- User asks to install, restore, or swap a `dt13_all.cpk`, `dt18_all.cpk`, or `FL_2026.exe`
- User asks to enable or disable a gameplay-related sider entry (`lua.module`, `cpk.root` tagged as gameplay)
- Reverting a failed playtest

## Steps

1. **Run [[pes-gameplay-status]] first.** Capture current hashes and the active sider config. The switch is reversible only if the prior state is recorded.

2. **Inventory the full gameplay stack** to confirm what the switch will actually change. A vanilla `dt18` swap is meaningless if a livecpk root or a `GamePlay-v2.lua` module is still routing the gameplay through someone's mod.

3. **Ask the user to approve the exact change.** State, in one block:

    - which file(s) will be replaced
    - source hash -> destination hash
    - whether SYSTEM cache invalidation is warranted (it is for dt13 / dt18 / exe; it is not for kits / cameras / luas, see [[feedback_system_cache_scope]])
    - that you will ASK the user before any SYSTEM delete, never run it as part of the switch

4. For a vanilla switch, prefer the helper:

    ```powershell
    pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla
    pwsh -File .\tools\fl-gameplay.ps1 switch-dt18-vanilla
    ```

    The helper backs up the current live file under the canonical backup root, restores the vanilla copy (loose files under `%USERPROFILE%\MEGA\gaming\pes\gameplay\vanilla\`, falling back to `dt13 & dt18 vanilla.rar`), and updates the `; gameplay:` tracking comment in `SiderAddons\sider.ini`.

5. For a non-vanilla switch (e.g. install a new gameplay release), extract directly to the final live path; no temp dirs, no intermediate copies. See [[feedback_extract_directly]].

6. After the switch, re-run [[pes-gameplay-status]] to confirm the live hashes match the intended destination.

## Pitfalls

- **Never partial-switch.** Replacing `dt18` while leaving a previous combo's livecpk root or `lua.module` entry active produces a mixed state the user did not ask for. If a switch can leave the stack mixed, stop and ask.
- **Propose, do not execute, the `SYSTEM00000000` delete.** For dt13 / dt18 / exe switches the cache invalidation is warranted; surface it as an explicit step and ask the user before running. Never bundle it silently. See [[feedback_system_cache_scope]].
- **No file is "obviously unrelated".** If a mod readme bundles an animation pack, a hook, or a livecpk root as part of its gameplay, treat all of it as the gameplay stack. Verify from the mod's instructions before excluding anything.
- **The user-provided install path wins.** If the user names an install path, use it; do not search broader drives. The current path on this machine is `C:\Program Files (x86)\SP Football Life 2026\`. See [[reference_fl26_install]].
- **Never `rm` files outside a git repo without a verified backup.** The helper makes its own backup for vanilla switches; for any non-helper switch, write the prior file to `%USERPROFILE%\MEGA\gaming\pes\gameplay\` first. See [[feedback_never_rm_user_data]].
- **Senior-developer judgement.** No "temporary workaround" gameplay configs. If a switch produces a mixed state, fix the mix; do not hide it behind a wrapper lua. See AGENTS.md "Root causes, not symptoms".

## Canonical references

- Sider config layout (section order, `lua.module` semantics, `cpk.root` semantics, cache behavior): [SOK Unleashed v9 thread](https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/).
- Gameplay archives: `%USERPROFILE%\MEGA\gaming\pes\gameplay\` ([[reference_gameplay_mods_path]]).

## Out of scope

- Reading the current state: see [[pes-gameplay-status]].
- Researching whether a new release is worth installing: see [[pes-evoweb-research]].
- Logging the playtest verdict: see [[pes-aikb-log]].
