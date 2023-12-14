package priorities

//dominantShare(=‘우선자원량/전체자원량’)이 작은 클러스터를 선호함
import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"time"
)

type Locationaffinity struct {
	prescoring map[string]int64

	betweenScore int64
}

func (pl *Locationaffinity) Name() string {
	return "Locationaffinity"
}
func contains(arr []string, str string) bool {
	for _, a := range arr {
		if str == a {
			return true
		}
	}
	return false
}

func (pl *Locationaffinity) PreScore(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, check bool) int64 {
	startTime := time.Now()
	var clusterScore int64
	clusterScore = 3
	//
	for _, node := range clusterInfo.Nodes {
		// node.PreFilter = true
		node_result := true
		for key, pod_values := range pod.Affinity {

			// compare Node's Affinity and Pod's Affinity
			if node_value, ok := node.Affinity[key]; ok {

				// if node's affinity has pod's affinity, new deployment can be deploymented
				if contains(pod_values, node_value) == false {
					clusterScore += -1
					node_result = false
				}

			} else {
				node_result = false
			}
			if node_result == true {
				clusterScore += 3
			}

		}
	}
	omcplog.V(4).Info("Locationaffinity score = ", clusterScore)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("Locationaffinity Time [%v]", elapsedTime)
	return clusterScore
}
func (pl *Locationaffinity) Score(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, replicas int32, clustername string) int64 {

	startTime := time.Now()
	var clusterScore int64
	clusterScore = 3

	for _, node := range clusterInfo.Nodes {

		node_result := true
		for key, pod_values := range pod.Affinity {

			if node_value, ok := node.Affinity[key]; ok {
				if contains(pod_values, node_value) == false {
					clusterScore += -1
					node_result = false
				}

			} else {
				node_result = false
			}
			if node_result == true {
				clusterScore += 3
			}

		}
	}
	omcplog.V(4).Info("Locationaffinity score = ", clusterScore)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("Locationaffinity Time [%v]", elapsedTime)
	return clusterScore
}
