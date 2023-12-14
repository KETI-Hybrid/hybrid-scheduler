/*
Copyright 2018 The Multicluster-Controller Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"math"
	"openmcp/openmcp/apis"
	resourcev1alpha1 "openmcp/openmcp/apis/resource/v1alpha1"
	"openmcp/openmcp/omcplog"
	openmcpscheduler "openmcp/openmcp/openmcp-scheduler/src"
	"openmcp/openmcp/util/clusterManager"
	"openmcp/openmcp/util/controller/logLevel"
	"openmcp/openmcp/util/controller/reshape"
	"sort"

	"admiralty.io/multicluster-controller/pkg/cluster"
	"admiralty.io/multicluster-controller/pkg/controller"
	"admiralty.io/multicluster-controller/pkg/manager"
	"admiralty.io/multicluster-controller/pkg/reconcile"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//"k8s.io/client-go/util/retry"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
)

type reconciler struct {
	live           client.Client
	ghosts         []client.Client
	ghostNamespace string
	scheduler      *openmcpscheduler.OpenMCPScheduler
}

type patchUInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value uint32 `json:"value"`
}

func NewControllers(cm *clusterManager.ClusterManager, scheduler *openmcpscheduler.OpenMCPScheduler) {
	host_ctx := "hcp-master-cluster"
	namespace := "hcp"

	host_cfg := cm.Host_config
	live := cluster.New(host_ctx, host_cfg, cluster.Options{})

	ghosts := []*cluster.Cluster{}

	for _, ghost_cluster := range cm.Cluster_list.Items {
		ghost_ctx := ghost_cluster.Name
		ghost_cfg := cm.Cluster_configs[ghost_ctx]

		ghost := cluster.New(ghost_ctx, ghost_cfg, cluster.Options{})
		ghosts = append(ghosts, ghost)
	}

	sched_cont, err := NewController(live, ghosts, namespace, scheduler, cm)
	if err != nil {
		omcplog.V(0).Info("err New Controller - Scheduler", err)
	}
	reshape_cont, err := reshape.NewController(live, ghosts, namespace, cm)
	if err != nil {
		omcplog.V(0).Info("err New Controller - Reshape", err)
	}
	loglevel_cont, err := logLevel.NewController(live, ghosts, namespace)
	if err != nil {
		omcplog.V(0).Info("err New Controller - logLevel", err)
	}

	m := manager.New()
	m.AddController(sched_cont)
	m.AddController(reshape_cont)
	m.AddController(loglevel_cont)

	stop := reshape.SetupSignalHandler()

	if err := m.Start(stop); err != nil {
		omcplog.V(0).Info(err)
	}

}

var cm *clusterManager.ClusterManager

func NewController(live *cluster.Cluster, ghosts []*cluster.Cluster, ghostNamespace string, scheduler *openmcpscheduler.OpenMCPScheduler, clusterManager *clusterManager.ClusterManager) (*controller.Controller, error) {
	omcplog.V(4).Info("[OpenMCP Schduler] NewController Function Called")
	cm = clusterManager
	liveclient, err := live.GetDelegatingClient()
	if err != nil {
		return nil, fmt.Errorf("getting delegating client for live cluster: %v", err)
	}
	ghostclients := []client.Client{}
	for _, ghost := range ghosts {
		ghostclient, err := ghost.GetDelegatingClient()
		if err != nil {
			return nil, fmt.Errorf("getting delegating client for ghost cluster: %v", err)
		}
		ghostclients = append(ghostclients, ghostclient)
	}
	recon := &reconciler{live: liveclient, ghosts: ghostclients, ghostNamespace: ghostNamespace, scheduler: scheduler}
	co := controller.New(recon, controller.Options{})
	if err := apis.AddToScheme(live.GetScheme()); err != nil {
		return nil, fmt.Errorf("adding APIs to live cluster's scheme: %v", err)
	}

	if err := co.WatchResourceReconcileObject(context.TODO(), live, &resourcev1alpha1.OpenMCPDeployment{}, controller.WatchOptions{}); err != nil {
		return nil, fmt.Errorf("setting up Pod watch in live cluster: %v", err)
	}

	for _, ghost := range ghosts {
		if err := co.WatchResourceReconcileController(context.TODO(), ghost, &appsv1.Deployment{}, controller.WatchOptions{}); err != nil {
			return nil, fmt.Errorf("setting up PodGhost watch in ghost cluster: %v", err)
		}
	}
	recon.scheduler.Live = &recon.live
	return co, nil
}
func (r *reconciler) JoinReconcile(newDeployment *resourcev1alpha1.OpenMCPDeployment) *resourcev1alpha1.OpenMCPDeployment {

	return newDeployment

}
func (r *reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	// Fetch the OpenMCPDeployment instance
	r.scheduler.Live = &r.live
	newDeployment := &resourcev1alpha1.OpenMCPDeployment{}
	err := r.live.Get(context.TODO(), req.NamespacedName, newDeployment)
	IsExist := false
	clusters := newDeployment.Spec.Clusters
	if err != nil && errors.IsNotFound(err) {
		omcplog.V(5).Info("Not Found::", newDeployment.GetName)
		return reconcile.Result{}, nil
	}

	// apply &&create 처리
	if newDeployment.Status.SchedulingNeed == true && newDeployment.Status.SchedulingComplete == false {
		lastspec := newDeployment.Status.LastSpec
		temp := newDeployment.Spec.Replicas
		omcplog.V(0).Info("################## Scheduling Start #####################")
		omcplog.V(0).Infof("Start Deployment : %v", newDeployment.Name)
		omcplog.V(0).Infof("Namespace : %v", newDeployment.Namespace)

		if r.scheduler.Live == nil {
			omcplog.V(0).Info("sched.Live NIL")
		}
		// unjoin join 요청된경우에는 클러스터맵이 있음
		if newDeployment.Status.ClusterMaps != nil {
			count := 0
			chagnedp := map[string]int32{}
			//replica 개수와 배포된 포드의 개수확인
			for key, val := range newDeployment.Status.ClusterMaps {
				_, exists := newDeployment.Status.ClusterMaps[key]
				if !exists {
					count = 1
					chagnedp[key] = 1
				} else {
					count += int(val)
					chagnedp[key] += val
				}
			}
			//allreschedcheck := false
			//두조건이 같다면 전체 리스케줄링 따라서 아닌경우에만 처리해주면 됨
			if int32(count) != temp && (lastspec.Replicas < newDeployment.Spec.Replicas) {
				//allreschedcheck = true
				extra_replicas := temp - int32(count)
				if r.scheduler.SchdPolicy == "RR" {
					newDeployment.Spec.Replicas = newDeployment.Spec.Replicas
				} else {
					newDeployment.Spec.Replicas = extra_replicas
				}

				omcplog.V(5).Infof("  extra_replicas =>", extra_replicas)
				//omcplog.V(0).Infof("  ClusterJoin Resource Get => [Name] : %v, [Namespace]  : %v", newDeployment.Name, newDeployment.Namespace)
				cluster_replicas_map, _ := r.scheduler.Scheduling(newDeployment, false, clusters)
				if r.scheduler.SchdPolicy != "RR" {
					for key, val := range cluster_replicas_map {
						_, exists := cluster_replicas_map[key]
						if !exists {
							chagnedp[key] = 1
						} else {
							chagnedp[key] += val
						}
					}

					newDeployment.Status.ClusterMaps = chagnedp
				} else {
					newDeployment.Status.ClusterMaps = cluster_replicas_map
				}
				newDeployment.Status.Replicas = temp
				newDeployment.Status.SchedulingNeed = false
				newDeployment.Status.SchedulingComplete = true
				//		omcplog.V(5).Infof("cluster map =", cluster_replicas_map)
				if !(len(cluster_replicas_map) == 0) {
					omcplog.V(0).Info("ClusterJoin Scheduling Result : ", chagnedp)
				}
				err := r.live.Status().Update(context.TODO(), newDeployment)
				if err != nil {
					omcplog.V(0).Infof("ClusterJoin Failed to update instance status, %v", err)
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, nil
			} else {
				//조인 언조인이후 동일한 경우 전체 리스케줄링하기 때문에 아래 단계로 전체 수행하면됨
			}
		}

		//ex 래플리카 8 에서 7로 변경될경우 수행 기존보다 현재가 더작은경우  수행
		if lastspec.Replicas != 0 && lastspec.Replicas > newDeployment.Spec.Replicas {
			decre := lastspec.Replicas - newDeployment.Spec.Replicas
			omcplog.V(5).Infof("replica 감소량 =", decre)
			for i := 0; i < int(decre); i++ {

				reasecluster := r.scheduler.EraseScheduling(newDeployment, lastspec.Replicas-newDeployment.Spec.Replicas, r.scheduler.ClusterInfos, newDeployment.Status.ClusterMaps)
				//RR일경우 erase를 할 수없음
				if reasecluster == "RR" {
					omcplog.V(5).Info("=> D RR ..")
					cluster_replicas_map, _ := r.scheduler.Scheduling(newDeployment, false, clusters)
					if !(len(cluster_replicas_map) == 0) {
						omcplog.V(0).Info("Scheduling Result :", cluster_replicas_map)
					}
					newDeployment.Status.ClusterMaps = cluster_replicas_map
					newDeployment.Status.Replicas = newDeployment.Status.Replicas
					newDeployment.Status.SchedulingNeed = false
					newDeployment.Status.SchedulingComplete = true
					//		omcplog.V(5).Infof("cluster map =", cluster_replicas_map)

					err := r.live.Status().Update(context.TODO(), newDeployment)
					if err != nil {
						omcplog.V(0).Infof("Failed to update instance status, %v", err)
						return reconcile.Result{}, err
					}
					return reconcile.Result{}, nil

				}
				if reasecluster == "" {
					omcplog.V(5).Info("error")
				}

				omcplog.V(5).Info("=> reasecluster : ", newDeployment.Status.ClusterMaps)
				newDeployment.Status.ClusterMaps[reasecluster] -= 1
			}
			newDeployment.Status.Replicas = newDeployment.Spec.Replicas
			newDeployment.Status.SchedulingNeed = false
			newDeployment.Status.SchedulingComplete = true

			if !(len(newDeployment.Status.ClusterMaps) == 0) {

				omcplog.V(0).Info("Scheduling Result : ", newDeployment.Status.ClusterMaps)
			}
			err := r.live.Status().Update(context.TODO(), newDeployment)
			if err != nil {
				omcplog.V(0).Infof("Failed to update instance status, %v", err)
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		if lastspec.Replicas != 0 && lastspec.Replicas < newDeployment.Spec.Replicas {
			IsExist = true
			//남은 개수만큼  분리
			//omcplog.V(5).Infof("남은개수만큼 지정", lastspec.Replicas, newDeployment.Spec.Replicas)
			if r.scheduler.SchdPolicy != "RR" {
				newDeployment.Spec.Replicas = newDeployment.Spec.Replicas - lastspec.Replicas
			}
		}
		// omcplog.V(0).Infof("  Resource Get => [Name] : %v, [Namespace]  : %v", newDeployment.Name, newDeployment.Namespace)
		cluster_replicas_map, _ := r.scheduler.Scheduling(newDeployment, false, clusters)

		if IsExist {
			IsExist = false
			//omcplog.V(5).Infof("리소스 변경 기존보다 커짐")
			for clustername, cnt := range newDeployment.Status.ClusterMaps {
				cluster_replicas_map[clustername] = cluster_replicas_map[clustername] + cnt
				omcplog.V(5).Infof("clustermap =", cluster_replicas_map)
			}
		}
		newDeployment.Status.ClusterMaps = cluster_replicas_map
		newDeployment.Status.Replicas = temp
		newDeployment.Status.SchedulingNeed = false
		newDeployment.Status.SchedulingComplete = true
		//		omcplog.V(5).Infof("cluster map =", cluster_replicas_map)
		if !(len(cluster_replicas_map) == 0) {
			omcplog.V(0).Infof("Scheduling Result : %v", cluster_replicas_map)
		}
		err := r.live.Status().Update(context.TODO(), newDeployment)
		if err != nil {
			omcplog.V(0).Infof("Failed to update instance status, %v", err)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func RRScheduling(cm *clusterManager.ClusterManager, replicas int32) map[string]int32 {

	cluster_replicas_map := make(map[string]int32)

	remain_rep := replicas
	rep := 0
	namespace := "kube-federation-system"
	cluster_len := len(cm.Cluster_list.Items)
	for i, cluster := range cm.Cluster_list.Items {
		if remain_rep == 0 {
			break
		}
		// except := false
		joined_cluster := &fedv1b1.KubeFedCluster{}
		err := cm.Host_client.Get(context.TODO(), joined_cluster, namespace, cluster.Name)
		if err != nil {
			return nil
		}
		// for k, v := range joined_cluster.Labels {
		// 	if k == "openmcp" && v == "true" {
		// 		omcplog.V(0).Info("Scheduling Except Cluster !! Include OpenMCP Label : ", k, v)
		// 		except = true
		// 		break
		// 	}
		// }
		// if except {
		// 	continue
		// }

		if i == cluster_len-1 {
			rep = int(remain_rep)
		} else {
			// rep = int(math.Ceil(float64(replicas) / cluster_len))
			reqd := math.Ceil((float64(remain_rep) / float64(cluster_len-i)))
			rep = int(reqd)
		}

		remain_rep = remain_rep - int32(rep)
		cluster_replicas_map[cluster.Name] = int32(rep)
		omcplog.V(4).Info("cluster.Name : ", cluster.Name)
		omcplog.V(4).Info("remain_rep : ", remain_rep)
		omcplog.V(4).Info("cluster_replicas_map : ", cluster_replicas_map)

	}
	keys := make([]string, 0)
	for k, _ := range cluster_replicas_map {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := cluster_replicas_map[k]
		omcplog.V(0).Info("  ", k, ": ", v)
	}
	omcplog.V(0).Info("Scheduling Result: ", cluster_replicas_map)
	return cluster_replicas_map
}
