package algorithms

import (
	"context"
	"fmt"
	"hybrid-scheduler/pkg/util/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) VolumeRestrictions(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** VolumeRestrictions **")
	fmt.Println("Checking provider-specific restrictions...")
	storageResource := int64(0)
	var err error
	if a.kubeClient == nil {
		a.kubeClient, err = client.NewClient()
		if err != nil {
			klog.Errorln(err)
		}
	}
	if len(args.Pod.Spec.Volumes) > 0 {
		for _, volume := range args.Pod.Spec.Volumes {
			if volume.PersistentVolumeClaim == nil {
				continue
			}
			volClaim, err := a.kubeClient.CoreV1().PersistentVolumeClaims(args.Pod.Namespace).Get(context.Background(), volume.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
			if err != nil {
				klog.Errorln(err)
			}
			containerStorageResource, _ := volClaim.Spec.Resources.Requests.Storage().AsInt64()
			storageResource += containerStorageResource
		}
	}

	if storageResource > 0 {
		fmt.Printf("%s volumes : %d\n", args.Pod.Name, storageResource)
	} else {
		fmt.Printf("%s volumes : None\n", args.Pod.Name)
	}

	for _, node := range args.Nodes.Items {
		if node.Status.Allocatable.Storage().CmpInt64(storageResource) >= 0 {
			nodeName = append(nodeName, node.Name)
		} else {
			continue
		}

	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
