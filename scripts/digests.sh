#!/bin/bash

# inspiration shell script

while read in
do
    docker pull "$in"
    repo=$(echo $in | cut -f1 -d/)
    image_tmp=$(echo $in | cut -f2 -d/)
    image=$(echo $image_tmp | cut -f1 -d:)
    tag=$(echo $image_tmp | cut -f2 -d:)
    docker images --digests | grep "$image" | grep "$repo" | grep "$tag" |  awk '{print "| " $1 ":" $2 " | " $3 " |"}' | sed 's/| //g' | sed 's/ |//g'  >> image-digests.txt
    docker rmi "$in"
done < images.txt