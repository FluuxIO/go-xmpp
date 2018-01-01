FROM golang:1.9
WORKDIR /go/src/fluux.io/xmpp
COPY . ./
RUN apt-get update \
    && apt-get install -y \
    git \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*
RUN go get -t -v ./...