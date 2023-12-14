// 네트워크 사용량이 적은 클러스터 선호함
package priorities

import (
	//"openmcp/openmcp/omcplog"

	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"time"
)

type BalancedNetworkAllocation struct {
	prescoring   map[string]int64
	betweenScore int64
}

func (pl *BalancedNetworkAllocation) Name() string {

	return "BalancedNetworkAllocation"
}

func (pl *BalancedNetworkAllocation) PreScore(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, check bool) int64 {
	// startTime := time.Now()
	var clusterScore int64
	clusterScore = 0

	for _, node := range clusterInfo.Nodes {
		clusterScore += node.NodeScore
	}
	// OelapsedTime := time.Since(startTime)
	if !check {
		if len(pl.prescoring) == 0 {
			pl.prescoring = make(map[string]int64)
		}
		pl.prescoring[clusterInfo.ClusterName] = clusterScore
	} else {
		pl.betweenScore = pl.prescoring[clusterInfo.ClusterName] - int64(clusterScore)
		if pl.betweenScore <= 0 {
			pl.betweenScore = 5
		}
		pl.prescoring[clusterInfo.ClusterName] = (clusterScore - pl.betweenScore) * weight

	}
	return clusterScore

}
func (pl *BalancedNetworkAllocation) Score(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, replicas int32, clustername string) int64 {
	startTime := time.Now()
	if clustername == clusterInfo.ClusterName {
		pl.prescoring[clusterInfo.ClusterName] = pl.prescoring[clusterInfo.ClusterName] - pl.betweenScore
		return pl.prescoring[clusterInfo.ClusterName]
	}
	score := pl.prescoring[clusterInfo.ClusterName]
	omcplog.V(4).Info("BalancedNetworkAllocation score = ", score)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("BalancedNetworkAllocation Time [%v]", elapsedTime)
	return score
}
