package algorithms

import (
	"fmt"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) SelectSpread(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	//nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** SelectSpread Algorithm **")

	// kind := args.Pod.Kind

	// for _, node := range args.Nodes.Items {
	// 	if kind == ? {
	// 		nameNameList = append(nodeNameList, node.Name)
	// 	}
	// }

	fmt.Printf("Selected Node(s) : %s\n", *args.NodeNames)

	return nil, nil
}
