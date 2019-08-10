#!/bin/bash

##########
# very unfinished. mostly a note for myself for the params to openssl and curl to make a bash script console client
#########
 
 
# bash can't store null bytes in a var. we need the base64 and hex
# representation of the key. Pipes can handle null bytes, so we get
# the key as base64 and then decode it and pipe it to get the hex value
 
KEYBASE=$(openssl rand -base64 32)
KEYASHEX=$(echo -n $KEYBASE | base64 -d | xxd -p -c 10000) # linebreak after 10000 chars. i.e. never
echo $KEYBASE
openssl enc -aes-256-cbc -K "$KEYASHEX" -iv $(printf '0%.0s' {1..32}) |  # print '0' 32 times = 16 bytes as hex
    curl -X POST -H "Content-Type: application/octet-stream" --data-binary \
        @- "https://host.name/store?exp=1" # @- is the curl way to specify stdin
