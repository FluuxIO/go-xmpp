FROM golang:1.13
WORKDIR /xmpp
RUN curl -o codecov.sh -s https://codecov.io/bash && chmod +x codecov.sh
COPY . ./
