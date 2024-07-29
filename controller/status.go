package controller

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/openshift-console/console-application-operator/api/v1alpha1"
)

// SetDegraded sets the Operator Degraded condition with the provided reason and message.
func SetDegraded(operatorCR *appsv1alpha1.ConsoleApplication, reason, message string) {
	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionOperatorDegraded.String(),
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            message,
	})
}

// SetGitServiceCondition sets the GitService condition with the provided status and reason.
func SetGitServiceCondition(operatorCR *appsv1alpha1.ConsoleApplication, status metav1.ConditionStatus, reason string) {
	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionGitRepoReachable.String(),
		Status:             status,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            fmt.Sprintf("Git Repository Reachable: %s", string(status)),
	})
}

// SetStarted sets the Operator Progressing condition to True.
func SetStarted(operatorCR *appsv1alpha1.ConsoleApplication) {
	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionProgressing.String(),
		Status:             metav1.ConditionTrue,
		Reason:             "Reconciling",
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            "Reconciliation in progress",
	})
}

// SetFailed sets the Operator Progressing and Application Ready conditions to False with the provided reason and message.
func SetFailed(operatorCR *appsv1alpha1.ConsoleApplication, reason, message string) {
	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionApplicationReady.String(),
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            message,
	})
	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionProgressing.String(),
		Status:             metav1.ConditionFalse,
		Reason:             appsv1alpha1.ReasonReconcileFailed.String(),
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            "Reconciliation failed",
	})
}

// SetSucceeded sets the Operator Progressing and Application Ready conditions to True.
func SetSucceeded(operatorCR *appsv1alpha1.ConsoleApplication) {
	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionApplicationReady.String(),
		Status:             metav1.ConditionTrue,
		Reason:             appsv1alpha1.ReasonAllResourcesReady.String(),
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            "All resources are successfully created and ready",
	})
	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionProgressing.String(),
		Status:             metav1.ConditionFalse,
		Reason:             appsv1alpha1.ReasonReconcileCompleted.String(),
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            "Reconciliation completed successfully",
	})
}

// SetCondition sets a generic condition, overwriting existing one by type if present.
func SetCondition(operatorCR *appsv1alpha1.ConsoleApplication, typ string, status metav1.ConditionStatus, reason string, message string) {
	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:    typ,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}
