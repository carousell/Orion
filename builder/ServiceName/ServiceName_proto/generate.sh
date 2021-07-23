#!/bin/bash
echo "generating proto"
protoc -I . ServiceName.proto --go_out=plugins=grpc:$GOPATH/src/. --orion_out=.