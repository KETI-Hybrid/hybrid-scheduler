package priorities

//기존 쿠버네티스의 노드 레벨 RequestedToCapacityRatioPriority을 클러스터 레벨로 확장
import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"time"
)

type RequestedToCapacityRatio struct {
	prescoring   map[string]int64
	betweenScore int64
}
type FunctionShape []FunctionShapePoint

type FunctionShapePoint struct {
	// Utilization is function argument
	Utilization int64
	// Score is function value
	Score int64
}

var (
	// give priority to least utilized nodes by default
	defaultFunctionShape = NewFunctionShape([]FunctionShapePoint{
		{
			Utilization: 0,
			Score:       minScore,
		},
		{
			Utilization: 100,
			Score:       maxScore,
		},
	})
)

const (
	minUtilization = 0
	maxUtilization = 100
)

func (pl *RequestedToCapacityRatio) Name() string {
	return "RequestedToCapacityRatio"
}
func (pl *RequestedToCapacityRatio) PreScore(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, check bool) int64 {
	var clusterScore int64

	for _, node := range clusterInfo.Nodes {
		requested := node.RequestedResource.MilliCPU + pod.RequestedResource.MilliCPU
		nodeScore := RunRequestedToCapacityRatioScorerFunction(node.CapacityResource.MilliCPU, requested)
		requested = node.RequestedResource.Memory + pod.RequestedResource.Memory
		nodeScore += RunRequestedToCapacityRatioScorerFunction(node.CapacityResource.Memory, requested)
		requested = node.RequestedResource.EphemeralStorage + pod.RequestedResource.EphemeralStorage
		nodeScore += RunRequestedToCapacityRatioScorerFunction(node.CapacityResource.EphemeralStorage, requested)

		node.NodeScore = nodeScore * weight * 10
		clusterScore += nodeScore
	}
	if !check {
		if len(pl.prescoring) == 0 {
			pl.prescoring = make(map[string]int64)
		}
		pl.prescoring[clusterInfo.ClusterName] = clusterScore
	} else {
		pl.betweenScore = pl.prescoring[clusterInfo.ClusterName] - int64(clusterScore)
		if pl.betweenScore <= 0 {
			pl.betweenScore = 5
		}
		pl.prescoring[clusterInfo.ClusterName] = clusterScore - pl.betweenScore
		// omcplog.V(0).Infof("2["+clusterInfo.ClusterName+"]노드"+pl.Name()+" 노드차이 =", pl.betweenScore)
	}
	return clusterScore
}

func (pl *RequestedToCapacityRatio) Score(pod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, replicas int32, clustername string) int64 {
	startTime := time.Now()
	if clustername == clusterInfo.ClusterName {
		pl.prescoring[clusterInfo.ClusterName] = pl.prescoring[clusterInfo.ClusterName] - pl.betweenScore
		return pl.prescoring[clusterInfo.ClusterName]
	}
	score := pl.prescoring[clusterInfo.ClusterName]
	omcplog.V(4).Info("RequestedToCapacityRatio score = ", score)
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("RequestedToCapacityRatio Time [%v]", elapsedTime)
	return score
}
func RunRequestedToCapacityRatioScorerFunction(capacity, requested int64) int64 {
	scoringFunctionShape := defaultFunctionShape
	rawScoringFunction := buildBrokenLinearFunction(scoringFunctionShape)
	resourceScoringFunction := func(requested, capacity int64) int64 {
		if capacity == 0 || requested > capacity {
			return rawScoringFunction(maxUtilization)
		}
		return rawScoringFunction(maxUtilization - (capacity-requested)*maxUtilization/capacity)
	}

	return int64(resourceScoringFunction(requested, capacity))
}

func buildBrokenLinearFunction(shape FunctionShape) func(int64) int64 {
	n := len(shape)
	return func(p int64) int64 {
		for i := 0; i < n; i++ {
			if p <= shape[i].Utilization {
				if i == 0 {
					return shape[0].Score
				}
				return shape[i-1].Score + (shape[i].Score-shape[i-1].Score)*(p-shape[i-1].Utilization)/(shape[i].Utilization-shape[i-1].Utilization)
			}
		}
		return shape[n-1].Score
	}
}

func NewFunctionShape(points []FunctionShapePoint) FunctionShape {
	n := len(points)

	if n == 0 {
		omcplog.V(0).Info("at least one point must be specified")
		return nil
	}

	for i := 1; i < n; i++ {
		if points[i-1].Utilization >= points[i].Utilization {
			omcplog.V(0).Infof("utilization values must be sorted. Utilization[%v]==%v >= Utilization[%v]==%v", i-1, points[i-1].Utilization, i, points[i].Utilization)
			return nil
		}
	}

	for i, point := range points {
		if point.Utilization < minUtilization {
			omcplog.V(0).Infof("utilization values must not be less than %v. Utilization[%v]==%v", minUtilization, i, point.Utilization)
			return nil
		}
		if point.Utilization > maxUtilization {
			omcplog.V(0).Infof("utilization values must not be greater than %v. Utilization[%v]==%v", maxUtilization, i, point.Utilization)
			return nil
		}
		if point.Score < minScore {
			omcplog.V(0).Infof("score values must not be less than %v. Score[%v]==%v", minScore, i, point.Score)
			return nil
		}
		if point.Score > maxScore {
			omcplog.V(0).Infof("score valuses not be greater than %v. Score[%v]==%v", maxScore, i, point.Score)
			return nil
		}
	}

	// We make defensive copy so we make no assumption if array passed as argument is not changed afterwards
	pointsCopy := make(FunctionShape, n)
	copy(pointsCopy, points)
	return pointsCopy
}
