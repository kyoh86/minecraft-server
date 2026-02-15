# Lobby gate at around (-8, 63, -2)
# Safe mode: additive placement only.
execute in minecraft:overworld run fill -8 62 -4 -8 67 0 obsidian outline
execute in minecraft:overworld run fill -8 63 -3 -8 66 -1 air
execute in minecraft:overworld run setblock -8 61 -2 obsidian
execute in minecraft:overworld run setblock -8 62 -2 obsidian
execute in minecraft:overworld run setblock -8 63 -2 polished_blackstone_pressure_plate
execute in minecraft:overworld run kill @e[type=marker,tag=gatebridge]
execute in minecraft:overworld run summon marker -7.5 63.0 -1.5 {Tags:["gatebridge","gate_to_survival"]}
