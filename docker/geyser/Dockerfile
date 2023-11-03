FROM amazoncorretto:17

# Install dependencies
RUN yum update -y \
  && yum install -y udev \
  && yum clean all \
  && rm -rf /var/cache/yum

# equal it with server-port in server.properties
EXPOSE 19132

WORKDIR /geyser

# Install GeyserMC
RUN curl -Lo ./Geyser-Standalone.jar https://download.geysermc.org/v2/projects/geyser/versions/latest/builds/latest/downloads/standalone

# Setup GeyserMC
ADD ./start.sh .

CMD /geyser/start.sh
