package predicates

import (
	"openmcp/openmcp/omcplog"
	_ "openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"time"
)

type MatchClusterAffinity struct{}

func (pl *MatchClusterAffinity) Name() string {
	return "MatchClusterAffinity"
}
func (pl *MatchClusterAffinity) PreFilter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster) bool {
	startTime := time.Now()
	// Node must have all of the additional resource
	// Examples of *.yaml for a new OpenMCPDeployemt as folllow:
	// # Example 01 #
	//   spec:
	//     affinity:
	//       region:
	//         -AS
	//         -EU
	//       zone:
	//         -KR
	//         -DE
	//         -PT
	// In this case, selected node must have "KR:AS" or "DE:EU" or "PT:EU"
	//
	// # Example 02 #
	//   spec:
	//     affinity:
	//       zone:
	//         -KR
	//         -CH
	// In this case, selected node must have "KR" or "CH"

	for _, node := range clusterInfo.Nodes {
		// node.PreFilter = true
		node_result := true

		for key, pod_values := range newPod.Affinity {

			// compare Node's Affinity and Pod's Affinity
			if node_value, ok := node.Affinity[key]; ok {

				// if node's affinity has pod's affinity, new deployment can be deploymented
				if contains(pod_values, node_value) == false {
					node_result = false
				}

			} else {
				node_result = false
			}

			if node_result == false {
				break
			}
		}
		if node_result == true {
			// omcplog.V(0).Info(clusterInfo.ClusterName + "True")
			clusterInfo.PreFilterTwoStep = true
			elapsedTime := time.Since(startTime)
			omcplog.V(3).Infof("MatchClusterAffinity Time [%v]", elapsedTime)
			return true
		}
	}
	omcplog.V(4).Info("MatchClusterAffinity : false")
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("MatchClusterAffinity Time [%v]", elapsedTime)
	return false

}
func (pl *MatchClusterAffinity) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {

	if len(newPod.Affinity) == 0 {
		omcplog.V(3).Info("MatchClusterAffinity : true ")
		return true
	}
	if clusterInfo.PreFilterTwoStep == true {
		omcplog.V(3).Info("MatchClusterAffinity : true ")
		return true
	}
	omcplog.V(4).Info("MatchClusterAffinity : false ")
	return false
}
