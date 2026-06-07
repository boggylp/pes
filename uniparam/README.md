# uniparam

Edits the PES / Football Life uniform databases `UniColor.bin` and `UniformParameter.bin` (both Konami WESYS + zlib envelopes), to extend a team's pre-match kit-slot count.

The pre-match "Kits" picker shows as many kits per team as the team's slot count in **`UniColor.bin`**. `UniformParameter.bin` only holds the kit config records. Adding a kit slot means: bump the UniColor count + add a UniformParameter config + supply the per-team `realUni` file, then deliver all three via Sider livecpk (no cpk repack). Kitserver fills the new slot's texture from its `pN` folder.

## Build

```sh
cd uniparam
go build .
```

## Commands

```sh
# Add a player kit slot to a team in UniColor.bin (in-place; bumps count, inserts
# a player entry where the GK sits, shifts GK into the padding).
uniparam unicolor-addslot UniColor.bin UniColor.bin <teamId>

# Add a <teamId>_DEF_<dst>th config to UniformParameter.bin. With a donor record
# file (a real Nth-slot realUni, carrying the correct selectable-kit flag bytes),
# that record is embedded; otherwise the team's srcSlot record is cloned+retextured.
uniparam add-slot UniformParameter.bin UniformParameter.bin <teamId> <srcSlot> <dstSlot> [donorRecord.bin]

# Rewrite the texture base name inside a raw realUni .bin (same length required).
uniparam retex <src>.bin <out>.bin u<id>p<src> u<id>p<dst>

# Dump the inflated body of a WESYS+zlib file (for inspection).
uniparam inflate <in> <out>
```

## End-to-end: extend a team to N player kits

```sh
G="/path/to/SP Football Life 2026"; UT="$G/SiderAddons/livecpk/root/common/character0/model/character/uniform/team"
ID=272; N=4
# donor: this team's real 3rd realUni from the cpk, retextured to pN
../cpk/cpk extract --file "common/character0/model/character/uniform/team/$ID/${ID}_DEF_3rd_realUni.bin" --out 3rd.bin "$G/Data/dt34_g4.cpk"
uniparam retex 3rd.bin "$UT/$ID/${ID}_DEF_${N}th_realUni.bin" u0${ID}p3 u0${ID}p${N}
uniparam add-slot "$UT/UniformParameter.bin" "$UT/UniformParameter.bin" $ID 3 $N "$UT/$ID/${ID}_DEF_${N}th_realUni.bin"
uniparam unicolor-addslot "$UT/UniColor.bin" "$UT/UniColor.bin" $ID
```

The kserv pack must already have the team's `pN` kit folder/texture. Verified on `BogambeDesktop` (FL26 v2.2-26.2.0.3): Hajduk 2->3, Partizan/Dinamo 3->4.

## File formats

`UniColor.bin`: per-team 85-byte records — `id u32`, `count u8` (used slots incl. GK), then 10 entries of `[marker u16][color1 3B][color2 3B]`, padded. Markers `0x00..0x03` player, `0x10` GK1st.

`UniformParameter.bin`: `entry_count u32`, then a 12-byte index (`recOff,recSize,nameOff`, absolute) and a name table + 120-byte slot records. See `wiki/pes/kitserver-uniform-slots.md` in the AI knowledge base for the full mechanism.

## Caveat

For some licensed teams in community patches the slot count is synced to `EDIT00000000` and read from there; a UniColor edit won't show. Re-check per team. Not the case for the FL26 base/Balkan teams tested here.
