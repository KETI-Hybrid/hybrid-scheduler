package predicates

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"strings"
	"time"
)

type Nodeunschedulable struct {
}

func (pl *Nodeunschedulable) Name() string {

	return "Nodeunschedulable"
}

func (pl *Nodeunschedulable) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {
	startTime := time.Now()
	for _, node := range clusterInfo.Nodes {
		if len(node.Node.Spec.Taints) > 0 {
			for _, avail := range node.Node.Spec.Taints {
				if strings.Compare(avail.Key, "NoSchedule") == 0 {
					omcplog.V(4).Info("Nodeunschedulable false  ")
					return false
				}
			}
		}

	}
	// omcplog.V(4).Info("Nodeunschedulable true  ")
	// elapsedTime := time.Since(startTime)
	// omcplog.V(3).Infof("Nodeunschedulable [%v]", elapsedTime)
	return true
}
