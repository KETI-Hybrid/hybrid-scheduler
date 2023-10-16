package algorithms

import (
	"fmt"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) NodePreferAvoidPods(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := ""
	maxScore := 0
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** NodePreferAvoidPods Algorithm **")
	fmt.Println("Check node annotation")

	nodeList := args.Nodes.Items

	for _, node := range nodeList {
		score := 0
		annos := node.Annotations

		score = score - len(annos)

		fmt.Println(node.Name, ":", score)

		if score > maxScore {
			nodeName = node.Name
			maxScore = score
		}
	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &[]string{nodeName}}, nil
}
