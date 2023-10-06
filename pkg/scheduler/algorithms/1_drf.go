package algorithms

import (
	"fmt"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) DRF(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := ""
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** DRF Algorithm **")

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return nil, nil
}
