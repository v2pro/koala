#!/bin/bash
set -e
set -x

case $1 in
    "tracer" )
        # trace & record to file, only for testing purpose
        export CGO_CFLAGS="-DKOALA_LIBC_NETWORK_HOOK -DKOALA_LIBC_FILE_HOOK"
        export CGO_CPPFLAGS=$CGO_CFLAGS
        exec go build -tags="koala_tracer koala_recorder" -buildmode=c-shared -o output/koala-tracer.so github.com/v2pro/koala/cmd/recorder
        ;;
    "recorder" )
        # record to file, only for testing purpose
        export CGO_CFLAGS="-DKOALA_LIBC_NETWORK_HOOK -DKOALA_LIBC_FILE_HOOK"
        export CGO_CPPFLAGS=$CGO_CFLAGS
        exec go build -tags="koala_recorder" -buildmode=c-shared -o output/koala-recorder.so github.com/v2pro/koala/cmd/recorder
        ;;
    "vendor" )
        if [ ! -d /tmp/build-golang/src/github.com/v2pro ]; then
            mkdir -p /tmp/build-golang/src/github.com/v2pro
            ln -s $PWD /tmp/build-golang/src/github.com/v2pro/koala
        fi
        export GOPATH=/tmp/build-golang
        go get github.com/Masterminds/glide
        cd /tmp/build-golang/src/github.com/v2pro/koala
        exec $GOPATH/bin/glide i
        ;;
esac

# build replayer by default
export CGO_CFLAGS="-DKOALA_LIBC_NETWORK_HOOK -DKOALA_LIBC_FILE_HOOK -DKOALA_LIBC_TIME_HOOK -DKOALA_LIBC_PATH_HOOK"
export CGO_CPPFLAGS=$CGO_CFLAGS
go build -tags="koala_replayer" -buildmode=c-shared -o output/koala-replayer.so github.com/v2pro/koala/cmd/replayer

cp output/koala-replayer.so ~/DiDiCode/Midi/res/replayer/koala-replayer.so
rm -rf /tmp/nuwa_dev/replayer/koala-replayer.so
