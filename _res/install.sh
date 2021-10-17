#!/bin/bash

deploymentDir="/opt/bin"
serverExe="fx"


logFilePrefix="${deploymentDir}/fx"


if [ -f "${logFilePrefix}.log" ] ; then
    if [ -f "${logFilePrefix}_prev.log" ] ; then
        rm -f "${logFilePrefix}_prev.log"
    fi
    mv "${logFilePrefix}.log" "${logFilePrefix}_prev.log"
fi
touch "${logFilePrefix}.log"

# May be use PID file later
killall "${serverExe}"

nohup "${deploymentDir}/${serverExe}" > "${logFilePrefix}.log" 2>&1 &
