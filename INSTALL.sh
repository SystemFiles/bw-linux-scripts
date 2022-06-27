#!/bin/bash
# Installation Script for all bw-linux-scripts

set -e pipefail

RELEASE_VERSION=`curl -s https://api.github.com/repos/systemfiles/bw-linux-scripts/releases/latest | grep "tag_name" | cut -d: -f2 | tr -d \", | xargs`

REQ_CMDS=(tar curl git jq)
for c in "${REQ_CMDS[@]}"; do
  if ! command -v $c &> /dev/null
  then
    echo "the bw-linux-scripts installer requires $c to be installed on the system. Please install it and try running the installer again"
    exit 1
  fi
done

SCRIPTS=`curl https://api.github.com/repos/systemfiles/bw-linux-scripts/releases/latest | jq '.assets[]|{name,browser_download_url}| select( .browser_download_url | contains("md5")|not)|select( .browser_download_url | contains("zip")|not)'`
SCRIPT_NAMES=`echo $SCRIPTS | jq '.name'`
SCRIPT_URLS=`echo $SCRIPTS | jq '.browser_download_url'`

for (( i=0; i<${#SCRIPTS[@]}; i++ )); do
  if [[ -f "$HOME/scripts/${SCRIPT_NAMES[$i]}" ]]; then
    echo "One or more scripts are already installed."
    exit 0
  fi
done

[ "$(uname -s)" == "Darwin" ] && INSTALL_OS="darwin"
[ "$(uname -s)" == "Linux" ] && INSTALL_OS="linux"

if [[ -z $INSTALL_OS ]]; then
  echo "Current OS not supported by installation script ... exiting!"
  exit 1
fi

[ "$(uname -m)" == "x86_64" ] && INSTALL_ARCH="amd64"
[ "$(uname -m)" == "armv7l" ] && INSTALL_ARCH="arm64"
[ "$(uname -m)" == "i386" ] && INSTALL_ARCH="386"

if [[ -z $INSTALL_ARCH ]]; then
  echo "Current OS Architecture not supported by installation script ... exiting!"
  exit 1
fi

if [[ ! -d "$HOME/bin" ]]; then
  mkdir -p "$HOME/bin"
fi

if [[ ! -d "$HOME/bwtmp" ]]; then
  mkdir -p "$HOME/bwtmp"
fi

for (( i=0; i<${#SCRIPTS[@]}; i++ )); do
  BIN_NAME=$(echo ${SCRIPT_NAMES[$i]} | cut -d'-' -f1 | tr -d '"' | xargs | awk '{print tolower($0)}')
  cd $HOME/bwtmp
  curl -sSLO ${SCRIPT_URLS[$i]}
  tar -zxf ./${SCRIPT_NAMES[$i]}
  mv ./$BIN_NAME $HOME/bin/$BIN_NAME
done
cd; rm -rf $HOME/bwtmp/

echo "Success. Remember to set your PATH properly to point to $HOME/bin in profile or rc config files."