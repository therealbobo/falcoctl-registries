#!/bin/bash

REGISTRY="$1"
TMPD=$(mktemp -d)
FILE=$(basename "$2")

cp -v "$2" $TMPD

shift 2

cd $TMPD

tar -caf "$FILE".tar.gz "$FILE"

sudo falcoctl registry push "$REGISTRY" "$FILE".tar.gz $@

cd -

rm -vfr $TMPD
