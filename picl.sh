#!/bin/bash

scriptDir="$(cd "$(dirname "$0")" || exit ; pwd -P)"
root=$(readlink -f "$scriptDir")

cmd="picl"
cmdDir="${root}/cmd/${cmd}"
if [ ! -d "$cmdDir" ] ; then
    echo "Command directory $cmdDir does not exist"
fi
cd "$cmdDir" || exit 1
echo "Building...."

go build -ldflags "-s -w" -race -o "$root/_local/bin/picl" || exit 1

echo "Running...."
echo
# shift
"$root/_local/bin/$cmd" "$@"