#!/bin/bash
echo "generating proto"
#protoc -I . echo.proto --go_out=plugins=grpc:.
protoc -I . echo.proto --go_out=plugins=grpc:. --orion_out=.
#sed -i "" "s/,omitempty//" echo.pb.go
#python -m grpc.tools.protoc -I. --python_out=. --grpc_python_out=. echo.proto
