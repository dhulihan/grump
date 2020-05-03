#!/bin/bash

OUTPUT="`ps aux | grep "go-build"`"

# override IFS
IFS='
'
for LINE in $OUTPUT; do
  KPID=$(echo $LINE | awk '{print $2}')
  echo "killing ${KPID}"
  kill -9 $KPID
done
