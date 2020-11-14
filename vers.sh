#!/bin/sh

version=$(git describe --tags | sed -n '1p' | tr - + | tr -d '\n')
if [ $? -eq 0 ]; then
  printf ${version}
else
  printf unknown
fi
