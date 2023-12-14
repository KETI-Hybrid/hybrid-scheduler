/*
Copyright 2018 The Multicluster-Controller Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"openmcp/openmcp/omcplog"
	"openmcp/openmcp/openmcp-analytic-engine/src/protobuf"
	openmcpscheduler "openmcp/openmcp/openmcp-scheduler/src"
	"openmcp/openmcp/openmcp-scheduler/src/controller"
	"openmcp/openmcp/util/clusterManager"
	"openmcp/openmcp/util/controller/logLevel"
	"os"

	"google.golang.org/grpc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	// Setup KetiLog
	logLevel.KetiLogInit()

	// Setup gRPC communication with openmcp-analyticEngine
	SERVER_IP := os.Getenv("GRPC_SERVER")
	SERVER_PORT := os.Getenv("GRPC_PORT")
	host := SERVER_IP + ":" + SERVER_PORT

	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		omcplog.V(0).Info("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := protobuf.NewRequestAnalysisClient(conn)
	// Start Openmcp-Scheduler
	for {
		omcplog.V(0).Info("***** [START] OpenMCP Scheduler *****")

		// Get Federated Cluster Information
		cm := clusterManager.NewClusterManager()

		// Init openmcp-scheduler
		scheduler := openmcpscheduler.NewScheduler(cm, grpcClient)

		// Init Controllers for openmcp-scheduler
		controller.NewControllers(cm, scheduler)
	}
}
