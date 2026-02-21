FROM maven:3.9.9-eclipse-temurin-21 AS plugin-build
WORKDIR /src
COPY infra/world/plugins/src/clickmobs-region-guard /src/clickmobs-region-guard
RUN mvn -f /src/clickmobs-region-guard/pom.xml -DskipTests package

FROM itzg/minecraft-server:java25
COPY --from=plugin-build /src/clickmobs-region-guard/target/clickmobs-region-guard-0.1.0.jar /plugins/ClickMobsRegionGuard.jar
COPY infra/world/plugins/dist/ClickMobs/config.yml /plugins/ClickMobs/config.yml
COPY infra/world/plugins/dist/ClickMobsRegionGuard/config.yml /plugins/ClickMobsRegionGuard/config.yml
