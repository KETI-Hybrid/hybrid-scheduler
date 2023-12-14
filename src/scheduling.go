/*
log 1레벨 :	결과
2레벨     : 필터 & 스코어 결과 추가
3레벨     : 필터 & 스코어 연산 과정 추가
4레벨     : 연산과정에 대한 모든 로깅
5레벨     : 디버깅관련 모든 로깅
*/

package openmcpscheduler

import (
	"container/list"
	"context"
	"fmt"
	"openmcp/openmcp/apis/cluster/v1alpha1"
	resourcev1alpha1 "openmcp/openmcp/apis/resource/v1alpha1"
	"sync"

	//clusterv1alpha1 "openmcp/openmcp/apis/cluster/v1alpha1"
	"openmcp/openmcp/omcplog"
	"openmcp/openmcp/openmcp-analytic-engine/src/protobuf"
	ketiframework "openmcp/openmcp/openmcp-scheduler/src/framework/v1alpha1"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OpenMCPScheduler struct {
	ClusterClients      map[string]*kubernetes.Clientset
	ClusterInfos        map[string]*ketiresource.Cluster
	Framework           ketiframework.OpenmcpFramework
	ClusterManager      *clusterManager.ClusterManager
	GRPC_Client         protobuf.RequestAnalysisClient
	ClusterList         *v1alpha1.OpenMCPClusterList
	IsNetwork           bool
	IsResource          bool
	PostDeployments     *list.List
	PostThread          bool
	Selectpolicy        string
	Live                *client.Client
	SchdPolicy          string
	Mutex               *sync.Mutex
	origin_clusternames *[]string
	requestclusters     []string
}

func NewScheduler(cm *clusterManager.ClusterManager, grpcClient protobuf.RequestAnalysisClient) *OpenMCPScheduler {
	sched := &OpenMCPScheduler{}
	sched.ClusterClients = make(map[string]*kubernetes.Clientset)
	sched.ClusterInfos = make(map[string]*ketiresource.Cluster)
	sched.Framework = ketiframework.NewFramework(grpcClient)
	sched.ClusterManager = cm
	sched.GRPC_Client = grpcClient
	sched.PostDeployments = list.New()
	sched.Mutex = new(sync.Mutex)
	if !sched.IsNetwork {
		omcplog.V(0).Infof("sched.ClusterInfos loading ...")
		sched.Mutex.Lock()
		sched.SetupResources()
		sched.Mutex.Unlock()
		go sched.LocalNetworkAnalysis()
		go sched.Postschduling()
		go sched.SchedulingPolicyMonitoring()
		sched.SchdPolicy = "None"
		sched.IsNetwork = true
		omcplog.V(0).Infof("LocalNetworkAnalysis Start")
	}
	return sched
}

/*
*
/*@brief 스케쥴링 정책에 대한 모니터링
/*@details 스케줄링 정책 라운드로빈 RR , OpenMCP Filter & Scoring 기법 openmcp
*
*/
func (sched *OpenMCPScheduler) SchedulingPolicyMonitoring() {
	for {

		openmcpPolicyInstance, perr := sched.ClusterManager.Crd_client.OpenMCPPolicy("openmcp").Get("scheduling-policy", metav1.GetOptions{})
		if perr == nil {
			policies := openmcpPolicyInstance.Spec.Template.Spec.Policies
			for _, policy := range policies {
				if policy.Type == "algorithm" {
					sched.SchdPolicy = policy.Value[0]
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func (sched *OpenMCPScheduler) PostsMonitoring() {

	sched.Mutex.Lock()
	sched.SetupResources()
	sched.Mutex.Unlock()
	postlist := sched.PostDeployments
	osvc_list := &resourcev1alpha1.OpenMCPDeploymentList{}
	if sched.Live == nil {
		omcplog.V(5).Info("sched.Live NIL")
		return
	}
	openmcpPolicyInstance, perr := sched.ClusterManager.Crd_client.OpenMCPPolicy("openmcp").Get("post-schduling", metav1.GetOptions{})
	if perr == nil {
		policies := openmcpPolicyInstance.Spec.Template.Spec.Policies
		for _, policy := range policies {
			if policy.Type == "priority" {
				sched.Selectpolicy = policy.Value[0]
			}
		}
	}
	(*(sched.Live)).List(context.TODO(), osvc_list)
	for e := postlist.Front(); e != nil; e = postlist.Front().Next() {
		deploy := e.Value.(*ketiresource.PostDelployment)
		checked := false

		//omcplog.V(0).Info("postlist e =", deploy.NewDeployment.GetName())
		for _, has := range osvc_list.Items {
			if has.GetName() == deploy.NewDeployment.GetName() {
				//omcplog.V(0).Info("has name =", has.GetName())
				checked = true
			}
		}
		if checked == false {
			//removename := deploy.NewDeployment.GetName()
			postlist.Remove(e)
			//omcplog.V(0).Info("remove postdeploy", removename)
			return
		}
	}

}

func (sched *OpenMCPScheduler) Postschduling() {
	omcplog.V(5).Info("postschduling start")
	for {

		sched.PostsMonitoring()
		time.Sleep(3 * time.Second)
		//newDeployment := &resourcev1alpha1.OpenMCPDeployment{}
		postlist := sched.PostDeployments
		//omcplog.V(0).Info(" postlist len", postlist)
		if postlist.Len() > 0 {
			var firstdeploy *ketiresource.PostDelployment
			if sched.Selectpolicy == "FIFO" {
				firstdeploy = (postlist.Front()).Value.(*ketiresource.PostDelployment)

			} else if sched.Selectpolicy == "OPENMCP" {
				minresource := (postlist.Front()).Value.(*ketiresource.PostDelployment)
				for e := postlist.Front(); e != nil; e = postlist.Front().Next() {
					deploy := e.Value.(*ketiresource.PostDelployment)
					newPod := newPodFromOpenMCPDeployment(deploy.NewDeployment)
					oldPod := newPodFromOpenMCPDeployment(minresource.NewDeployment)
					if newPod.RequestedResource.MilliCPU < oldPod.RequestedResource.Memory {
						minresource = deploy
					}
				}
				firstdeploy = minresource
			} else {
				//잘못된 policy 정책이 들어왔을 경우 FIFO로 수행
				firstdeploy = (postlist.Front()).Value.(*ketiresource.PostDelployment)
			}

			postdeployment := firstdeploy.NewDeployment
			omcplog.V(4).Info("Post Scheduling RemainReplica", firstdeploy.RemainReplica)
			firstdeploy.NewDeployment.Status.Replicas = firstdeploy.RemainReplica
			exist := postdeployment.Status.ClusterMaps
			backup := map[string]int32{}
			backup = exist
			cluster_replicas_map, _ := sched.Scheduling(postdeployment, true, postdeployment.Spec.Clusters)

			replicacount := 0
			chagnedp := map[string]int32{}
			for key, val := range exist {
				_, exists := exist[key]
				if !exists {
					chagnedp[key] = 1
				} else {
					chagnedp[key] += val
				}
			}
			for key, val := range cluster_replicas_map {
				_, exists := cluster_replicas_map[key]
				if !exists {
					chagnedp[key] = 1
					replicacount++
				} else {
					chagnedp[key] += val
					replicacount += int(val)
				}
			}
			if len(cluster_replicas_map) == 0 {
				continue
			}
			opts := &client.PatchOptions{DryRun: []string{"All"}}
			(*sched.Live).Status().Patch(context.TODO(), postdeployment, client.MergeFrom(postdeployment), opts)
			postdeployment.Status.ClusterMaps = chagnedp
			postdeployment.Status.SchedulingNeed = false
			postdeployment.Status.SchedulingComplete = true
			postdeployment.Status.CreateSyncRequestComplete = false

			err := (*sched.Live).Status().Update(context.TODO(), postdeployment)
			if err != nil {
				omcplog.V(0).Infof("Failed to update instance status, %v", err)
				postdeployment.Status.ClusterMaps = backup
			} else {
				firstdeploy.RemainReplica = firstdeploy.RemainReplica - int32(replicacount)
				if firstdeploy.RemainReplica <= 0 {
					postlist.Remove(postlist.Front())
				}
				omcplog.V(4).Infof("Cluster MAP : %v", chagnedp)
				omcplog.V(4).Infof("Remain count : %v exist count : %v", firstdeploy.RemainReplica, replicacount)
			}

		} else {
			//omcplog.V(0).Infof("len 0")
			continue
		}
	}

}

/*
*
/*@brief RR 스케쥴링 수행하는 함수
*
*/
func (sched *OpenMCPScheduler) RRScheduling(clusters map[string]*ketiresource.Cluster, replicas int32, requestclusters []string) map[string]int32 {
	cluster_replicas_map := make(map[string]int32)
	remain_rep := replicas
	sched.requestclusters = requestclusters
	filteredCluster := make(map[string]*ketiresource.Cluster)
	cluster_count := 0
	queue_cluster := make(map[int]string)
	for clusterName, cluster := range clusters {
		omcplog.V(5).Infof("RRScheduling:", clusterName)
		if remain_rep == 0 {
			break
		}
		if requestclusters == nil {
			filteredCluster[clusterName] = cluster
			queue_cluster[cluster_count] = clusterName
			cluster_count++
		}
		if requestclusters != nil {
			for _, s := range requestclusters {

				if s == clusterName {
					filteredCluster[clusterName] = cluster
					queue_cluster[cluster_count] = clusterName
					cluster_count++
				}
			}
		}
	}
	if len(filteredCluster) == 0 {
		omcplog.V(5).Infof("error :RRScheduling cluster_count =0")
	}
	for _, s := range queue_cluster {
		cluster_replicas_map[s] = 0
	}
	for i := 0; i < int(replicas); i++ {
		if cluster_count == 0 {
			omcplog.V(1).Info(cluster_count)
			break
		}
		index := i % cluster_count
		//omcplog.V(2).Info("index=", int(index))
		//omcplog.V(2).Info("queue_cluster[index]=", queue_cluster[index])
		cluster_replicas_map[queue_cluster[index]] = cluster_replicas_map[queue_cluster[index]] + 1
	}

	return cluster_replicas_map
}

/*
*
@brief 임시로 작성
*
*/
func PrintFilterString(datas map[string]*ketiresource.Cluster) []string {
	returndata := make([]string, 0)
	for clustername, _ := range datas {
		returndata = append(returndata, clustername)
	}
	return returndata

}

/*
*
/*@params requestclusters : define cluster yaml파일에 기록된 클러스터들
/*@params requestclusters : posted: 더이상 배포될수 있는 환경이 아닐때 true
/*@params posted :
*
*/
func (sched *OpenMCPScheduler) Complite_Scheduing(selectcluster string) {
	cm := sched.ClusterManager
	sched.ClusterList, _ = cm.Crd_cluster_client.OpenMCPCluster("openmcp").List(v1.ListOptions{})
	// Setup Clusters
	omcplog.V(2).Infof("selectcluster = %v", selectcluster)
	updatecluster, exist := sched.ClusterClients[selectcluster]
	if exist {

		omcplog.V(2).Infof("exist update ...")
		pods, _ := updatecluster.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		// informations on cluster level
		allPods := make([]*ketiresource.Pod, 0)
		allNodes := make([]*ketiresource.NodeInfo, 0)
		cluster_request := ketiresource.NewResource()
		cluster_allocatable := ketiresource.NewResource()
		for _, pod := range pods.Items {
			// add Stroage
			pod_request := &ketiresource.Resource{0, 0, 0}
			pod_additionalResource := make([]string, 0)

			for _, container := range pod.Spec.Containers {
				for rName, rQuant := range container.Resources.Requests {
					switch rName {
					case corev1.ResourceCPU:
						pod_request.MilliCPU = rQuant.MilliValue()
					case corev1.ResourceMemory:
						pod_request.Memory = rQuant.Value()
					case corev1.ResourceEphemeralStorage:
						pod_request.EphemeralStorage = rQuant.Value()
					default:
						// Casting from ResourceName to stirng because rName is ResourceName type
						resourceName := fmt.Sprintf("%s", rName)
						pod_additionalResource = append(pod_additionalResource, resourceName)
					}
				}
			}
			newPod := &ketiresource.Pod{
				Pod:                &pod,
				ClusterName:        selectcluster,
				NodeName:           pod.Spec.NodeName,
				PodName:            pod.Name,
				RequestedResource:  pod_request,
				AdditionalResource: pod_additionalResource,
			}
			allPods = append(allPods, newPod)
		}

		// Setup Nodes
		nodes, _ := sched.ClusterClients[selectcluster].CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		for _, node := range nodes.Items {

			// Get v1.Pod, corev1.ContainerPort and RequestResource
			podsInNode := make([]*ketiresource.Pod, 0)
			node_request := ketiresource.NewResource()

			for _, pod := range allPods {
				if strings.Compare(pod.NodeName, node.Name) == 0 {
					podsInNode = append(podsInNode, pod)
					node_request = ketiresource.AddResources(node_request, pod.RequestedResource)
				}
			}

			// Get capacity, Additional Resource from node Spec
			node_additionalResource := make([]string, 0)
			node_capacity := &ketiresource.Resource{}

			for rName, rQuant := range node.Status.Capacity {
				switch rName {
				case corev1.ResourceCPU:
					node_capacity.MilliCPU = rQuant.MilliValue()
				case corev1.ResourceMemory:
					node_capacity.Memory = rQuant.Value()
				case corev1.ResourceEphemeralStorage:
					node_capacity.EphemeralStorage = rQuant.Value()
				default:
					// Casting from ResourceName to stirng because rName is ResourceName type
					resourceName := fmt.Sprintf("%s", rName)
					node_additionalResource = append(node_additionalResource, resourceName)
				}
			}

			// Get allocatable Resource based on capacity and request
			node_allocatable := ketiresource.GetAllocatable(node_capacity, node_request)

			// Get Affinity
			node_affinity := make(map[string]string)

			for key, value := range node.Labels {
				switch key {
				case "failure-domain.beta.kubernetes.io/region":
					if _, ok := node_affinity["region"]; !ok {
						node_affinity["region"] = value
					}

				case "failure-domain.beta.kubernetes.io/zone":
					if _, ok := node_affinity["zone"]; !ok {
						node_affinity["zone"] = value
					}
				}
			}

			// make new Node
			newNode := &ketiresource.NodeInfo{
				ClusterName: selectcluster,
				NodeName:    node.Name,
				Node:        &node,
				Pods:        podsInNode,
				// UsedPorts:				node_usedPorts,
				CapacityResource:    node_capacity,
				RequestedResource:   node_request,
				AllocatableResource: node_allocatable,
				AdditionalResource:  node_additionalResource,
				Affinity:            node_affinity,
				NodeScore:           0,
			}
			allNodes = append(allNodes, newNode)
			cluster_request = ketiresource.AddResources(cluster_request, node_request)
			cluster_allocatable = ketiresource.AddResources(cluster_allocatable, node_allocatable)
		}

		// Setup Cluster

		sched.ClusterInfos[selectcluster] = &ketiresource.Cluster{
			ClusterName:         selectcluster,
			Nodes:               allNodes,
			RequestedResource:   cluster_request,
			AllocatableResource: cluster_allocatable,
			ClusterList:         sched.ClusterList,
		}
	}
}
func (sched *OpenMCPScheduler) Scheduling(dep *resourcev1alpha1.OpenMCPDeployment, posted bool, requestclusters []string) (map[string]int32, error) {
	startTime := time.Now()
	// Get CLusterClients from clusterManager
	//clusterv1alpha1.OpenMCPCluster
	cm := sched.ClusterManager
	depReplicas := dep.Spec.Replicas

	// Get CLusterClients from clusterManager
	sched.ClusterClients = cm.Cluster_kubeClients

	// Return scheduling result (ex. cluster1:2, cluster2:1)
	totalSchedulingResult := map[string]int32{}

	if len(sched.ClusterInfos) == 0 {
		omcplog.V(0).Infof("sched.ClusterInfos loading ...")
		sched.Mutex.Lock()
		sched.SetupResources()
		sched.Mutex.Unlock()
		sched.IsResource = true
	}
	// omcplog.V(0).Infof("sched.ClusterInfos kcp test loading ...")
	// sched.SetupResources()
	// RR 정책일경우 처리
	if sched.SchdPolicy == "RR" {
		omcplog.V(0).Infof("Round Robin Scheduling ...")
		totalSchedulingResult = sched.RRScheduling(sched.ClusterInfos, depReplicas, requestclusters)
		if len(totalSchedulingResult) == 0 {
			return totalSchedulingResult, fmt.Errorf("There is postpods error")

		}
		sched.Framework.EndPod()
		return totalSchedulingResult, nil

	} else {
		// Get Data from Node&Pod Spec
		//sched.SetupResources()
		// Make resource to schedule pod into cluster
		newPod := newPodFromOpenMCPDeployment(dep)
		// Scheduling one pod
		for i := int32(0); i < depReplicas; i++ {
			//sched.SetupCapacityResource()
			// If there is no proper cluster to deploy Pod,
			// stop scheduling and return scheduling result
			omcplog.V(0).Info("################################################")
			omcplog.V(0).Infof("A Deploy / Replica : %v / %v", i+1, depReplicas)
			schedulingResult, err := sched.ScheduleOne(newPod, depReplicas-i, dep, posted, requestclusters)
			if err != nil {
				return totalSchedulingResult, fmt.Errorf("There is no proper cluster to deploy Pod(%d)~Pod(%d)", i, depReplicas)
			}
			if schedulingResult == "" {
				continue

				//If schdulingResult is post,
			} else if schedulingResult == "post" {
				backdeploy := (sched.PostDeployments.Back()).Value.(*ketiresource.PostDelployment)
				backdeploy.Replica = depReplicas
				backdeploy.NewDeployment.Status.ClusterMaps = totalSchedulingResult
				dep.Spec.Replicas = dep.Spec.Replicas - backdeploy.RemainReplica
				return totalSchedulingResult, nil
			}

			_, exists := totalSchedulingResult[schedulingResult]
			if !exists {
				totalSchedulingResult[schedulingResult] = 1
			} else {
				totalSchedulingResult[schedulingResult] += 1
			}

			sched.UpdateResources(newPod, schedulingResult)
		}
		sched.Framework.EndPod()
		elapsedTime := time.Since(startTime)
		omcplog.V(0).Infof("Scheduling Time %v", elapsedTime)

		return totalSchedulingResult, nil
	}
}

/*
*
/*@brief EraseScheduling 레플리카 개수가 감소했을 경우에 수행하는 함수
/*@todo Filter 이전에 request_cluster 적용 지금은 필터이후에 추출하는방식
*
*/
func (sched *OpenMCPScheduler) EraseScheduling(dep *resourcev1alpha1.OpenMCPDeployment, replicas int32, clusters map[string]*ketiresource.Cluster, request map[string]int32) string {
	// Make resource to schedule pod into cluster
	if sched.SchdPolicy == "RR" {
		return "RR"
	}
	newPod := newPodFromOpenMCPDeployment(dep)
	filterdResult := sched.Framework.EraseFilterPluginsOnClusters(newPod, clusters, request)

	return filterdResult

}

func (sched *OpenMCPScheduler) ScheduleOne(newPod *ketiresource.Pod, replicas int32, dep *resourcev1alpha1.OpenMCPDeployment, posted bool, requestclusters []string) (string, error) {
	startTime := time.Now()
	TempClusterInfos := make(map[string]*ketiresource.Cluster)

	if requestclusters != nil {

		for _, s := range requestclusters {
			TempClusterInfos[s] = sched.ClusterInfos[s]
		}
	} else {
		TempClusterInfos = sched.ClusterInfos
	}

	filterdResult := sched.Framework.RunFilterPluginsOnClusters(newPod, TempClusterInfos, sched.ClusterManager)

	filteredCluster := make(map[string]*ketiresource.Cluster)
	for clusterName, isfiltered := range filterdResult {
		if isfiltered {
			filteredCluster[clusterName] = sched.ClusterInfos[clusterName]
		}
	}
	// //define cluster 설정시 수행
	// for clusterName, isfiltered := range filterdResult {
	// 	if isfiltered {
	// 		if requestclusters == nil {
	// 			filteredCluster[clusterName] = sched.ClusterInfos[clusterName]
	// 		}
	// 		if requestclusters != nil {
	// 			for _, s := range requestclusters {
	// 				if s == clusterName {
	// 					filteredCluster[clusterName] = sched.ClusterInfos[clusterName]
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	if len(filteredCluster) == 0 {
		if replicas < 0 {
			return "", fmt.Errorf("There is no cluster")
		}
		postresult := sched.Framework.RunPostFilterPluginsOnClusters(newPod, sched.ClusterInfos, sched.PostDeployments)

		if postresult["unscheduable"] {
			return "", fmt.Errorf("There is postpods error")
		}
		if postresult["error"] {
			return "", fmt.Errorf("There is PostFilter Error")
		}
		if postresult["success"] && !posted {
			post := new(ketiresource.PostDelployment)
			post.NewPod = newPod
			post.RemainReplica = replicas
			post.Fcnt = 5
			post.NewDeployment = dep.DeepCopy()
			omcplog.V(4).Infof("Posting Resource Get => [Name] : %v, [Namespace]  : %v [replicas] : %v", post.NewDeployment.Name, post.NewDeployment.Namespace, post.RemainReplica)
			sched.PostDeployments.PushBack(post)

			return "post", nil
		}
		return "", fmt.Errorf("There is no cluster")
	}
	elapsedTime := time.Since(startTime)
	omcplog.V(2).Infof("     filter Time [%v]", elapsedTime)
	omcplog.V(2).Infof("     Existing_cluster [%v]", *sched.origin_clusternames)
	omcplog.V(2).Infof("     FilteredResultMap [%v]", PrintFilterString(filteredCluster))
	selectedCluster := sched.Framework.RunScorePluginsOnClusters(newPod, filteredCluster, sched.ClusterInfos, replicas)

	omcplog.V(2).Infof("     SelectedCluster [%v]", selectedCluster)
	omcplog.V(2).Infof("     Scoring Time [%v]", time.Since(startTime))

	sched.Complite_Scheduing(selectedCluster)
	omcplog.V(2).Infof("     Complite_Scheduing")
	return selectedCluster, nil
}

func (sched *OpenMCPScheduler) LocalNetworkAnalysis() {
	clusters := sched.ClusterInfos
	for {

		for _, cluster := range clusters {
			for _, node := range cluster.Nodes {
				var nodeScore int64
				node_info := &protobuf.NodeInfo{ClusterName: cluster.ClusterName, NodeName: node.NodeName}
				client := sched.GRPC_Client
				result, err := client.SendNetworkAnalysis(context.TODO(), node_info)
				if err != nil || result == nil {
					//omcplog.V(0).Infof("cannot get %v's data from openmcp-analytic-engine", node.NodeName)
					//omcplog.V(0).Info(err)
					continue
				}
				if result.RX == -1 || result.TX == -1 {
					continue
				}

				node.UpdateRX = result.RX
				node.UpdateTX = result.TX
				if node.UpdateRX == 0 && node.UpdateTX == 0 {
					nodeScore = 100
				} else {
					nodeScore = int64((1 / float64(node.UpdateRX+node.UpdateTX)) * float64(100))
				}
				node.NodeScore = nodeScore
			}
		}
		time.Sleep(10 * time.Second)
	}

}

func (sched *OpenMCPScheduler) UpdateResources(newPod *ketiresource.Pod, schedulingResult string) {

	var maxScoreNode *ketiresource.NodeInfo
	maxScore := int64(0)

	for _, node := range sched.ClusterInfos[schedulingResult].Nodes {
		//omcplog.V(0).Info("count", node.NodeScore)
		if maxScore <= node.NodeScore {
			maxScoreNode = node
			maxScore = node.NodeScore
		}
	}
	if maxScoreNode == nil {
		omcplog.V(0).Info("UpdateResources  return back")
		return
	}

	maxScoreNode.RequestedResource = ketiresource.AddResources(maxScoreNode.RequestedResource, newPod.RequestedResource)
	maxScoreNode.AllocatableResource = ketiresource.GetAllocatable(maxScoreNode.CapacityResource, maxScoreNode.RequestedResource)

}

// Returns ketiresource.Resource if specified
func newPodFromOpenMCPDeployment(dep *resourcev1alpha1.OpenMCPDeployment) *ketiresource.Pod {
	res := ketiresource.NewResource()
	additionalResource := make([]string, 0)
	affinities := make(map[string][]string)

	for _, container := range dep.Spec.Template.Spec.Template.Spec.Containers {
		for rName, rQuant := range container.Resources.Requests {
			switch rName {
			case corev1.ResourceCPU:
				res.MilliCPU = rQuant.MilliValue()
			case corev1.ResourceMemory:
				res.Memory = rQuant.Value()
			case corev1.ResourceEphemeralStorage:
				res.EphemeralStorage = rQuant.Value()
			default:
				// Casting from ResourceName to stirng because rName is ResourceName type
				resourceName := fmt.Sprintf("%s", rName)
				additionalResource = append(additionalResource, resourceName)
			}
		}
		for key, values := range dep.Spec.Affinity {
			for _, value := range values {
				affinities[key] = append(affinities[key], value)
			}
		}
	}

	return &ketiresource.Pod{
		Pod: &corev1.Pod{
			Spec: openmcpPodSpecToPodSpec(dep.Spec.Template.Spec.Template.Spec),
		},
		RequestedResource:  res,
		AdditionalResource: additionalResource,
		Affinity:           affinities,
	}
}
func (sched *OpenMCPScheduler) SetupResources() error {
	cm := sched.ClusterManager
	sched.ClusterList, _ = cm.Crd_cluster_client.OpenMCPCluster("openmcp").List(v1.ListOptions{})
	// Setup Clusters
	for clusterName, _ := range sched.ClusterClients {
		pods, _ := sched.ClusterClients[clusterName].CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		// informations on cluster level
		allPods := make([]*ketiresource.Pod, 0)
		allNodes := make([]*ketiresource.NodeInfo, 0)
		cluster_request := ketiresource.NewResource()
		cluster_allocatable := ketiresource.NewResource()

		// Setup Pods
		for _, pod := range pods.Items {
			// add Stroage
			pod_request := &ketiresource.Resource{0, 0, 0}
			pod_additionalResource := make([]string, 0)

			for _, container := range pod.Spec.Containers {
				for rName, rQuant := range container.Resources.Requests {
					switch rName {
					case corev1.ResourceCPU:
						pod_request.MilliCPU = rQuant.MilliValue()
					case corev1.ResourceMemory:
						pod_request.Memory = rQuant.Value()
					case corev1.ResourceEphemeralStorage:
						pod_request.EphemeralStorage = rQuant.Value()
					default:
						// Casting from ResourceName to stirng because rName is ResourceName type
						resourceName := fmt.Sprintf("%s", rName)
						pod_additionalResource = append(pod_additionalResource, resourceName)
					}
				}
			}
			newPod := &ketiresource.Pod{
				Pod:                &pod,
				ClusterName:        clusterName,
				NodeName:           pod.Spec.NodeName,
				PodName:            pod.Name,
				RequestedResource:  pod_request,
				AdditionalResource: pod_additionalResource,
			}
			allPods = append(allPods, newPod)
		}

		// Setup Nodes
		nodes, _ := sched.ClusterClients[clusterName].CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		for _, node := range nodes.Items {

			// Get v1.Pod, corev1.ContainerPort and RequestResource
			podsInNode := make([]*ketiresource.Pod, 0)
			node_request := ketiresource.NewResource()

			for _, pod := range allPods {
				if strings.Compare(pod.NodeName, node.Name) == 0 {
					podsInNode = append(podsInNode, pod)
					node_request = ketiresource.AddResources(node_request, pod.RequestedResource)
				}
			}

			// Get capacity, Additional Resource from node Spec
			node_additionalResource := make([]string, 0)
			node_capacity := &ketiresource.Resource{}

			for rName, rQuant := range node.Status.Capacity {
				switch rName {
				case corev1.ResourceCPU:
					node_capacity.MilliCPU = rQuant.MilliValue()
				case corev1.ResourceMemory:
					node_capacity.Memory = rQuant.Value()
				case corev1.ResourceEphemeralStorage:
					node_capacity.EphemeralStorage = rQuant.Value()
				default:
					// Casting from ResourceName to stirng because rName is ResourceName type
					resourceName := fmt.Sprintf("%s", rName)
					node_additionalResource = append(node_additionalResource, resourceName)
				}
			}

			// Get allocatable Resource based on capacity and request
			node_allocatable := ketiresource.GetAllocatable(node_capacity, node_request)

			// Get Affinity
			node_affinity := make(map[string]string)

			for key, value := range node.Labels {
				switch key {
				case "failure-domain.beta.kubernetes.io/region":
					if _, ok := node_affinity["region"]; !ok {
						node_affinity["region"] = value
					}

				case "failure-domain.beta.kubernetes.io/zone":
					if _, ok := node_affinity["zone"]; !ok {
						node_affinity["zone"] = value
					}
				}
			}

			// make new Node
			newNode := &ketiresource.NodeInfo{
				ClusterName: clusterName,
				NodeName:    node.Name,
				Node:        &node,
				Pods:        podsInNode,
				// UsedPorts:				node_usedPorts,
				CapacityResource:    node_capacity,
				RequestedResource:   node_request,
				AllocatableResource: node_allocatable,
				AdditionalResource:  node_additionalResource,
				Affinity:            node_affinity,
				NodeScore:           0,
			}
			allNodes = append(allNodes, newNode)
			cluster_request = ketiresource.AddResources(cluster_request, node_request)
			cluster_allocatable = ketiresource.AddResources(cluster_allocatable, node_allocatable)
		}

		// Setup Cluster

		sched.ClusterInfos[clusterName] = &ketiresource.Cluster{
			ClusterName:         clusterName,
			Nodes:               allNodes,
			RequestedResource:   cluster_request,
			AllocatableResource: cluster_allocatable,
			ClusterList:         sched.ClusterList,
		}
	}
	temp := make([]string, 0)
	sched.origin_clusternames = &temp
	for clustername, _ := range sched.ClusterInfos {
		*sched.origin_clusternames = append(*sched.origin_clusternames, clustername)
	}
	return nil
}

func openmcpContainersToContainers(containers []resourcev1alpha1.OpenMCPContainer) []corev1.Container {
	var newContainers []corev1.Container

	for _, container := range containers {
		newContainer := corev1.Container{
			Name:       container.Name,
			Image:      container.Image,
			Command:    container.Command,
			Args:       container.Args,
			WorkingDir: container.WorkingDir,
			Ports:      container.Ports,
			EnvFrom:    container.EnvFrom,
			Env:        container.Env,
			Resources: corev1.ResourceRequirements{
				Limits:   container.Resources.Limits,
				Requests: container.Resources.Requests,
			},
			VolumeMounts:             container.VolumeMounts,
			VolumeDevices:            container.VolumeDevices,
			LivenessProbe:            container.LivenessProbe,
			ReadinessProbe:           container.ReadinessProbe,
			Lifecycle:                container.Lifecycle,
			TerminationMessagePath:   container.TerminationMessagePath,
			TerminationMessagePolicy: container.TerminationMessagePolicy,
			ImagePullPolicy:          container.ImagePullPolicy,
			SecurityContext:          container.SecurityContext,
			Stdin:                    container.Stdin,
			StdinOnce:                container.StdinOnce,
			TTY:                      container.TTY,
		}
		newContainers = append(newContainers, newContainer)
	}

	return newContainers
}

func openmcpPodSpecToPodSpec(spec resourcev1alpha1.OpenMCPPodSpec) corev1.PodSpec {
	return corev1.PodSpec{
		Volumes:                       spec.Volumes,
		InitContainers:                openmcpContainersToContainers(spec.InitContainers),
		Containers:                    openmcpContainersToContainers(spec.Containers),
		RestartPolicy:                 spec.RestartPolicy,
		TerminationGracePeriodSeconds: spec.TerminationGracePeriodSeconds,
		ActiveDeadlineSeconds:         spec.ActiveDeadlineSeconds,
		DNSPolicy:                     spec.DNSPolicy,
		NodeSelector:                  spec.NodeSelector,
		ServiceAccountName:            spec.ServiceAccountName,
		DeprecatedServiceAccount:      spec.DeprecatedServiceAccount,
		AutomountServiceAccountToken:  spec.AutomountServiceAccountToken,
		NodeName:                      spec.NodeName,
		HostNetwork:                   spec.HostNetwork,
		HostPID:                       spec.HostPID,
		HostIPC:                       spec.HostIPC,
		ShareProcessNamespace:         spec.ShareProcessNamespace,
		SecurityContext:               spec.SecurityContext,
		ImagePullSecrets:              spec.ImagePullSecrets,
		Hostname:                      spec.Hostname,
		Subdomain:                     spec.Subdomain,
		Affinity:                      spec.Affinity,
		SchedulerName:                 spec.SchedulerName,
		Tolerations:                   spec.Tolerations,
		HostAliases:                   spec.HostAliases,
		PriorityClassName:             spec.PriorityClassName,
		Priority:                      spec.Priority,
		DNSConfig:                     spec.DNSConfig,
		ReadinessGates:                spec.ReadinessGates,
		RuntimeClassName:              spec.RuntimeClassName,
		EnableServiceLinks:            spec.EnableServiceLinks,
	}
}

func openmcpPodTemplateSpecToPodTemplateSpec(template resourcev1alpha1.OpenMCPPodTemplateSpec) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: template.ObjectMeta,
		Spec:       openmcpPodSpecToPodSpec(template.Spec),
	}
}

func openmcpDeploymentTemplateSpecToDeploymentSpec(templateSpec resourcev1alpha1.OpenMCPDeploymentTemplateSpec) appsv1.DeploymentSpec {
	return appsv1.DeploymentSpec{
		Replicas:                templateSpec.Replicas,
		Selector:                templateSpec.Selector,
		Template:                openmcpPodTemplateSpecToPodTemplateSpec(templateSpec.Template),
		Strategy:                templateSpec.Strategy,
		MinReadySeconds:         templateSpec.MinReadySeconds,
		RevisionHistoryLimit:    templateSpec.RevisionHistoryLimit,
		Paused:                  templateSpec.Paused,
		ProgressDeadlineSeconds: templateSpec.ProgressDeadlineSeconds,
	}
}
