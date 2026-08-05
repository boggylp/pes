---
name: pes-reshade
description: 'Invoke BEFORE installing, updating, removing, configuring, or diagnosing ReShade or a ReShade preset for PES 2021 or Football Life. Covers the official installer, preset shader resolution, Sider-safe wiring, rollback, crash isolation, and runtime verification.'
metadata:
  trusted_sources:
    - https://reshade.me/
    - https://github.com/crosire/reshade/blob/main/setup/MainWindow.xaml.cs
---

# PES ReShade

Install one verified ReShade runtime and one preset without changing gameplay files. Prove startup, team selection, and match loading before declaring it stable.

## Steps

1. Read the preset instructions and the current official ReShade installer documentation or source. Extract the preset's active `Techniques` list. Resolve each active shader to its owning repository.

2. Use the user-provided game root. Read its `AGENTS.md` and `README.md` when present. Target the real game executable, such as `FL_2026.exe`, not its launcher.

3. Ask the user to close the game and Sider normally before changing a hook DLL, configuration, preset, or shader. Do not launch or stop the game, Sider, an installer, a terminal, or a subagent.

4. Record hashes and versions for the game executable and any existing `dxgi.dll`, `d3d11.dll`, `ReShade.ini`, preset, and `reshade-shaders` directory. Create a rollback archive with mem-backup. Include the official headless uninstall command and any custom files it does not remove.

5. Check the official ReShade site at runtime. Compare the setup file's product version and signer with the advertised release. Treat package-manager versions as cache entries; use the official installer when the package manifest lags.

6. Select DirectX 10/11/12 for PES 2021 and Football Life. This installs the DXGI hook. Prefer the official interactive setup with the preset selected so setup resolves effect packages. Headless setup installs the runtime and base configuration only; install the active shaders and shared includes separately from their owning repositories.

7. Install every active shader in `Techniques`, its includes, and required textures. Set the preset path. Enable performance mode and skip disabled effects. Keep Home for the overlay and assign an unused effects-toggle key.

8. Ask the user to launch with effects disabled and repeat the normal team-selection-to-match path. Stop if the game exits.

9. Enable the preset. Confirm every active shader compiled successfully in `ReShade.log`, with no shader error. Ask the user to repeat the same team-selection-to-match path and compare effects on and off in the same scene. Read the resulting logs and measurements yourself.

10. Disable only the ReShade hook and repeat the same user path. If that passes, keep other visual roots unchanged and test the runtime without effects, then with the preset. Change one variable per run. Capture `ReShade.log`, `SiderAddons/sider.log`, and the Windows Application event before the next change.

## Pitfalls

- Validate this skill inline. Do not open a fresh session, reviewer, pane, or terminal.
- Startup and shader compilation do not prove match loading is stable.
- A preset can reference a shader absent from the current setup package list. Resolve it from the owning repository.
- Slow multi-pass sharpening is the first effect to disable when frame time rises.
- ReShade modifies the rendered image. Stadium light files and LUT roots remain separate layers.

## Verification

- The hook version equals the current official stable release.
- The game executable hash is unchanged.
- Every active technique compiles successfully.
- The same team-selection-to-match path passes with the hook and preset enabled.
- The effects toggle changes the image and the measured performance cost is acceptable.
