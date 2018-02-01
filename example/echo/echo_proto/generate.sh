#!/bin/bash
echo "generating proto"
protoc -I . echo.proto --go_out=plugins=grpc:. --orion_out=.
