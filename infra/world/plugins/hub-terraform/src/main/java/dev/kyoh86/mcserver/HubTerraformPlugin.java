package dev.kyoh86.mcserver;

import com.sk89q.worldedit.EditSession;
import com.sk89q.worldedit.MaxChangedBlocksException;
import com.sk89q.worldedit.WorldEdit;
import com.sk89q.worldedit.bukkit.BukkitAdapter;
import com.sk89q.worldedit.math.BlockVector3;
import com.sk89q.worldedit.regions.CuboidRegion;
import com.sk89q.worldedit.world.block.BlockState;
import com.sk89q.worldedit.world.block.BlockType;
import com.sk89q.worldedit.world.block.BlockTypes;
import org.bukkit.HeightMap;
import org.bukkit.Material;
import org.bukkit.World;
import org.bukkit.command.Command;
import org.bukkit.command.CommandExecutor;
import org.bukkit.command.CommandSender;
import org.bukkit.plugin.java.JavaPlugin;

public class HubTerraformPlugin extends JavaPlugin implements CommandExecutor {
  private static final int CORE = 32;
  private static final int OUTER = 64;
  private static final int BLUR_RADIUS = 2;

  @Override
  public void onEnable() {
    if (getCommand("hubterraform") != null) {
      getCommand("hubterraform").setExecutor(this);
    }
  }

  @Override
  public boolean onCommand(CommandSender sender, Command command, String label, String[] args) {
    if (args.length != 3 || !"apply".equalsIgnoreCase(args[0])) {
      sender.sendMessage("Usage: /hubterraform apply <world> <surfaceY>");
      return true;
    }

    World world = getServer().getWorld(args[1]);
    if (world == null) {
      sender.sendMessage("world not found: " + args[1]);
      return true;
    }

    final int surfaceY;
    try {
      surfaceY = Integer.parseInt(args[2]);
    } catch (NumberFormatException e) {
      sender.sendMessage("invalid surfaceY: " + args[2]);
      return true;
    }

    try {
      int changed = terraform(world, surfaceY);
      sender.sendMessage("hubterraform applied: world=" + world.getName() + " surfaceY=" + surfaceY + " columns=" + changed);
    } catch (Exception e) {
      getLogger().severe("hubterraform failed: " + e.getMessage());
      sender.sendMessage("hubterraform failed: " + e.getMessage());
    }
    return true;
  }

  private int terraform(World world, int surfaceY) throws Exception {
    int size = OUTER * 2 + 1;
    int[][] originalTop = new int[size][size];
    int[][] smoothedTop = new int[size][size];
    Material[][] originalTopMaterial = new Material[size][size];
    Material[][] originalFillMaterial = new Material[size][size];

    for (int x = -OUTER; x <= OUTER; x++) {
      for (int z = -OUTER; z <= OUTER; z++) {
        int topY = world.getHighestBlockYAt(x, z, HeightMap.MOTION_BLOCKING_NO_LEAVES);
        originalTop[x + OUTER][z + OUTER] = topY;
        originalTopMaterial[x + OUTER][z + OUTER] = sanitizeTopMaterial(world.getBlockAt(x, topY, z).getType());
        originalFillMaterial[x + OUTER][z + OUTER] = sanitizeFillMaterial(world.getBlockAt(x, topY - 1, z).getType());
      }
    }
    smoothHeights(originalTop, smoothedTop);

    BlockState stone = BlockTypes.STONE.getDefaultState();
    BlockState air = BlockTypes.AIR.getDefaultState();

    int baseY = surfaceY - 1;
    int floorMinY = surfaceY - 16;
    int clearMaxY = surfaceY + 24;

    int columns = 0;
    try (EditSession edit = WorldEdit.getInstance()
      .newEditSessionBuilder()
      .world(BukkitAdapter.adapt(world))
      .build()) {

      for (int x = -OUTER; x <= OUTER; x++) {
        for (int z = -OUTER; z <= OUTER; z++) {
          double radius = Math.sqrt((double) x * x + (double) z * z);
          int targetTopY;
          if (radius <= CORE) {
            targetTopY = baseY;
          } else {
            double t = (radius - CORE) / (double) (OUTER - CORE);
            t = smoothstep(t);
            int outsideY = smoothedTop[x + OUTER][z + OUTER];
            double edgeBlend = clamp((radius - (OUTER - 4.0)) / 4.0);
            outsideY = (int) Math.round((1.0 - edgeBlend) * outsideY + edgeBlend * originalTop[x + OUTER][z + OUTER]);
            targetTopY = (int) Math.round((1.0 - t) * baseY + t * outsideY);
          }

          int foundationTop = targetTopY - 2;
          int seabedY = world.getHighestBlockYAt(x, z, HeightMap.OCEAN_FLOOR);
          int foundationBottom = Math.min(floorMinY, seabedY);
          fill(edit, x, foundationBottom, z, x, foundationTop, z, stone);

          BlockState fillMat = toBlockStateOrDefault(originalFillMaterial[x + OUTER][z + OUTER], BlockTypes.DIRT.getDefaultState());
          BlockState topMat = toBlockStateOrDefault(originalTopMaterial[x + OUTER][z + OUTER], BlockTypes.GRASS_BLOCK.getDefaultState());
          fill(edit, x, foundationTop + 1, z, x, targetTopY - 1, z, fillMat);
          fill(edit, x, targetTopY, z, x, targetTopY, z, topMat);
          fill(edit, x, targetTopY + 1, z, x, clearMaxY, z, air);
          columns++;
        }
      }

      edit.flushSession();
    }

    return columns;
  }

  private void smoothHeights(int[][] src, int[][] dst) {
    int size = src.length;
    for (int ix = 0; ix < size; ix++) {
      for (int iz = 0; iz < size; iz++) {
        int sum = 0;
        int count = 0;
        for (int dx = -BLUR_RADIUS; dx <= BLUR_RADIUS; dx++) {
          for (int dz = -BLUR_RADIUS; dz <= BLUR_RADIUS; dz++) {
            int nx = ix + dx;
            int nz = iz + dz;
            if (nx < 0 || nx >= size || nz < 0 || nz >= size) {
              continue;
            }
            sum += src[nx][nz];
            count++;
          }
        }
        dst[ix][iz] = count == 0 ? src[ix][iz] : (int) Math.round((double) sum / (double) count);
      }
    }
  }

  private void fill(EditSession edit, int minX, int minY, int minZ, int maxX, int maxY, int maxZ, BlockState block)
    throws MaxChangedBlocksException {
    if (minY > maxY) {
      return;
    }
    CuboidRegion region = new CuboidRegion(
      BlockVector3.at(minX, minY, minZ),
      BlockVector3.at(maxX, maxY, maxZ)
    );
    edit.setBlocks(region, block);
  }

  private double smoothstep(double t) {
    t = clamp(t);
    return t * t * (3.0 - 2.0 * t);
  }

  private double clamp(double t) {
    if (t <= 0.0) {
      return 0.0;
    }
    if (t >= 1.0) {
      return 1.0;
    }
    return t;
  }

  private BlockState toBlockStateOrDefault(Material material, BlockState fallback) {
    if (material == null || material.isAir()) {
      return fallback;
    }
    BlockType blockType = BlockTypes.get(material.name().toLowerCase());
    if (blockType == null) {
      blockType = BlockTypes.get("minecraft:" + material.name().toLowerCase());
    }
    if (blockType == null) {
      return fallback;
    }
    return blockType.getDefaultState();
  }

  private Material sanitizeTopMaterial(Material material) {
    if (material == null || material.isAir() || material == Material.WATER || material == Material.LAVA) {
      return Material.GRASS_BLOCK;
    }
    return material;
  }

  private Material sanitizeFillMaterial(Material material) {
    if (material == null || material.isAir() || material == Material.WATER || material == Material.LAVA) {
      return Material.DIRT;
    }
    return material;
  }
}
