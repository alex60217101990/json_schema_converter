#!/bin/bash

depsPath="$(dirname $BASH_SOURCE)"/tmp-scripts
mkdir -p "$depsPath"

curl -J -s \
  -H 'Accept: application/vnd.github.v3.raw' \
  -o "$depsPath/colors.sh"\
  -L 'https://api.github.com/repos/alex60217101990/bash-tools/contents/colors.sh'

curl -J -s \
  -H 'Accept: application/vnd.github.v3.raw' \
  -o "$depsPath/logger.sh"\
  -L 'https://api.github.com/repos/alex60217101990/bash-tools/contents/logger.sh'

. "$depsPath"/logger.sh -c=true

INFO "$(helm3 env | grep HELM_PLUGIN)"


