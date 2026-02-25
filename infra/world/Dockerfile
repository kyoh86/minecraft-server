FROM maven:3.9.9-eclipse-temurin-21 AS plugin-build
WORKDIR /src
COPY infra/world/plugins/clickmobs-region-guard /src/clickmobs-region-guard
COPY infra/world/plugins/hub-terraform /src/hub-terraform
RUN mvn -f /src/clickmobs-region-guard/pom.xml -DskipTests package
RUN mvn -f /src/hub-terraform/pom.xml -DskipTests package

FROM itzg/minecraft-server:java25
COPY --from=plugin-build /src/clickmobs-region-guard/target/clickmobs-region-guard-0.1.0.jar /plugins/ClickMobsRegionGuard.jar
COPY --from=plugin-build /src/hub-terraform/target/hub-terraform-0.1.0.jar /plugins/HubTerraform.jar
COPY infra/world/plugins/clickmobs/config/config.yml /plugins/ClickMobs/config.yml
COPY infra/world/plugins/clickmobs-region-guard/config/config.yml /plugins/ClickMobsRegionGuard/config.yml
