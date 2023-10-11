package algorithms

import (
	"fmt"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) JoinCheck(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {

	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** Join Check **")
	fmt.Print("Joined node list : ")
	for i, nodeName := range *args.NodeNames {
		fmt.Print(nodeName)
		if i < len(*args.NodeNames)-1 {
			fmt.Print(", ")
		} else {
			fmt.Print("\n")
		}
	}

	fmt.Printf("Selected Node(s) :")
	for i, nodeName := range *args.NodeNames {
		fmt.Print(nodeName)
		if i < len(*args.NodeNames)-1 {
			fmt.Print(", ")
		} else {
			fmt.Print("\n")
		}
	}

	return &extenderv1.ExtenderFilterResult{NodeNames: args.NodeNames}, nil
}
