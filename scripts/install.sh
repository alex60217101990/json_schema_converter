#!/bin/bash

depsPath="$(cd "$(dirname "$(readlink -f "$BASH_SOURCE" || echo "$BASH_SOURCE")")" && pwd)"/tmp-scripts
mkdir -p "$depsPath"

curl -J -s \
  -H 'Accept: application/vnd.github.v3.raw' \
  -o "$depsPath/colors.sh"\
  -L 'https://api.github.com/repos/alex60217101990/bash-tools/contents/colors.sh'

curl -J -s \
  -H 'Accept: application/vnd.github.v3.raw' \
  -o "$depsPath/logger.sh"\
  -L 'https://api.github.com/repos/alex60217101990/bash-tools/contents/logger.sh'

curl -J -s \
  -H 'Accept: application/vnd.github.v3.raw' \
  -o "$depsPath/helpers.sh"\
  -L 'https://api.github.com/repos/alex60217101990/bash-tools/contents/helpers.sh'

. "$depsPath"/helpers.sh
. "$depsPath"/logger.sh -c=true

INFO "${HELM_PLUGIN_DIR} => ${HELM_PLUGINS}/${HELM_PLUGIN_NAME}"

HELM_PLUGIN_DIR="$(echo "$(helm env | grep HELM_PLUGIN)" | cut -d '=' -f 2- | tr -d '"')/json_schema_converter"

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

function is_command() {
  command -v "$1" >/dev/null
}

function untar() {
  tarball=$1
  dir=$2
  case "${tarball}" in
    *.tar.gz | *.tgz | *.tar.gz.sha256 | *.tgz.sha256) tar --no-same-owner -xzf "${tarball}" --directory "${dir}" ;;
    *.tar | *.tar.sha256) tar --no-same-owner -xf "${tarball}" --directory "${dir}" ;;
    *.zip | *.zip.sha256) unzip "${tarball}" -d "${dir}" ;;
    *)
      ERROR "untar unknown archive format for ${tarball}"
      exit 1
      ;;
  esac
}

function hash_sha256() {
  TARGET=${1:-/dev/stdin}
  if is_command gsha256sum; then
    hash=$(gsha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command sha256sum; then
    hash=$(sha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command shasum; then
    hash=$(shasum -a 256 "$TARGET" 2>/dev/null) || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command openssl; then
    hash=$(openssl -dst openssl dgst -sha256 "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f a
  else
    ERROR "hash_sha256 unable to find command to compute sha-256 hash"
    return 1
  fi
}

function hash_sha256_verify() {
  TARGET=$1
  checksum=$2
  if [ -n "$checksums" ]; then
    ERROR "hash_sha256_verify checksum string not specified in arg2"
    exit 1
  fi

  got=$(hash_sha256 "$TARGET")
  if [ "$checksum" != "$got" ]; then
    ERROR "hash_sha256_verify checksum for '$TARGET' did not verify ${want} vs $got"
    exit 1
  fi
}

OWNER=alex60217101990
REPO="json_schema_converter"

VERSION=$(lastrelease "https://github.com/${OWNER}/${REPO}.git")
BINARY=schema-generator
FORMAT=tar.gz
OS=$(uname_os)
ARCH=$(uname_arch)

INFO "Version: $VERSION"

uname_os_check "$OS"
uname_arch_check "$ARCH"
adjust_format

INFO "found version: ${VERSION} for OS: ${OS}, with arch.: ${ARCH}"

NAME="schema-generator-${VERSION}-${OS}-${ARCH}"
TARBALL="${NAME}.${FORMAT}"
CHECKSUM_ARCH="${NAME}.${FORMAT}"

tmpdir=$(mktemp -d)

DEBUG "downloading files into ${tmpdir}"
http_download_curl "${tmpdir}/${TARBALL}" "https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}/${TARBALL}"
http_download_curl "${tmpdir}/${TARBALL}.sha256" "https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}/${TARBALL}.sha256"

DEBUG "sha256 verify: ${tmpdir}"
hash_sha256_verify "${tmpdir}/${TARBALL}" "$(cat "${tmpdir}/${TARBALL}.sha256")"

DEBUG "untar files into ${tmpdir}"
untar "${tmpdir}/${TARBALL}" "${tmpdir}"
BlueStr "$(ls -hla "$tmpdir")\n"

DEBUG "move binary file into ${HELM_PLUGIN_DIR}/bin"
bash -c "${tmpdir}/schema-generator -h"
BlueStr "$(ls -hla "${HELM_PLUGIN_DIR}/bin")\n"
cp -fL "${tmpdir}/schema-generator" "${HELM_PLUGIN_DIR}/bin"

rm -rf "${tmpdir}"
rm -rf "${depsPath}"