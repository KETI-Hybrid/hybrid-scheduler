package algorithms

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) TaintToleration(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** TaintToleration Algorithm **")
	for _, node := range args.Nodes.Items {
		if len(node.Spec.Taints) > 0 {
			for _, taint := range node.Spec.Taints {
				if taint.Effect == v1.TaintEffectNoSchedule {
					fmt.Printf("%s : Exist\n", node.Name)
					fmt.Println("Toleration match: False")
					break
				} else if taint.Effect == v1.TaintEffectPreferNoSchedule {
					fmt.Printf("%s : Exist\n", node.Name)
					fmt.Println("Toleration match: False")
					break
				}
			}
		} else {
			fmt.Printf("%s : None\n", node.Name)
			nodeName = append(nodeName, node.Name)
		}
	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
