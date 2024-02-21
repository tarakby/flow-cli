#!/bin/sh

# Exit as soon as any command fails
set -e

# INPUT_GOARCH and INPUT_GOOS are set by the wangyoucao577 action:
# https://github.com/wangyoucao577/go-release-action/blob/v1.40/action.yml#L109
# it represents the target arch and target os

export GOOS=${INPUT_GOOS} 
export GOARCH=${INPUT_GOARCH}

C_COMP=""
if [ "${GOOS}" = "linux" ]; then
    if [ "${GOARCH}" = "arm64" ]; then
        C_COMP="aarch64-linux-gnu-gcc"
    elif [ "${GOARCH}" = "amd64" ]; then
        C_COMP="x86_64-linux-gnu-gcc"
    fi
elif [ "${GOOS}" = "windows" ]; then
    if [ "${GOARCH}" = "amd64" ]; then
        C_COMP="x86_64-w64-mingw-gcc"
    else 
        { echo "arm64 on windows isn't supported"; exit 1; }
    fi
fi


export GO111MODULE=on
# enable CGO because it is requiired by the onflow/crypto package
export CGO_ENABLED=1
# set the correct C compiler
export CC=${C_COMP}
# this flag disables non-portable code that requires specific CPU features.
export CGO_FLAGS=\""-O2 -D__BLST_PORTABLE__"\"

