package algorithms

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) NodeUnschedulable(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** NodeUnschedulable Algorithm **")
	for _, node := range args.Nodes.Items {
		if len(node.Spec.Taints) > 0 {
			for _, taint := range node.Spec.Taints {
				if taint.Effect == v1.TaintEffectNoSchedule {
					break
				} else if taint.Effect == v1.TaintEffectPreferNoSchedule {
					break
				} else {
					nodeName = append(nodeName, node.Name)
				}
			}
		} else {
			continue
		}
	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
