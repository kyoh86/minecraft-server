execute in minecraft:overworld run fill -12 -60 -12 12 -43 12 air
execute in minecraft:overworld run fill -12 -60 -12 12 -60 12 polished_andesite
execute in minecraft:overworld run fill -11 -59 -11 11 -59 11 smooth_quartz

execute in minecraft:overworld run fill -8 -59 -8 8 -51 8 smooth_quartz
execute in minecraft:overworld run fill -7 -58 -7 7 -52 7 air

execute in minecraft:overworld run fill -1 -58 -8 1 -55 -8 air
execute in minecraft:overworld run fill 8 -58 -1 8 -55 1 air
execute in minecraft:overworld run fill -8 -58 -1 -8 -55 1 air

execute in minecraft:overworld run fill -1 -58 -10 1 -56 -10 smooth_quartz
execute in minecraft:overworld run fill 10 -58 -1 10 -56 1 smooth_quartz
execute in minecraft:overworld run fill -10 -58 -1 -10 -56 1 smooth_quartz

execute in minecraft:overworld run fill -2 -59 -9 2 -55 -9 obsidian
execute in minecraft:overworld run fill -1 -58 -9 1 -56 -9 air
execute in minecraft:overworld run fill 9 -59 -2 9 -55 2 obsidian
execute in minecraft:overworld run fill 9 -58 -1 9 -56 1 air
execute in minecraft:overworld run fill -9 -59 -2 -9 -55 2 obsidian
execute in minecraft:overworld run fill -9 -58 -1 -9 -56 1 air

execute in minecraft:overworld run fill -9 -59 -9 9 -59 -9 polished_deepslate
execute in minecraft:overworld run fill 9 -59 -9 9 -59 9 polished_deepslate
execute in minecraft:overworld run fill -9 -59 -9 -9 -59 9 polished_deepslate
execute in minecraft:overworld run fill -9 -59 9 9 -59 9 polished_deepslate

execute in minecraft:overworld run setblock 0 -58 -7 oak_sign[rotation=0]
execute in minecraft:overworld run data merge block 0 -58 -7 {front_text:{messages:[{text:"Residence",color:"green",bold:1b},"","",""]}}
execute in minecraft:overworld run setblock 7 -58 0 oak_sign[rotation=4]
execute in minecraft:overworld run data merge block 7 -58 0 {front_text:{messages:[{text:"Resource",color:"gold",bold:1b},"","",""]}}
execute in minecraft:overworld run setblock -7 -58 0 oak_sign[rotation=12]
execute in minecraft:overworld run data merge block -7 -58 0 {front_text:{messages:[{text:"Factory",color:"aqua",bold:1b},"","",""]}}

execute in minecraft:overworld run setblock 0 -59 0 lodestone
execute in minecraft:overworld run setblock 0 -58 0 sea_lantern
execute in minecraft:overworld run setblock 0 -57 0 lightning_rod[facing=up]

execute in minecraft:overworld run setblock 0 -59 -9 waxed_copper_bulb
execute in minecraft:overworld run setblock 0 -55 -9 waxed_copper_bulb
execute in minecraft:overworld run setblock 0 -60 -9 redstone_torch
execute in minecraft:overworld run setblock 0 -54 -9 redstone_block

execute in minecraft:overworld run setblock 9 -59 0 waxed_copper_bulb
execute in minecraft:overworld run setblock 9 -55 0 waxed_copper_bulb
execute in minecraft:overworld run setblock 9 -60 0 redstone_torch
execute in minecraft:overworld run setblock 9 -54 0 redstone_block

execute in minecraft:overworld run setblock -9 -59 0 waxed_copper_bulb
execute in minecraft:overworld run setblock -9 -55 0 waxed_copper_bulb
execute in minecraft:overworld run setblock -9 -60 0 redstone_torch
execute in minecraft:overworld run setblock -9 -54 0 redstone_block

execute in minecraft:overworld run kill @e[type=minecraft:block_display,tag=mcserver_gate_glass,distance=..64]

# Gate for Regidence
execute in minecraft:overworld run summon minecraft:block_display -1.0 -58.0 -8.6 {Tags:["mcserver_gate_glass","mcserver_gate_glass_north"],block_state:{Name:"minecraft:tinted_glass"}}
execute in minecraft:overworld run data merge entity @e[type=minecraft:block_display,tag=mcserver_gate_glass_north,limit=1] {transformation:{scale:[3.0f,3.0f,0.2f]}}
execute in minecraft:overworld run setblock -3 -61 11 repeating_command_block[facing=up]{auto:1b,TrackOutput:0b,Command:"execute in minecraft:overworld run particle minecraft:end_rod 0 -56 -9 1.0 1.0 0.05 0.01 8 normal"}
execute in minecraft:overworld run setblock -2 -61 11 repeating_command_block[facing=up]{auto:1b,TrackOutput:0b,Command:"execute in minecraft:overworld run particle minecraft:enchant 0 -56 -9 1.0 1.0 0.05 0.05 24 normal"}

# Gate for Resources
execute in minecraft:overworld run summon minecraft:block_display 9.4 -58.0 -1.0 {Tags:["mcserver_gate_glass","mcserver_gate_glass_east"],block_state:{Name:"minecraft:tinted_glass"}}
execute in minecraft:overworld run data merge entity @e[type=minecraft:block_display,tag=mcserver_gate_glass_east,limit=1] {transformation:{scale:[0.2f,3.0f,3.0f]}}
execute in minecraft:overworld run setblock -1 -61 11 repeating_command_block[facing=up]{auto:1b,TrackOutput:0b,Command:"execute in minecraft:overworld run particle minecraft:end_rod 9 -56 0 0.05 1.0 1.0 0.01 8 normal"}
execute in minecraft:overworld run setblock 0 -61 11 repeating_command_block[facing=up]{auto:1b,TrackOutput:0b,Command:"execute in minecraft:overworld run particle minecraft:enchant 9 -56 0 0.05 1.0 1.0 0.05 24 normal"}

# Gate for Factory
execute in minecraft:overworld run summon minecraft:block_display -8.6 -58.0 -1.0 {Tags:["mcserver_gate_glass","mcserver_gate_glass_west"],block_state:{Name:"minecraft:tinted_glass"}}
execute in minecraft:overworld run data merge entity @e[type=minecraft:block_display,tag=mcserver_gate_glass_west,limit=1] {transformation:{scale:[0.2f,3.0f,3.0f]}}
execute in minecraft:overworld run setblock 1 -61 11 repeating_command_block[facing=up]{auto:1b,TrackOutput:0b,Command:"execute in minecraft:overworld run particle minecraft:end_rod -9 -56 0 0.05 1.0 1.0 0.01 8 normal"}
execute in minecraft:overworld run setblock 2 -61 11 repeating_command_block[facing=up]{auto:1b,TrackOutput:0b,Command:"execute in minecraft:overworld run particle minecraft:enchant -9 -56 0 0.05 1.0 1.0 0.05 24 normal"}
