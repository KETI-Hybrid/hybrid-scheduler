package algorithms

import (
	"fmt"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) NodeName(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := args.Pod.Spec.NodeName
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** NodeName Algorithm **")

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &[]string{nodeName}}, nil
}
