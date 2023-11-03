FROM amazoncorretto:17

# Install dependencies
RUN yum update -y \
  && yum install -y udev \
  && yum clean all \
  && rm -rf /var/cache/yum

ARG MINECRAFT_VERSION=1.20.2
ARG FABRIC_LOADER_VERSION=0.14.24
ARG FABRIC_API_VERSION=0.90.7
ARG FABRIC_INSTALLER_VERSION=0.11.1

# equal it with server-port in server.properties
EXPOSE 25565
EXPOSE 19132

WORKDIR /minecraft

# Install Minecraft Server
RUN curl -Lo fabric-server-mc.jar https://meta.fabricmc.net/v2/versions/loader/${MINECRAFT_VERSION}/${FABRIC_LOADER_VERSION}/${FABRIC_INSTALLER_VERSION}/server/jar \
  && java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting
RUN echo "java -Xmx2G -jar fabric-server-mc.jar --nogui" >> ./start.sh \
  && chmod +x ./start.sh

# Install GeyserMC
RUN mkdir -p ./mods \
  && curl -Lo ./mods/Geyser-Fabric.jar https://download.geysermc.org/v2/projects/geyser/versions/latest/builds/latest/downloads/fabric

# Install Fabric API
RUN mkdir -p mods \
  && curl -Lo ./mods/fabric-api.jar https://github.com/FabricMC/fabric/releases/download/${FABRIC_API_VERSION}+${MINECRAFT_VERSION}/fabric-api-${FABRIC_API_VERSION}+${MINECRAFT_VERSION}.jar

RUN java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting

# Install Floodgate
RUN mkdir -p ./mods \
  && curl -Lo ./mods/floodgate-fabric.jar https://ci.opencollab.dev/job/GeyserMC/job/Floodgate-Fabric/job/master/lastSuccessfulBuild/artifact/build/libs/floodgate-fabric.jar

# Setup Minecraft Server
ADD eula.txt .
ADD server.properties .
ADD whitelist.json .

# Setup GeyserMC
ADD config/geyser.yml config/Geyser-Fabric/config.yml
# Setup Floodgate
ADD config/floodgate.yml config/floodgate/config.yml

CMD /minecraft/start.sh
