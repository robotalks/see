#!/bin/bash

set -ex

test -d "$HMAKE_PROJECT_DIR"

LDFLAGS=
if [ -n "$RELEASE" ]; then
  LDFLAGS="$LDFLAGS -X main.VersionSuffix="
else
  VERSION_SUFFIX=$(git log -1 --format=%h || true)
  if [ -n "$VERSION_SUFFIX" ]; then
    test -z "$(git status --porcelain || true)" || VERSION_SUFFIX="${VERSION_SUFFIX}+"
    LDFLAGS="$LDFLAGS -X main.VersionSuffix=-g${VERSION_SUFFIX}"
  fi
fi

build-go() {
    local cmd=$1
    CGO_ENABLED=0 go build -o $HMAKE_PROJECT_DIR/bin/$cmd \
        -a -tags 'static_build netgo' -installsuffix netgo \
        -ldflags "$LDFLAGS -extldflags -static" \
        github.com/robotalks/simulator/cmd/$cmd
}

mkdir -p $HMAKE_PROJECT_DIR/bin
build-go sim-ng
rice append -i github.com/robotalks/simulator/vis \
    --exec $HMAKE_PROJECT_DIR/bin/sim-ng
