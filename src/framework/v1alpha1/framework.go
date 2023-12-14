package v1alpha1

import (
	// "openmcp/openmcp/omcplog"

	"container/list"
	"openmcp/openmcp/omcplog"
	"openmcp/openmcp/openmcp-analytic-engine/src/protobuf"
	"openmcp/openmcp/openmcp-scheduler/src/framework/plugins/predicates"
	"openmcp/openmcp/openmcp-scheduler/src/framework/plugins/priorities"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
)

const (
	defMaxNice    int = 100
	defMinNice    int = 0
	defniceweight int = 10
)

type openmcpFramework struct {
	filterPlugins     []OpenmcpFilterPlugin
	scorePlugins      []OpenmcpScorePlugin
	prefilterPlugins  []OpenmcpPreFilterPlugin
	postfilterPlugins []OpenmcpPostFilterPlugin
	erasePlugins      []OpenmcpEraseScorePlugin
	IspreScore        bool
	preScore          int64
	betweenScores     int64
	preselectedName   string
	preClusterName    string
	Clusters_nice     map[string]OpenmcpNiceScore
	NiceScore         int
	NiceMax           int
	NiceMin           int
}

func (f *openmcpFramework) Set_NiceVaule(minScore int64, maxScore int64, selectcluster string) {

}

// The appearance of the blank identifier in this construct indicates
// that the declaration exists only for the type checking, not to create a variable.
var _ OpenmcpFramework = &openmcpFramework{}

func (f *openmcpFramework) EndPod() {
	f.IspreScore = false
	f.preScore = 0
	f.betweenScores = 0
	f.preselectedName = ""
	f.preClusterName = ""
}
func NewFramework(grpcClient protobuf.RequestAnalysisClient) OpenmcpFramework {

	f := &openmcpFramework{
		filterPlugins: []OpenmcpFilterPlugin{
			&predicates.MatchClusterSelector{},
			&predicates.PodFitsResources{},
			&predicates.CheckNeededResources{},
			&predicates.MatchClusterAffinity{},
			&predicates.PodFitsHostPorts{},
			&predicates.NoDiskConflict{},
			// &predicates.ClusterJoninCheck{},
			&predicates.DRF{},
			&predicates.Selectspread{},
			&predicates.Nodeunschedulable{},
			&predicates.Tainttoleration{},
		},
		scorePlugins: []OpenmcpScorePlugin{
			&priorities.MostRequested{},
			&priorities.DominantResource{},
			&priorities.RequestedToCapacityRatio{},
			&priorities.BalancedNetworkAllocation{},
			&priorities.QosPriority{},
			&priorities.Optimizationcount{},
			&priorities.NodePreferAvoidPods{},
			&priorities.Locationaffinity{},
		},
		prefilterPlugins: []OpenmcpPreFilterPlugin{
			&predicates.PodFitsResources{},
			&predicates.MatchClusterAffinity{},
		},
		postfilterPlugins: []OpenmcpPostFilterPlugin{
			&predicates.PodFitsResources{},
		},
		erasePlugins: []OpenmcpEraseScorePlugin{
			&priorities.DominantResource{},
		},
	}
	return f
}

func (f *openmcpFramework) RunPostFilterPluginsOnClusters(pod *ketiresource.Pod, clusters map[string]*ketiresource.Cluster, postdeployments *list.List) OpenmcpClusterPostFilteredStatus {
	result := make(map[string]bool)

	result["unscheduable"] = false
	result["error"] = false
	result["success"] = false
	sucess := false
	var err error
	for _, cluster := range clusters {
		cluster.PreFilterTwoStep = false
		cluster.PreFilter = false
		result[cluster.ClusterName] = true
		for _, pl := range f.postfilterPlugins {
			sucess, err = pl.PostFilter(pod, cluster, postdeployments)
			if sucess || err == nil {
				result["success"] = true
				return result
			}
		}
	}
	return result
}
func (f *openmcpFramework) EraseFilterPluginsOnClusters(pod *ketiresource.Pod, clusters map[string]*ketiresource.Cluster, requestclusters map[string]int32) string {
	preresult := make(map[string]OpenmcpPluginScoreList)
	for _, cluster := range clusters {
		for r_name, count := range requestclusters {
			if r_name == cluster.ClusterName && count > 0 {

				preresult[cluster.ClusterName] = make([]OpenmcpPluginScore, 0)
				for _, pl := range f.erasePlugins {
					scoring := pl.PreScore(pod, cluster, false)
					transScore := OpenmcpPluginScore{
						Name:  pl.Name(),
						Score: scoring,
					}
					preresult[cluster.ClusterName] = append(preresult[cluster.ClusterName], transScore)
				}
			}
		}

	}
	//omcplog.V(4).Infof("before eraserCluster =, %v", preresult)
	pr := eraserCluster(preresult)
	//omcplog.V(4).Infof("eraserCluster =, %v", pr)
	return pr
}

// func PrintFilterResult(datas map[string]bool) {
// 	for clustername, scores := range datas {
// 		omcplog.V(2).Infof("[%v] Filters {", clustername
// 			omcplog.V(2).Infof("    %v = %v", scores[i].Name, scores[i].Score)
// 	}
// 	omcplog.V(2).Info("}")

// }

func PrintFilterResult(datas map[string]OpenmcpPluginFilterList) {
	for clustername, filters := range datas {
		omcplog.V(2).Infof("[%v] Filters {", clustername)
		for i := 0; i < len(filters); i++ {
			omcplog.V(2).Infof("    %v = %v", filters[i].Name, filters[i].Filter)
		}
		omcplog.V(2).Info("}")
	}
}
func (f *openmcpFramework) RunFilterPluginsOnClusters(pod *ketiresource.Pod, clusters map[string]*ketiresource.Cluster, cm *clusterManager.ClusterManager) OpenmcpClusterFilteredStatus {
	Filters := make(map[string]OpenmcpPluginFilterList)
	result := make(map[string]bool)
	if clusters == nil {
		omcplog.V(3).Infof("clusters NILL")
		return nil
	}
	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}
		cluster.PreFilterTwoStep = false
		cluster.PreFilter = false
		result[cluster.ClusterName] = true
		for _, pl := range f.prefilterPlugins {
			pl.PreFilter(pod, cluster)

		}
		if cluster.PreFilter == false || cluster.PreFilterTwoStep == false {
			result[cluster.ClusterName] = false
			continue
		}
		for _, pl := range f.filterPlugins {
			isFiltered := pl.Filter(pod, cluster, cm)
			result[cluster.ClusterName] = result[cluster.ClusterName] && isFiltered
			filterindex := OpenmcpPluginFilter{
				Name:   pl.Name(),
				Filter: isFiltered,
			}
			Filters[cluster.ClusterName] = append(Filters[cluster.ClusterName], filterindex)
			if !result[cluster.ClusterName] {
				break
			}
		}
	}
	PrintFilterResult(Filters)
	//omcplog.V(0).Info("Filter Info=>", result)
	return result
}
func eraserCluster(scoreResult OpenmcpPluginToClusterScores) string {
	var selectedCluster string
	var minScore int64
	minScore = 1000
	for clusterName, scoreList := range scoreResult {
		var clusterScore int64
		for _, score := range scoreList {
			clusterScore += score.Score
		}

		if clusterScore < minScore {
			selectedCluster = clusterName
			minScore = clusterScore
		}
	}
	//omcplog.V(0).Info("selected clustet ==", selectedCluster)
	return selectedCluster
}
func (f *openmcpFramework) selectCluster(scoreResult OpenmcpPluginToClusterScores) string {
	var selectedCluster string
	var maxScore int64
	for clusterName, scoreList := range scoreResult {
		var clusterScore int64
		if selectedCluster == "" {
			selectedCluster = clusterName
		}
		for _, score := range scoreList {
			clusterScore += score.Score
		}

		if clusterScore > maxScore {
			selectedCluster = clusterName
			maxScore = clusterScore
			f.NiceMax = int(maxScore)
			if maxScore == 0 {
				f.NiceMin = 1
			} else {
				f.NiceMin = int(maxScore / int64(defniceweight))
			}
			if f.NiceMin == 0 {
				f.NiceMin = 1
			}
		}
	}
	//Nice값 계산
	for cluster, _ := range scoreResult {
		temp := f.Clusters_nice[cluster]
		temp.NiceValue = temp.CluersterScore / defniceweight
		if selectedCluster == cluster {
			omcplog.V(2).Infof("[%v] NiceScore Update %v(-%v)", cluster, temp.NiceScore, f.NiceMin)
			temp.NiceScore = temp.NiceScore - f.NiceMin
			if temp.NiceScore < 0 {
				temp.NiceScore = 0
			}
			omcplog.V(2)
		} else {
			omcplog.V(2).Infof("[%v] NiceScore Update %v(+%v)", cluster, temp.NiceScore, temp.NiceValue)
			temp.NiceScore = temp.NiceScore + temp.NiceValue
		}
		f.Clusters_nice[cluster] = temp

	}
	//omcplog.V(0).Info("selected clustet ==", selectedCluster)
	return selectedCluster
}
func PrintScoreResult(datas map[string]OpenmcpPluginScoreList) {
	for clustername, scores := range datas {
		omcplog.V(2).Infof("[%v] Scores {", clustername)
		for i := 0; i < len(scores); i++ {
			omcplog.V(2).Infof("    %v = %v", scores[i].Name, scores[i].Score)
		}
		omcplog.V(2).Info("}")

	}

}

/*
**brief Nice값을 Score에 추가해주는 함수
 */
func (f *openmcpFramework) NiceScoreCul(Scorelist *map[string]OpenmcpPluginScoreList, clstername string) {

	transScore := OpenmcpPluginScore{
		Name:  "Nice",
		Score: int64(f.Clusters_nice[clstername].NiceScore),
	}
	(*Scorelist)[clstername] = append((*Scorelist)[clstername], transScore)
}

// func (f *openmcpFramework) RunScorePluginsOnClusters(pod *ketiresource.Pod, clusters map[string]*ketiresource.Cluster, replicas int32) OpenmcpPluginToClusterScores {
func (f *openmcpFramework) RunScorePluginsOnClusters(pod *ketiresource.Pod, clusters map[string]*ketiresource.Cluster, allclusters map[string]*ketiresource.Cluster, replicas int32) string {
	// nice 값 계산전 기전에 클러스터가 있는지 없는지 다시 확인
	// 나이스 추가
	if f.NiceMax == 0 {
		f.NiceMax = defMaxNice
		f.NiceMin = defMinNice
	}
	omcplog.V(2).Infof("kcp test start")
	if f.Clusters_nice == nil {
		f.Clusters_nice = make(map[string]OpenmcpNiceScore)
	}
	for clustername, _ := range clusters {
		_, exist := f.Clusters_nice[clustername]
		if exist {
			if f.Clusters_nice[clustername].NiceScore > f.NiceMax {
				str_nice := f.Clusters_nice[clustername]
				str_nice.NiceScore = 5
				f.Clusters_nice[clustername] = str_nice
			}
			continue
		} else {
			str_nice := OpenmcpNiceScore{}
			str_nice.NiceScore = f.NiceMax / 4
			f.Clusters_nice[clustername] = str_nice
		}
	}
	if !f.IspreScore {
		f.preScore = 0
		preresult := make(map[string]OpenmcpPluginScoreList)
		for _, cluster := range clusters {
			perClusterScore := 0
			preresult[cluster.ClusterName] = make([]OpenmcpPluginScore, 0)
			for _, pl := range f.scorePlugins {
				scoring := pl.PreScore(pod, cluster, false)
				transScore := OpenmcpPluginScore{
					Name:  pl.Name(),
					Score: scoring,
				}
				f.preScore += scoring
				perClusterScore += int(scoring)
				preresult[cluster.ClusterName] = append(preresult[cluster.ClusterName], transScore)
			}
			// Nice 값 추가 //prescore는 한번만 수행하기 때문에(디플로이먼트당)이때 계산
			temp := f.Clusters_nice[cluster.ClusterName]
			temp.CluersterScore = perClusterScore
			f.Clusters_nice[cluster.ClusterName] = temp
			perClusterScore = 0
			f.NiceScoreCul(&preresult, cluster.ClusterName)
		}

		f.IspreScore = true
		PrintScoreResult(preresult)
		pr := f.selectCluster(preresult)
		f.preselectedName = pr
		f.preClusterName = pr

		return pr
	}
	if f.IspreScore && f.preselectedName != "" {
		for _, pl := range f.scorePlugins {
			pl.PreScore(pod, allclusters[f.preselectedName], true)
		}
		f.preselectedName = ""
	}
	result := make(map[string]OpenmcpPluginScoreList)
	for _, cluster := range clusters {

		result[cluster.ClusterName] = make([]OpenmcpPluginScore, 0)

		for _, pl := range f.scorePlugins {

			plScore := OpenmcpPluginScore{
				Name:  pl.Name(),
				Score: pl.Score(pod, cluster, replicas, f.preClusterName),
			}
			// Update the result of this cluster
			result[cluster.ClusterName] = append(result[cluster.ClusterName], plScore)
		}
		f.NiceScoreCul(&result, cluster.ClusterName)
	}
	PrintScoreResult(result)
	pr := f.selectCluster(result)
	omcplog.V(5).Info("RunScorePluginsOnClusters pr ", pr)
	//	omcplog.V(0).Info("Score Info=>", result)
	f.preClusterName = pr
	return pr
}
func (f *openmcpFramework) HasFilterPlugins() bool {
	return len(f.filterPlugins) > 0
}

func (f *openmcpFramework) HasScorePlugins() bool {
	return len(f.scorePlugins) > 0
}
