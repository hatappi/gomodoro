#!/bin/bash

pkgs=`grep -E "^\s+_\s" tools.go | sed -E 's/^[[:space:]]+_[[:space:]]"([^"]+)"/\1/'`

for pkg in ${pkgs[@]}
do
  echo "install ${pkg} ..."
  go install ${pkg}
done
