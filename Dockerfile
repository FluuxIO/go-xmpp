FROM golang:1.9
WORKDIR /go/src/fluux.io/xmpp
COPY . ./
RUN go get github.com/processone/mpg123 github.com/processone/soundcloud