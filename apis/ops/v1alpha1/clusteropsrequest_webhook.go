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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var opslog = logf.Log.WithName("ops-manager-resource")

func (c *ClusterOpsRequest) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-ops-klusters-dev-v1alpha1-clusteropsrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=ops.klusters.dev,resources=clusteropsrequests,verbs=create;update,versions=v1alpha1,name=mclusteropsrequest.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterOpsRequest{}

func (c *ClusterOpsRequest) Default() {
}

//+kubebuilder:webhook:path=/validate-ops-klusters-dev-v1alpha1-clusteropsrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=ops.klusters.dev,resources=clusteropsrequests,verbs=create;update,versions=v1alpha1,name=vclusteropsrequest.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterOpsRequest{}

func (c *ClusterOpsRequest) ValidateCreate() (admission.Warnings, error) {
	opslog.Info("validate create", "name", c.Name)
	return c.ValidateCreateOrUpdate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (c *ClusterOpsRequest) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	opslog.Info("validate update", "name", c.Name)
	return c.ValidateCreateOrUpdate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (c *ClusterOpsRequest) ValidateDelete() (admission.Warnings, error) {
	opslog.Info("validate delete", "name", c.Name)

	return nil, nil
}

func (c *ClusterOpsRequest) ValidateCreateOrUpdate() (admission.Warnings, error) {
	var allErr field.ErrorList

	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kluster.dev", Kind: "ClusterOpsRequest"}, c.Name, allErr)
}
