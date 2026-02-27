FROM maven:3.9.9-eclipse-temurin-21 AS plugin-build
WORKDIR /src
COPY infra/world/plugins/clickmobs-region-guard /src/clickmobs-region-guard
COPY infra/world/plugins/region-status-ui /src/region-status-ui
COPY infra/world/plugins/spawn-safety-guard /src/spawn-safety-guard
COPY infra/world/plugins/hub-terraform /src/hub-terraform
RUN mvn -f /src/clickmobs-region-guard/pom.xml -DskipTests package
RUN mvn -f /src/region-status-ui/pom.xml -DskipTests package
RUN mvn -f /src/spawn-safety-guard/pom.xml -DskipTests package
RUN mvn -f /src/hub-terraform/pom.xml -DskipTests package

FROM itzg/minecraft-server:java25
COPY --from=plugin-build /src/clickmobs-region-guard/target/clickmobs-region-guard-0.1.0.jar /plugins/ClickMobsRegionGuard.jar
COPY --from=plugin-build /src/region-status-ui/target/region-status-ui-0.1.0.jar /plugins/RegionStatusUI.jar
COPY --from=plugin-build /src/spawn-safety-guard/target/spawn-safety-guard-0.1.0.jar /plugins/SpawnSafetyGuard.jar
COPY --from=plugin-build /src/hub-terraform/target/hub-terraform-0.1.0.jar /plugins/HubTerraform.jar
COPY infra/world/plugins/clickmobs/config/config.yml /plugins/ClickMobs/config.yml
COPY infra/world/plugins/clickmobs-region-guard/config/config.yml /plugins/ClickMobsRegionGuard/config.yml
COPY infra/world/plugins/region-status-ui/config/config.yml /plugins/RegionStatusUI/config.yml
COPY infra/world/plugins/spawn-safety-guard/config/config.yml /plugins/SpawnSafetyGuard/config.yml
