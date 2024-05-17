package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

type ClusterOpsRequestPhase string

const (
	ClusterOpsRequestPhasePending     ClusterOpsRequestPhase = "Pending"
	ClusterOpsRequestPhaseInProgress  ClusterOpsRequestPhase = "InProgress"
	ClusterOpsRequestPhaseSuccessful  ClusterOpsRequestPhase = "Successful"
	ClusterOpsRequestPhaseSkipped     ClusterOpsRequestPhase = "Skipped"
	ClusterOpsRequestPhaseFailed      ClusterOpsRequestPhase = "Failed"
	ClusterOpsRequestPhaseTerminating ClusterOpsRequestPhase = "Terminating"
)

const (
	ClusterOpsRequestConditionTypeReady       kmapi.ConditionType = "Ready"
	ClusterOpsRequestConditionTypeProgressing kmapi.ConditionType = "Progressing"
)

const (
	ControlPlaneUpdateCondition     kmapi.ConditionType = "ControlPlaneUpdate"
	ControlPlaneUpdateStartedReason string              = "ControlPlaneUpdateStarted"
	ControlPlaneUpdateFailedReason  string              = "ControlPlaneUpdateFailed"
)

const (
	MachinePoolUpdateCondition     kmapi.ConditionType = "MachinePoolUpdate"
	MachinePoolUpdateStartedReason string              = "MachinePoolUpdateStarted"
	MachinePoolUpdateFailedReason  string              = "MachinePoolUpdateFailed"
)

func ConditionsOrder() []kmapi.ConditionType {
	return []kmapi.ConditionType{
		MachinePoolUpdateCondition,
		ControlPlaneUpdateCondition,
	}
}

func GetPhase(obj *ClusterOpsRequest) ClusterOpsRequestPhase {
	if !obj.GetDeletionTimestamp().IsZero() {
		return ClusterOpsRequestPhaseTerminating
	}
	conditions := obj.GetConditions()
	if len(conditions) == 0 {
		return ClusterOpsRequestPhasePending
	}
	var cond kmapi.Condition
	for i := range conditions {
		c := conditions[i]
		if c.Type == kmapi.ReadyCondition {
			cond = c
			break
		}
	}
	if cond.Type != kmapi.ReadyCondition {
		fmt.Printf("no Ready condition in the status for %s/%s", obj.GetNamespace(), obj.GetName())
		return ClusterOpsRequestPhaseInProgress
	}
	if cond.Status == metav1.ConditionTrue {
		return ClusterOpsRequestPhaseSuccessful
	}

	return ClusterOpsRequestPhaseInProgress
}
