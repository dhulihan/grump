FROM golang:1.14-stretch

RUN apt-get update && apt-get install libasound2-dev build-essential -y -q

#RUN go get -u -v github.com/goreleaser/goreleaser
RUN curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh

WORKDIR /data

ENTRYPOINT ['./scripts/build.sh']
