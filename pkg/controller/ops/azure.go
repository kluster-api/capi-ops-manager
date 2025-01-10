/*
Copyright AppsCode Inc. and Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ops

import (
	opsapi "go.klusters.dev/capi-ops-manager/apis/ops/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	clientutil "kmodules.xyz/client-go/client"
	"kmodules.xyz/client-go/conditions"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ClusterOpsRequestReconciler) updateAzureManagedControlPlane(namespacedName types.NamespacedName, clusterOps *opsapi.ClusterOpsRequest) (bool, error) {
	if conditions.IsConditionTrue(clusterOps.GetConditions(), string(opsapi.ControlPlaneUpdateCondition)) {
		return false, nil
	}
	azureManagedCP := &capz.AzureManagedControlPlane{}
	err := r.KBClient.Get(r.ctx, namespacedName, azureManagedCP)
	if err != nil {
		return false, err
	}
	_, err = clientutil.CreateOrPatch(r.ctx, r.KBClient, azureManagedCP, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*capz.AzureManagedControlPlane)
		in.Spec.Version = *clusterOps.Spec.UpdateVersion.TargetVersion.Cluster
		return in
	})
	if err != nil {
		return false, err
	}

	if !r.isAzureManagedControlPlaneReady(azureManagedCP) || !isVersionEqual(azureManagedCP.Status.Version, ptr.Deref(clusterOps.Spec.UpdateVersion.TargetVersion.Cluster, "")) {
		r.Log.Info("Waiting for AzureManagedControlPlane to be ready")
		return true, nil
	}
	r.Log.Info("Successfully updated AzureManagedControlPlane version")
	return false, nil
}

func (r *ClusterOpsRequestReconciler) isAzureManagedControlPlaneReady(azureManagedCP *capz.AzureManagedControlPlane) bool {
	conds := azureManagedCP.GetConditions()
	for _, cond := range conds {
		if cond.Type == capi.ReadyCondition {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}
