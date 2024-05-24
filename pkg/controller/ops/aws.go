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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientutil "kmodules.xyz/client-go/client"
	"kmodules.xyz/client-go/conditions"
	capa "sigs.k8s.io/cluster-api-provider-aws/v2/controlplane/eks/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ClusterOpsRequestReconciler) updateAWSManagedControlPlane(namespacedName types.NamespacedName) (bool, error) {
	if conditions.IsConditionTrue(r.ClusterOps.GetConditions(), string(opsapi.ControlPlaneUpdateCondition)) {
		return false, nil
	}
	awsManagedCP := &capa.AWSManagedControlPlane{}
	err := r.KBClient.Get(r.ctx, namespacedName, awsManagedCP)
	if err != nil {
		return false, err
	}
	_, err = clientutil.CreateOrPatch(r.ctx, r.KBClient, awsManagedCP, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*capa.AWSManagedControlPlane)
		in.Spec.Version = r.ClusterOps.Spec.UpdateVersion.TargetVersion
		return in
	})
	if err != nil {
		return false, err
	}

	if !r.isAWSManagedControlPlaneReady(awsManagedCP) {
		r.Log.Info("Waiting for AWSManagedControlPlane to be ready")
		return true, nil
	}
	r.Log.Info("Successfully updated AWSManagedControlPlane version")
	return false, nil
}

func (r *ClusterOpsRequestReconciler) isAWSManagedControlPlaneReady(awsManagedCP *capa.AWSManagedControlPlane) bool {
	conds := awsManagedCP.GetConditions()
	cpLastTransTime := metav1.Time{}
	for _, c := range conds {
		if c.Type == capa.EKSControlPlaneUpdatingCondition && c.Status == corev1.ConditionFalse {
			cpLastTransTime = c.LastTransitionTime
		}
	}
	opsLastTransTime := metav1.Time{}
	opsConds := r.ClusterOps.GetConditions()
	for _, c := range opsConds {
		if c.Type == opsapi.ControlPlaneUpdateCondition {
			opsLastTransTime = c.LastTransitionTime
		}
	}
	return cpLastTransTime.After(opsLastTransTime.Time)
}
