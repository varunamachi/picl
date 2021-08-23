#!/bin/bash

scriptDir="$(cd "$(dirname "$0")" || exit ; pwd -P)"
root=$(readlink -f $scriptDir)

cmd="fx"
cmdDir="${root}/cmd/${cmd}"
if [ ! -d "$cmdDir" ] ; then
    echo "Command directory $cmdDir does not exist"
fi
cd "$cmdDir" || exit -1
echo "Building...."
go build -o "$root/_local/bin/fx" || exit -1
echo "Running...."
echo
# shift
"$root/_local/bin/$cmd" "$@"