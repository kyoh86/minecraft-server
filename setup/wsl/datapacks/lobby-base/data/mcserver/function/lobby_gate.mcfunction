# Lobby dual-gate around x=-7, y=63..65
# Built from the two placeholder regions created in-world:
# - Right gate (from player's perspective): z=-3..-2 (to survival)
# - Left gate (future route): z=2..3

# 0) Build lobby room
execute in minecraft:overworld run fill -10 62 -10 10 72 10 minecraft:quartz_block
execute in minecraft:overworld run fill -9 63 -9 9 71 9 minecraft:air
execute in minecraft:overworld run fill -10 72 -10 10 72 10 minecraft:glass
execute in minecraft:overworld run fill -6 63 0 -3 63 0 minecraft:lime_carpet
execute in minecraft:overworld run fill 6 63 0 3 63 0 minecraft:orange_carpet
execute in minecraft:overworld run fill -2 63 -2 2 63 2 minecraft:light_gray_carpet

# 1) Right gate frame and trigger stack
execute in minecraft:overworld run fill -7 62 -1 -7 66 1 obsidian outline
execute in minecraft:overworld run fill -7 63 0 -7 65 0 air
execute in minecraft:overworld run setblock -7 61 0 obsidian
execute in minecraft:overworld run setblock -7 62 0 obsidian
execute in minecraft:overworld run setblock -7 63 0 polished_blackstone_pressure_plate

# 2) Left gate frame and trigger stack (reserved route)
execute in minecraft:overworld run fill 7 62 -1 7 66 1 obsidian outline
execute in minecraft:overworld run fill 7 63 0 7 65 0 air
execute in minecraft:overworld run setblock 7 61 0 obsidian
execute in minecraft:overworld run setblock 7 62 0 obsidian
execute in minecraft:overworld run setblock 7 63 0 polished_blackstone_pressure_plate

# 3) Visual panels via block_display (1x3x2 each), aligned to block grid
kill @e[tag=gate_gate]
execute in minecraft:overworld run summon block_display -6.5 63.0 0 {Tags:["gate_gate"],block_state:{Name:"minecraft:tinted_glass"},transformation:{left_rotation:[0f,0f,0f,1f],scale:[1f,3f,1f],right_rotation:[0f,0f,0f,1f],translation:[-0.5f,0f,-0.5f]}}
execute in minecraft:overworld run summon block_display 7.5 63.0 0 {Tags:["gate_gate"],block_state:{Name:"minecraft:tinted_glass"},transformation:{left_rotation:[0f,0f,0f,1f],scale:[1f,3f,1f],right_rotation:[0f,0f,0f,1f],translation:[-0.5f,0f,-0.5f]}}

# 4) GateBridge markers
kill @e[tag=gatebridge]
execute in minecraft:overworld run summon marker -6.5 63.0 0 {Tags:["gatebridge","gate_to_survival"]}
execute in minecraft:overworld run summon marker 6.5 63.0 0 {Tags:["gatebridge","gate_to_future"]}
