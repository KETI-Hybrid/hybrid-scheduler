package priorities

//dominantShare(=‘우선자원량/전체자원량’)이 작은 클러스터를 선호함
import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"time"
)

type NodePreferAvoidPods struct {
	prescoring map[string]int64

	betweenScore int64
}

func (pl *NodePreferAvoidPods) Name() string {
	return "NodePreferAvoidPods"
}
func (pl *NodePreferAvoidPods) PreScore(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, check bool) int64 {
	startTime := time.Now()
	var clusterScore int64
	for _, node := range clusterInfo.Nodes {
		annos := node.Node.Annotations
		clusterScore += clusterScore - int64(len(annos))
	}
	omcplog.V(4).Info("NodePreferAvoidPods score = ", clusterScore)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("NodePreferAvoidPods Time [%v]", elapsedTime)
	return clusterScore
}
func (pl *NodePreferAvoidPods) Score(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, replicas int32, clustername string) int64 {

	startTime := time.Now()
	var clusterScore int64
	for _, node := range clusterInfo.Nodes {
		annos := node.Node.Annotations
		clusterScore += clusterScore - int64(len(annos))
	}
	omcplog.V(4).Info("NodePreferAvoidPods score = ", clusterScore)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("NodePreferAvoidPods Time [%v]", elapsedTime)
	return clusterScore
}
