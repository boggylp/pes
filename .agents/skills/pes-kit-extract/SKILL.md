---
name: pes-kit-extract
description: 'Extract existing PES kits from a uniform CPK or assemble a self-contained kitserver pack.'
---

# PES kit extract

## Steps

1. Build the tool: `cd cpk && go build . && cd ..`.
2. Confirm the source patch's own uniform CPK holds uniform entries: `./cpk/cpk list <uniform.cpk> | rg -c 'uniform'`.
3. Resolve each team ID from the source patch's kitserver `map.txt` or team list, and confirm the destination game uses the same ID.
   - A patch-added team ID stays patch-specific until verified.
4. For one team, extract its definitions and textures, padding the texture team ID to four digits.

   ```sh
   ./cpk/cpk extract --prefix "common/character0/model/character/uniform/team/<id>/" --out <dir> <uniform.cpk>
   ./cpk/cpk extract --prefix "Asset/model/character/uniform/texture/#windx11/u<PAD>" --out <dir> <uniform.cpk>
   ```

   - The texture prefix covers player slots `p1` to `p4` and goalkeeper slots `g1` to `g3`, so keep them together unless the request narrows it.
   - The team's `realUni` definitions hold its colors and slot configuration.

5. For a self-contained kitserver pack, combine the uniform CPK with the patch's kitserver tree, per `cpk/README.md`:

   ```sh
   ./cpk/cpk kits --kserv-src <patch>/sider/content/kit-server --leagues "<league-one>,<league-two>" --out <pack> <uniform.cpk>
   ```

6. Verify the result:
   - Each slot holds its `config.txt` and every FTEX its `KitFile` names.
   - Each FTEX starts with `FTEX`.
   - The generated `map.txt` matches the destination game's team IDs.
