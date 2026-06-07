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

# Extract a subset by inner-path prefix (one pass; exits non-zero if nothing matches)
./cpk extract --prefix "Asset/model/character/uniform/texture/#windx11/u0272" --out extracted/ Data/dt34_g4.cpk
```

The matcher used by `--file` is case-insensitive and accepts both forward and backslashes, so `common/etc/pesdb/Player.bin` and `Common\Etc\Pesdb\Player.bin` resolve to the same entry.

## Assembling a kitserver pack

`cpk kits` reads uniform textures straight from a patch's uniform cpk and merges them with a kitserver config tree (the per-team `config.txt` / `order.ini` folders a patch ships under `sider/content/kit-server/`) into a self-contained pack: each `<League>/<Team>/{p1..,g1..}` slot folder gets its config plus the `.ftex` files that config references, and a pack `map.txt` is written.

```sh
cpk kits \
  --kserv-src "/path/to/patch/sider/content/kit-server" \
  --leagues "Mozzart Bet Super liga Srbije,SuperSport HNL" \
  --out pack/ \
  "/path/to/patch/Data/dt34_g4.cpk"
```

- `--kserv-src` holds `map.txt` (lines `team-id, "League\Team"`) and the `<League>/<Team>/` config folders. Override the map with `--map`.
- `--leagues` is comma-separated and repeatable; omit to include every league in the map.
- Team IDs in the output `map.txt` are the source patch's IDs. Importing into a different game needs IDs that game actually uses (base-Konami clubs share IDs; patch-custom clubs do not).
- The texture source must be the cpk that holds those teams' textures (the patch's own `dt34_g4.cpk`), not another game's.

## Where Player.bin lives

In a vanilla PES 2021 install Konami's player base is at `common/etc/pesdb/Player.bin` inside `Data/dt00_x64.cpk` (the base game data) and is overridden by `Data/dt10_x64.cpk` and the DLC `download/dt80_*E_x64.cpk` packs in load order.

The file is encrypted at-rest with Konami's pesdb key. The cpk tool extracts the raw encrypted bytes; decryption is handled by external editors (ejogc327's PES 2020 Editor, kisni07's PESDatabase) which apply Konami's key on load. See `wiki/pes/bpb-2026.md` in the AI knowledge base for the BPB 2026 / FL26 specifics.

## ContentOffset normalisation

Modern PES season packs (e.g. FL26's `download/data_s2526*.cpk`, BPB's `Data/dt*.cpk`) put the TOC at the end of the file with `ContentOffset == TocOffset` and store FileOffsets as absolute. The reader detects this case (any computed offset would land past EOF) and re-bases entries with `contentOffset=0`. Without this fix the tool silently extracts zero bytes from past-EOF reads — the symptom that masked real `Player.bin` data as "zero-filled" before 2026-05-06.
