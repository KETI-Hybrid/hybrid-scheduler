#!/bin/bash
docker_id="ketidevit2"
controller_name="hybrid.hybrid-scheduler"
export GO111MODULE=on
go mod vendor
go build -o /root/workspace/hth/dev/hybrid-scheduler/docker/_output/bin/$controller_name -mod=vendor /root/workspace/hth/dev/hybrid-scheduler/cmd/main.go
docker build -t $docker_id/$controller_name:latest /root/workspace/hth/dev/hybrid-scheduler/build && \
docker push $docker_id/$controller_name:latest

#0.0.2 => 운용중인 metric collector member
#0.0.3 => test 용 버전 (제병)