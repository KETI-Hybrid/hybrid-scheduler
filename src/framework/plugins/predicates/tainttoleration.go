package predicates

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"time"

	v1 "k8s.io/api/core/v1"
)

type Tainttoleration struct {
}

func (pl *Tainttoleration) Name() string {

	return "Tainttoleration"
}

func (pl *Tainttoleration) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {
	startTime := time.Now()
	for _, node := range clusterInfo.Nodes {
		if len(node.Node.Spec.Taints) > 0 {
			for _, taint := range node.Node.Spec.Taints {
				if taint.Effect == v1.TaintEffectNoSchedule {
					omcplog.V(4).Info("Tainttoleration false   ")
					return false
				} else if taint.Effect == v1.TaintEffectPreferNoSchedule {
					omcplog.V(4).Info("Tainttoleration false  ")
					return false
				}
			}
		}
	}
	omcplog.V(4).Info("Tainttoleration true  ")
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("Tainttoleration [%v]", elapsedTime)
	return true
}
