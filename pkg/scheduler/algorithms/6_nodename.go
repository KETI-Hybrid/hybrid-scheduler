package algorithms

import (
	"fmt"
	"strings"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) NodeName(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	var nodeName []string = nil
	pod := args.Pod
	fmt.Printf("pod add -> name : %s\n", pod.Name)
	fmt.Println("** NodeName Algorithm **")

	fmt.Println("pod spec node name :", pod.Spec.NodeName)

	// get node name
	nodeList := args.Nodes.Items

	for _, node := range nodeList {
		strings.Compare(pod.Spec.NodeName, node.Name)
		nodeName = append(nodeName, node.Name)
	}

	fmt.Printf("Selected Node(s) :", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
