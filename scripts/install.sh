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

HELM_PLUGIN_DIR=$(echo "$(helm env | grep HELM_PLUGIN)" | cut -d '=' -f 2-)

INFO "$HELM_PLUGIN_DIR"

function uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    cygwin_nt*) os="windows" ;;
    mingw*) os="windows" ;;
    msys_nt*) os="windows" ;;
  esac
  echo "$os"
}

function uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86_64) arch="amd64" ;;
    x86) arch="386" ;;
    i686) arch="386" ;;
    i386) arch="386" ;;
    aarch64) arch="arm64" ;;
    armv5*) arch="armv5" ;;
    armv6*) arch="armv6" ;;
    armv7*) arch="armv7" ;;
  esac
  echo ${arch}
}

function uname_os_check() {
  os=$(uname_os)
  case "$os" in
    darwin) return 0 ;;
    dragonfly) return 0 ;;
    freebsd) return 0 ;;
    linux) return 0 ;;
    android) return 0 ;;
    nacl) return 0 ;;
    netbsd) return 0 ;;
    openbsd) return 0 ;;
    plan9) return 0 ;;
    solaris) return 0 ;;
    windows) return 0 ;;
  esac
  ERROR "uname_os_check '$(uname -s)' got converted to '$os' which is not a GOOS value. Please file bug at https://github.com/client9/shlib"
  exit 1
}

function uname_arch_check() {
  arch=$(uname_arch)
  case "$arch" in
    386) return 0 ;;
    amd64) return 0 ;;
    arm64) return 0 ;;
    armv5) return 0 ;;
    armv6) return 0 ;;
    armv7) return 0 ;;
    ppc64) return 0 ;;
    ppc64le) return 0 ;;
    mips) return 0 ;;
    mipsle) return 0 ;;
    mips64) return 0 ;;
    mips64le) return 0 ;;
    s390x) return 0 ;;
    amd64p32) return 0 ;;
  esac
  ERROR "uname_arch_check '$(uname -m)' got converted to '$arch' which is not a GOARCH value.  Please file bug report at https://github.com/client9/shlib"
  exit 1
}

function lastrelease() {
  git ls-remote --refs --sort="version:refname" --tags "$1" | cut -d/ -f3- | tail -n1;
}

function adjust_format() {
  # change format (tar.gz or zip) based on OS
  case ${OS} in
    windows) FORMAT=zip ;;
  esac
  true
}

function http_download_curl() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    code=$(curl -w '%{http_code}' -sL -o "$local_file" "$source_url")
  else
    code=$(curl -w '%{http_code}' -sL -H "$header" -o "$local_file" "$source_url")
  fi
  if [ "$code" != "200" ]; then
    DEBUG "http_download_curl received HTTP status $code"
    exit 1
  fi
  return 0
}

PROJECT_NAME="helm-schema-gen"
OWNER=alex60217101990
REPO="json_schema_converter"

VERSION=$(lastrelease "https://github.com/alex60217101990/json_schema_converter.git")
BINARY=schema-generator
FORMAT=tar.gz
OS=$(uname_os)
ARCH=$(uname_arch)

INFO "Version: $VERSION"

#PREFIX="$OWNER/$REPO"
#
#NAME=${PROJECT_NAME}_${VERSION}_${OS}_${ARCH}
#TARBALL=${NAME}.${FORMAT}
#TARBALL_URL=${GITHUB_DOWNLOAD}/${TAG}/${TARBALL}
#CHECKSUM=${PROJECT_NAME}_${VERSION}_checksums.txt
#CHECKSUM_URL=${GITHUB_DOWNLOAD}/${TAG}/${CHECKSUM}

uname_os_check "$OS"
uname_arch_check "$ARCH"
adjust_format

INFO "found version: ${VERSION} for OS: ${OS}, with arch.: ${ARCH}"

tmpdir=$(mktemp -d)
DEBUG "downloading files into ${tmpdir}"
http_download_curl "schema-generator-${VERSION}-${OS}-${ARCH}.${FORMAT}" "${tmpdir}/${TARBALL}" "${TARBALL_URL}"