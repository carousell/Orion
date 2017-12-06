#!/bin/bash
protoc -I . simple.proto --go_out=plugins=grpc:. --orion_out=.
