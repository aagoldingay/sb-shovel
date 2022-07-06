#!/bin/bash

# usage : ./releaseBundle.sh <version>

os=("windows" "darwin" "linux")

for goos in "${os[@]}"
do
    GOOS=$goos go build
    filename=""
    if [ $goos = "windows" ]; then
        filename=".exe"
    fi

    zip "release/sb-shovel_$1_${goos}_amd64.zip" "sb-shovel$filename"

    rm "sb-shovel$filename"
done