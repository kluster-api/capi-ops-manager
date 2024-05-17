package ops

import (
	"sort"

	opsapi "go.klusters.dev/capi-ops-manager/apis/ops/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	clientutil "kmodules.xyz/client-go/client"
	"kmodules.xyz/client-go/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ClusterOpsRequestReconciler) updateClusterMachinePoolVersion(clusterName string) (bool, error) {
	if !conditions.HasCondition(r.ClusterOps.Status.Conditions, string(opsapi.MachinePoolUpdateCondition)) {
		conditions.MarkFalse(r.ClusterOps, opsapi.MachinePoolUpdateCondition, opsapi.MachinePoolUpdateStartedReason, kmapi.ConditionSeverityInfo, "")
		return false, nil
	}
	machinePools := &capiexp.MachinePoolList{}

	err := r.KBClient.List(r.ctx, machinePools, client.MatchingLabels{
		capi.ClusterNameLabel: clusterName,
	})
	if err != nil {
		return false, err
	}
	sort.Slice(machinePools.Items, func(i, j int) bool {
		return machinePools.Items[i].Name < machinePools.Items[j].Name
	})
	var reKey bool

	for _, mp := range machinePools.Items {
		reKey, err = r.updateMachinePoolVersion(&mp)
		if err != nil {
			conditions.MarkFalse(r.ClusterOps, opsapi.MachinePoolUpdateCondition, opsapi.MachinePoolUpdateFailedReason, kmapi.ConditionSeverityInfo, err.Error())
			return false, err
		}
		if reKey {
			return true, nil
		}
	}
	r.Log.Info("Successfully Updated all MachinePools Version")
	conditions.MarkTrue(r.ClusterOps, opsapi.MachinePoolUpdateCondition)
	return false, nil
}

func (r *ClusterOpsRequestReconciler) updateMachinePoolVersion(mp *capiexp.MachinePool) (bool, error) {
	vt, err := r.patchMachinePoolVersion(mp)
	if err != nil {
		return false, err
	}

	if vt == kutil.VerbPatched || !r.isMachinePoolConditionUpdatedForOps(mp) || capiexp.MachinePoolPhase(mp.Status.Phase) != capiexp.MachinePoolPhaseRunning {
		r.Log.Info("Waiting for MachinePool to be Running", "Name", mp.GetName(), "Namespace", mp.GetNamespace())
		return true, nil
	}
	r.Log.Info("Successfully updated MachinePool version", "Name", mp.GetName(), "Namespace", mp.GetNamespace())
	return false, nil
}

func (r *ClusterOpsRequestReconciler) patchMachinePoolVersion(mp *capiexp.MachinePool) (kutil.VerbType, error) {
	if ptr.Deref(mp.Spec.Template.Spec.Version, "0") == ptr.Deref(r.ClusterOps.Spec.UpdateVersion.TargetVersion, "0") {
		return kutil.VerbUnchanged, nil
	}
	r.Log.Info("Patching MachinePool Version", "Name", mp.GetName(), "Namespace", mp.GetNamespace())
	return clientutil.CreateOrPatch(r.ctx, r.KBClient, mp, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*capiexp.MachinePool)
		in.Spec.Template.Spec.Version = r.ClusterOps.Spec.UpdateVersion.TargetVersion
		return in
	})
}

func (r *ClusterOpsRequestReconciler) isMachinePoolConditionUpdatedForOps(mp *capiexp.MachinePool) bool {
	conds := mp.GetConditions()
	mpLastTransTime := metav1.Time{}
	for _, c := range conds {
		if c.Type == capi.ReadyCondition && c.Status == corev1.ConditionTrue {
			mpLastTransTime = c.LastTransitionTime
		}
	}
	opsLastTransTime := metav1.Time{}
	opsConds := r.ClusterOps.GetConditions()
	for _, c := range opsConds {
		if c.Type == opsapi.MachinePoolUpdateCondition {
			opsLastTransTime = c.LastTransitionTime
		}
	}
	return mpLastTransTime.After(opsLastTransTime.Time)
}
