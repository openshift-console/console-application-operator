package controller

import (
	"fmt"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/openshift-console/console-application-operator/api/v1alpha1"
)

// TODO: Implement "Progressing" status condition in the future.

// SetDegraded sets the Operator Degraded condition with the provided reason and message.
func SetDegraded(consoleApplication *appsv1alpha1.ConsoleApplication, reason, message string) {
	meta.SetStatusCondition(&consoleApplication.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionOperatorDegraded.String(),
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            message,
	})
}

// SetGitServiceCondition sets the GitService condition with the provided status and reason.
func SetGitServiceCondition(consoleApplication *appsv1alpha1.ConsoleApplication, status metav1.ConditionStatus, reason string) {
	meta.SetStatusCondition(&consoleApplication.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionGitRepoReachable.String(),
		Status:             status,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            fmt.Sprintf("Git Repository Reachable: %s", string(status)),
	})
}

// SetResourceCondition sets a condition for a specific resource type.
func SetResourceCondition(consoleApplication *appsv1alpha1.ConsoleApplication, conditionType string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&consoleApplication.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            message,
	})
}

// SetStatusField sets a status field with the provided value.
func SetStatusField(consoleApplication *appsv1alpha1.ConsoleApplication, fieldName string, value interface{}) error {
	statusValue := reflect.ValueOf(&consoleApplication.Status).Elem()
	fieldValue := statusValue.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in status", fieldName)
	}
	if !fieldValue.CanSet() {
		return fmt.Errorf("cannot set field: %s in status", fieldName)
	}

	val := reflect.ValueOf(value)
	if fieldValue.Type() != val.Type() {
		return fmt.Errorf("provided value type didn't match status field type")
	}

	fieldValue.Set(val)
	return nil
}

// SetReady sets the Operator Ready condition.
func SetReady(consoleApplication *appsv1alpha1.ConsoleApplication, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&consoleApplication.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionReady.String(),
		Status:             status,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            message,
	})
}

// SetProgressing sets the Operator Progressing condition.
func SetProgressing(consoleApplication *appsv1alpha1.ConsoleApplication, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&consoleApplication.Status.Conditions, metav1.Condition{
		Type:               appsv1alpha1.ConditionProgressing.String(),
		Status:             status,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            message,
	})
}

// SetStarted sets the Operator Progressing condition to True and Ready condition to Unknown.
func SetStarted(consoleApplication *appsv1alpha1.ConsoleApplication) {
	if consoleApplication.Status.Conditions == nil {
		consoleApplication.Status.Conditions = make([]metav1.Condition, 0)
		SetReady(consoleApplication, metav1.ConditionUnknown, appsv1alpha1.ReasonInit.String(), "Initializing ConsoleApplication")
	}
	SetProgressing(consoleApplication, metav1.ConditionTrue, appsv1alpha1.ReasonRequirementsBeingMet.String(), "Requirements are being met")
}

// SetFailed sets the Operator Progressing and Application Ready conditions to False with the provided reason and message.
func SetFailed(consoleApplication *appsv1alpha1.ConsoleApplication, reason, message string) {
	SetReady(consoleApplication, metav1.ConditionFalse, reason, message)
	SetProgressing(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonRequirementsNotMet.String(), "Requirements are not met")
}

// SetSucceeded sets the Operator Progressing and Application Ready conditions to True.
func SetSucceeded(consoleApplication *appsv1alpha1.ConsoleApplication) {
	SetReady(consoleApplication, metav1.ConditionTrue, appsv1alpha1.ReasonAllResourcesReady.String(), "All resources are successfully created and ready")
	SetProgressing(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonRequirementsMet.String(), "All requirements are met")
}
