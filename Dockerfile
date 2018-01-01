FROM golang:1.9
WORKDIR /go/src/fluux.io/xmpp
COPY . ./
RUN go get