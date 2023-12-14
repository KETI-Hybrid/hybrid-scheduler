package priorities

//dominantShare(=‘우선자원량/전체자원량’)이 작은 클러스터를 선호함
import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"strconv"
	"time"
)

type Optimizationcount struct {
	prescoring map[string]int64

	betweenScore int64
}

func (pl *Optimizationcount) Name() string {
	return "Optimizationcount"
}
func (pl *Optimizationcount) PreScore(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, check bool) int64 {
	startTime := time.Now()
	var clusterScore int64
	max := 0
	for _, node := range clusterInfo.Nodes {
		cntStr := node.Node.Annotations["optimazationCount"]
		cnt, _ := strconv.Atoi(cntStr)
		if cnt > 10 {
			cnt = 10
		}
		if max > cnt {
			max = cnt
		}
		clusterScore += int64(cnt / 10)
	}
	omcplog.V(4).Info("Optimizationcount score = ", clusterScore)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("Optimizationcount Time [%v]", elapsedTime)
	return clusterScore
}
func (pl *Optimizationcount) Score(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, replicas int32, clustername string) int64 {

	startTime := time.Now()
	var clusterScore int64
	max := 0
	for _, node := range clusterInfo.Nodes {
		cntStr := node.Node.Annotations["optimazationCount"]
		cnt, _ := strconv.Atoi(cntStr)
		if cnt > 5 {
			cnt = 5
		}
		if max > cnt {
			max = cnt
		}
		clusterScore += int64(cnt / 10)
	}
	omcplog.V(4).Info("Optimizationcount score = ", clusterScore)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("Optimizationcount Time [%v]", elapsedTime)
	return clusterScore
}
