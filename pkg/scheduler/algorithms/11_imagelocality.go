package algorithms

import (
	"fmt"
	"strings"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) ImageLocality(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	nodeImageMap := make(map[string]map[string]bool)
	nodeMap := make(map[string]bool)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** ImageLocality Algorithm **")
	for _, node := range args.Nodes.Items {
		images := node.Status.Images
		nodeMap[node.Name] = true
		nodeImageMap[node.Name] = make(map[string]bool)
		for _, image := range images {
			imageName := strings.Split(image.Names[0], "@")[0]
			nodeImageMap[node.Name][imageName] = true
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
	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
