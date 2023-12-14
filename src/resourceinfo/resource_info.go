package resourceinfo

import (
	"openmcp/openmcp/apis/cluster/v1alpha1"
	resourcev1alpha1 "openmcp/openmcp/apis/resource/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

var (
	emptyResource = Resource{}
)

type RequestScheduler struct {
	Replicas int32    `json:"replicas" protobuf:"varint,1,opt,name=replicas"`
	Clusters []string `json:"clusters,omitempty" protobuf:"bytes,11,opt,name=clusters"`
}

// cluster Level
type Cluster struct {
	ClusterName         string
	Nodes               []*NodeInfo
	RequestedResource   *Resource
	AllocatableResource *Resource
	ClusterList         *v1alpha1.OpenMCPClusterList
	PreFilter           bool
	PreFilterTwoStep    bool
}

// NodeInfo is node level aggregated information.
type NodeInfo struct {
	// Overall node information.
	ClusterName string
	NodeName    string

	Node *corev1.Node
	Pods []*Pod

	// Capacity
	CapacityResource *Resource
	// Total requested resource of all pods on this node
	RequestedResource *Resource
	// Total allocatable resource of all pods on this node
	AllocatableResource *Resource
	// Additional resource like nvidia/gpu
	AdditionalResource []string
	// Affinity(Region/Zone)
	Affinity map[string]string
	// Score to Update Resourcese
	NodeScore int64
	UpdateTX  int64
	UpdateRX  int64
	//if PreFilter is true, return Nodeis false

}

type Pod struct {
	// Overall pod informtation.
	ClusterName string
	NodeName    string
	PodName     string

	Pod                *corev1.Pod
	RequestedResource  *Resource
	AdditionalResource []string
	Affinity           map[string][]string
}

type Resource struct {
	MilliCPU         int64
	Memory           int64
	EphemeralStorage int64
}

func NewResource() *Resource {
	return &Resource{
		MilliCPU:         0,
		Memory:           0,
		EphemeralStorage: 0,
	}
}

func AddResources(res, new *Resource) *Resource {

	return &Resource{
		MilliCPU:         res.MilliCPU + new.MilliCPU,
		Memory:           res.Memory + new.Memory,
		EphemeralStorage: res.EphemeralStorage + new.EphemeralStorage,
	}
}

func GetAllocatable(capacity, request *Resource) *Resource {
	return &Resource{
		MilliCPU:         capacity.MilliCPU - request.MilliCPU,
		Memory:           capacity.Memory - request.Memory,
		EphemeralStorage: capacity.EphemeralStorage - request.EphemeralStorage,
	}
}

// Remaining replica
type PostDelployment struct {
	Fcnt          int64
	RemainReplica int32
	NewPod        *Pod
	NewDeployment *resourcev1alpha1.OpenMCPDeployment
	Replica       int32
}
