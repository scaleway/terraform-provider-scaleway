# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.231.6/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version (use -bullseye variants on local arm64/Apple Silicon): 1, 1.16, 1.17, 1-bullseye, 1.16-bullseye, 1.17-bullseye, 1-buster, 1.16-buster, 1.17-buster
ARG VARIANT="1.18-bullseye"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

COPY .terraformrc /home/vscode/.terraformrc

# Check out the latest release available on github <https://github.com/scaleway/scaleway-cli/releases/latest>
ARG SCW_CLI_VERSION="2.5.1"
# Download the release from github
RUN curl -o /usr/local/bin/scw -L "https://github.com/scaleway/scaleway-cli/releases/download/v${SCW_CLI_VERSION}/scaleway-cli_${SCW_CLI_VERSION}_linux_amd64"
# Allow executing file as program
RUN chmod +x /usr/local/bin/scw
