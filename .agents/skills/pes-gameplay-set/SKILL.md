---
name: pes-gameplay-set
description: 'Set the complete live PES or Football Life gameplay stack from local archives or researched release instructions.'
---

# Set PES gameplay

> Simplicity is the soul of efficiency. — Austin Freeman

Apply one complete requested stack. Never leave an unrequested mixed state.

## 1. Resolve the request

- Use the live-install path that the user gave.
- Reuse a known active path only when the request omits one.
- Search for another install only when that path fails or the user asks.
- Treat preview, version, and issue labels as package names unless the user names a separate install root.

A direct install or switch command approves the necessary live changes for that named stack in the current turn.
Ask one focused question only when the package, destination stack, overwrite, or rollback source is unresolved.

## 2. Inventory the current stack

Load `pes-gameplay-status`. Record:

- dt13 and dt18 hashes;
- gameplay livecpk roots;
- gameplay `lua.module` entries;
- the executable and hooks;
- cache state.

Stop when the request would leave an unresolved mixed stack.

## 3. Resolve the target stack

Look first under `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.

- Prefer the highest matching version or newest matching issue already present.
- Read the package's install instructions and matching memory-wiki article.
- Use external research only when local evidence cannot identify the release or the user requests it.
- Install only components required by the requested package.
- Preserve verified vanilla components that the package does not replace.
- Exclude gameplay files, livecpk roots, and Lua modules outside the requested stack.
- Treat `pure` or `nothing else` as excluding optional support files and add-ons.
- Confirm an archive can restore every replaced or removed component.

Stop before an overwrite when the prior live file has no rollback source.

## 4. Apply the stack

For a vanilla dt13 or dt18 switch, use the helper:

```powershell
pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla
pwsh -File .\tools\fl-gameplay.ps1 switch-dt18-vanilla
```

For another release:

1. Follow its bundled install instructions.
2. Apply every required file and Sider entry in one pass.
3. Remove active gameplay layers excluded by the approved stack.
4. Mark each manual `sider.ini` block with one `; [GB-CUSTOM]` line.
5. Update the `; gameplay:` tracking line.

Leave the SYSTEM cache present.

## 5. Verify the result

Load `pes-gameplay-status` again.

- Verify every active component by hash.
- Confirm all Sider entries match the requested destination.
- Confirm no unrequested gameplay layer remains.
- Report the result before removing temporary extraction files.
- Keep downloaded archives in the MEGA gameplay location.

## Review

- Treat every file named as gameplay by the mod instructions as part of the stack.
- Use existing archives as rollback sources.
- Create an additional backup only when the user requests it.
- Store requested gameplay backups under `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.
- Keep state and history in the `; gameplay:` line and memory wiki.
- Follow the repository cache-safety rule.

## Boundaries

- Use `pes-gameplay-status` to read current state.
- Use `pes-evoweb-research` when current community evidence is required.
- Use `pes-wiki-log` to record a playtest verdict.
