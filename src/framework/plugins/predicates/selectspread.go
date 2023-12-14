package predicates

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"time"
)

type Selectspread struct {
}

func (pl *Selectspread) Name() string {

	return "Selectspread"
}

func (pl *Selectspread) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {
	startTime := time.Now()
	omcplog.V(4).Info("Selectspread true  ")
	elapsedTime := time.Since(startTime)
	// omcplog.V(3).Infof("Selectspread [%v]", elapsedTime)
	return true
}
