sudo su - ec2-user

sudo yum update -y
sudo yum install -y xfsprogs
sudo yum clean all

sudo mkfs -t xfs /dev/sdh

sudo amazon-linux-extras install -y nginx1
sudo cp -a /etc/nginx/nginx.conf /etc/nginx/nginx.conf.back
sudo systemctl enable nginx

# -----

sudo su - ec2-user

sudo yum update -y
sudo yum install -y udev java-17-amazon-corretto
sudo yum clean all

sudo adduser minecraft --gid wheel
sudo mkdir -p /minecraft
sudo mount /dev/sdh /minecraft
sudo chown minecraft:wheel /minecraft

# ----
sudo su - minecraft

cd /minecraft

MINECRAFT_VERSION=1.20.2

# Install "FabricMC" as a Minecraft Server
FABRIC_LOADER_VERSION=0.14.24
FABRIC_INSTALLER_VERSION=0.11.1

curl -Lo ./fabric-server-mc.jar https://meta.fabricmc.net/v2/versions/loader/${MINECRAFT_VERSION}/${FABRIC_LOADER_VERSION}/${FABRIC_INSTALLER_VERSION}/server/jar
java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting
mkdir -p mods

# Install "Fabric API" to run Fabric Mods
FABRIC_API_VERSION=0.90.7
curl -Lo ./mods/fabric-api.jar https://github.com/FabricMC/fabric/releases/download/${FABRIC_API_VERSION}+${MINECRAFT_VERSION}/fabric-api-${FABRIC_API_VERSION}+${MINECRAFT_VERSION}.jar

# Install "GeyserMC" to cross-play between Bedrock & Java
curl -Lo ./mods/Geyser-Fabric.jar https://download.geysermc.org/v2/projects/geyser/versions/latest/builds/latest/downloads/fabric

# Install "DiscordIntegration" to integrate with Discord
DISCORD_INTEGRATION_VERSION=3.0.3
curl -Lo ./mods/dcintegration-fabric.jar https://cdn.modrinth.com/data/rbJ7eS5V/versions/ZlLJC9ox/dcintegration-fabric-${DISCORD_INTEGRATION_VERSION}-${MINECRAFT_VERSION}.jar

# Install "Floodgate" to support login with Bedrock account
curl -Lo ./mods/floodgate-fabric.jar https://ci.opencollab.dev/job/GeyserMC/job/Floodgate-Fabric/job/master/lastSuccessfulBuild/artifact/build/libs/floodgate-fabric.jar

# Install "Dynmap" to show world map
curl -Lo ./mods/Dynmap-fabric.jar https://www.curseforge.com/api/v1/mods/59433/files/4765921/download

# Initialize Fabric server settings
timeout 1m java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting || :

java -Xmx3G -jar fabric-server-mc.jar --nogui
