package algorithms

import (
	"context"
	"encoding/json"
	"fmt"
	"hybrid-scheduler/pkg/util/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

type ImageData struct {
	ID string `json:"id"`
}

func (a *AlgoManager) ImageLocality(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	nodeImageMap := make(map[string]map[string]bool)
	nodeMap := make(map[string]bool)
	var err error
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** ImageLocality Algorithm **")
	if a.kubeClient == nil {
		a.kubeClient, err = client.NewClient()
		if err != nil {
			klog.Errorln(err)
		}
	}
	configMaps, err := a.kubeClient.CoreV1().ConfigMaps("keti-system").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Errorln(err)
	}
	configMap := corev1.ConfigMap{}

	for _, configmap := range configMaps.Items {
		if configmap.Name == "imagelist" {
			configMap = configmap
		}
	}
	for _, node := range args.Nodes.Items {
		nodeMap[node.Name] = true
		imageData := make([]ImageData, 0)

		err = json.Unmarshal([]byte(configMap.Data[node.Name+".json"]), &imageData)
		if err != nil {
			klog.Errorln(err)
		}

		nodeImageMap[node.Name] = make(map[string]bool)
		for _, image := range imageData {
			nodeImageMap[node.Name][image.ID] = true
		}
	}
	for _, container := range args.Pod.Spec.Containers {
		for _, node := range args.Nodes.Items {
			if !nodeImageMap[node.Name][container.Image] {
				nodeMap[node.Name] = false
			}
		}
	}

	for node, isImage := range nodeMap {
		if isImage {
			nodeName = append(nodeName, node)
		}
	}
	fmt.Print("Selecting node with existing container for pod list : ")
	for i, name := range nodeName {
		if i == len(nodeName)-1 {
			fmt.Printf("%s\n", name)
		} else {
			fmt.Printf("%s ,", name)
		}
	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
