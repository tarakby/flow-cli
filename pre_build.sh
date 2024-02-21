#!/bin/sh

# Exit as soon as any command fails
set -e

# INPUT_GOARCH is a variable set by the wangyoucao577 action
# https://github.com/wangyoucao577/go-release-action/blob/v1.40/action.yml#L109
C_COMP=""
if [ "${INPUT_GOARCH}" = "arm64" ]; then
    C_COMP="aarch64-linux-gnu-gcc"
elif [ "${INPUT_GOARCH}" = "amd64" ]; then
    C_COMP="x86_64-linux-gnu-gcc"
fi

export GO111MODULE=on
export CGO_ENABLED=1
export CC=${C_COMP}
export CGO_FLAGS=\""-O2 -D__BLST_PORTABLE__"\"