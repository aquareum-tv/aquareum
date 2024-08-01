
FROM ubuntu:24.04
ARG AQUAREUM_URL
RUN apt update && apt install -y curl
ENV AQUAREUM_URL $AQUAREUM_URL
RUN echo "downloading $AQUAREUM_URL" && cd /usr/local/bin && curl -L "$AQUAREUM_URL" | tar xzv
CMD aquareum
