---
name: pes-reshade
description: 'Install, update, remove, configure, or diagnose ReShade for PES 2021 or Football Life.'
metadata:
  trusted_sources:
    - https://reshade.me/
    - https://github.com/crosire/reshade/blob/main/setup/MainWindow.xaml.cs
---

# PES ReShade

## Steps

1. Read the preset instructions.

2. Read the current official ReShade installer documentation or source.

3. Extract the preset's active `Techniques` list. Resolve each active shader to its source repository.

4. Use the user-provided game root. Read its `AGENTS.md` and `README.md` when present.

5. Select the game executable, not its launcher.

6. Ask the user to close the game and Sider before changing ReShade files. Do not stop unrelated processes.

7. Record versions and hashes for the game executable, hook DLLs, `ReShade.ini`, preset, and `reshade-shaders` directory.

8. Create a rollback archive with `mem-backup`. Record the official headless uninstall command and custom files that it does not remove.

9. Download the current official installer. Verify its signer and product version against the advertised release.

10. Select DirectX 10, 11, or 12 for PES 2021 and Football Life. This installs the DXGI hook.

11. Prefer interactive setup with the preset selected. For headless setup, install active shaders, includes, and textures from their source repositories.

12. Set the preset path. Enable performance mode. Skip disabled effects. Keep Home for the overlay and assign an unused effects-toggle key.

13. Ask the user to launch with effects disabled and repeat the normal team-selection-to-match path. Stop when the game exits.

14. Enable the preset. Confirm that every active shader compiles in `ReShade.log`.

15. Ask the user to repeat the same path. Compare effects on and off in the same scene. Read the resulting logs and measurements.

16. For crash isolation, disable only the ReShade hook and repeat the same path. Then test the runtime without effects and with the preset. Change one variable per run.

17. Capture `ReShade.log`, `SiderAddons/sider.log`, and the Windows Application event before the next change.

## Review

- Confirm that the hook version is the current official stable release.
- Confirm that the game executable hash is unchanged.
- Confirm that every active technique compiles.
- Confirm that the team-selection-to-match path passes with the hook and preset enabled.
- Confirm that the effects toggle changes the image.
- Report the measured performance cost.

## Boundaries

- Treat startup and shader compilation as partial checks, not match-load verification.
- Keep stadium light files and lookup-table roots unchanged during ReShade isolation.
- Disable expensive effects first when frame time increases.
