package algorithms

import (
	"fmt"
	"time"

	"k8s.io/klog"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) OptimizationTime(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** OptimizationTime Algorithm **")
	nodeName, minCnt := "", float64(2147483647)
	for _, node := range args.Nodes.Items {
		cntStr := node.Annotations["optimazationTime"]
		lastTime, err := time.Parse(time.DateTime, cntStr)
		if err != nil {
			klog.Errorln(err)
		}

		cnt := time.Since(lastTime).Seconds()
		if minCnt > cnt {
			nodeName = node.Name
			minCnt = cnt
		}
	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &[]string{nodeName}}, nil
}
