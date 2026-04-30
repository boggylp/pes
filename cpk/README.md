# cpk

CRI Middleware CPK archive reader. Lists the table of contents and extracts files (with CRILAYLA decompression when present).

PES 2021 / Football Life / BPB cpks store files uncompressed, so the decompressor is a fallback for older patches and CRI-based games beyond PES.

## Build

```sh
cd cpk
go build .
```

## Usage

```sh
# Print TOC
./cpk list "/d/SteamLibrary/steamapps/common/eFootball PES 2021/Data/dt40_all.cpk"

# Print TOC with offsets and sizes (compressed and decompressed)
./cpk list -l Data/dt40_all.cpk

# Extract one file by inner path (DirName + FileName as listed in TOC)
./cpk extract --file common/etc/pesdb/Player.bin --out Player.bin Data/dt00_x64.cpk

# Extract everything to a directory
./cpk extract --out extracted/ Data/dt40_all.cpk
```

The matcher used by `--file` is case-insensitive and accepts both forward and backslashes, so `common/etc/pesdb/Player.bin` and `Common\Etc\Pesdb\Player.bin` resolve to the same entry.

## Where Player.bin lives

In a vanilla PES 2021 install Konami's player base is at `common/etc/pesdb/Player.bin` inside `Data/dt00_x64.cpk` (the base game data) and is overridden by `Data/dt10_x64.cpk` and the DLC `download/dt80_*E_x64.cpk` packs in load order.

In the BPB 2026 install on this machine both `dt00_x64.cpk` and `dt10_x64.cpk` ship a fully-zeroed `Player.bin` (1,751,422 bytes, all `00`). The actual player database lives in the EDIT save file at:

```
~/Documents/KONAMI/eFootball PES 2021 SEASON UPDATE/<SteamID>/save/EDIT00000000
```

That file is encrypted and out of scope for this tool — see `wiki/pes/bpb-2026.md` in the AI knowledge base for the full picture.
