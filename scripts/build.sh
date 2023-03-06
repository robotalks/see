#!/bin/bash

set -ex

# Target must be OS-ARCH, like linux-amd64
TARGET="$1"
test -n "$TARGET"

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
    arm64)
        export GOARCH=arm64
        ;;
    amd64|x86_64)
        export GOARCH=amd64
        ;;
    *)
        echo "unsupported arch" >&2
        exit 1
        ;;
esac

OUTDIR="bin/$TARGET_OS/$TARGET_ARCH"

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
        -ldflags "$LDFLAGS" \
        ./cmd/$cmd
}

mkdir -p $OUTDIR
build-go see
