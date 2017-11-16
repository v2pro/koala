#!/bin/bash
set -e
set -x

if [ ! -d /tmp/build-golang/src/github.com/v2pro ]; then
    mkdir -p /tmp/build-golang/src/github.com/v2pro
    ln -s $PWD /tmp/build-golang/src/github.com/v2pro/koala
fi
export GOPATH=/tmp/build-golang
rm -rf output
mkdir output
go get github.com/Masterminds/glide
pushd /tmp/build-golang/src/github.com/v2pro/koala
/tmp/build-golang/bin/glide i
popd

case $1 in
    "tracer" )
        export CGO_CFLAGS="-DKOALA_LIBC_NETWORK_HOOK -DKOALA_LIBC_FILE_HOOK"
        export CGO_CPPFLAGS=$CGO_CFLAGS
        exec go build -tags="koala_tracer koala_recorder" -buildmode=c-shared -o output/koala-tracer.so github.com/v2pro/koala/cmd/recorder
        ;;
esac

export CGO_CFLAGS="-DKOALA_LIBC_NETWORK_HOOK -DKOALA_LIBC_FILE_HOOK -DKOALA_LIBC_TIME_HOOK -DKOALA_LIBC_PATH_HOOK"
export CGO_CPPFLAGS=$CGO_CFLAGS
go build -tags="koala_replayer" -buildmode=c-shared -o output/koala-replayer.so github.com/v2pro/koala/cmd/replayer
