# Uniform parameters

Edit PES and Football Life `UniColor.bin` and `UniformParameter.bin` files to add a pre-match kit slot.

`UniColor.bin` controls the visible kit count. `UniformParameter.bin` supplies kit configuration. Add the team `realUni` file and load all three files through Sider livecpk. Kitserver supplies the texture from the team's `pN` folder.

## Build

```sh
cd uniparam
go build .
```

## Commands

Add a player kit entry to `UniColor.bin`. Use the same input and output path for an in-place update.

```sh
uniparam unicolor-addslot <in> <out> <team-id>
```

Add a 1st to 4th kit configuration to `UniformParameter.bin`. With a donor record, the command embeds that record. Otherwise, it clones the source record and changes its texture name.

```sh
uniparam add-slot <in> <out> <team-id> <source-slot> <destination-slot> [donor-record.bin]
```

Rewrite every matching texture base name in a raw `realUni` file. The old and new names must have the same length.

```sh
uniparam retex <source.bin> <output.bin> <old-texture> <new-texture>
```

Inflate a WESYS and zlib file for inspection:

```sh
uniparam inflate <input> <output>
```

## Example

Extend team 272 to four player kits:

```sh
G="/path/to/SP Football Life 2026"
UT="$G/SiderAddons/livecpk/root/common/character0/model/character/uniform/team"
ID=272
N=4

../cpk/cpk extract \
  --file "common/character0/model/character/uniform/team/$ID/${ID}_DEF_3rd_realUni.bin" \
  --out 3rd.bin \
  "$G/Data/dt34_g4.cpk"

uniparam retex 3rd.bin "$UT/$ID/${ID}_DEF_${N}th_realUni.bin" u0${ID}p3 u0${ID}p${N}
uniparam add-slot "$UT/UniformParameter.bin" "$UT/UniformParameter.bin" $ID 3 $N "$UT/$ID/${ID}_DEF_${N}th_realUni.bin"
uniparam unicolor-addslot "$UT/UniColor.bin" "$UT/UniColor.bin" $ID
```

The Kitserver pack must already contain the team's `pN` folder and texture.

## Formats

`UniColor.bin` contains 85-byte team records. Each record contains a team ID, a used-slot count, ten color entries, and padding. Player markers range from `0x00` to `0x03`. `0x10` marks the first goalkeeper slot.

`UniformParameter.bin` contains an entry count, a 12-byte index per record, a name table, and 120-byte slot records.

For some licensed teams in community patches, the EDIT save controls the slot count. A `UniColor.bin` change will then not appear. Check the target team in the game.
