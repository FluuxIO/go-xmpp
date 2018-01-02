FROM golang:1.9
WORKDIR /go/src/fluux.io/xmpp
RUN curl -o codecov.sh -s https://codecov.io/bash && chmod +x codecov.sh
COPY . ./
RUN go get -t  ./...
