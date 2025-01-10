/*
Copyright 2024.

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

	opsapi "go.klusters.dev/capi-ops-manager/apis/ops/v1alpha1"

	"github.com/go-logr/logr"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"kmodules.xyz/client-go/conditions/committer"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClusterOpsRequestReconciler reconciles a ClusterOpsRequest object
type ClusterOpsRequestReconciler struct {
	ctx       context.Context
	committer func(ctx context.Context, old, obj committer.StatusGetter[*opsapi.ClusterOpsRequestStatus]) error
	Log       logr.Logger
	KBClient  client.Client
	Scheme    *runtime.Scheme
}

//+kubebuilder:rbac:groups=ops.klusters.dev,resources=clusteropsrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ops.klusters.dev,resources=clusteropsrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ops.klusters.dev,resources=clusteropsrequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ClusterOpsRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = log.FromContext(ctx)
	r.ctx = ctx
	r.committer = committer.NewStatusCommitter[*opsapi.ClusterOpsRequest, *opsapi.ClusterOpsRequestStatus](r.KBClient.Status())

	r.Log.Info("Started reconciling ClusterOpsRequest")

	clusterOps := &opsapi.ClusterOpsRequest{}
	err := r.KBClient.Get(ctx, req.NamespacedName, clusterOps)
	if err != nil {
		if kerr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return r.requeueWithError("Failed to get ClusterOps", err)
	}

	if clusterOps.Status.Phase == opsapi.ClusterOpsRequestPhaseSuccessful || clusterOps.Status.Phase == opsapi.ClusterOpsRequestPhaseFailed || clusterOps.Status.Phase == opsapi.ClusterOpsRequestPhaseSkipped {
		return r.reconciled()
	}
	if clusterOps.Status.Phase == "" {
		return ctrl.Result{}, r.updateClusterOpsRequestStatus(req.NamespacedName, clusterOps)
	}

	cluster := &capi.Cluster{}
	err = r.KBClient.Get(ctx, types.NamespacedName{Name: clusterOps.Spec.ClusterRef.Name, Namespace: clusterOps.Spec.ClusterRef.Namespace}, cluster)
	if err != nil {
		return r.requeueWithError("failed to get cluster", err)
	}

	if clusterOps.Status.Phase != opsapi.ClusterOpsRequestPhaseInProgress {
		if capi.ClusterPhase(cluster.Status.Phase) != capi.ClusterPhaseProvisioned {
			return ctrl.Result{
				RequeueAfter: retryInterval,
			}, nil
		}
	}
	var reKey bool
	if clusterOps.GetRequestType().(opsapi.ClusterOpsRequestType) == opsapi.ClusterOpsRequestTypeUpdateVersion {
		reKey, err = r.updateClusterVersion(cluster, clusterOps)
		if err != nil {
			return r.requeueWithError("failed to update version", err)
		}
	}

	reconciledResult := ctrl.Result{}
	if reKey {
		reconciledResult.RequeueAfter = retryInterval
	}

	return reconciledResult, r.updateClusterOpsRequestStatus(req.NamespacedName, clusterOps)
}
