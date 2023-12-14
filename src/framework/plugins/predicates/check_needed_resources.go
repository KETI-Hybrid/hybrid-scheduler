package predicates

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"time"
)

type CheckNeededResources struct{}

func (pl *CheckNeededResources) Name() string {
	return "CheckNeededResources"
}

// Return true if there is at least 1 node that have AdditionalResources
func (pl *CheckNeededResources) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {
	startTime := time.Now()
	// Node must have all of the additional resource
	// Example of *.yaml for a new OpenMCPDeployemt as folllow:
	//     resource:
	//       request:
	//         nvidia.com/gpu: 1
	//         amd.com/gpu: 1
	// In this case, selected node must have both of "nvidia.com/gpu, amd.com/gpu"

	if len(newPod.AdditionalResource) == 0 {
		elapsedTime := time.Since(startTime)
		omcplog.V(3).Infof("CheckNeededResources Time [%v]", elapsedTime)
		return true
	}

	for _, node := range clusterInfo.Nodes {
		node_result := true
		for _, resource := range newPod.AdditionalResource {
			if contains(node.AdditionalResource, resource) == false {
				node_result = false
				break
			}
		}

		if node_result == true {
			//omcplog.V(3).Info("CheckNeededResources True ")
			elapsedTime := time.Since(startTime)
			omcplog.V(3).Infof("CheckNeededResources Time [%v]", elapsedTime)
			return true
		}
	}
	omcplog.V(3).Info("CheckNeededResources False ")
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("CheckNeededResources Time [%v]", elapsedTime)
	return false
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if str == a {
			return true
		}
	}
	return false
}
