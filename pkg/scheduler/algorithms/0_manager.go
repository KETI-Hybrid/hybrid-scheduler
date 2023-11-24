package scheduler

import (
	"context"

	p "hcp-pkg/hcp-resource/hcppolicy"
	"hcp-pkg/util/clusterManager"

	"hcp-pkg/apis/resource/v1alpha1"

	f "hcp-scheduler/src/framework/v1alpha1"
	"hcp-scheduler/src/resourceinfo"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var cm, _ = clusterManager.NewClusterManager()

// Scheduler watches for new unscheduled pods. It attempts to find
// nodes that they fit on and writes bindings back to the api server.
type Scheduler struct {
	SchedulingResource *v1.Pod
	HCPFramework       f.HCPFramework
	ClusterClients     map[string]*kubernetes.Clientset
	ClusterInfoList    resourceinfo.ClusterInfoList
	ClusterInfoMap     map[string]*resourceinfo.ClusterInfo
	SchedulingResult   []v1alpha1.Target
	//SchedPolicy        []string
}

func NewScheduler() *Scheduler {
	hcpFramework := f.NewFramework()
	clusterInfoList := resourceinfo.NewClusterInfoList()
	clusterInfoMap := resourceinfo.CreateClusterInfoMap(clusterInfoList)
	klog.Infoln(clusterInfoMap)
	//schedPolicy, _ := policy.GetAlgorithm()
	// if schedPolicy == nil {
	// 	// default algorithm
	// }

	schd := Scheduler{
		HCPFramework:    hcpFramework,
		ClusterInfoList: *clusterInfoList,
		ClusterInfoMap:  clusterInfoMap,
		//SchedPolicy:     schedPolicy,
	}

	return &schd
}

func (sched *Scheduler) Scheduling(deployment *v1alpha1.HCPDeployment) []v1alpha1.Target {

	klog.Infoln("[scheduling start]")
	sched.ClusterInfoList = *resourceinfo.NewClusterInfoList()
	sched.SchedulingResult = make([]v1alpha1.Target, 0)
	schedPolicy, err := p.GetHCPPolicy(*cm.HCPPolicy_Client, "scheduling-policy")
	if err != nil {
		klog.Errorln(err)
		return nil
	}
	var filter []string
	var score []string
	for _, policy := range schedPolicy.Spec.Template.Spec.Policies {
		if policy.Type == "filter" {
			filter = append(filter, policy.Value...)
		} else if policy.Type == "score" {
			score = append(score, policy.Value...)
		}
	}

	var cnt int32 = 0
	status := resourceinfo.NewCycleStatus(sched.getTotalNumNodes())
	sched.SchedulingResource = newPodFromHCPDeployment(deployment)
	replicas := *deployment.Spec.RealDeploymentSpec.Replicas

	for i := 0; i < int(replicas); i++ {
		sched.HCPFramework.RunFilterPluginsOnClusters(filter, sched.SchedulingResource, status, &sched.ClusterInfoList)
		sched.HCPFramework.RunScorePluginsOnClusters(score, sched.SchedulingResource, status, &sched.ClusterInfoList)
		best_cluster := sched.getMaxScoreCluster()
		if best_cluster != "" {
			if sched.updateSchedulingResult(best_cluster) {
				cnt += 1
				klog.Infof("[Scheduling] %d/%d pod / TargetCluster : %s\n", i+1, replicas, best_cluster)
				klog.Infoln()
			} else {
				klog.Infoln("ERROR: No cluster to be scheduled")
				klog.Infoln("Scheduling failed")
				break
			}
		} else {
			klog.Infoln("Scheduling failed")
			return nil
		}
	}

	if cnt == replicas {
		klog.Infoln("Scheduling succeeded")
		sched.printSchedulingResult()
		return sched.SchedulingResult
	} else {
		klog.Infoln("Scheduling failed")
		return nil
	}
}

func newPodFromHCPDeployment(deployment *v1alpha1.HCPDeployment) *v1.Pod {

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deployment.GetObjectMeta().GetName() + "-pod",
			Annotations: deployment.Annotations,
			Labels:      deployment.Labels,
		},
		Spec: deployment.Spec.RealDeploymentSpec.Template.Spec,
	}

	return pod
}

func (sched *Scheduler) updateSchedulingResult(cluster string) bool {

	for i, target := range sched.SchedulingResult {
		// 이미 target cluster 목록에 cluster가 있는 경우
		if target.Cluster == cluster {
			// replicas 개수 증가
			temp := *target.Replicas
			temp += 1
			target.Replicas = &temp
			sched.SchedulingResult[i] = target
			return true
		}
	}

	// target cluster 목록에 cluster가 없는 경우

	// replicas 개수 1로 설정
	var new_target v1alpha1.Target
	new_target.Cluster = cluster
	var one int32 = 1
	new_target.Replicas = &one
	sched.SchedulingResult = append(sched.SchedulingResult, new_target)
	return true
}

func (sched *Scheduler) printSchedulingResult() {
	klog.Infoln("========scheduling result========")
	targets := sched.SchedulingResult
	for _, i := range targets {
		klog.Infoln("target cluster :", i.Cluster)
		klog.Infoln("replicas       :", *i.Replicas)
		klog.Infoln()
	}
}

func (sched *Scheduler) getTotalNumNodes() int {
	clusterinfoList := sched.ClusterInfoList
	cnt := 0

	for _, clusterinfo := range clusterinfoList {
		cnt += len(clusterinfo.Nodes)
	}

	return cnt
}

func (sched *Scheduler) getMaxScoreCluster() string {
	var max_score int32 = 0
	var best_cluster string = ""
	_ = best_cluster

	for key, value := range sched.ClusterInfoMap {
		//klog.Infoln((*sched.ClusterInfoMap[key]).ClusterScore)
		if !(*sched.ClusterInfoMap[key]).IsFiltered && (*sched.ClusterInfoMap[key]).ClusterScore >= int32(max_score) {
			max_score = (*sched.ClusterInfoMap[key]).ClusterScore
			best_cluster = value.ClusterName
		}
	}

	return best_cluster
}

func (sched *Scheduler) scheduleOne(ctx context.Context) {

}

/*
func (sched *Scheduler) Filtering(algorithm string) {
	var pod = &sched.SchedulingResource
	switch algorithm {
	case "CheckNodeUnschedulable":
		for i, _ := range sched.ClusterInfoList {
			klog.Infoln(sched.ClusterInfoList[i].ClusterName)
			klog.Infoln("==before filtering==")
			klog.Infoln(sched.ClusterInfoList[i].Nodes)
			predicates.CheckNodeUnschedulable(*pod, &sched.ClusterInfoList[i])
			klog.Infoln("==after filtering==")
			klog.Infoln(&sched.ClusterInfoList[i].Nodes)
		}
	}
}

func (sched *Scheduler) Scoring(algorithm string) {

	klog.Infoln("[scoring start]")
	var pod = &sched.SchedulingResource
	var score int32

	//clusterInfoMap := resourceinfo.CreateClusterInfoMap(&sched.ClusterInfoList)

	switch algorithm {

	case "Affinity":
		for _, clusterinfo := range sched.ClusterInfoList {
			klog.Infoln("==>", clusterinfo.ClusterName)
			score = 0
			for _, node := range clusterinfo.Nodes {
				var node_score int32 = priorities.NodeAffinity(*pod, node.Node)
				if node_score == -1 {
					klog.Infoln("fail to scoring node")
					return
				} else {
					node.NodeScore = node_score
					klog.Infoln(node.NodeName, "score :", node_score)
					score += node_score
				}
			}
			sched.ClusterInfoMap[clusterinfo.ClusterName].ClusterScore = score
			klog.Infoln("*", clusterinfo.ClusterName, "total score :", score)
		}
	case "TaintToleration":
		var node_score int32
		var result []int32

		// Get intolerable taints count
		for _, clusterinfo := range sched.ClusterInfoList {
			for _, node := range clusterinfo.Nodes {
				node_score = priorities.TaintToleration(*pod, node.Node)
				if node_score == -1 {
					klog.Infoln("fail to scoring node")
					return
				} else {
					node.NodeScore = node_score
					result = append(result, node_score)
				}
			}
		}

		// sort intolerable taints count and get max value
		sort.Slice(result, func(i, j int) bool {
			return result[i] > result[j]
		})
		max := result[0]

		// scoring - normalize for intolerable taints count
		klog.Infoln("[REAL SCORE]")
		for _, clusterinfo := range sched.ClusterInfoList {
			klog.Infoln("==>", clusterinfo.ClusterName)
			score = 0
			for _, node := range clusterinfo.Nodes {
				if node.NodeScore == 0 {
					node.NodeScore = int32(scoretable.MaxNodeScore)
				} else {
					node.NodeScore = int32(100 * ((float32(max) - float32(node.NodeScore)) / float32(max)))
				}
				score += node.NodeScore
				klog.Infoln("===>", node.NodeName, node.NodeScore)
			}

			if int32(len(clusterinfo.Nodes)) > 0 {
				// clusterInfoMap[clusterinfo.ClusterName].ClusterScore = score / int32(len((*clusterinfo).Nodes))
				sched.ClusterInfoMap[clusterinfo.ClusterName].ClusterScore = score
				klog.Infoln("*", clusterinfo.ClusterName, "total score :", sched.ClusterInfoMap[clusterinfo.ClusterName].ClusterScore)
			} else {
				sched.ClusterInfoMap[clusterinfo.ClusterName].ClusterScore = 0
				klog.Infoln("*", clusterinfo.ClusterName, "total score :", sched.ClusterInfoMap[clusterinfo.ClusterName].ClusterScore)
			}
		}
	case "NodeResourcesBalancedAllocation":
		// 현재 pod이 배치 된 후, CPU와 Memory 사용률이 균형을 검사
		for _, clusterinfo := range sched.ClusterInfoList {
			klog.Infoln("==>", clusterinfo.ClusterName)
			score = 0
			for _, node := range clusterinfo.Nodes {
				var node_score int32 = int32(priorities.NodeResourcesBalancedAllocation(*pod, node.Node))
				if node_score == -1 {
					klog.Infoln("fail to scoring node")
					return
				} else {
					node.NodeScore = node_score
					klog.Infoln(node.NodeName, "score :", node_score)
					score += node_score
				}
			}

			//klog.Infoln(sched.ClusterInfoMap)
			(*sched.ClusterInfoMap[clusterinfo.ClusterName]).ClusterScore = score
			klog.Infoln("*", clusterinfo.ClusterName, "total score :", score)
			klog.Infoln()
		}
	case "ImageLocality":
		for _, clusterinfo := range sched.ClusterInfoList {
			klog.Infoln("==>", clusterinfo.ClusterName)
			score = 0
			for _, node := range clusterinfo.Nodes {
				klog.Infoln(node.ImageStates)
				var node_score int32 = int32(priorities.ImageLocality(*pod, node, sched.getTotalNumNodes()))
				if node_score == -1 {
					klog.Infoln("fail to scoring node")
					return
				} else {
					node.NodeScore = node_score
					klog.Infoln(node.NodeName, "score :", node_score)
					score += node_score
				}
			}
			sched.ClusterInfoMap[clusterinfo.ClusterName].ClusterScore = score
			klog.Infoln("*", clusterinfo.ClusterName, "total score :", score)
		}
	}

}
*/

/*
func (sched *Scheduler) NewScoreTable() *scoretable.ScoreTable {
	var score_table scoretable.ScoreTable

	for _, cluster_info := range *sched.ClusterInfoList {
		var node_score_list scoretable.NodeScoreList

		// create node_score_list
		for _, node_info := range cluster_info.Nodes {
			node_score := &scoretable.NodeScore{
				Name:  node_info.NodeName,
				Score: 0,
			}
			node_score_list = append(node_score_list, *node_score)
		}
		cluster_score := &scoretable.ClusterScore{
			Cluster:       cluster_info.ClusterName,
			NodeScoreList: node_score_list,
			Score:         0,
		}
		score_table = append(score_table, *cluster_score)
	}

	return &score_table
}
*/
