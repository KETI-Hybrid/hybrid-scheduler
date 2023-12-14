#!/bin/bash
docker_id="ketidevit2"
image_name="hcp-scheduler"

export GO111MODULE=on
go mod vendor


go build -o build/_output/bin/$image_name -mod=vendor openmcp/openmcp/openmcp-scheduler/src/main && \

docker build -t $docker_id/$image_name:v0.0.1 build && \
docker push $docker_id/$image_name:v0.0.1

