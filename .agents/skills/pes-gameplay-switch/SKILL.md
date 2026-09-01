---
name: pes-gameplay-switch
description: 'Prepare or perform a live PES or Football Life gameplay-stack switch when the user asks to replace gameplay files or Sider entries.'
---

# PES gameplay switch

## Steps

1. When the user names a complete target stack and matching archives already exist in the MEGA gameplay location, load `pes-set-gameplay` and stop.

2. Run `pes-gameplay-status`. Record current hashes and active Sider entries.

3. Inventory every component that can affect gameplay. Determine the complete intended stack. Stop when the request would leave an unapproved mixed stack.

4. Resolve the source and destination files. Calculate their hashes.

5. Ask the user to approve the exact change. State each target path and its source-hash to destination-hash transition.

6. For a vanilla dt13 or dt18 switch, use the helper.

   ```powershell
   pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla
   pwsh -File .\tools\fl-gameplay.ps1 switch-dt18-vanilla
   ```

   The helper saves a timestamped copy under `<GameRoot>\.backup\Data\`, restores the selected vanilla source, and updates the `; gameplay:` tracking line.

7. For another release, follow its install instructions and write directly to the approved live paths.

8. Apply every file and Sider entry required for the approved stack.

9. Run `pes-gameplay-status` again. Confirm that all live hashes and entries match the approved destination.

## Review

- Use the install path that the user gave.
- Treat every file bundled as gameplay by the mod instructions as part of the stack.
- Use existing archives as rollback sources.
- Create an additional backup only when the user requests it. Store a requested gameplay backup under `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.
- Stop before an overwrite when the prior live file has no rollback source.
- Follow the repository cache-safety rule.
- Mark each contiguous manual `sider.ini` block with one `; [GB-CUSTOM]` line. Put state and history in the `; gameplay:` tracking line and the memory wiki.

## References

- [Sider configuration and cache behavior](https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/)
- Gameplay archives: `%USERPROFILE%\MEGA\gaming\pes\gameplay\`

## Boundaries

- Use `pes-set-gameplay` for a named stack whose archives are already local.
- Use `pes-gameplay-status` to read current state.
- Use `pes-evoweb-research` when local evidence cannot establish the requested release or the user asks for current community evidence.
- Use `pes-wiki-log` to record a playtest verdict.
