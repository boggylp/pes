---
name: pes-gameplay-status
description: 'Invoke BEFORE answering any question about which gameplay, dt13/dt18, exe, sider modules, or livecpk roots are active on the live PES / Football Life install. Covers `tools/fl-gameplay.ps1 status`, hostname tagging, the full-stack inventory rule, and the verify-against-hash convention. Triggers: "what gameplay is installed?", "which dt18 am I running?", "check the live install", "is the exe vanilla?", any question that needs hash-level proof of what is loaded.'
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

   Known machines: `BogambeDesktop` (desktop), `BogambeLegion5` (Legion 5 laptop, formerly `DESKTOP-J0MDFMU`). Every AIKB log entry, every commit message, every diary entry must name the host.

3. Resolve hashes against `~/dev/priv/ai-knowledge-base/wiki/pes/gameplay-combos-fl26.md` (or `bpb-2026.md` for the laptop). The AIKB tables list known hashes per release; match by hash, not by filename.

4. Inventory the **full gameplay stack** before drawing any conclusion:

   - `Data/dt13_all.cpk` and `Data/dt18_all.cpk`
   - `FL_2026.exe`
   - `SiderAddons\sider.ini`: every `lua.module` entry tagged or commented as gameplay, every gameplay-related `cpk.root` entry
   - Hook files, exe replacements
   - `SYSTEM00000000` cache state

   Effective gameplay = what wins after load order, not just everything wired. State both when they diverge.

## Pitfalls

- **Filename is not identity.** A file called `dt18_all.cpk.f4l` proves nothing; hash it.
- **Never edit individual mod values.** Add or remove mods as whole units; do not tweak a single line inside someone else's lua.
- **Anti-Cheat.lua is debunked.** Holland's `Anti-Cheat.lua` has no valid AOB addresses; use `Dynamic_Difficulty` instead.
- **DT18 also unlocks animations**, not just DT13. alexfe87's DT18 is the animation source for that combo.
- **Do not propose changes after a status check** unless the user explicitly approves the install/change step. Evidence-first only.
- **Status checks are read-only.** Report `SYSTEM00000000` presence as a fact; its delete is hard-banned (repo `AGENTS.md`).

## Out of scope

- Performing a gameplay switch: see pes-gameplay-switch.
- Writing the verdict to the AIKB: see pes-aikb-log.
- Editing sider config to add or remove modules.
