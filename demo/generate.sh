#!/bin/bash
protoc -I ./ --go_out=plugins=grpc:. --orion_out=. foo/foo.proto
