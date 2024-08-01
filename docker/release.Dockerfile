
FROM ubuntu:24.04@sha256:010f94447a26deb0dcdbbeb08d7dfcd87b64c40b4f25d0cf4d582949b735039d
ARG AQUAREUM_URL
RUN apt update && apt install -y curl
ENV AQUAREUM_URL $AQUAREUM_URL
RUN echo "downloading $AQUAREUM_URL" && cd /usr/local/bin && curl -L "$AQUAREUM_URL" | tar xzv
CMD aquareum
