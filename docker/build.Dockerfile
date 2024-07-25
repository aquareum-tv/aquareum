FROM ubuntu:22.04 AS builder

ARG TARGETARCH
ENV TARGETARCH $TARGETARCH

ENV GO_VERSION 1.22.4
ENV NODE_VERSION 22.3.0

RUN apt update \
  && apt install -y build-essential curl git openjdk-17-jdk unzip jq g++ python3-pip ninja-build gcc-aarch64-linux-gnu g++-aarch64-linux-gnu clang lld qemu-user-static \
  && pip install meson \
  && curl -L --fail https://go.dev/dl/go$GO_VERSION.linux-$TARGETARCH.tar.gz -o go.tar.gz \
  && tar -C /usr/local -xf go.tar.gz \
  && rm go.tar.gz
ENV PATH $PATH:/usr/local/go/bin:/root/go/bin

RUN  echo 'deb [arch=arm64] http://ports.ubuntu.com/ jammy main multiverse universe' >> /etc/apt/sources.list \
  && echo 'deb [arch=arm64] http://ports.ubuntu.com/ jammy-security main multiverse universe' >> /etc/apt/sources.list \
  && echo 'deb [arch=arm64] http://ports.ubuntu.com/ jammy-backports main multiverse universe' >> /etc/apt/sources.list \
  && echo 'deb [arch=arm64] http://ports.ubuntu.com/ jammy-updates main multiverse universe' >> /etc/apt/sources.list \
  && dpkg --add-architecture arm64 \
  && bash -c "apt update || echo 'ignoring errors'" \
  && apt install -y libc6:arm64 libstdc++6:arm64

RUN export NODEARCH="$TARGETARCH" \
  && if [ "$TARGETARCH" = "amd64" ]; then export NODEARCH="x64"; fi \
  && curl -L --fail https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-$NODEARCH.tar.xz -o node.tar.gz \
  && tar -xf node.tar.gz \
  && cp -r node-v$NODE_VERSION-linux-$NODEARCH/* /usr/local \
  && rm -rf node.tar.gz node-v$NODE_VERSION-linux-$NODEARCH
  
RUN npm install -g yarn

ARG ANDROID_SDK_VERSION=11076708
ENV ANDROID_HOME /opt/android-sdk
RUN mkdir -p ${ANDROID_HOME}/cmdline-tools && \
    curl -L -O https://dl.google.com/android/repository/commandlinetools-linux-${ANDROID_SDK_VERSION}_latest.zip && \
    unzip *tools*linux*.zip -d ${ANDROID_HOME}/cmdline-tools && \
    mv ${ANDROID_HOME}/cmdline-tools/cmdline-tools ${ANDROID_HOME}/cmdline-tools/tools && \
    rm *tools*linux*.zip && \
    curl -L https://raw.githubusercontent.com/thyrlian/AndroidSDK/bfcbf0cdfd6bb1ef45579e6ddc4d3876264cbdd1/android-sdk/license_accepter.sh | bash

FROM builder AS cached-builder
WORKDIR /cached-build
RUN git clone https://git.aquareum.tv/aquareum-tv/aquareum && cd aquareum && make all -j$(nproc)
