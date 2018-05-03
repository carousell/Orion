#!/bin/bash
echo "generating proto"
protoc -I . ServiceName.proto --go_out=plugins=grpc:. --orion_out=.
