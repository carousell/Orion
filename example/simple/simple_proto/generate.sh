#!/bin/bash
protoc -I . simple.proto --go_out="paths=source_relative:." --go-grpc_out="paths=source_relative:." --orion_out=.
