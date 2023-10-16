package algorithms

import (
	"fmt"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) NodeResourceFit(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** NodeResourceFit Algorithm **")
	cpuResource := int64(0)
	memoryResource := int64(0)
	storageResource := int64(0)
	for _, container := range args.Pod.Spec.Containers {
		containerResource := container.Resources.Requests
		containerCPUResource, _ := containerResource.Cpu().AsInt64()
		containerMemoryResource, _ := containerResource.Memory().AsInt64()
		containerStorageResource, _ := containerResource.Storage().AsInt64()
		cpuResource += containerCPUResource
		memoryResource += containerMemoryResource
		storageResource += containerStorageResource
	}

	for _, node := range args.Nodes.Items {
		if node.Status.Allocatable.Cpu().CmpInt64(cpuResource) >= 0 {
			if node.Status.Allocatable.Memory().CmpInt64(memoryResource) >= 0 {
				if node.Status.Allocatable.Storage().CmpInt64(storageResource) >= 0 {
					nodeName = append(nodeName, node.Name)
				} else {
					continue
				}
			} else {
				continue
			}
		} else {
			continue
		}

	}
	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
