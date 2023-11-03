link_data() {
  local mounted="/data/$1"
  if [ -e $mounted ]; then
    rm -rf "./$1"
    mkdir -p "$(dirname $1)"
    ln -s "$mounted" "./$1"
  fi
}

if [ -e /data ]; then
  link_data eula.txt
  link_data server.properties
  link_data ops.json
  link_data whitelist.json
  link_data banned-ips.json
  link_data banned-players.json  

  # Configure GeyserMC
  link_data config/Geyser-Fabric/config.yml

  # Configure DiscordIntegration
  link_data DiscordIntegration-Data
  link_data config/Discord-Integration.toml

  # Link world data
  rm -rf world
  mkdir -p /data/world
  ln -s /data/world world
fi

java -Xmx2G -jar fabric-server-mc.jar --nogui
