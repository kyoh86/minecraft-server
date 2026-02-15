# Lobby gate at around (-8, 63, -2)
# Safe mode: additive placement only.
execute in minecraft:overworld run fill -8 62 -4 -8 67 0 obsidian outline
execute in minecraft:overworld run fill -8 63 -3 -8 66 -1 purple_stained_glass
execute in minecraft:overworld run setblock -8 62 -2 sea_lantern
execute in minecraft:overworld run setblock -8 67 -2 crying_obsidian
execute in minecraft:overworld run setblock -7 62 -2 command_block{Command:"execute in minecraft:overworld run vsend @r survival",auto:0b,TrackOutput:0b}
execute in minecraft:overworld run setblock -7 63 -2 polished_blackstone_pressure_plate
execute in minecraft:overworld run kill @e[type=area_effect_cloud,tag=lobby_gate_portal,limit=5]
execute in minecraft:overworld run summon area_effect_cloud -8.0 64.0 -2.0 {Tags:["lobby_gate_portal"],Duration:2147483647,Radius:1.2f,RadiusPerTick:0f,WaitTime:0,ReapplicationDelay:20,Particle:{type:"minecraft:portal"}}
