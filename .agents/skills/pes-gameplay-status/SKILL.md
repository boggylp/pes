---
name: pes-gameplay-status
description: '**Invoke this skill BEFORE answering any question about which gameplay, dt13/dt18, exe, sider modules, or livecpk roots are active on the live PES / Football Life install.** Covers `tools/fl-gameplay.ps1 status` (hashes `dt13_all.cpk`, `dt18_all.cpk`, `FL_2026.exe`, dumps Sider tracking comment, active gameplay-related sider entries, SYSTEM cache presence), hostname tagging (`BogambeDesktop` desktop / `DESKTOP-J0MDFMU` laptop), the inventory rule (always check the full gameplay stack: dt13, dt18, livecpk roots, lua.module entries, exe, hooks, cache), and the verify-against-hash convention. Triggers: "what gameplay is installed?", "which dt18 am I running?", "check the live install", "is the exe vanilla?", "status of FL26", any question that needs hash-level proof of what is loaded.'
metadata:
  trusted_sources:
    - https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/
---

# PES live install gameplay status

The user's recurring entry point for any gameplay question: "what's running on this machine right now?". Identify by **hash, not filename**.

## When to use

- Any question about which dt13 / dt18 / exe is active
- Before proposing any gameplay switch or install change (the user requires evidence first)
- Cross-machine comparison ("does the laptop have the same combo as the desktop?")
- Logging a playtest verdict to the AIKB (run status, capture hashes, then log)
- **Verifying the gameplay stack survived an installer run** (FL26 patch reinstall or update, a third-party mod installer, anything that touches `Data/` or the exe). See "Post-installer verification" below.

## Steps

1. Run the helper:

    ```powershell
    pwsh -File .\tools\fl-gameplay.ps1 status
    ```

    Prints SHA-256 for `dt13_all.cpk`, `dt18_all.cpk`, `FL_2026.exe`; the `; gameplay:` tracking comment from `SiderAddons\sider.ini`; active `lua.module` and `cpk.root` entries that relate to gameplay; `SYSTEM00000000` cache presence.

2. Capture hostname for any artefact that leaves this session:

    ```powershell
    hostname
    ```

    Expected: `BogambeDesktop` (FL26 desktop) or `DESKTOP-J0MDFMU` (BPB laptop). Every AIKB log entry, every commit message, every diary entry must name the host. See [[feedback_pes_log_hostname]].

3. Resolve hashes against `~/dev/priv/ai-knowledge-base/wiki/pes/gameplay-combos-fl26.md` (or `bpb-2026.md` for the laptop). The AIKB tables list known hashes per release; match by hash, not by filename.

4. Inventory the **full gameplay stack** before drawing any conclusion:

    - `Data/dt13_all.cpk` and `Data/dt18_all.cpk`
    - `FL_2026.exe`
    - `SiderAddons\sider.ini`: every `lua.module` entry tagged or commented as gameplay, every gameplay-related `cpk.root` entry
    - Hook files, exe replacements
    - `SYSTEM00000000` cache state

    Effective gameplay = what wins after load order, not just everything wired. State both when they diverge.

## Post-installer verification

After any FL26 patch reinstall / update, or any third-party installer that could write to `Data/`, `FL_2026.exe`, `SiderAddons/`, or `dataSP/`, verify the gameplay stack is intact instead of trusting the installer to be a no-op.

1. **Snapshot the expected state before the installer runs**, from the AIKB combo verdict in `~/dev/priv/ai-knowledge-base/wiki/pes/gameplay-combos-fl26.md` (or `bpb-2026.md`). Capture the three SHA-256s for `dt13_all.cpk`, `dt18_all.cpk`, `FL_2026.exe`, plus the sider tracking comment and `dataSP/version.txt`. If the AIKB is stale, run a pre-install status first and use that.

2. **Run the installer** (user action, not yours — see [[pes-gameplay-switch]] for the approval gate).

3. **Re-run status after the installer finishes**:

    ```powershell
    pwsh -File .\tools\fl-gameplay.ps1 status
    Get-Content "$env:ProgramFiles(x86)\SP Football Life 2026\dataSP\version.txt"
    ```

4. **Diff hash-by-hash**: dt13, dt18, exe, plus `dataSP/version.txt` and the sider tracking comment. A clean reinstall (no version change, no stack damage) leaves all four identical. Any divergence means the installer overwrote something:

    | What changed                                  | What that means                                                                                   |
    | --------------------------------------------- | ------------------------------------------------------------------------------------------------- |
    | dt13 / dt18 / exe reverted to stock           | Installer wiped the modded stack. Reinstall the mod from `MEGA/gaming/pes/gameplay/` archives.   |
    | `dataSP/version.txt` bumped                   | Patch updated. Expected for an update; unexpected for a reinstall of the same version.            |
    | Sider tracking comment reset                  | `SiderAddons/sider.ini` was rewritten. Restore the comment block manually.                        |
    | `SYSTEM00000000` deleted                      | Save cache wiped by installer. The user always handles cache state — flag and stop, do not delete or restore it autonomously. |

5. **Report the diff plainly**: every hash and its prior value, side by side. Do not paraphrase "all good" when one or more lines diverged.

## Pitfalls

- **Filename is not identity.** A file called `dt18_all.cpk.f4l` proves nothing; hash it. See [[feedback_investigate_before_assuming]].
- **Never edit individual mod values.** Add or remove mods as whole units; do not tweak a single line inside someone else's lua. See [[feedback_gameplay_mods]].
- **Anti-Cheat.lua is debunked.** Holland's `Anti-Cheat.lua` has no valid AOB addresses; use `Dynamic_Difficulty` instead. See [[feedback_anticheat_debunked]].
- **DT18 also unlocks animations**, not just DT13. alexfe87's DT18 is the animation source for that combo. See [[feedback_pes_dt_files]].
- **Do not propose changes after a status check** unless the user explicitly approves the install/change step. Evidence-first only.
- **Never delete `SYSTEM00000000`** as part of any verification step. Cache invalidation is the user's manual job. See [[feedback_system_cache_scope]].

## Out of scope

- Performing a gameplay switch: see [[pes-gameplay-switch]].
- Writing the verdict to the AIKB: see [[pes-aikb-log]].
- Editing sider config to add or remove modules.
