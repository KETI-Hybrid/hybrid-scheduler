package algorithms

import (
	"fmt"
	"strings"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) NodeUnschedulable(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	var availScheduleNode []string = nil
	var unavailScheduleNode []string = nil
	nodeList := args.Nodes.Items

	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** NodeUnschedulable Algorithm **")

	for _, node := range nodeList {
		for _, avail := range node.Spec.Taints {

			if strings.Compare(avail.Key, "NoSchedule") == 0 {
				unavailScheduleNode = append(unavailScheduleNode, node.Name)
			} else {
				availScheduleNode = append(availScheduleNode, node.Name)
			}
		}
	}

	fmt.Println("Unschedulable node list :", unavailScheduleNode)
	fmt.Println("Schedulable node list :", availScheduleNode)

	fmt.Println("Selected Node(s) :", availScheduleNode)

	return &extenderv1.ExtenderFilterResult{NodeNames: &availScheduleNode}, nil
}
