# CPK

Read CRI Middleware CPK packages. List table contents, extract files, and assemble kitserver packs. The tool decompresses CRILAYLA files when necessary. PES 2021, Football Life, and BPB CPKs normally store files uncompressed.

## Build

```sh
cd cpk
go build .
```

## Commands

```sh
# Print the table of contents
./cpk list "/path/to/Data/dt40_all.cpk"

# Print stored size, extracted size, offset, and path
./cpk list -l Data/dt40_all.cpk

# Extract one file by inner path
./cpk extract --file common/etc/pesdb/Player.bin --out Player.bin Data/dt00_x64.cpk

# Extract all files
./cpk extract --out extracted/ Data/dt40_all.cpk

# Extract files with an inner-path prefix
./cpk extract --prefix "Asset/model/character/uniform/texture/#windx11/u0272" --out extracted/ Data/dt34_g4.cpk
```

`--prefix` exits non-zero when no file matches. Inner-path matching ignores case and accepts forward or backslashes.

## Kitserver packs

`cpk kits` combines uniform textures from a CPK with a source kitserver configuration tree. It copies each selected slot configuration, adds the referenced FTEX files, and writes a pack `map.txt`.

```sh
cpk kits \
  --kserv-src "/path/to/patch/sider/content/kit-server" \
  --leagues "Mozzart Bet Super liga Srbije,SuperSport HNL" \
  --out pack/ \
  "/path/to/patch/Data/dt34_g4.cpk"
```

- `--kserv-src` contains `map.txt` and the `<League>/<Team>/` configuration folders.
- `--map` selects another map file.
- `--leagues` accepts a comma-separated list and can be repeated. Omit it to include all mapped leagues.
- `--label` sets the first comment line in the output `map.txt`.
- The output map keeps source-patch team IDs. Verify IDs before import into another game.
- Use the uniform CPK from the source patch because it contains the referenced textures.

## Player.bin

A standard PES 2021 load order can contain `common/etc/pesdb/Player.bin` in these locations:

1. `Data/dt00_x64.cpk`
2. `Data/dt10_x64.cpk`
3. `Data/download/dt80_*E_x64.cpk`

Later entries override earlier entries. An active livecpk database can override all archive copies. Use [pesdb](../pesdb/README.md) to decode the WESYS and zlib data.

## Offset handling

Some modern season packs store the table of contents at the end of the CPK. They can set `ContentOffset` equal to `TocOffset` and use absolute file offsets. When computed offsets exceed the file size but raw offsets are valid, the reader uses a content offset of zero.
