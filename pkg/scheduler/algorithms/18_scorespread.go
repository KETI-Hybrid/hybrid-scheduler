package algorithms

import (
	"fmt"
	"hybrid-scheduler/pkg/util"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) ScoreSpread(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	return_nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** ScoreSpread Algorithm **")
	nodeScores := util.GetScore()
	nodeID, minScore := "", float32(3.4e+38)
	for _, nodeName := range *args.NodeNames {
		score := nodeScores[nodeName]

		if score < float32(minScore) {
			minScore = score
			nodeID = nodeName
		}
		fmt.Printf("%s : %.1f\n", nodeName, score)
	}

	fmt.Printf("Selected Node(s) : %s\n", nodeID)
	return_nodeName = append(return_nodeName, nodeID)
	return &extenderv1.ExtenderFilterResult{NodeNames: &return_nodeName}, nil
}
