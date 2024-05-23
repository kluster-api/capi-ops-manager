package ops

import (
	opsapi "go.klusters.dev/capi-ops-manager/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	v1 "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/conditions"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

func (r *ClusterOpsRequestReconciler) updateVersion(cluster *capi.Cluster) (bool, error) {
	var err error
	if !conditions.HasCondition(r.ClusterOps.Status.Conditions, string(opsapi.ControlPlaneUpdateCondition)) {
		r.Log.Info("Started updating control plane version")
		conditions.MarkFalse(r.ClusterOps, opsapi.ControlPlaneUpdateCondition, opsapi.ControlPlaneUpdateStartedReason, v1.ConditionSeverityInfo, "")
		return false, nil
	}
	var reKey bool
	if cluster.Spec.ControlPlaneRef.Kind == capz.AzureManagedControlPlaneKind {
		namespacedName := types.NamespacedName{Namespace: cluster.Spec.ControlPlaneRef.Namespace, Name: cluster.Spec.ControlPlaneRef.Name}
		reKey, err = r.updateAzureManagedControlPlane(namespacedName)
		if err != nil {
			conditions.MarkFalse(r.ClusterOps, opsapi.ControlPlaneUpdateCondition, opsapi.ControlPlaneUpdateFailedReason, v1.ConditionSeverityError, err.Error())
			return false, err
		}
	} else if cluster.Spec.ControlPlaneRef.Kind == "GCPManagedControlPlane" {
		namespacedName := types.NamespacedName{Namespace: cluster.Spec.ControlPlaneRef.Namespace, Name: cluster.Spec.ControlPlaneRef.Name}
		reKey, err = r.updateGCPManagedControlPlane(namespacedName)
		if err != nil {
			conditions.MarkFalse(r.ClusterOps, opsapi.ControlPlaneUpdateCondition, opsapi.ControlPlaneUpdateFailedReason, v1.ConditionSeverityError, err.Error())
			return false, err
		}
	}
	if reKey {
		return true, nil
	}
	conditions.MarkTrue(r.ClusterOps, opsapi.ControlPlaneUpdateCondition)

	reKey, err = r.updateClusterMachinePoolVersion(cluster.Name)
	if err != nil {
		return false, err
	}
	return reKey, nil
}
