package algorithms

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) Affinity(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** Affinity Algorithm **")

	affinity := args.Pod.Spec.Affinity.NodeAffinity
	listoptions := metav1.ListOptions{}
	labelselector := metav1.LabelSelector{}

	for _, require := range affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, exp := range require.MatchExpressions {
			tmp := metav1.LabelSelectorRequirement{
				Key:      exp.Key,
				Operator: metav1.LabelSelectorOperator(exp.Operator),
				Values:   exp.Values,
			}
			labelselector.MatchExpressions = append(labelselector.MatchExpressions, tmp)
		}
		listoptions.LabelSelector = labelselector.String()
		labelselector = metav1.LabelSelector{}
		for _, exp := range require.MatchFields {
			tmp := metav1.LabelSelectorRequirement{
				Key:      exp.Key,
				Operator: metav1.LabelSelectorOperator(exp.Operator),
				Values:   exp.Values,
			}
			labelselector.MatchExpressions = append(labelselector.MatchExpressions, tmp)
		}
		listoptions.FieldSelector = labelselector.String()
		labelselector = metav1.LabelSelector{}
	}

	nodes, err := a.kubeClient.CoreV1().Nodes().List(context.Background(), listoptions)
	if err != nil {
		klog.Errorln(err)
	}
	if len(nodes.Items) > 0 {
		for _, node := range nodes.Items {
			nodeName = append(nodeName, node.Name)
		}
		fmt.Printf("Selected Node(s) : %s\n", nodeName)
		return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
	} else {
		fmt.Printf("Selected Node(s) : %s\n", nodeName)
		return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
	}
}
