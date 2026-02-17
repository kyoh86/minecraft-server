setworldspawn ~ ~ ~

# Margin
fill ~-16 ~-1 ~-16 ~16 ~10 ~16 air

# Room
fill ~-9 ~-2 ~-9 ~9 ~-2 ~9 polished_deepslate
fill ~-8 ~-1 ~-8 ~8 ~5 ~8 smooth_quartz outline
fill ~-7 ~ ~-7 ~7 ~4 ~7 air

# Lights
fill ~6 ~-1 ~6 ~6 ~-1 ~6 sea_lantern
fill ~6 ~-1 ~-6 ~6 ~-1 ~-6 sea_lantern
fill ~-6 ~-1 ~6 ~-6 ~-1 ~6 sea_lantern
fill ~-6 ~-1 ~-6 ~-6 ~-1 ~-6 sea_lantern
fill ~ ~-1 ~ ~ ~-1 ~ sea_lantern

# Doors
## East
setblock ~-8 ~0 ~ iron_door[facing=west,half=lower,open=false]
setblock ~-8 ~1 ~ iron_door[facing=west,half=upper,open=false]
setblock ~-7 ~0 ~ heavy_weighted_pressure_plate
### Circuit guard blocks
fill ~-9 ~-2 ~3 ~-3 ~-5 ~-5 minecraft:smooth_stone
fill ~-8 ~-2 ~2 ~-4 ~-4 ~-4 minecraft:air
### Circuit
fill ~-8 ~0 ~-1 ~-8 ~-2 ~-1 minecraft:observer[facing=up]
setblock ~-8 ~1 ~-1 minecraft:note_block
fill ~-8 ~-4 ~-1 ~-8 ~-4 ~-3 minecraft:smooth_stone
setblock ~-7 ~-4 ~-3 minecraft:smooth_stone
fill ~-6 ~-4 ~-3 ~-5 ~-4 ~-1 minecraft:smooth_stone
fill ~-6 ~-4 ~-0 ~-6 ~-4 ~1 minecraft:smooth_stone
setblock ~-7 ~-3 ~1 minecraft:target
setblock ~-8 ~-2 ~1 minecraft:smooth_stone
setblock ~-8 ~-1 ~-0 minecraft:smooth_stone
setblock ~-8 ~-3 ~1 minecraft:redstone_wall_torch[facing=west]
setblock ~-8 ~-2 ~-0 minecraft:redstone_wall_torch[facing=north]
fill ~-8 ~-3 ~-1 ~-8 ~-3 ~-3 minecraft:redstone_wire
setblock ~-7 ~-3 ~-3 minecraft:repeater[facing=west,delay=2]
fill ~-6 ~-3 ~-3 ~-5 ~-3 ~-3 minecraft:redstone_wire
setblock ~-6 ~-3 ~-2 minecraft:comparator[facing=north]
setblock ~-5 ~-3 ~-2 minecraft:comparator[facing=south]
fill ~-6 ~-3 ~-1 ~-5 ~-3 ~-1 minecraft:redstone_wire
fill ~-6 ~-3 ~-0 ~-6 ~-3 ~1 minecraft:redstone_wire

## West
setblock ~8 ~0 ~0 minecraft:iron_door[facing=east,half=lower]
setblock ~8 ~1 ~0 minecraft:iron_door[facing=east,half=upper]
setblock ~7 ~0 ~ heavy_weighted_pressure_plate
### Circuit guard blocks
fill ~9 ~-2 ~3 ~3 ~-5 ~-5 minecraft:smooth_stone
fill ~8 ~-2 ~2 ~4 ~-4 ~-4 minecraft:air
### Circuit
fill ~8 ~0 ~-1 ~8 ~-2 ~-1 minecraft:observer[facing=up]
setblock ~8 ~1 ~-1 minecraft:note_block
fill ~8 ~-4 ~-1 ~8 ~-4 ~-3 minecraft:smooth_stone
setblock ~7 ~-4 ~-3 minecraft:smooth_stone
fill ~6 ~-4 ~-3 ~5 ~-4 ~-1 minecraft:smooth_stone
fill ~6 ~-4 ~-0 ~6 ~-4 ~1 minecraft:smooth_stone
setblock ~7 ~-3 ~1 minecraft:target
setblock ~8 ~-2 ~1 minecraft:smooth_stone
setblock ~8 ~-1 ~-0 minecraft:smooth_stone
setblock ~8 ~-3 ~1 minecraft:redstone_wall_torch[facing=east]
setblock ~8 ~-2 ~-0 minecraft:redstone_wall_torch[facing=north]
fill ~8 ~-3 ~-1 ~8 ~-3 ~-3 minecraft:redstone_wire
setblock ~7 ~-3 ~-3 minecraft:repeater[facing=east,delay=2]
fill ~6 ~-3 ~-3 ~5 ~-3 ~-3 minecraft:redstone_wire
setblock ~6 ~-3 ~-2 minecraft:comparator[facing=south]
setblock ~5 ~-3 ~-2 minecraft:comparator[facing=north]
fill ~6 ~-3 ~-1 ~5 ~-3 ~-1 minecraft:redstone_wire
fill ~6 ~-3 ~-0 ~6 ~-3 ~1 minecraft:redstone_wire

fill ~-2 ~ ~-9 ~2 ~5 ~-9 polished_deepslate
fill ~-2 ~ ~-8 ~2 ~4 ~-8 obsidian outline
fill ~-1 ~1 ~-8 ~1 ~3 ~-8 air
setblock ~ ~ ~-6 oak_sign[rotation=0]
data merge block ~ ~ ~-6 {front_text:{messages:[{text:"To Mainhall",color:"gold",bold:1b},"","",""]}}
