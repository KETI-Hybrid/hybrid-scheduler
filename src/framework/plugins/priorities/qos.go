//QoSPriority : 기존 kubelet은 QoS 을 제공함 (Best-Effort/Burstable/Guaranteed)
//제공되는 QoS 를 기반으로 “Best-Effort < Burstable < Guaranteed” 우선순위를 부여하여
//총 우선순위 값이 큰 클러스터를 선호함

package priorities

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"time"

	v1 "k8s.io/api/core/v1"
)

type QosPriority struct {
	prescoring   map[string]int64
	betweenScore int64
}

const (
	minScore     int64 = 1
	midScore     int64 = (minScore + maxScore) / 2
	maxScore     int64 = 2
	maxRTCRScore int64 = 10
	weight       int64 = 3
)

func (pl *QosPriority) Name() string {
	return "scorespread"
}
func (pl *QosPriority) PreScore(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, check bool) int64 {
	var clusterScore int64

	for _, node := range clusterInfo.Nodes {
		var nodeScore int64
		for _, pod := range node.Pods {

			// get PodQOSClass from v1.Pod
			qos := pod.Pod.Status.QOSClass

			switch qos {
			case v1.PodQOSGuaranteed:
				nodeScore += minScore
			case v1.PodQOSBurstable:
				nodeScore += midScore
			case v1.PodQOSBestEffort:
				nodeScore += maxScore
			}
		}
		node.NodeScore = nodeScore
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

func (pl *QosPriority) Score(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, replicas int32, clustername string) int64 {
	startTime := time.Now()
	if clustername == clusterInfo.ClusterName {
		pl.prescoring[clusterInfo.ClusterName] = pl.prescoring[clusterInfo.ClusterName] - pl.betweenScore
		return pl.prescoring[clusterInfo.ClusterName]
	}
	score := pl.prescoring[clusterInfo.ClusterName]
	omcplog.V(4).Info("scorespread score = ", score)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("scorespread Time [%v]", elapsedTime)
	return score
}
