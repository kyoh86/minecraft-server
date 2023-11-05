FROM amazoncorretto:17

# Install dependencies
RUN yum update -y \
  && yum install -y udev \
  && yum clean all \
  && rm -rf /var/cache/yum

ARG MINECRAFT_VERSION=1.20.2

WORKDIR /minecraft

# Expose ports (NOTE: equal it with server-port in server.properties)
# Java
EXPOSE 25565
# Bedrock
EXPOSE 19132
# Dynmap
EXPOSE 8123

# Install "FabricMC" as a Minecraft Server
ARG FABRIC_LOADER_VERSION=0.14.24
ARG FABRIC_INSTALLER_VERSION=0.11.1
RUN curl -Lo fabric-server-mc.jar https://meta.fabricmc.net/v2/versions/loader/${MINECRAFT_VERSION}/${FABRIC_LOADER_VERSION}/${FABRIC_INSTALLER_VERSION}/server/jar \
  && java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting

# Install "Fabric API" to run Fabric Mods
ARG FABRIC_API_VERSION=0.90.7
RUN mkdir -p mods \
  && curl -Lo ./mods/fabric-api.jar https://github.com/FabricMC/fabric/releases/download/${FABRIC_API_VERSION}+${MINECRAFT_VERSION}/fabric-api-${FABRIC_API_VERSION}+${MINECRAFT_VERSION}.jar

# Install "GeyserMC" to cross-play between Bedrock & Java
RUN mkdir -p ./mods \
  && curl -Lo ./mods/Geyser-Fabric.jar https://download.geysermc.org/v2/projects/geyser/versions/latest/builds/latest/downloads/fabric

# Install "DiscordIntegration" to integrate with Discord
ARG DISCORD_INTEGRATION_VERSION=3.0.3
RUN mkdir -p ./mods \
  && curl -Lo ./mods/dcintegration-fabric.jar https://cdn.modrinth.com/data/rbJ7eS5V/versions/ZlLJC9ox/dcintegration-fabric-${DISCORD_INTEGRATION_VERSION}-${MINECRAFT_VERSION}.jar

# Install "Floodgate" to support login with Bedrock account
RUN mkdir -p ./mods \
  && curl -Lo ./mods/floodgate-fabric.jar https://ci.opencollab.dev/job/GeyserMC/job/Floodgate-Fabric/job/master/lastSuccessfulBuild/artifact/build/libs/floodgate-fabric.jar

# Install "Dynmap" to show world map
RUN mkdir -p ./mods \
  && curl -Lo ./mods/Dynmap-fabric.jar https://www.curseforge.com/api/v1/mods/59433/files/4765921/download

# Initialize Fabric server settings
RUN timeout 1m java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting || :

# Configure Minecraft Server
ADD ./data/eula.txt .
ADD ./data/server.properties .
ADD ./data/ops.json .
ADD ./data/whitelist.json .

# Configure GeyserMC
ADD ./data/config/Geyser-Fabric/config.yml ./config/Geyser-Fabric/config.yml

# Configure DiscordIntegration
ADD ./data/DiscordIntegration-Data ./DiscordIntegration-Data
ADD ./data/config/Discord-Integration.toml ./config/Discord-Integration.toml

ADD ./start.sh .
CMD /minecraft/start.sh
