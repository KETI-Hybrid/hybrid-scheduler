package algorithms

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) Affinity(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** Affinity Algorithm **")
	fmt.Println("Checking pod affinity")

	affinity := args.Pod.Spec.Affinity.NodeAffinity
	var nodeLabelSelector labels.Selector
	var err error

	for _, require := range affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		// TODO: Avoid computing it for all nodes if this becomes a performance problem.
		nodeLabelSelector, err = NodeSelectorRequirementsAsSelector(require.MatchExpressions)
		if err != nil {
			klog.Error(err)
		}
	}
	fmt.Print("keti-application4 affinity: ")

	for _, node := range args.Nodes.Items {
		if nodeLabelSelector.Matches(labels.Set(node.Labels)) {
			nodeName = append(nodeName, node.Name)
		}
	}

	for i, name := range nodeName {
		if i == len(nodeName)-1 {
			fmt.Printf("%s\n", name)
		} else {
			fmt.Printf("%s ,", name)
		}
	}
	fmt.Printf("Selected Node(s) : %s\n", nodeName)
	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}

func NodeSelectorRequirementsAsSelector(nsm []corev1.NodeSelectorRequirement) (labels.Selector, error) {
	if len(nsm) == 0 {
		return labels.Nothing(), nil
	}
	selector := labels.NewSelector()
	for _, expr := range nsm {
		var op selection.Operator
		switch expr.Operator {
		case corev1.NodeSelectorOpIn:
			op = selection.In
		case corev1.NodeSelectorOpNotIn:
			op = selection.NotIn
		case corev1.NodeSelectorOpExists:
			op = selection.Exists
		case corev1.NodeSelectorOpDoesNotExist:
			op = selection.DoesNotExist
		case corev1.NodeSelectorOpGt:
			op = selection.GreaterThan
		case corev1.NodeSelectorOpLt:
			op = selection.LessThan
		default:
			return nil, fmt.Errorf("%q is not a valid node selector operator", expr.Operator)
		}
		r, err := labels.NewRequirement(expr.Key, op, expr.Values)
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*r)
	}
	return selector, nil
}
