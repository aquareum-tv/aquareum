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
    --target cached-builder \
    --destination $CACHED_BUILD_IMAGE
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

MANIFEST_UNKNOWN="0"
filter() {
  while read line; do
    echo "$line"
    if echo $line | grep MANIFEST_UNKNOWN; then
      MANIFEST_UNKNOWN="1"
    fi
  done
}

build-cache-if-needed() {
  # lmao mad science https://unix.stackexchange.com/a/70675
  set +e
  ((((build; echo $? >&3) | filter >&4) 3>&1) | (read xs; exit $xs)) 4>&1
  status=$?
  set -e
  echo "exit status=$status MANIFEST_UNKNOWN=$MANIFEST_UNKNOWN"
}

if [ "$1" = "build" ]; then
  build
elif [ "$1" = "cache" ]; then
  cache
else
  echo 'unknown option'
  exit 1
fi
