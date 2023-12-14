package priorities

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"time"
)

type MostRequested struct {
	prescoring   map[string]int64
	betweenScore int64
}

func (pl *MostRequested) Name() string {
	return "MostRequested"
}
func (pl *MostRequested) PreScore(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, check bool) int64 {
	var clusterScore int64

	for _, node := range clusterInfo.Nodes {
		nodeScore := mostRequestedScore(pod.RequestedResource.MilliCPU, node.AllocatableResource.MilliCPU)
		nodeScore += mostRequestedScore(pod.RequestedResource.Memory, node.AllocatableResource.Memory)
		nodeScore += mostRequestedScore(pod.RequestedResource.EphemeralStorage, node.AllocatableResource.EphemeralStorage)

		node.NodeScore = nodeScore * weight
		clusterScore += nodeScore
	}
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
		pl.prescoring[clusterInfo.ClusterName] = clusterScore - pl.betweenScore

	}
	//omcplog.V(0).Infof("["+clusterInfo.ClusterName+"]노드"+pl.Name()+" 스코어 =", clusterScore)
	return clusterScore
}

func (pl *MostRequested) Score(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, replicas int32, clustername string) int64 {
	startTime := time.Now()
	if clustername == clusterInfo.ClusterName {
		pl.prescoring[clusterInfo.ClusterName] = pl.prescoring[clusterInfo.ClusterName] - pl.betweenScore
		return pl.prescoring[clusterInfo.ClusterName]
	}
	score := pl.prescoring[clusterInfo.ClusterName]
	omcplog.V(4).Info("MostRequested score = ", score)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("MostRequested Time [%v]", elapsedTime)
	return score
}

func mostRequestedScore(requested, allocable int64) int64 {
	if allocable == 0 {
		return 0
	}
	if requested > allocable {
		return 0
	}
	return (requested * int64(100)) / allocable
}
