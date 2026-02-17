FROM eclipse-temurin:21-jre-alpine

WORKDIR /opt/dataease

RUN mkdir -p /opt/dataease/drivers \
    /opt/dataease/cache \
    /opt/dataease/data/map \
    /opt/dataease/data/static-resource \
    /opt/dataease/data/appearance \
    /opt/dataease/data/exportData \
    /opt/dataease/data/excel \
    /opt/dataease/data/i8n \
    /opt/dataease/data/plugin \
    /opt/dataease/logs

COPY core/core-backend/target/CoreApplication.jar /opt/dataease/app.jar

ENV JAVA_OPTS="-Xms2g -Xmx4g -Dfile.encoding=utf-8"
ENV SERVER_PORT=8100

CMD ["java", "-jar", "/opt/dataease/app.jar", "--spring.profiles.active=standalone"]
