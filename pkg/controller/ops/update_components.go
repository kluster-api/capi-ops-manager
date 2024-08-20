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
	"os"

	opsapi "go.klusters.dev/capi-ops-manager/apis/ops/v1alpha1"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/conditions"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
)

func (r *ClusterOpsRequestReconciler) updateComponents() error {
	if r.ClusterOps.Spec.UpdateVersion.TargetVersion.Providers == nil {
		return nil
	}
	if !conditions.HasCondition(r.ClusterOps.Status.Conditions, string(opsapi.CapiProvidersUpdateCondition)) {
		r.Log.Info("Started updating Capi Providers version")
		conditions.MarkFalse(r.ClusterOps, opsapi.CapiProvidersUpdateCondition, opsapi.CapiProvidersUpdateStartedReason, kmapi.ConditionSeverityInfo, "")
		return nil
	}
	err := os.Setenv("AWS_B64ENCODED_CREDENTIALS", "")
	if err != nil {
		return err
	}
	err = os.Setenv("GCP_B64ENCODED_CREDENTIALS", "")
	if err != nil {
		return err
	}
	client, err := clusterctl.New(r.ctx, "")
	if err != nil {
		r.Log.Info("Failed to get clusterctl client")
		conditions.MarkFalse(r.ClusterOps, opsapi.CapiProvidersUpdateCondition, opsapi.CapiProvidersUpdateFailedReason, kmapi.ConditionSeverityInfo, "%s", err.Error())
		return err
	}

	err = client.ApplyUpgrade(r.ctx, clusterctl.ApplyUpgradeOptions{
		CoreProvider:            r.ClusterOps.Spec.UpdateVersion.TargetVersion.Providers.Core,
		BootstrapProviders:      []string{r.ClusterOps.Spec.UpdateVersion.TargetVersion.Providers.Bootstrap},
		ControlPlaneProviders:   []string{r.ClusterOps.Spec.UpdateVersion.TargetVersion.Providers.ControlPlane},
		InfrastructureProviders: []string{r.ClusterOps.Spec.UpdateVersion.TargetVersion.Providers.Infrastructure},
	})
	if err != nil {
		r.Log.Info("Failed to updated Capi provider versions")
		conditions.MarkFalse(r.ClusterOps, opsapi.CapiProvidersUpdateCondition, opsapi.CapiProvidersUpdateFailedReason, kmapi.ConditionSeverityInfo, "%s", err.Error())
		return err
	}
	r.Log.Info("Successfully Updated Capi Provider Version")
	conditions.MarkTrue(r.ClusterOps, opsapi.CapiProvidersUpdateCondition)
	return nil
}
