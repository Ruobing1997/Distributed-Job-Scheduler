#!/bin/bash
#build manager
# docker build --no-cache -t morefun-supernova-manager:4.1 -f cmd/task-manager/Dockerfile .
#run
#docker run -d -p 8080:8080 morefun-supernova-manager:0.1


docker build --no-cache -t morefun-supernova-manager:4.1 -f cmd/task-manager/Dockerfile .
docker build --no-cache -t morefun-supernova-worker:4.1 -f cmd/task-executor/Dockerfile .
docker build --no-cache -t morefun-supernova-api:4.1 -f cmd/api/Dockerfile .