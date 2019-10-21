#!/bin/bash
[ $# -gt 1 ] \
    && wrk -t1 -c"$2" -d60 --timeout 60 \
        -s scripts.lua http://localhost:8080 > "${1}_index_with_${2}c.txt" \
    || echo "Provide after/before parameter and number of connections"
