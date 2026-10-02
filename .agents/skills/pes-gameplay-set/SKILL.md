---
name: pes-gameplay-set
description: 'Set the complete live PES or Football Life gameplay stack from local archives or researched release instructions.'
---

# Set PES gameplay

> Simplicity is the soul of efficiency. — Austin Freeman

## 1. Inventory the current stack

Load `pes-gameplay-status`.

## 2. Resolve the target stack

- Take the package from `%USERPROFILE%\MEGA\gaming\pes\gameplay\`, highest matching version first, and load `pes-evoweb-research` only when no local package matches.
- Read its install instructions and its memory-wiki article.
- Install only the components the user named, and ask before adding one the package merely recommends.

## 3. Apply the stack

- For a vanilla dt13 or dt18, run `pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla` or `switch-dt18-vanilla`.
- For another release, follow its install instructions in one pass.
- Remove a layer by deleting its files and its `sider.ini` lines.
  - Comment a layer out only when the user asks for a trial.
- Mark each manual `sider.ini` block with one `; [GB-CUSTOM]` line.
- Update the `; gameplay:` tracking line.

## 4. Verify the result

- Load `pes-gameplay-status` again and check every active component by hash.
- List the livecpk and module directories, because the status helper reads only Sider entries.
- Delete the temporary extraction after reporting the result.
