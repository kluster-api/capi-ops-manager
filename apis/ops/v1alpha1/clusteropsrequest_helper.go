package v1alpha1

import kmapi "kmodules.xyz/client-go/api/v1"

func (c *ClusterOpsRequest) GetStatus() *ClusterOpsRequestStatus {
	return &c.Status
}

func (c *ClusterOpsRequest) GetConditions() kmapi.Conditions {
	return c.Status.Conditions
}

func (c *ClusterOpsRequest) SetConditions(conditions kmapi.Conditions) {
	c.Status.Conditions = conditions
}

func (c *ClusterOpsRequest) GetRequestType() any {
	return c.Spec.Type
}
