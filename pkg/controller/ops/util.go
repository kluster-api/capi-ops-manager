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
	"context"
	"time"

	opsapi "go.klusters.dev/capi-ops-manager/apis/ops/v1alpha1"

	cutil "kmodules.xyz/client-go/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const retryInterval = 1 * time.Minute

func (r *ClusterOpsRequestReconciler) updateClusterOpsRequestReconcile(ctx context.Context, namespacedName client.ObjectKey) (string, error) {
	clusterOps := &opsapi.ClusterOpsRequest{}
	if err := r.KBClient.Get(ctx, namespacedName, clusterOps); err != nil {
		return "Failed to get ClusterOps", err
	}
	r.ClusterOps = clusterOps
	r.ctx = ctx

	return "", nil
}

// reconciled returns an empty result with nil error to signal a successful reconcile
// to the controller manager
func (r *ClusterOpsRequestReconciler) reconciled() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// requeueWithError is a wrapper around logging an error message
// then passes the error through to the controller manager
func (r *ClusterOpsRequestReconciler) requeueWithError(msg string, err error) (ctrl.Result, error) {
	// Info Log the error message and then let the reconciler dump the stacktrace
	r.Log.Info(msg, "Reason : ", err.Error())
	return ctrl.Result{}, err
}

func (r *ClusterOpsRequestReconciler) updateClusterOpsRequestStatus(namespacedName client.ObjectKey) error {
	clusterOps := &opsapi.ClusterOpsRequest{}
	if err := r.KBClient.Get(r.ctx, namespacedName, clusterOps); err != nil {
		return err
	}
	cutil.SetSummary(r.ClusterOps, cutil.WithConditions(opsapi.ConditionsOrder()...))
	r.ClusterOps.Status.Phase = opsapi.GetPhase(r.ClusterOps)

	if err := r.committer(r.ctx, clusterOps, r.ClusterOps); err != nil {
		return err
	}
	return nil
}

func isVersionEqual(v1 string, v2 string) bool {
	if v1 == v2 || "v"+v1 == v2 || v1 == "v"+v2 {
		return true
	}
	return false
}
