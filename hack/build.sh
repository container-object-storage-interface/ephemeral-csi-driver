#! /usr/bin/env bash

set -e

REPO_ROOT=$(readlink -f "$(dirname ${BASH_SOURCE})"/../)

GOOS=linux
GOARCH=amd64
TARGET="$REPO_ROOT/main.go"
OUTPUT="$REPO_ROOT/build/csi-cosi-adapter"
TAG="quay.io/jcope/csi-cosi-adapter"

build() {
  (
    printf "cd'ing to dir: %s\n" $PWD
    cd "$REPO_ROOT" || exit 1
    printf "rm'ing stale binary\n"
    rm -f "$OUTPUT"
    printf "compiling <%s>\n\tto => %s\n" "$TARGET" "$OUTPUT"
    GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT" "$TARGET"
    printf "done compiling\n"
  )
}

image() {
  (
    cd "${REPO_ROOT}/build" || exit 2
    printf "building image: %s\n" $TAG
    docker build -t "$TAG" "$REPO_ROOT/build"
  )
}

build
image