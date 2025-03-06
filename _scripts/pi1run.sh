#!/bin/bash

scriptDir="$(cd "$(dirname "$0")" || exit ; pwd -P)"
root=$(readlink -f $scriptDir/../..)

HOST=${REMOTE_HOST:-"oldman"} # Change based on your setup
USER=${REMOTE_USER:-"pi"}

"${scriptDir}/pi1copy.sh" || exit 1

ssh "$USER@$HOST" "/opt/bin/picl" "$@"

