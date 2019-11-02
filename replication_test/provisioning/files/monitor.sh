#!/bin/bash
MYSQLD_PID="$(pgrep mysqld)"
while true; do
    echo "$(date +'%H:%M:%S')" \
        "$(uptime | cut -d: -f5 | cut -d, -f1)" \
        "$(ps q $MYSQLD_PID --format %cpu,%mem | tail -1)" \
        | tee -a monitoring.txt
    sleep 1
done
