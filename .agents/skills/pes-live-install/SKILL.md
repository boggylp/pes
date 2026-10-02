---
name: pes-live-install
description: 'Install or restore a PES or Football Life patch stack, write an EDIT save, or add a custom player to a live install.'
---

# PES live install

## Steps

1. Load `pes-gameplay-status`, and record its result with the `hostname` result.
2. List the patch parts in the order its README or Evoweb thread documents.
3. Apply one part at a time, and verify it before the next.
   - Skip a standalone EDIT save when a later part already contains it.
   - Stop when a write has no rollback source among the staged parts and backup trees.
4. Write an EDIT save only to `%USERPROFILE%\Documents\KONAMI\eFootball PES 2021 SEASON UPDATE\2026\save\EDIT00000000`, and treat MEGA copies as sources.
5. Import a custom player as a complete record with the configured EDIT-save editor: attributes, physical data, traits, name, and kit data.
6. When player IDs or face folders change, load `pes-roster-extract` and `pes-faces-set`.
7. Keep Hardware-accelerated GPU scheduling off.
8. When a step needs elevation, run the approved script from this session:

   ```powershell
   Start-Process (Get-Command pwsh.exe).Source -Verb RunAs -Wait -ArgumentList @('-NoProfile','-ExecutionPolicy','Bypass','-File',<script>)
   ```

9. Load `pes-gameplay-status` again, and confirm the intended hashes and Sider entries.
10. Ask the user to launch the game, and confirm the save loads without a create-edit-data prompt.
