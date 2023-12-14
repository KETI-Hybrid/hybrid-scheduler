package priorities

//dominantShare(=‘우선자원량/전체자원량’)이 작은 클러스터를 선호함
import (
	"math"
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"time"
)

type DominantResource struct {
	prescoring map[string]int64

	betweenScore int64
}

func (pl *DominantResource) Name() string {
	return "DominantResource"
}

func (pl *DominantResource) PreScore(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, check bool) int64 {
	dominantShareArr := make([]float64, 0)
	var clusterScore int64

	for _, node := range clusterInfo.Nodes {
		dominantShare := float64(0)

		// get Dominant share
		tmp := (float64(node.RequestedResource.MilliCPU) / float64(node.CapacityResource.MilliCPU)) * 100
		math.Max(dominantShare, tmp)

		tmp = (float64(node.RequestedResource.Memory) / float64(node.CapacityResource.Memory)) * 100
		math.Max(dominantShare, tmp)

		tmp = (float64(node.RequestedResource.EphemeralStorage) / float64(node.CapacityResource.EphemeralStorage)) * 100
		math.Max(dominantShare, tmp)

		dominantShareArr = append(dominantShareArr, dominantShare)
		nodeScore := int64(math.Round((1/getMinDominantShare(dominantShareArr))*math.MaxFloat64) * float64(maxScore))

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
	return clusterScore
}

func (pl *DominantResource) Score(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, replicas int32, clustername string) int64 {
	startTime := time.Now()
	if clustername == clusterInfo.ClusterName {
		pl.prescoring[clusterInfo.ClusterName] = pl.prescoring[clusterInfo.ClusterName] - pl.betweenScore
		return pl.prescoring[clusterInfo.ClusterName]
	}
	score := pl.prescoring[clusterInfo.ClusterName]
	omcplog.V(4).Info("DominantResource score = ", score)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("DominantResource Time [%v]", elapsedTime)
	return score
}

func getMinDominantShare(arr []float64) float64 {
	min := math.MaxFloat64

	for _, a := range arr {
		if a == 0 {
			continue
		}
		min = math.Min(min, a)
	}
	return min
}
