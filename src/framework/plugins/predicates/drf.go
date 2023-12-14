package predicates

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"time"
)

type DRF struct {
}

func (pl *DRF) Name() string {

	return "DRF"
}

func (pl *DRF) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {
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
			omcplog.V(4).Info("DRF true ")
			elapsedTime := time.Since(startTime)
			omcplog.V(3).Infof("DRF Time [%v]", elapsedTime)
			return true

		}
	}
	omcplog.V(4).Info("DRF false  ")
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("DRF [%v]", elapsedTime)
	return false
}
