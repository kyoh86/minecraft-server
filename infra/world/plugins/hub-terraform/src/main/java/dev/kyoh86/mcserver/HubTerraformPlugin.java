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
import java.util.ArrayDeque;
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
  private static final int CLEAR_MARGIN = 96;
  private static final int SURFACE_FROZEN_COVER_IGNORE_MIN_Y = 64;
  private static final int WATER_FILL_TOP_OFFSET = 4;

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
    int[][] targetTop = new int[size][size];
    int[][] originalFluidSurfaceY = new int[size][size];
    int[][] propagatedWaterSurfaceY = new int[size][size];
    Material[][] originalTopMaterial = new Material[size][size];
    Material[][] originalFillMaterial = new Material[size][size];
    Material[][] originalFluidMaterial = new Material[size][size];

    for (int x = -OUTER; x <= OUTER; x++) {
      for (int z = -OUTER; z <= OUTER; z++) {
        int topY = world.getHighestBlockYAt(x, z, HeightMap.MOTION_BLOCKING_NO_LEAVES);
        TerrainColumn terrainColumn = resolveTerrainColumn(world, x, z, topY);
        originalTop[x + OUTER][z + OUTER] = terrainColumn.terrainY;
        originalTopMaterial[x + OUTER][z + OUTER] = terrainColumn.top;
        originalFillMaterial[x + OUTER][z + OUTER] = terrainColumn.fill;
        originalFluidSurfaceY[x + OUTER][z + OUTER] = terrainColumn.fluidSurfaceY;
        originalFluidMaterial[x + OUTER][z + OUTER] = terrainColumn.fluid;
      }
    }
    smoothHeights(originalTop, smoothedTop);
    initIntMatrix(propagatedWaterSurfaceY, Integer.MIN_VALUE);

    BlockState stone = BlockTypes.STONE.getDefaultState();
    BlockState air = BlockTypes.AIR.getDefaultState();
    BlockState water = BlockTypes.WATER.getDefaultState();

    int baseY = surfaceY - 1;
    int floorMinY = surfaceY - 16;
    int clearMaxY = Math.min(world.getMaxHeight() - 1, maxHeight(originalTop) + CLEAR_MARGIN);

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
          targetTop[x + OUTER][z + OUTER] = targetTopY;
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

      propagateWaterSurfaceY(targetTop, originalFluidSurfaceY, originalFluidMaterial, clearMaxY, propagatedWaterSurfaceY);
      for (int x = -OUTER; x <= OUTER; x++) {
        for (int z = -OUTER; z <= OUTER; z++) {
          int ix = x + OUTER;
          int iz = z + OUTER;
          int waterTop = propagatedWaterSurfaceY[ix][iz];
          int targetY = targetTop[ix][iz];
          int waterFillTop = waterTop - WATER_FILL_TOP_OFFSET;
          if (waterFillTop <= targetY) {
            continue;
          }
          fill(edit, x, targetY + 1, z, x, Math.min(clearMaxY, waterFillTop), z, water);
        }
      }

      edit.flushSession();
    }

    return columns;
  }

  private int maxHeight(int[][] heights) {
    int max = Integer.MIN_VALUE;
    for (int[] row : heights) {
      for (int h : row) {
        if (h > max) {
          max = h;
        }
      }
    }
    return max;
  }

  private void initIntMatrix(int[][] matrix, int value) {
    for (int i = 0; i < matrix.length; i++) {
      for (int j = 0; j < matrix[i].length; j++) {
        matrix[i][j] = value;
      }
    }
  }

  private void propagateWaterSurfaceY(
    int[][] targetTop,
    int[][] originalFluidSurfaceY,
    Material[][] originalFluidMaterial,
    int clearMaxY,
    int[][] propagatedWaterSurfaceY
  ) {
    int size = targetTop.length;
    ArrayDeque<int[]> queue = new ArrayDeque<>();

    for (int ix = 0; ix < size; ix++) {
      for (int iz = 0; iz < size; iz++) {
        if (originalFluidMaterial[ix][iz] != Material.WATER) {
          continue;
        }
        int level = Math.min(clearMaxY, originalFluidSurfaceY[ix][iz]);
        if (level <= targetTop[ix][iz]) {
          continue;
        }
        if (level <= propagatedWaterSurfaceY[ix][iz]) {
          continue;
        }
        propagatedWaterSurfaceY[ix][iz] = level;
        queue.addLast(new int[] {ix, iz});
      }
    }

    int[][] dirs = new int[][] {{1, 0}, {-1, 0}, {0, 1}, {0, -1}};
    while (!queue.isEmpty()) {
      int[] p = queue.removeFirst();
      int ix = p[0];
      int iz = p[1];
      int level = propagatedWaterSurfaceY[ix][iz];

      for (int[] d : dirs) {
        int nx = ix + d[0];
        int nz = iz + d[1];
        if (nx < 0 || nx >= size || nz < 0 || nz >= size) {
          continue;
        }
        if (targetTop[nx][nz] >= level) {
          continue;
        }
        if (level <= propagatedWaterSurfaceY[nx][nz]) {
          continue;
        }
        propagatedWaterSurfaceY[nx][nz] = level;
        queue.addLast(new int[] {nx, nz});
      }
    }
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

  private Material sanitizeTopMaterial(Material material, int y) {
    if (!isTerrainTopCandidate(material, y)) {
      return Material.GRASS_BLOCK;
    }
    return material;
  }

  private Material sanitizeFillMaterial(Material material) {
    if (!isTerrainFillCandidate(material)) {
      return Material.DIRT;
    }
    return material;
  }

  private TerrainColumn resolveTerrainColumn(World world, int x, int z, int startY) {
    int minY = world.getMinHeight();
    Material top = Material.GRASS_BLOCK;
    Material fill = Material.DIRT;
    int terrainY = startY;
    int fluidSurfaceY = Integer.MIN_VALUE;
    Material fluid = null;

    int worldSurfaceY = world.getHighestBlockYAt(x, z, HeightMap.WORLD_SURFACE);
    Material worldSurface = world.getBlockAt(x, worldSurfaceY, z).getType();
    if (worldSurface == Material.WATER || worldSurface == Material.LAVA) {
      fluidSurfaceY = worldSurfaceY;
      fluid = worldSurface;
    }

    for (int y = startY; y >= minY; y--) {
      Material mat = world.getBlockAt(x, y, z).getType();
      if (!isTerrainTopCandidate(mat, y)) {
        continue;
      }
      top = sanitizeTopMaterial(mat, y);
      fill = findFillMaterial(world, x, z, y - 1, minY);
      terrainY = y;
      return new TerrainColumn(terrainY, top, fill, fluidSurfaceY, fluid);
    }
    return new TerrainColumn(terrainY, top, fill, fluidSurfaceY, fluid);
  }

  private Material findFillMaterial(World world, int x, int z, int startY, int minY) {
    for (int y = startY; y >= minY; y--) {
      Material mat = world.getBlockAt(x, y, z).getType();
      if (isTerrainFillCandidate(mat)) {
        return sanitizeFillMaterial(mat);
      }
    }
    return Material.DIRT;
  }

  private boolean isTerrainTopCandidate(Material material, int y) {
    if (material == null || !material.isBlock() || material.isAir()) {
      return false;
    }
    if (!material.isSolid() || !material.isOccluding()) {
      return false;
    }
    if (isHighAltitudeFrozenCover(material, y)) {
      return false;
    }
    if (isNonTerrainMaterial(material)) {
      return false;
    }
    return true;
  }

  private boolean isHighAltitudeFrozenCover(Material material, int y) {
    if (y < SURFACE_FROZEN_COVER_IGNORE_MIN_Y) {
      return false;
    }
    return material == Material.ICE
      || material == Material.PACKED_ICE
      || material == Material.BLUE_ICE
      || material == Material.SNOW
      || material == Material.SNOW_BLOCK;
  }

  private boolean isTerrainFillCandidate(Material material) {
    if (material == null || !material.isBlock() || material.isAir()) {
      return false;
    }
    if (!material.isSolid()) {
      return false;
    }
    if (isNonTerrainMaterial(material)) {
      return false;
    }
    return true;
  }

  private boolean isNonTerrainMaterial(Material material) {
    if (material == Material.WATER || material == Material.LAVA) {
      return true;
    }
    if (material == Material.MOSS_BLOCK
      || material == Material.MUDDY_MANGROVE_ROOTS
      || material == Material.MANGROVE_ROOTS
      || material == Material.GRAVEL
      || material == Material.CLAY) {
      return true;
    }
    String name = material.name();
    if (name.endsWith("_ORE")) {
      return true;
    }
    if (name.endsWith("_CARPET")
      || name.endsWith("_RAIL")
      || name.endsWith("_TORCH")
      || name.endsWith("_WALL_TORCH")
      || name.endsWith("_BUTTON")
      || name.endsWith("_PRESSURE_PLATE")
      || name.endsWith("_SIGN")
      || name.endsWith("_WALL_SIGN")
      || name.endsWith("_BANNER")
      || name.endsWith("_WALL_BANNER")) {
      return true;
    }
    if (name.contains("FENCE")
      || name.contains("DOOR")
      || name.contains("TRAPDOOR")
      || name.contains("LEAVES")
      || name.contains("LOG")
      || name.contains("_WOOD")
      || name.contains("CORAL")
      || name.contains("KELP")
      || name.contains("SEAGRASS")
      || name.contains("SEA_PICKLE")
      || name.contains("STEM")
      || name.contains("VINE")
      || name.contains("CHAIN")
      || name.contains("LANTERN")
      || name.contains("SAPLING")
      || name.contains("FLOWER")
      || name.contains("MUSHROOM")) {
      return true;
    }
    return false;
  }

  private static final class TerrainColumn {
    private final int terrainY;
    private final Material top;
    private final Material fill;
    private final int fluidSurfaceY;
    private final Material fluid;

    private TerrainColumn(int terrainY, Material top, Material fill, int fluidSurfaceY, Material fluid) {
      this.terrainY = terrainY;
      this.top = top;
      this.fill = fill;
      this.fluidSurfaceY = fluidSurfaceY;
      this.fluid = fluid;
    }
  }
}
