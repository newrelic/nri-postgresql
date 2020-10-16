FROM golang:1.15.2-buster

RUN apt-get update \
    && apt-get -y install \
        rpm \
        hub \
        gnupg2 \
        gpg-agent \
        debsigs \
        unzip \
        zip \
    && go get golang.org/dl/go1.9.7 \
    && /go/bin/go1.9.7 download
