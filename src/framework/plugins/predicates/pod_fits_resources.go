package predicates

import (
	"container/list"
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"time"
)

type PodFitsResources struct{}

func (pl *PodFitsResources) Name() string {
	return "PodFitsResources"
}
func (pl *PodFitsResources) PreFilter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster) bool {
	for _, node := range clusterInfo.Nodes {
		// check if node has enough CPU
		if node.AllocatableResource.MilliCPU <= newPod.RequestedResource.MilliCPU {

			clusterInfo.PreFilter = false
			omcplog.V(4).Info("pod fits resource true")
			return true
		}

	}

	clusterInfo.PreFilter = true
	omcplog.V(4).Info("pod fits resource false  ")
	return false

}

func (pl *PodFitsResources) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {
	startTime := time.Now()
	for _, node := range clusterInfo.Nodes {
		node_result := true
		// check if node has enough Memory
		if node.AllocatableResource.Memory < newPod.RequestedResource.Memory {
			omcplog.V(5).Infof("node memory [%v]", node.AllocatableResource.Memory)
			node_result = false
		}
		// check if node has enough EphemeralStorage
		if node.AllocatableResource.EphemeralStorage < newPod.RequestedResource.EphemeralStorage {
			omcplog.V(5).Infof("EphemeralStorage [%v]", node.AllocatableResource.EphemeralStorage)
			node_result = false
		}
		if node.AllocatableResource.MilliCPU < newPod.RequestedResource.MilliCPU {
			omcplog.V(0).Infof("MilliCPU [%v]", node.AllocatableResource.MilliCPU)
			omcplog.V(0).Infof("------------CPU [%v]", newPod.RequestedResource.MilliCPU)
			node_result = false
		}
		if node_result == true {
			node.AllocatableResource.EphemeralStorage -= newPod.RequestedResource.EphemeralStorage
			node.AllocatableResource.MilliCPU -= newPod.RequestedResource.MilliCPU - (newPod.RequestedResource.MilliCPU / 6)

			node.AllocatableResource.Memory -= newPod.RequestedResource.Memory
			omcplog.V(4).Info("pod fits resource true ")
			elapsedTime := time.Since(startTime)
			omcplog.V(3).Infof("pod fits resource Time [%v]", elapsedTime)
			return true

		}
	}
	omcplog.V(4).Info("pod fits resource false  ")
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("pod fits resource Time [%v]", elapsedTime)
	return false
}

// Return true if there is at least 1 node that have AdditionalResources
func (pl *PodFitsResources) PostFilter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, postpods *list.List) (bool, error) {

	var postCPU int64
	var postMemory int64
	var postEphemeralStorage int64

	// for _, pod := range postpods {
	// 	postCPU += pod.NewPod.RequestedResource.MilliCPU
	// 	postMemory += pod.RequestedResource.Memory
	// 	postEphemeralStorage += pod.RequestedResource.EphemeralStorage
	// }
	for _, node := range clusterInfo.Nodes {

		node_result := true
		if node.CapacityResource.MilliCPU < newPod.RequestedResource.MilliCPU {
			if node.CapacityResource.MilliCPU < postCPU+newPod.RequestedResource.MilliCPU {
				node_result = false
			}

		}
		// check if node has enough Memory
		if node.CapacityResource.Memory < newPod.RequestedResource.Memory {
			if node.CapacityResource.Memory < postMemory+newPod.RequestedResource.Memory {
				node_result = false
			}
		}

		// check if node has enough EphemeralStorage
		if node.CapacityResource.EphemeralStorage < newPod.RequestedResource.EphemeralStorage {
			if node.CapacityResource.EphemeralStorage < postEphemeralStorage+newPod.RequestedResource.EphemeralStorage {
				node_result = false
			}
		}

		if node_result == true {
			return false, nil
		}
	}
	return true, nil
}
