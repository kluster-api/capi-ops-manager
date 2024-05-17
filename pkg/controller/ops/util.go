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
