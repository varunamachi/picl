#!/bin/bash

scriptDir="$(cd "$(dirname "$0")" || exit ; pwd -P)"
root=$(readlink -f "$scriptDir/../..")

HOST=${REMOTE_HOST:-"oldman"} # Change based on your setup
USER=${REMOTE_USER:-"pi"}

if [ $# -gt 0 ] ; then
    HOST="$1"
fi
if [ $# -gt 1 ] ; then 
    USER="$2"
fi

buildPath="$root/_local/bin/arm"
if [[ ! -d "$buildPath" ]]; then 
    mkdir -p "$buildPath" || exit 1
fi

echo "Bulding..."
cd "cmd/picl" || exit 1
GOARCH=arm GOOS=linux go build -ldflags "-s -w" -o  "$buildPath" || exit 1
echo "Generated at $buildPath"

rsync -avz -e ssh "$buildPath/picl" "$USER@$HOST:/opt/bin"

# ssh "$USER@$HOST" 'killall -9 teak'
# ssh "$USER@$HOST" 'nohup "/opt/bin/teak" serve --port 9999 > console.log 2>&1 &'