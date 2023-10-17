package algorithms

import (
	"fmt"
	"hybrid-scheduler/pkg/util"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) BalanceAllocation(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	return_nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** BalanceAllocation Algorithm **")
	nodeScores := util.GetScore()
	select_node := ""

	for _, nodeName := range *args.NodeNames {
		score := nodeScores[nodeName]

		if select_node == "" {
			select_node = nodeName
		}
		fmt.Printf("%s : %.1f\n", nodeName, score)

	}

	fmt.Printf("Selected Node(s) : %s\n", select_node)
	return_nodeName = append(return_nodeName, select_node)
	return &extenderv1.ExtenderFilterResult{NodeNames: &return_nodeName}, nil
}
