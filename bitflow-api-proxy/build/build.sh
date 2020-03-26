#!/bin/bash
set -x
home=`dirname $(readlink -f $0)`
root=`readlink -f "$home/../.."`

test $# -ge 2 || { echo "Parameters: <container to build for (arm32v7/arm64v8/alpine)> <Go-mod-cache directory> <Build args (optional)>"; exit 1; }
BUILD_TARGET="$1"
BUILD_IMAGE="teambitflow/golang-build:$BUILD_TARGET"
BUILD_DIR="bitflow-api-proxy/build/_output/bin/$BUILD_TARGET"
echo "Building into $BUILD_DIR"
shift

mod_cache_dir="$1/$BUILD_TARGET"
echo "Using Go-mod-cache directory: $mod_cache_dir"
mkdir -p "$mod_cache_dir"
shift

build_args="$@"


echo -e "\n==== 1"
pwd
ls

echo -e "\n==== 2"
ls $mod_cache_dir

echo -e "\n==== 3"
ls $root

# Build inside the container, but mount relevant directories to get access to the build results.
docker run -v "$mod_cache_dir:/go" -v "$root:/build/src" "$BUILD_IMAGE" \
  sh -c "
    set -x
    echo PWD
    pwd
    echo LISTING
    ls src

    # Copy entire source-tree in order to make changes to go.mod/go.sum
    cp -r src build
    cd build

  echo PWD2
  pwd
  ls

    # Prepare go.mod/go.sum files
    sed -i \$(find -name go.mod) -e '\_//.*gitignore\$_d' -e '\_#.*gitignore\$_d'
    find -name go.sum -delete

    # Build the collector and plugins, put the outputs in the mounted source folder
    cd bitflow-api-proxy
    go build -o ../../src/$BUILD_DIR/bitflow-api-proxy $build_args .
  "
