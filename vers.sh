#!/bin/sh

version=$(git describe --tags | sed -n '1p' | tr - + | tr -d '\n') >/dev/null 2>&1
if [ $? -eq 0 ]; then
  printf ${version}
else
  printf unknown
fi
