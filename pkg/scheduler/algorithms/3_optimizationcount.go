package algorithms

import (
	"fmt"
	"strconv"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) OptimizationCount(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** OptimizationCount Algorithm **")
	nodeName, minCnt := "", 2147483647
	for _, node := range args.Nodes.Items {
		cntStr := node.Annotations["optimazationCount"]
		cnt, _ := strconv.Atoi(cntStr)
		if minCnt > cnt {
			nodeName = node.Name
			minCnt = cnt
		}
	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &[]string{nodeName}}, nil
}
