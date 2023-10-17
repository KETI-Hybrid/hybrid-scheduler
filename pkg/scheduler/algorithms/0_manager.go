package algorithms

import (
	"hybrid-scheduler/pkg/util/client"

	keticlient "github.com/KETI-Hybrid/keti-controller/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

type AlgoManager struct {
	kubeClient kubernetes.Interface
	ketiClient keticlient.Interface
}
type Algorithms func(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)

// type Algorithms interface {
// 	DRF(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	NodeRegion(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	OptimizationCount(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	Affinity(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	LocationAffinity(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	NodeName(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	NodePorts(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	NodeUnschedulable(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	JoinCheck(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	VolumeRestrictions(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	ImageLocality(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	OptimizationTime(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	NodePreferAvoidPods(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	SelectSpread(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	NodeResourceFit(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	TaintToleration(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	VolumeBinding(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	ScoreSpread(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// 	BalanceAllocation(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
// }

func InitAlgoManager() *AlgoManager {
	kubeClient, err := client.NewClient()
	if err != nil {
		klog.Fatal(err)
	}
	ketiClient, err := client.NewKETIClient()
	if err != nil {
		klog.Fatal(err)
	}
	a := &AlgoManager{}
	a.kubeClient = kubeClient
	a.ketiClient = ketiClient
	// a.algoMap = make(map[string]Algorithms)
	// a.algoMap["drf"] = a.DRF
	// a.algoMap["noderegion"] = a.NodeRegion
	// a.algoMap["optimizationcount"] = a.OptimizationCount
	// a.algoMap["affinity"] = a.Affinity
	// a.algoMap["locationaffinity"] = a.LocationAffinity
	// a.algoMap["nodename"] = a.NodeName
	// a.algoMap["nodeports"] = a.NodePorts
	// a.algoMap["nodeunschedulable"] = a.NodeUnschedulable
	// a.algoMap["joincheck"] = a.JoinCheck
	// a.algoMap["volumerestrictions"] = a.VolumeRestrictions
	// a.algoMap["imagelocality"] = a.ImageLocality
	// a.algoMap["optimizationtime"] = a.OptimizationTime
	// a.algoMap["nodepreferavoidpods"] = a.NodePreferAvoidPods
	// a.algoMap["selectspread"] = a.SelectSpread
	// a.algoMap["noderesourcefit"] = a.NodeResourceFit
	// a.algoMap["tainttoleration"] = a.TaintToleration
	// a.algoMap["volumebinding"] = a.VolumeBinding
	// a.algoMap["scorespread"] = a.ScoreSpread
	// a.algoMap["balanceallocation"] = a.BalanceAllocation

	return &AlgoManager{}
}

func (a *AlgoManager) Do(algoName string, args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	switch algoName {
	case "drf":
		return a.DRF(args)
	case "noderegion":
		return a.NodeRegion(args)
	case "optimizationcount":
		return a.OptimizationCount(args)
	case "affinity":
		return a.Affinity(args)
	case "locationaffinity":
		return a.LocationAffinity(args)
	case "nodename":
		return a.NodeName(args)
	case "nodeports":
		return a.NodePorts(args)
	case "nodeunschedulable":
		return a.NodeUnschedulable(args)
	case "joincheck":
		return a.JoinCheck(args)
	case "volumerestrictions":
		return a.VolumeRestrictions(args)
	case "imagelocality":
		return a.ImageLocality(args)
	case "optimizationtime":
		return a.OptimizationTime(args)
	case "nodepreferavoidpods":
		return a.NodePreferAvoidPods(args)
	case "selectspread":
		return a.SelectSpread(args)
	case "noderesourcefit":
		return a.NodeResourceFit(args)
	case "tainttoleration":
		return a.TaintToleration(args)
	case "volumebinding":
		return a.VolumeBinding(args)
	case "scorespread":
		return a.ScoreSpread(args)
	case "balanceallocation":
		return a.BalanceAllocation(args)
	}
	return nil, nil
}
