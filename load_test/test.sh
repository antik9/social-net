#!/bin/bash
[ $# -gt 0 ] \
    && wrk -t1 -c"$1" -d60 --timeout 60 \
        -s scripts.lua http://localhost:8080 > "get_messages_tarantool_${1}c.txt" \
    || echo "Provide number of connections"
