package algorithms

import (
	"fmt"
	"strings"
	"time"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) NodeName(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	time.Sleep(time.Millisecond)
	var nodeName string
	pod := args.Pod
	fmt.Printf("pod add -> name : %s\n", pod.Name)
	fmt.Print("** NodeName Algorithm **\n")

	fmt.Print("pod spec node name :", pod.Labels["nodeName"], "\n")

	// get node name
	nodeList := args.Nodes.Items

	for _, node := range nodeList {
		if strings.Compare(pod.Labels["nodeName"], node.Name) == 0 {
			nodeName = node.Name
		}
	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)
	time.Sleep(time.Millisecond)
	return &extenderv1.ExtenderFilterResult{NodeNames: &[]string{nodeName}}, nil
}
