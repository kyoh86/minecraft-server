FROM amazoncorretto:17

# Install dependencies
RUN yum update -y \
  && yum install -y udev \
  && yum clean all \
  && rm -rf /var/cache/yum

ARG MINECRAFT_VERSION=1.20.2

WORKDIR /minecraft-bin

# Expose ports (NOTE: equal it with server-port in server.properties)
EXPOSE 25565
EXPOSE 19132

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

# Initialize Fabric server settings
RUN timeout java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting || :

# Put linker
ADD ./mount.sh .
RUN chmod +x ./mount.sh

# Configure Minecraft Server
ADD ./data/eula.txt /data/eula.txt
RUN ./mount.sh eula.txt

ADD ./data/server.properties /data/server.properties
RUN ./mount.sh server.properties

ADD ./data/ops.json /data/ops.json
RUN ./mount.sh ops.json

ADD ./data/whitelist.json /data/whitelist.json
RUN ./mount.sh whitelist.json

ADD ./data/start.sh /data/start.sh
RUN chmod +x /data/start.sh
RUN ./mount.sh start.sh

# Configure GeyserMC
ADD ./data/config/Geyser-Fabric/config.yml /data/config/Geyser-Fabric/config.yml
RUN ./mount.sh config/Geyser-Fabric/config.yml

# Configure DiscordIntegration
ADD ./data/config/discord_integration.toml /data/config/Discord-Integration.toml
RUN ./mount.sh config/Discord-Integration.toml

CMD /minecraft/start.sh