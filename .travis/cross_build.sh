#!/usr/bin/env bash

# Functions.
function info() {
    echo -e "\033[00;34mi\033[0m $1"
}

function fail() {
    echo -e "\033[00;31m!\033[0m $1"
    exit 1
}

function build() {
    goos=$1
    goarch=$2

    info "Building for $goos $goarch..."
    GOOS=$goos GOARCH=$goarch go build -o $goos_$goarch main.go
    if [[ $? -ne 0 ]]; then
        fail "Could not build for $goos $goarch. Aborting."
    fi
}

# Create build directory.
mkdir -p bin
cd bin

# Create go.mod
cat <<EOF > go.mod
module cross_build
require github.com/trangsystems/tpdf v3.0.0
EOF

echo "replace github.com/trangsystems/tpdf => $TRAVIS_BUILD_DIR" >> go.mod

# Create Go file.
cat <<EOF > main.go
package main

import (
	_ "github.com/trangsystems/tpdf/annotator"
	_ "github.com/trangsystems/tpdf/common"
	_ "github.com/trangsystems/tpdf/common/license"
	_ "github.com/trangsystems/tpdf/contentstream"
	_ "github.com/trangsystems/tpdf/contentstream/draw"
	_ "github.com/trangsystems/tpdf/core"
	_ "github.com/trangsystems/tpdf/core/security"
	_ "github.com/trangsystems/tpdf/core/security/crypt"
	_ "github.com/trangsystems/tpdf/creator"
	_ "github.com/trangsystems/tpdf/extractor"
	_ "github.com/trangsystems/tpdf/fdf"
	_ "github.com/trangsystems/tpdf/fjson"
	_ "github.com/trangsystems/tpdf/model"
	_ "github.com/trangsystems/tpdf/model/optimize"
	_ "github.com/trangsystems/tpdf/model/sighandler"
	_ "github.com/trangsystems/tpdf/ps"
	_ "github.com/trangsystems/tpdf/render"
)

func main() {}
EOF

# Build file.
for os in "linux" "darwin" "windows"; do
    for arch in "386" "amd64"; do
        build $os $arch
    done
done
