/*
 * Copyright © 2021 peizhaoyou <peizhaoyou@4paradigm.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scheduler

import (
	"context"
	"strings"
	"time"

	"hybrid-scheduler/pkg/scheduler/algorithms"
	"hybrid-scheduler/pkg/util"
	"hybrid-scheduler/pkg/util/client"
	"hybrid-scheduler/pkg/util/nodelock"
	podutil "hybrid-scheduler/pkg/util/podutil"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

type Scheduler struct {
	nodeManager
	podManager
	*algorithms.AlgoManager
	stopCh     chan struct{}
	kubeClient kubernetes.Interface
	podLister  listerscorev1.PodLister
	nodeLister listerscorev1.NodeLister
}

func NewScheduler() *Scheduler {
	klog.Infof("New Scheduler")
	s := &Scheduler{
		stopCh: make(chan struct{}),
	}
	s.nodeManager.init()
	s.podManager.init()
	s.AlgoManager = algorithms.InitAlgoManager()
	return s
}

func check(err error) {
	if err != nil {
		klog.Fatal(err)
	}
}

func (s *Scheduler) onAddPod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	klog.Infoln("onAddPod", pod.Name)
	if !ok {
		klog.Errorf("unknown add object type")
		return
	}
	nodeID := pod.Spec.NodeName
	if len(nodeID) == 0 {
		return
	}
	if podutil.IsPodInTerminatedState(pod) {
		s.delPod(pod)
		return
	}
	s.addPod(pod, nodeID)
}

func (s *Scheduler) onUpdatePod(_, newObj interface{}) {
	s.onAddPod(newObj)
}

func (s *Scheduler) onDelPod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		klog.Errorf("unknown add object type")
		return
	}
	s.delPod(pod)
}

func (s *Scheduler) Start() {
	kubeClient, err := client.NewClient()
	check(err)
	s.kubeClient = kubeClient
	informerFactory := informers.NewSharedInformerFactoryWithOptions(s.kubeClient, time.Hour*1)
	s.podLister = informerFactory.Core().V1().Pods().Lister()
	s.nodeLister = informerFactory.Core().V1().Nodes().Lister()

	informer := informerFactory.Core().V1().Pods().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.onAddPod,
		UpdateFunc: s.onUpdatePod,
		DeleteFunc: s.onDelPod,
	})

	informerFactory.Start(s.stopCh)
	informerFactory.WaitForCacheSync(s.stopCh)

}

func (s *Scheduler) RegisterFromNodeAnnotatons() error {
	klog.V(5).Infoln("Scheduler into RegisterFromNodeAnnotations")
	for {
		nodes, err := s.kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
			//LabelSelector: "gpu=on",
		})
		if err != nil {
			klog.Errorln("nodes list failed", err.Error())
			return err
		}
		for _, val := range nodes.Items {
			nodeInfo := &NodeInfo{}
			nodeInfo.ID = val.Name
			s.addNode(val.Name, nodeInfo)
			node, err := s.kubeClient.CoreV1().Nodes().Get(context.TODO(), val.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorln(err.Error())
			}
			if _, ok := node.Status.Allocatable["keti.hybrid/schedule"]; !ok {
				node.Status.Allocatable["keti.hybrid/schedule"] = resource.MustParse("100") // VALUE를 실제 값으로 대체
				_, err = s.kubeClient.CoreV1().Nodes().UpdateStatus(context.TODO(), node, metav1.UpdateOptions{})
				if err != nil {
					klog.Errorln(err.Error())
				}
			}
			if s.nodes[val.Name] != nil && nodeInfo != nil {
				klog.Infof("node %v come node info=%v", val.Name, nodeInfo)
			}
		}
		time.Sleep(time.Second * 15)
	}
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) Bind(args extenderv1.ExtenderBindingArgs) (*extenderv1.ExtenderBindingResult, error) {
	klog.InfoS("Bind", "pod", args.PodName, "namespace", args.PodNamespace, "podUID", args.PodUID, "node", args.Node)
	var err error
	var res *extenderv1.ExtenderBindingResult
	binding := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: args.PodName, UID: args.PodUID},
		Target:     v1.ObjectReference{Kind: "Node", Name: args.Node},
	}
	current, err := s.kubeClient.CoreV1().Pods(args.PodNamespace).Get(context.Background(), args.PodName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Get pod failed")
	}
	err = nodelock.LockNode(args.Node)
	if err != nil {
		klog.ErrorS(err, "Failed to lock node", "node", args.Node)
	}
	//defer util.ReleaseNodeLock(args.Node)

	tmppatch := make(map[string]string)
	tmppatch[util.DeviceBindPhase] = util.DeviceBindAllocating
	tmppatch[util.BindTimeAnnotations] = time.Now().Format("2006-01-02 15:04:05")

	err = util.PatchPodAnnotations(current, tmppatch)
	if err != nil {
		klog.ErrorS(err, "patch pod annotation failed")
	}
	if err = s.kubeClient.CoreV1().Pods(args.PodNamespace).Bind(context.Background(), binding, metav1.CreateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to bind pod", "pod", args.PodName, "namespace", args.PodNamespace, "podUID", args.PodUID, "node", args.Node)
	}
	if err == nil {
		res = &extenderv1.ExtenderBindingResult{
			Error: "",
		}
	} else {
		res = &extenderv1.ExtenderBindingResult{
			Error: err.Error(),
		}
	}
	klog.Infoln("After Binding Process")
	return res, nil
}

func (s *Scheduler) Filter(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	var err error
	klog.Infof("schedule pod %v/%v[%v]", args.Pod.Namespace, args.Pod.Name, args.Pod.UID)
	args.Nodes, err = s.kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Errorln(err)
	}
	nodeName := make([]string, 0)
	for _, node := range args.Nodes.Items {
		nodeName = append(nodeName, node.Name)
	}
	args.NodeNames = &nodeName
	s.delPod(args.Pod)
	res := &extenderv1.ExtenderFilterResult{}
	if algoName, ok := args.Pod.Annotations["schedulepolicy"]; ok {
		algoName = strings.ToLower(algoName)
		res, err = s.Do(algoName, args)
		if err != nil {
			klog.Errorln(err)
		}
	}
	nodeScores := util.GetScore()
	nodeID, minScore := "", float32(3.4e+38)
	for _, nodeName := range *res.NodeNames {
		score := nodeScores[nodeName]

		if score < float32(minScore) {
			minScore = score
			nodeID = nodeName
		}
	}
	klog.Infof("schedule %v/%v to %v", args.Pod.Namespace, args.Pod.Name, nodeID)
	s.addPod(args.Pod, nodeID)
	res = &extenderv1.ExtenderFilterResult{NodeNames: &[]string{nodeID}}
	return res, nil

}
