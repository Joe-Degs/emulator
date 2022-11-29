#!/bin/bash

INSTALL_DIR='/opt/riscv-newlib-toolchain'
TMP_DIR='/tmp/riscv-newlib'
OUT_FILE=$TMP_DIR/output.tar.xz
DL_LINK='https://random-oracles.org/riscv/riscv-gnu-toolchain-newlib-multilib-jul-10-2020.tar.xz'
FNAME=$(echo $DL_LINK | awk -F '/' '{print $NF}' | awk -F '.' '{print $1F}')
EXT_DIR=$TMP_DIR/$FNAME

USAGE="USAGE: $0 [(install|add)|(remove|rm)]"

failfunc() {
  echo "command failed: cleaning up $TMP_DIR"
  rm -rf $TMP_DIR
}

cleanup() {
  echo "cleaning up $TMP_DIR"
  rm -rf $TMP_DIR
  exit 0
}

trap cleanup SIGINT

install() {
  mkdir -p $TMP_DIR
  
  if ! curl $DL_LINK -o $OUT_FILE; then
    failfunc
    exit 1
  fi
  
  if ! tar -C $TMP_DIR -xvf $OUT_FILE; then
    failfunc
    exit 1
  fi
  
  if ! sudo mv $EXT_DIR $INSTALL_DIR; then
    failfunc
    exit 1
  fi

  rm -rf $TMP_DIR
}

uninstall() {
  sudo rm -rf $INSTALL_DIR
}

case "$1" in
  install|add)
    install
    ;;
    
  remove|rm)
    uninstall
    ;;
    
  *)
    echo "$USAGE" >&2
    ;;
esac
