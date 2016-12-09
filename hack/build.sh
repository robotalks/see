#!/bin/bash

set -ex

TARGET="$1"
test -n "$TARGET"
test -d "$HMAKE_PROJECT_DIR"

TARGET_OS="${TARGET%%-*}"
TARGET_ARCH="${TARGET##*-}"
test -n "$TARGET_OS"
test -n "$TARGET_ARCH"

case "$TARGET_ARCH" in
    armhf)
        export GOARCH=arm
        export GOARM=7
        ;;
    arm)
        export GOARCH=arm
        ;;
    amd64)
        export GOARCH=amd64
        ;;
    *)
        echo "unsupported arch" >&2
        exit 1
        ;;
esac

OUTDIR="$HMAKE_PROJECT_DIR/bin/$TARGET_OS/$TARGET_ARCH"

LDFLAGS="-s -w"
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
    CGO_ENABLED=0 GOOS=$TARGET_OS go build -o $OUTDIR/$cmd \
        -a -tags 'static_build netgo' -installsuffix netgo \
        -ldflags "$LDFLAGS -extldflags -static" \
        ./cmd/$cmd
}

mkdir -p $OUTDIR
build-go see
rice append -i github.com/robotalks/see/vis \
    --exec $OUTDIR/see
