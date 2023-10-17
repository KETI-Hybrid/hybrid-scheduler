package algorithms

import (
	"fmt"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) DRF(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeName := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** DRF Algorithm **")
	fmt.Println("Checking node's priority resources")
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

	if cpuResource > 0 {
		if memoryResource > 0 {
			if storageResource > 0 {
				fmt.Printf("%s priority : high-priority\n", args.Pod.Name)
			} else {
				fmt.Printf("%s priority : middle-priority\n", args.Pod.Name)
			}
		} else {
			fmt.Printf("%s priority : low-priority\n", args.Pod.Name)
		}
	} else {
		fmt.Printf("%s priority : low-priority\n", args.Pod.Name)
	}

	for _, node := range args.Nodes.Items {
		if node.Status.Allocatable.Cpu().CmpInt64(cpuResource) >= 0 {
			if node.Status.Allocatable.Memory().CmpInt64(memoryResource) >= 0 {
				if node.Status.Allocatable.Storage().CmpInt64(storageResource) >= 0 {
					fmt.Printf("%s : True\n", node.Name)
					nodeName = append(nodeName, node.Name)
				} else {
					fmt.Printf("%s : False\n", node.Name)
					continue
				}
			} else {
				fmt.Printf("%s : False\n", node.Name)
				continue
			}
		} else {
			fmt.Printf("%s : False\n", node.Name)
			continue
		}

	}

	fmt.Printf("Selected Node(s) : %s\n", nodeName)

	return &extenderv1.ExtenderFilterResult{NodeNames: &nodeName}, nil
}
