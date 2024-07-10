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
	"fmt"

	opsapi "go.klusters.dev/capi-ops-manager/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	v1 "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/conditions"
	capa "sigs.k8s.io/cluster-api-provider-aws/v2/controlplane/eks/api/v1beta2"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

const GCPManagedControlPlaneKind = "GCPManagedControlPlane"

func (r *ClusterOpsRequestReconciler) updateClusterVersion(cluster *capi.Cluster) (bool, error) {
	var reKey bool
	var err error
	reKey, err = r.updateControlPlaneVersion(cluster)
	if err != nil || reKey {
		return reKey, err
	}
	reKey, err = r.updateClusterMachinePoolVersion(cluster.Name)
	if err != nil || reKey {
		return reKey, err
	}
	return false, r.updateComponents()
}

func (r *ClusterOpsRequestReconciler) updateControlPlaneVersion(cluster *capi.Cluster) (bool, error) {
	if !conditions.HasCondition(r.ClusterOps.Status.Conditions, string(opsapi.ControlPlaneUpdateCondition)) {
		r.Log.Info("Started updating control plane version")
		conditions.MarkFalse(r.ClusterOps, opsapi.ControlPlaneUpdateCondition, opsapi.ControlPlaneUpdateStartedReason, v1.ConditionSeverityInfo, "")
		return false, nil
	}
	if conditions.IsConditionTrue(r.ClusterOps.Status.Conditions, string(opsapi.ControlPlaneUpdateCondition)) {
		return false, nil
	}
	var reKey bool
	var err error
	namespacedName := types.NamespacedName{Namespace: cluster.Spec.ControlPlaneRef.Namespace, Name: cluster.Spec.ControlPlaneRef.Name}
	if cluster.Spec.ControlPlaneRef.Kind == capz.AzureManagedControlPlaneKind {
		reKey, err = r.updateAzureManagedControlPlane(namespacedName)
	} else if cluster.Spec.ControlPlaneRef.Kind == GCPManagedControlPlaneKind {
		reKey, err = r.updateGCPManagedControlPlane(namespacedName)
	} else if cluster.Spec.ControlPlaneRef.Kind == capa.AWSManagedControlPlaneKind {
		reKey, err = r.updateAWSManagedControlPlane(namespacedName)
	} else {
		err = fmt.Errorf("unknown Control Plane Kind")
	}
	if err != nil {
		conditions.MarkFalse(r.ClusterOps, opsapi.ControlPlaneUpdateCondition, opsapi.ControlPlaneUpdateFailedReason, v1.ConditionSeverityError, err.Error())
		return false, err
	}
	if reKey {
		return true, nil
	}
	conditions.MarkTrue(r.ClusterOps, opsapi.ControlPlaneUpdateCondition)
	return false, nil
}
