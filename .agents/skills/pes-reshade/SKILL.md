---
name: pes-reshade
description: 'Install, update, remove, configure, or diagnose ReShade for PES 2021 or Football Life.'
metadata:
  trusted_sources:
    - https://reshade.me/
    - https://github.com/crosire/reshade/blob/main/setup/MainWindow.xaml.cs
---

# PES ReShade

The **match path** is the user's normal route from team selection into a match. Startup and shader compilation alone do not verify it.

## Prepare

1. Read the preset instructions and the current official ReShade installer documentation or source.
2. Resolve each shader in the preset's active `Techniques` list to its source repository.
3. Target the game executable, not its launcher.
4. Ask the user to close the game and Sider, and stop no other process.
5. Record versions and hashes of the executable, hook DLLs, `ReShade.ini`, preset, and `reshade-shaders`.
6. Archive them through `mem-artifact`, and record the official headless uninstall command and the custom files it leaves behind.

## Install

1. Download the current official stable installer, and verify its signer and product version.
2. Select DirectX 10, 11, or 12, which installs the DXGI hook.
3. Prefer interactive setup with the preset selected.
   - Headless, install the active shaders, includes, and textures from their source repositories.
4. Set the preset path, enable performance mode, skip disabled effects, keep Home for the overlay, and bind an unused effects-toggle key.
5. Treat the setup package and the per-game runtime as separate.
   - Uninstall a stale setup package only with approval, then confirm the hook, configuration, shaders, and preset are unchanged.

## Verify

1. Ask the user to run the match path with effects off, and stop when the game exits.
2. Enable the preset, and confirm every active shader compiles in `ReShade.log`.
3. Ask the user to run the match path again, then compare effects on and off in one scene.
4. Confirm the executable hash is unchanged and the toggle changes the image.
5. Report the measured cost, and disable the expensive effects first when frame time rises.

## Isolate a crash

- Change one variable per run: hook disabled, then runtime without effects, then the preset.
- Keep stadium light files and lookup-table roots unchanged.
- Capture `ReShade.log`, `SiderAddons/sider.log`, and the Windows Application event before the next change.
