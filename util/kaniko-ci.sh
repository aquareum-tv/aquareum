#!/bin/sh

set -eu

dockerfile="$CI_PROJECT_DIR/docker/build.Dockerfile"

dockerhash() {
  cat $dockerfile | sha256sum | awk '{ print $1 }'
}

CACHED_BUILD_IMAGE="$CI_REGISTRY_IMAGE:buildcache-`dockerhash`"
echo "CACHED_BUILD_IMAGE=$CACHED_BUILD_IMAGE"

cache() {
  /kaniko/executor \
    --build-arg TARGETARCH=amd64 \
    --cache=true \
    --context "$CI_PROJECT_DIR" \
    --dockerfile "$CI_PROJECT_DIR/docker/build.Dockerfile" \
    --cache-repo "$CI_REGISTRY_IMAGE" \
    --use-new-run \
    --target cached-builder
}

build() {
  hash=`dockerhash`
  /kaniko/executor \
    --build-arg CACHED_BUILD_IMAGE="$CACHED_BUILD_IMAGE" \
    --build-arg TARGETARCH=amd64 \
    --build-arg CI_API_V4_URL=$CI_API_V4_URL \
    --build-arg CI_COMMIT_SHA=$CI_COMMIT_SHA \
    --build-arg CI_JOB_TOKEN=$CI_JOB_TOKEN \
    --build-arg CI_PROJECT_DIR=$CI_PROJECT_DIR \
    --build-arg CI_PROJECT_ID=$CI_PROJECT_ID \
    --build-arg CI_REGISTRY_IMAGE=$CI_REGISTRY_IMAGE \
    --build-arg CI_REPOSITORY_URL=$CI_REPOSITORY_URL \
    --build-arg CI_COMMIT_TAG=${CI_COMMIT_TAG:-} \
    --build-arg CI_COMMIT_BRANCH=${CI_COMMIT_BRANCH:-} \
    --cache=true \
    --cache-repo "$CI_REGISTRY_IMAGE" \
    --context "$CI_PROJECT_DIR" \
    --dockerfile "$CI_PROJECT_DIR/docker/build.Dockerfile" \
    --use-new-run \
    --no-push \
    --no-push-cache \
    --skip-unused-stages
}

if [ "$1" = "build" ]; then
  build
elif [ "$1" = "cache" ]; then
  cache
else
  echo 'unknown option'
  exit 1
fi
