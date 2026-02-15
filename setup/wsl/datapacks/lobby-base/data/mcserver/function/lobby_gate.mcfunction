# Lobby dual-gate around x=-7, y=63..65
# Built from the two placeholder regions created in-world:
# - Right gate (from player's perspective): z=-3..-2 (to survival)
# - Left gate (future route): z=2..3

# 0) Cleanup old logic artifacts
execute in minecraft:overworld run kill @e[type=marker,tag=gatebridge]
execute in minecraft:overworld run kill @e[type=block_display,tag=gate_gate]

# 1) Right gate frame and trigger stack
execute in minecraft:overworld run fill -7 62 -4 -7 66 -1 obsidian outline
execute in minecraft:overworld run fill -7 63 -3 -7 65 -2 air
execute in minecraft:overworld run setblock -7 61 -2 obsidian
execute in minecraft:overworld run setblock -7 62 -2 obsidian
execute in minecraft:overworld run setblock -7 63 -2 polished_blackstone_pressure_plate

# 2) Left gate frame and trigger stack (reserved route)
execute in minecraft:overworld run fill -7 62 1 -7 66 4 obsidian outline
execute in minecraft:overworld run fill -7 63 2 -7 65 3 air
execute in minecraft:overworld run setblock -7 61 2 obsidian
execute in minecraft:overworld run setblock -7 62 2 obsidian
execute in minecraft:overworld run setblock -7 63 2 polished_blackstone_pressure_plate

# 3) Visual panels via block_display (1x3x2 each), aligned to block grid
execute in minecraft:overworld run summon block_display -6.5 63.0 -2.5 {Tags:["gate_gate"],block_state:{Name:"minecraft:tinted_glass"},transformation:{left_rotation:[0f,0f,0f,1f],scale:[1f,3f,2f],right_rotation:[0f,0f,0f,1f],translation:[-0.5f,0f,-0.5f]}}
execute in minecraft:overworld run summon block_display -6.5 63.0 2.5 {Tags:["gate_gate"],block_state:{Name:"minecraft:tinted_glass"},transformation:{left_rotation:[0f,0f,0f,1f],scale:[1f,3f,2f],right_rotation:[0f,0f,0f,1f],translation:[-0.5f,0f,-0.5f]}}

# 4) GateBridge markers
execute in minecraft:overworld run summon marker -6.5 63.0 -2.5 {Tags:["gatebridge","gate_to_survival"]}
execute in minecraft:overworld run summon marker -6.5 63.0 2.5 {Tags:["gatebridge","gate_to_future"]}
