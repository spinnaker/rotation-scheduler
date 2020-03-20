#!/usr/bin/env bash
set -x
set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)
pushd "${SCRIPT_DIR}"

GOPATH=${GOPATH:=$HOME/go}
GOBIN=${GOBIN:=$GOPATH/bin}

go get github.com/golang/protobuf/protoc-gen-go
go get github.com/mitchellh/protoc-gen-go-json

if [[ ":$PATH:" != *":${GOBIN}:"* ]]; then
	echo "Adding $GOBIN to PATH"
	export PATH=$PATH:$GOBIN
fi

./protoc \
  --proto_path=./schedule/ \
  --go_out=paths=source_relative:$SCRIPT_DIR/schedule/ \
  --go-json_out=$SCRIPT_DIR/schedule/ \
  ./**/*.proto

popd
