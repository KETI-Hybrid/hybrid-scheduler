package predicates

import (
	"openmcp/openmcp/omcplog"
	ketiresource "openmcp/openmcp/openmcp-scheduler/src/resourceinfo"
	"openmcp/openmcp/util/clusterManager"
	"time"

	v1 "k8s.io/api/core/v1"
)

type NoDiskConflict struct {
}

func (pl *NoDiskConflict) Name() string {
	return "volumerestrictions"
}

func (pl *NoDiskConflict) Filter(newPod *ketiresource.Pod, clusterInfo *ketiresource.Cluster, cm *clusterManager.ClusterManager) bool {
	startTime := time.Now()
	// check all nodes in this cluster
	for _, node := range clusterInfo.Nodes {
		// if node.PreFilter == false || node.PreFilterTwoStep == false {
		// 	omcplog.V(0).Infof("preFilter True", pl.Name(), node.PreFilter)
		// 	continue
		// }
		node_result := true

		for _, v := range newPod.Pod.Spec.Volumes {

			for _, ev := range node.Pods {
				if isVolumeConflict(v, ev.Pod) {
					node_result = false
					break
				}
			}
			if !node_result {
				break
			}
		}

		if node_result {
			// omcplog.V(4).Info("volumerestrictions true ")
			elapsedTime := time.Since(startTime)
			omcplog.V(3).Infof("volumerestrictions Time [%v]", elapsedTime)
			return true
		}
	}
	elapsedTime := time.Since(startTime)
	omcplog.V(3).Infof("volumerestrictions Time [%v]", elapsedTime)
	omcplog.V(4).Info("volumerestrictions false ")
	return false
}

// If you want to detail, check "kubernetes/pkg/scheduler/algorithm/predicates/predicates.go"
func isVolumeConflict(volume v1.Volume, pod *v1.Pod) bool {
	// fast path if there is no conflict checking targets.
	if volume.GCEPersistentDisk == nil && volume.AWSElasticBlockStore == nil && volume.RBD == nil && volume.ISCSI == nil {
		return false
	}

	for _, existingVolume := range pod.Spec.Volumes {
		// case 1) GCEPersistentDisk
		if volume.GCEPersistentDisk != nil && existingVolume.GCEPersistentDisk != nil {
			disk, existingDisk := volume.GCEPersistentDisk, existingVolume.GCEPersistentDisk
			if disk.PDName == existingDisk.PDName && !(disk.ReadOnly && existingDisk.ReadOnly) {
				return true
			}
		}

		// case 2) AWSElasticBlockStore
		if volume.AWSElasticBlockStore != nil && existingVolume.AWSElasticBlockStore != nil {
			if volume.AWSElasticBlockStore.VolumeID == existingVolume.AWSElasticBlockStore.VolumeID {
				return true
			}
		}

		// case 3) ISCSI
		if volume.ISCSI != nil && existingVolume.ISCSI != nil {
			iqn := volume.ISCSI.IQN
			eiqn := existingVolume.ISCSI.IQN
			if iqn == eiqn && !(volume.ISCSI.ReadOnly && existingVolume.ISCSI.ReadOnly) {
				return true
			}
		}

		// case 4) RBD
		if volume.RBD != nil && existingVolume.RBD != nil {
			mon, pool, image := volume.RBD.CephMonitors, volume.RBD.RBDPool, volume.RBD.RBDImage
			emon, epool, eimage := existingVolume.RBD.CephMonitors, existingVolume.RBD.RBDPool, existingVolume.RBD.RBDImage
			if haveOverlap(mon, emon) && pool == epool && image == eimage && !(volume.RBD.ReadOnly && existingVolume.RBD.ReadOnly) {
				return true
			}
		}
	}

	return false
}

// If you want to detail, check " kubernetes/pkg/scheduler/algorithm/predicates/predicates.go"
// haveOverlap searches two arrays and returns true
// if they have at least one common element; returns false otherwise.
func haveOverlap(a1, a2 []string) bool {
	if len(a1) > len(a2) {
		a1, a2 = a2, a1
	}
	m := map[string]bool{}

	for _, val := range a1 {
		m[val] = true
	}
	for _, val := range a2 {
		if _, ok := m[val]; ok {
			return true
		}
	}

	return false
}
