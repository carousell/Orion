#!/bin/bash
protoc -I . stringproto.proto --go_out=plugins=grpc:. --orion_out=.
