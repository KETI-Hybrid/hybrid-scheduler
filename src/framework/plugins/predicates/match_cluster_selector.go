package predicates

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"strings"
	"time"
)

type MatchClusterSelector struct {
}

func (pl *MatchClusterSelector) Name() string {

	return "MatchClusterSelector"
}

func (pl *MatchClusterSelector) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {
	startTime := time.Now()
	// Node must have labels correspoinding to new Pod's NodeSelector
	// Example of *.yaml for a new OpenMCPDeployemt as folllow:
	//  spec:
	//     nodeSelector:
	//       disktype: ssd
	//		 key: worker
	// In this case, selected node must have both of "disktype: ssd" and "key:worker" in Labels

	if len(newPod.Pod.Spec.NodeSelector) == 0 {
		elapsedTime := time.Since(startTime)
		omcplog.V(3).Infof("MatchClusterSelector Time [%v]", elapsedTime)
		return true
	}

	for _, node := range clusterInfo.Nodes {
		// if node.PreFilter == false || node.PreFilterTwoStep == false {
		// 	omcplog.V(0).Infof("preFilter True", pl.Name(), node.PreFilter)
		// 	continue
		// }
		node_result := true
		// NodeSelector's type is map[string]string
		// if you want to more information, check "k8s.io/api/core/v1"
		for key, pod_value := range newPod.Pod.Spec.NodeSelector {

			//klog.Infof("pod_value:%v", pod_value)

			if node_value, ok := node.Node.Labels[key]; !ok {
				//klog.Infof("n./4ode_value:%v", node_value)
				node_result = false
			} else {
				// Check if value is the same
				if strings.Compare(pod_value, node_value) != 0 {
					node_result = false
				}
			}

			// if node doesnt have key or the value is not correct
			// stop checking this node to reduce scheduling time
			if node_result == false {
				break
			}
		}

		if node_result == true {
			//omcplog.V(3).Info("MatchClusterSelector true ")
			elapsedTime := time.Since(startTime)
			omcplog.V(3).Infof("MatchClusterSelector Time [%v]", elapsedTime)
			return true
		}
	}
	omcplog.V(4).Info("MatchClusterSelector false ")
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("MatchClusterSelector Time [%v]", elapsedTime)
	return false
}
