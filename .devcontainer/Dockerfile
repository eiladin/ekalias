# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.194.0/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version: 1, 1.16, 1.17
ARG VARIANT="1"
ARG FZF_VERSION
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT} as base

# [Choice] Node.js version: none, lts/*, 16, 14, 12, 10
ARG NODE_VERSION="none"
RUN if [ "${NODE_VERSION}" != "none" ]; then su vscode -c "umask 0002 && . /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

FROM ubuntu:focal AS tools
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends gnupg software-properties-common curl unzip git
    
FROM tools AS kubectl
RUN curl -sSLJO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x ./kubectl

FROM tools AS awscli
RUN curl -sSLJ "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"\
    && unzip awscliv2.zip \
    && ./aws/install

FROM tools AS fzf
ARG FZF_VERSION
RUN curl -sSLJ "https://github.com/junegunn/fzf/releases/download/${FZF_VERSION}/fzf-${FZF_VERSION}-linux_amd64.tar.gz" -o "fzf.tar.gz"\
    && tar zxf fzf.tar.gz \
    && chmod +x ./fzf

FROM tools as krew
RUN groupadd vscode
RUN useradd -s /bin/bash --gid vscode -m vscode
USER vscode
WORKDIR /home/vscode
RUN curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/krew.tar.gz" \
    && tar zxf krew.tar.gz \
    && ./krew-linux_amd64 install krew

FROM base AS final
COPY --from=awscli "/usr/local/aws-cli/" "/usr/local/aws-cli/"
RUN ln -s /usr/local/aws-cli/v2/current/bin/aws /usr/local/bin/aws
RUN ln -s /usr/local/aws-cli/v2/current/bin/aws_completer /usr/local/bin/aws_completer
COPY --from=kubectl "kubectl" "/usr/local/bin/kubectl"
COPY --from=fzf "fzf" "/usr/local/bin/fzf"
USER vscode
COPY --from=krew "/home/vscode/.krew/" "/home/vscode/.krew"
RUN sed -i 's/plugins=(git)/plugins=(git kubectl)/g' ~/.zshrc
ENV PATH="/home/vscode/.krew/bin:${PATH}"