FROM maven:3.9.9-eclipse-temurin-21 AS plugin-build
WORKDIR /src
COPY infra/velocity/plugins/src/link-code-gate /src/link-code-gate
RUN mvn -f /src/link-code-gate/pom.xml -DskipTests package

FROM itzg/mc-proxy:java25
COPY --from=plugin-build /src/link-code-gate/target/link-code-gate-0.1.0.jar /plugins/LinkCodeGate.jar
