
ARG TARGETARCH
FROM --platform=linux/$TARGETARCH ubuntu:24.04
RUN apt update && apt install -y curl
ARG AQUAREUM_URL
ENV AQUAREUM_URL $AQUAREUM_URL
RUN echo "downloading $AQUAREUM_URL" && cd /usr/local/bin && curl -L "$AQUAREUM_URL" | tar xzv
RUN aquareum --version
CMD aquareum
