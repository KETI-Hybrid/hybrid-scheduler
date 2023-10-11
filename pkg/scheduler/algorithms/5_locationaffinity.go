package algorithms

import (
	"fmt"
	"strings"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) LocationAffinity(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	zoneNodeName := make([]string, 0)
	RegionNodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** LocationAffinity Algorithm **")
	zone, region := "", ""
	for _, node := range args.Nodes.Items {
		zone = node.Annotations["zone"]
		region = node.Annotations["region"]
		if strings.Compare(args.Pod.Annotations["region"], region) == 0 {
			if strings.Compare(args.Pod.Annotations["zone"], zone) == 0 {
				zoneNodeName = append(zoneNodeName, node.Name)
			} else {
				RegionNodeName = append(RegionNodeName, node.Name)
			}
		} else {
			continue
		}
	}
	if len(zoneNodeName) > 0 {
		fmt.Printf("Selected Node(s) : %s\n", zoneNodeName)
		return &extenderv1.ExtenderFilterResult{NodeNames: &zoneNodeName}, nil
	} else {
		fmt.Printf("Selected Node(s) : %s\n", RegionNodeName)
		return &extenderv1.ExtenderFilterResult{NodeNames: &RegionNodeName}, nil
	}
}
