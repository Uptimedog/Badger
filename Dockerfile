FROM golang:1.23.2

ARG BADGER_VERSION=0.3.0

RUN mkdir -p /app/configs
RUN mkdir -p /app/var/logs
RUN apt-get update

WORKDIR /app

RUN curl -sL https://github.com/Uptimedog/Badger/releases/download/v${BADGER_VERSION}/badger_Linux_x86_64.tar.gz | tar xz
RUN rm LICENSE
RUN rm README.md

COPY ./config.dist.yml /app/configs/

EXPOSE 8080

VOLUME /app/configs
VOLUME /app/var

RUN ./badger version

CMD ["./badger", "server", "-c", "/app/configs/config.dist.yml"]
