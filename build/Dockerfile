FROM golang:1.20.6-bookworm

ARG GH_VERSION='1.9.2'

RUN apt-get update \
    && apt-get -y install \
        rpm \
        gnupg2 \
        gpg-agent \
        debsigs \
        unzip \
        zip

# Since the user does not match the owners of the repo "git rev-parse --is-inside-work-tree" fails and goreleaser does not populate projectName
# https://stackoverflow.com/questions/72978485/git-submodule-update-failed-with-fatal-detected-dubious-ownership-in-repositor
RUN git config --global --add safe.directory '*'
RUN curl -L https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_amd64.deb -o gh_${GH_VERSION}_linux_amd64.deb
RUN dpkg -i gh_${GH_VERSION}_linux_amd64.deb
