#!/bin/bash
echo "generating proto"
protoc -I . ServiceName.proto --go_out=plugins=grpc:. --orion_out=.
sed -i "" "s/,omitempty//" ServiceName.pb.go
#python -m grpc.tools.protoc -I. --python_out=. --grpc_python_out=. groups.proto
