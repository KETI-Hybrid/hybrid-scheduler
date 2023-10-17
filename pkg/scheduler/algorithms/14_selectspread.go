package algorithms

import (
	"context"
	"fmt"
	"hybrid-scheduler/pkg/util/client"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) SelectSpread(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** SelectSpread Algorithm **")
	fmt.Println("Prefer checking...")
	var err error
	nodeMap := make(map[string]bool)
	if a.kubeClient == nil {
		a.kubeClient, err = client.NewClient()
		if err != nil {
			klog.Errorln(err)
		}
	}
	kind := args.Pod.Labels["kind"]
	podPrefix := a.kubeClient.CoreV1().Pods("")
	labelMap := make(map[string]string)
	labelMap["kind"] = kind

	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	}
	kindPods, err := podPrefix.List(context.Background(), options)
	if err != nil {
		klog.Errorln(err)
	}
	for _, pod := range kindPods.Items {
		if pod.Name == args.Pod.Name {
			continue
		}
		nodeMap[pod.Spec.NodeName] = true
	}
	for name, _ := range nodeMap {
		nodeName = append(nodeName, strings.TrimSpace(name))
	}

	fmt.Print("list : ")
	if len(nodeName) > 0 {
		for i, name := range nodeName {
			if i == len(nodeName) {
				fmt.Printf("%s\n", name)
			} else {
				fmt.Printf("%s ,", name)
			}
		}
	} else {
		fmt.Print("None\n")
		for _, node := range args.Nodes.Items {
			nodeName = append(nodeName, node.Name)
		}
	}
	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
