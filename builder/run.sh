#!/bin/bash
docker build . -t service_name|| exit
docker stop service_name
docker rm service_name
docker run --privileged --name service_name -p 9281:9281 -p 9282:9282 -p 9283:9283 -p 9284:9284 -h `whoami`-`hostname` service_name server
