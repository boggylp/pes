---
name: pes-set-gameplay
description: 'Set the active Football Life gameplay from existing MEGA archives when the user names the desired release or issue and expects it installed immediately.'
---

# Set PES gameplay

> Simplicity is the soul of efficiency. — Austin Freeman

A **set-gameplay request** installs a named local gameplay package into the known live install without release research or another approval round.

## Steps

1. Resolve the live install.
   - Treat preview, version, and issue labels as package names unless the user explicitly supplies a separate install root.
   - Reuse the known active live-install path.
   - Search for another install only when that path fails or the user asks.

2. Load `pes-gameplay-status`. Record dt13, dt18, gameplay livecpk roots, gameplay `lua.module` entries, the executable, hooks, and cache state.

3. Resolve the named package under `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.
   - Prefer the highest matching version or newest matching issue archive already present.
   - Read its bundled install instructions and the matching memory-wiki article.
   - Use external research only when local evidence cannot identify the requested package or the user asks for current community evidence.

4. Define the complete destination stack.
   - Install only the requested package components.
   - Preserve verified vanilla components the request does not replace.
   - Remove gameplay files, livecpk roots, and gameplay Lua files outside the requested stack.
   - Treat `pure` or `nothing else` as excluding optional support files and add-ons.
   - Confirm that existing archives can restore every replaced or removed component.

5. Treat the user's direct install or switch command as approval for the necessary live file and configuration changes.
   - Ask one focused question only when the package, destination stack, or rollback source remains unresolved.

6. Apply the whole destination stack in one pass. Update the `; gameplay:` line and mark each manual `sider.ini` block with one `; [GB-CUSTOM]` line.

7. Load `pes-gameplay-status` again. Verify every active component by hash and confirm that no unrequested gameplay layer remains.

## Review

- Leave the SYSTEM cache present.
- Keep downloaded archives in the MEGA gameplay location.
- Report the result before removing temporary extraction files.

## Transitions

- When the requested package is missing or the destination stack remains incomplete, load `pes-gameplay-switch` and stop.
- When the user gives a playtest verdict, load `pes-wiki-log` and stop.
- Otherwise, stop.
