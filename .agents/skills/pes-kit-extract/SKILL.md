---
name: pes-kit-extract
description: 'Extract existing PES kits from a uniform CPK or assemble a self-contained kitserver pack.'
---

# PES kit extract

## Steps

1. Build the CPK tool from the repository root.

   ```sh
   cd cpk
   go build .
   cd ..
   ```

2. Identify the uniform CPK for the source patch. Confirm that it contains uniform entries.

   ```sh
   ./cpk/cpk list <uniform.cpk> | rg -c 'uniform'
   ```

3. Resolve each team ID from the source patch's kitserver `map.txt` or team list. Verify the destination game uses the intended ID.

4. Extract one team's definitions and textures when a full pack is not required. Pad the texture team ID to four digits.

   ```sh
   ./cpk/cpk extract --prefix "common/character0/model/character/uniform/team/<id>/" --out <dir> <uniform.cpk>
   ./cpk/cpk extract --prefix "Asset/model/character/uniform/texture/#windx11/u<PAD>" --out <dir> <uniform.cpk>
   ```

   The texture prefix includes player slots `p1` to `p4` and goalkeeper slots `g1` to `g3`.

5. Use the team `realUni` definitions as the source for colors and slot configuration.

6. To build a self-contained kitserver pack, combine the source uniform CPK with its kitserver configuration tree.

   ```sh
   ./cpk/cpk kits \
     --kserv-src <patch>/sider/content/kit-server \
     --leagues "<league-one>,<league-two>" \
     --out <pack> \
     <uniform.cpk>
   ```

7. Verify that each selected slot contains its `config.txt` and every FTEX file that `KitFile` references.

8. Verify that each extracted FTEX starts with `FTEX`.

9. Verify the generated `map.txt` against the destination game's team IDs before installation.

## Review

- Use the source patch's uniform CPK. Do not substitute another game's CPK.
- Keep player and goalkeeper slots together unless the request limits the scope.
- Treat patch-added team IDs as patch-specific until verified.
- A config-only source folder depends on textures in its source CPK. Add the extracted textures to make it self-contained.

## Boundaries

- Use `pes-kit-ftex` to convert a PNG to DXT5 FTEX.
- Use `cpk/README.md` for all `cpk kits` options.
- This skill does not edit kitserver slot configuration.
