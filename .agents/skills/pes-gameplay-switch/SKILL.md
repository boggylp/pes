---
name: pes-gameplay-switch
description: 'Prepare or perform a live PES or Football Life gameplay-stack switch when the user asks to replace gameplay files or Sider entries.'
---

# PES gameplay switch

## Steps

1. Run `pes-gameplay-status`. Record current hashes and active Sider entries.

2. Inventory every component that can affect gameplay. Determine the complete intended stack. Stop when the request would leave an unapproved mixed stack.

3. Resolve the source and destination files. Calculate their hashes.

4. Ask the user to approve the exact change. State each target path and its source-hash to destination-hash transition.

5. For a vanilla dt13 or dt18 switch, use the helper.

   ```powershell
   pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla
   pwsh -File .\tools\fl-gameplay.ps1 switch-dt18-vanilla
   ```

   The helper saves a timestamped copy under `<GameRoot>\.backup\Data\`, restores the selected vanilla source, and updates the `; gameplay:` tracking line.

6. For another release, follow its install instructions and write directly to the approved live paths.

7. Apply every file and Sider entry required for the approved stack.

8. Run `pes-gameplay-status` again. Confirm that all live hashes and entries match the approved destination.

## Review

- Use the install path that the user gave.
- Treat every file bundled as gameplay by the mod instructions as part of the stack.
- Use existing archives as rollback sources.
- Create an additional backup only when the user requests it. Store a requested gameplay backup under `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.
- Stop before an overwrite when the prior live file has no rollback source.
- Follow the repository cache-safety rule.
- Mark each contiguous manual `sider.ini` block with one `; [GB-CUSTOM]` line. Put state and history in the `; gameplay:` tracking line and the AI Knowledge Base.

## References

- [Sider configuration and cache behavior](https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/)
- Gameplay archives: `%USERPROFILE%\MEGA\gaming\pes\gameplay\`

## Boundaries

- Use `pes-gameplay-status` to read current state.
- Use `pes-evoweb-research` to assess a release.
- Use `pes-wiki-log` to record a playtest verdict.
