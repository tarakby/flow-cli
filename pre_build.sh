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

GO111MODULE=on
CGO_ENABLED=1
CC=${C_COMP}
CGO_FLAGS=\""-O2 -D__BLST_PORTABLE__"\"
