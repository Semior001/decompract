#!/bin/sh
echo "Running application"
export LOCATION=/db
/go/build/app $@
