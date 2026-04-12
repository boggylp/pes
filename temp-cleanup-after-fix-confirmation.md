# Temporary Cleanup After Fix Confirmation

Delete these temporary investigation artifacts after the fix is confirmed:

- `pes-trace.etl`
- `pes-trace.csv`
- `pes-trace-summary.xml`
- `summary.txt`
- `pes-trace.log`
- `pes-trace-stop.log`
- `presentmon.csv`
- `presentmon.log`
- `sudo-test.txt`

Keep these if the investigation notes and reusable capture scripts are still wanted:

- `investigation-2026-04-12-pes-multi-monitor-stutter.md`
- `tools/presentmon-capture.ps1`
- `tools/presentmon-capture-current.ps1`
- `tools/wpr-capture-pes-trace.ps1`
- `tools/wpr-list-status.ps1`
- `tools/wpr-stop-pes-trace.ps1`

Current reversible per-game experiment in the game folder:

- `C:\Program Files (x86)\SP Football Life 2026\dxgi.dll` (Special K local shim)
- delete it to remove the Special K experiment
- if a backup file ever appears as `dxgi.dll.pre-specialk.bak`, restore that instead of deleting blindly
