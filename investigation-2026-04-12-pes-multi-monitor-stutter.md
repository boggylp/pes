---
path: ~/veemon/journal/priv/2026-04-12_pes-multi-monitor-stutter.md
session_id: 2026-04-12_priv_pes-multi-monitor-stutter
started_at: 2026-04-12T00:00:00Z
updated_at: 2026-04-12T00:00:00Z
workspace: priv
repos:
  - pes
  - ai-knowledge-base
scope: Investigate PES 2021 / Football Life multi-monitor stutter root cause
status: completed
ticket:
---

# Investigate PES 2021 / Football Life Multi-Monitor Stutter Root Cause

## Current

Investigation completed. PresentMon finally provided a concrete root cause: `FL_2026.exe` is presenting through `Composed: Copy with GPU GDI`, while the browser is using `Composed: Flip`. That bad PES present path is the strongest confirmed explanation for the second-monitor stutter. The user also confirmed the same issue existed on the original PES 2021 exe, which points to stock engine behavior instead of a Smoke Team-specific change.

## Done

- Checked AIKB article and compared it against live system telemetry instead of trusting it blindly.
- Sampled `nvidia-smi`, per-process GPU engine counters, CPU counters, and process/window state while `FL_2026.exe` was running.
- Confirmed `FL_2026.exe` is the dominant 3D GPU workload and browser video decode is tiny by comparison.
- Confirmed the issue reproduces across fullscreen, windowed, and borderless configurations.
- Confirmed the user had the same issue earlier with two identical monitors, so mixed refresh is not the root cause.
- Found similar reports online for PES 2021 stutter and for Windows 11 second-monitor video stutter while a focused game is running.
- Captured a PresentMon trace showing `FL_2026.exe` on `Composed: Copy with GPU GDI`, `brave.exe` on `Composed: Flip`, and `dwm.exe` on `Hardware: Legacy Flip`.
- Verified that the high-DPI compatibility override did not change the present mode.
- Ruled out `dgVoodoo2` as the wrong tool for this game because the active API stack is `D3D11 + DXGI`, not D3D9 or older.
- Tried a local-only `Special K` `dxgi.dll` shim and confirmed it conflicts with `sider.dll` on this setup, preventing normal launch.
- Corrected stale or misleading AIKB wording about MPO in the PES index and article.

## Next

- If continuing later, investigate whether the game can be launched without `sider.dll` purely as a diagnostic to see whether stock PES and `sider` differ in present path behavior.

## Blockers

- None

## Decisions

- Treat the root cause as a bad compositor path, `Composed: Copy with GPU GDI`, rather than a generic performance problem, because PresentMon directly observed the path difference between PES and the browser.
- Update AIKB in place instead of adding a new article, because the existing article was close but had stale system details and an inconsistent MPO summary in the index.

## Log

### 2026-04-12T00:00:00Z

Early telemetry showed `FL_2026.exe` holding most of the 3D engine and pushing GPU utilization to `98% to 100%`, which initially looked like simple starvation.

### 2026-04-12T00:30:00Z

Later corrections from the user ruled out borderless-only behavior, fullscreen-only behavior, and mixed-refresh-only behavior. The strongest remaining explanation is foreground presentation scheduling: when PES has focus the second-monitor video stutters, when focus moves away both become smooth. Browser decode load remains small, so the browser is not the heavy part of the interaction.

### 2026-04-12T01:42:00Z

An elevated PresentMon capture finally gave a hard result. `FL_2026.exe` presented only as `Composed: Copy with GPU GDI` in the sample, while `brave.exe` presented as `Composed: Flip` and `dwm.exe` as `Hardware: Legacy Flip`. That gives a concrete bad path to target and makes DPI-scaling or bitmap-scaling style causes more plausible than earlier generic scheduler guesses.

### 2026-04-12T01:55:00Z

The user confirmed the same symptom existed on the original PES 2021 exe. That makes the root cause much more likely to be stock engine presentation behavior rather than Smoke Team packaging or mod-specific behavior.

### 2026-04-12T02:05:00Z

High-DPI compatibility override was tested and did not change the PresentMon mode. A fresh capture still showed `FL_2026.exe` on `Composed: Copy with GPU GDI`.

### 2026-04-12T02:12:00Z

`dgVoodoo2` was ruled out because the game is running on `D3D11 + DXGI`, not the older APIs dgVoodoo wraps.

### 2026-04-12T02:18:00Z

A local-only `Special K` `dxgi.dll` shim was tested as a reversible per-game experiment. It caused the game to fail to launch correctly. `logs\dxgi.log` showed repeated non-continuable exceptions from `SiderAddons\sider.dll` after the hook chain initialized, so `Special K + sider` is not a safe combination on this setup.
