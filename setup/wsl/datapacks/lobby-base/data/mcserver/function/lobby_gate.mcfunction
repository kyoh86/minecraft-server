# Lobby gate at around (-8, 63, -2)
# Safe mode: additive placement only.
execute in minecraft:overworld run fill -8 63 -4 -8 68 0 obsidian outline
execute in minecraft:overworld run fill -8 64 -3 -8 67 -1 purple_stained_glass
execute in minecraft:overworld run setblock -8 63 -2 sea_lantern
execute in minecraft:overworld run setblock -8 68 -2 crying_obsidian
execute in minecraft:overworld run kill @e[type=area_effect_cloud,tag=lobby_gate_portal,limit=5]
execute in minecraft:overworld run summon area_effect_cloud -8.0 65.0 -2.0 {Tags:["lobby_gate_portal"],Duration:2147483647,Radius:1.2f,RadiusPerTick:0f,WaitTime:0,ReapplicationDelay:20,Particle:{type:"minecraft:portal"}}
tellraw @a [{"text":"[Lobby Gate] ","color":"light_purple"},{"text":"サバイバルへ移動","color":"aqua","clickEvent":{"action":"run_command","value":"/server survival"},"hoverEvent":{"action":"show_text","contents":"クリックで /server survival"}}]
